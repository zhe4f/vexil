package sender

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"vexil_go/network"
	"vexil_go/protocol"
	"vexil_go/tracker"
	"vexil_go/util"
)

type Config struct {
	Files []string
	Host  string
	Port  int
	PeerName string
}

type TransferStats struct {
	TotalSize int64
	Duration  float64
	SpeedMBps float64
	FileCount int
	FileNames []string
}

type Sender struct {
	cfg        Config
	tcfg       protocol.TransferConfig
	scanResult *ScanResult
	Stats      TransferStats
	eventCh    chan<- protocol.TaskEvent
	taskID     string
	errOnce    sync.Once
	startTime  time.Time
    startOnce  sync.Once
}

type chunkInfo struct {
	offset int64
	hash   [32]byte
}

type chunkMeta struct {
	offset int64
	size   int64
}

func New(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) sendEvent(ev protocol.TaskEvent) {
	select {
	case s.eventCh <- ev:
	default:
	}
}

func (s *Sender) sendError(err error) {
	s.errOnce.Do(func() {
		s.eventCh <- protocol.TaskEvent{
			TaskID: s.taskID,
			Type:   protocol.EventError,
			Error:  err,
		}
	})
}

func (s *Sender) Run(ctx context.Context, tcfg protocol.TransferConfig, eventCh chan<- protocol.TaskEvent, taskID string) error {
	s.tcfg = tcfg
	s.eventCh = eventCh
	s.taskID = taskID

	var tkr *tracker.Tracker
	defer func() {
		if tkr != nil {
			tkr.Stop()
		}
	}()

	s.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskPreparing,
	})

	result, err := ScanFiles(s.cfg.Files)
	if err != nil {
		s.sendError(fmt.Errorf("扫描文件失败: %w", err))
		return fmt.Errorf("扫描文件失败: %w", err)
	}
	s.scanResult = result

	if len(result.Files) == 0 {
		err := fmt.Errorf("没有找到文件")
		s.sendError(err)
		return err
	}

	s.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskConnecting,
	})

	var tlsCfg *tls.Config
	if tcfg.TLSEnabled {
		builder := network.NewClientTLSBuilder(tcfg.TLSInsecureSkipVerify, tcfg.TLSSessionCacheSize)
		tlsCfg, err = builder.Build()
		if err != nil {
			s.sendError(fmt.Errorf("TLS 配置失败: %w", err))
			return fmt.Errorf("TLS 配置失败: %w", err)
		}
	}

	pool := NewConnPool(s.cfg.Host, s.cfg.Port,
		tcfg.DialTimeout, tcfg.DialRetries, tcfg.DialRetryBaseDelay,
		tcfg.ConnEstablishGap, tlsCfg)
	defer pool.Close()

	controlConn, err := pool.DialOne(tcfg.TCPReadBufSize, tcfg.TCPWriteBufSize)
	if err != nil {
		s.sendError(fmt.Errorf("dial control: %w", err))
		return fmt.Errorf("dial control: %w", err)
	}

	roleFrame, _ := protocol.EncodeConnRole(protocol.ConnRole{Role: "control"})
	if err := controlConn.SendFromMemory(roleFrame); err != nil {
		s.sendError(fmt.Errorf("发送 CONN_ROLE: %w", err))
		return fmt.Errorf("发送 CONN_ROLE: %w", err)
	}

	helloFrame, _ := protocol.EncodeHello(protocol.Hello{NumConns: tcfg.NumConns})
	if err := controlConn.SendFromMemory(helloFrame); err != nil {
		s.sendError(fmt.Errorf("发送 HELLO: %w", err))
		return fmt.Errorf("发送 HELLO: %w", err)
	}
	fmt.Printf("  控制连接已建立，连接数: %d\n", tcfg.NumConns)

	readyFrame, err := controlConn.ReadFrame(tcfg.ReadTimeout)
	if err != nil {
		s.sendError(fmt.Errorf("等待 READY: %w", err))
		return fmt.Errorf("等待 READY: %w", err)
	}
	if readyFrame.Type != protocol.MSG_READY {
		s.sendError(fmt.Errorf("expected READY, got %s", protocol.MessageName(readyFrame.Type)))
		return fmt.Errorf("expected READY")
	}

	conns := make([]*network.FramedConn, tcfg.NumConns)
	conns[0] = controlConn

	for i := 1; i < tcfg.NumConns; i++ {
		c, err := pool.DialOne(tcfg.TCPReadBufSize, tcfg.TCPWriteBufSize)
		if err != nil {
			for j := 0; j < i; j++ {
				conns[j].Close()
			}
			s.sendError(fmt.Errorf("dial data %d: %w", i, err))
			return fmt.Errorf("dial data %d: %w", i, err)
		}
		conns[i] = c

		roleFrame, _ := protocol.EncodeConnRole(protocol.ConnRole{Role: "data"})
		if err := c.SendFromMemory(roleFrame); err != nil {
			s.sendError(fmt.Errorf("发送 CONN_ROLE: %w", err))
			return fmt.Errorf("发送 CONN_ROLE: %w", err)
		}
		fmt.Printf("  数据连接 %d/%d 已建立\n", i, tcfg.NumConns-1)

		if i < tcfg.NumConns-1 {
			time.Sleep(tcfg.ConnEstablishGap)
		}
	}

	goFrame, err := controlConn.ReadFrame(tcfg.ReadTimeout)
	if err != nil {
		s.sendError(fmt.Errorf("等待 GO: %w", err))
		return fmt.Errorf("等待 GO: %w", err)
	}
	if goFrame.Type != protocol.MSG_GO {
		s.sendError(fmt.Errorf("expected GO, got %s", protocol.MessageName(goFrame.Type)))
		return fmt.Errorf("expected GO")
	}

	s.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskRunning,
	})

	rootName := ""
	isDir := false
	var topNames []string
	for _, p := range s.cfg.Files {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			topNames = append(topNames, filepath.Base(p)+"/")
		} else {
			topNames = append(topNames, filepath.Base(p))
		}
	}
	if len(s.cfg.Files) == 1 {
		info, _ := os.Stat(s.cfg.Files[0])
		if info != nil && info.IsDir() {
			rootName = filepath.Base(s.cfg.Files[0])
			isDir = true
		}
	}

	manifests := make([]protocol.FileManifest, len(result.Files))
	var totalSize int64
	for i, f := range result.Files {
		name := filepath.ToSlash(f.RelPath)
		if rootName != "" {
			name = filepath.ToSlash(filepath.Join(rootName, f.RelPath))
		}
		manifests[i] = protocol.FileManifest{Name: name, Size: f.Size}
		totalSize += f.Size
	}

	allChunks := precalculateChunks(result.Files, totalSize, tcfg.MaxChunk)

	batchStart := protocol.BatchStart{
		TotalUnits: uint32(len(result.Files)),
		Files:      manifests,
	}
	startFrame, err := protocol.EncodeBatchStart(batchStart)
	if err != nil {
		s.sendError(fmt.Errorf("构造 BATCH_START 失败: %w", err))
		return fmt.Errorf("构造 BATCH_START 失败: %w", err)
	}
	if err := conns[0].SendFromMemory(startFrame); err != nil {
		s.sendError(fmt.Errorf("发送 BATCH_START 失败: %w", err))
		return fmt.Errorf("发送 BATCH_START 失败: %w", err)
	}

	openFiles := make([]*os.File, len(result.Files))
	for i, f := range result.Files {
		file, err := os.Open(f.Path)
		if err != nil {
			for j := 0; j < i; j++ {
				openFiles[j].Close()
			}
			s.sendError(fmt.Errorf("打开文件 %s 失败: %w", f.Path, err))
			return fmt.Errorf("打开文件 %s 失败: %w", f.Path, err)
		}

		stat, err := file.Stat()
		if err != nil {
			file.Close()
			for j := 0; j < i; j++ {
				openFiles[j].Close()
			}
			s.sendError(fmt.Errorf("获取文件信息失败 %s: %w", f.Path, err))
			return fmt.Errorf("获取文件信息失败 %s: %w", f.Path, err)
		}
		if stat.Size() != f.Size {
			file.Close()
			for j := 0; j < i; j++ {
				openFiles[j].Close()
			}
			err := fmt.Errorf("文件大小已变化 %s", f.Path)
			s.sendError(err)
			return err
		}

		openFiles[i] = file
	}
	defer func() {
		for _, f := range openFiles {
			if f != nil {
				f.Close()
			}
		}
	}()

	hostname, _ := os.Hostname()
	req := protocol.FileRequest{
		FileName: "__stream__",
		FileSize: totalSize,
		NumConns: tcfg.NumConns,
		SenderName: hostname,
	}
	frame, err := protocol.EncodeFileRequest(req)
	if err != nil {
		s.sendError(fmt.Errorf("构造 FILE_REQUEST 失败: %w", err))
		return fmt.Errorf("构造 FILE_REQUEST 失败: %w", err)
	}
	if err := conns[0].SendFromMemory(frame); err != nil {
		s.sendError(fmt.Errorf("发送 FILE_REQUEST 失败: %w", err))
		return fmt.Errorf("发送 FILE_REQUEST 失败: %w", err)
	}

	infoCh := make(chan chunkInfo, 256)
	w := NewWindow(allChunks, tcfg.WindowSize, tcfg.NumConns, infoCh)

	var ah *AckHandler
	ah = NewAckHandler(ctx, conns[0], w, tcfg.ReadTimeout)

	finalHashCh := make(chan string, 1)

	go func() {
		var entries []chunkInfo
		for info := range infoCh {
			entries = append(entries, info)
		}

		seen := make(map[int64]bool)
		unique := make([]chunkInfo, 0, len(entries))
		for _, e := range entries {
			if !seen[e.offset] {
				unique = append(unique, e)
				seen[e.offset] = true
			}
		}

		sort.Slice(unique, func(i, j int) bool {
			return unique[i].offset < unique[j].offset
		})

		hasher := sha256.New()
		for _, e := range unique {
			hasher.Write(e.hash[:])
		}
		finalHashCh <- hex.EncodeToString(hasher.Sum(nil))
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ah.Done()
	}()

	for i := 0; i < tcfg.NumConns; i++ {
		wg.Add(1)
		go s.streamWorker(ctx, i, conns[i], result.Files, openFiles, w, &wg, infoCh)
	}

	tkr = tracker.New(totalSize, len(allChunks), int(tcfg.MaxChunk),
		func() int {
			return int(w.SentBytes()) / int(tcfg.MaxChunk)
		},
		func() int64 {
			return w.SentBytes()
		},
		w.Done(),
		func(p tracker.Progress) {
			select {
			case s.eventCh <- protocol.TaskEvent{
				TaskID:    taskID,
				Type:      protocol.EventProgress,
				State:     protocol.TaskRunning,
				Percent:   p.Percent,
				SpeedMBps: p.SpeedMBps,
				Sent:      p.Sent,
				Total:     p.Total,
				ETA:       p.ETA,
			}:
			default:
			}
		},
		tcfg.TrackerInterval,
	)

	go func() {
		<-ctx.Done()
		pool.Close()
	}()

	w.Run(ctx)
	wg.Wait()
	close(infoCh)

	// 检查连接是否异常断开
	if ah.Err() != nil {
		s.sendEvent(protocol.TaskEvent{
			TaskID: taskID,
			Type:   protocol.EventState,
			State:  protocol.TaskFailed,
		})
		s.sendError(fmt.Errorf("连接断开: %w", ah.Err()))
		return ah.Err()
	}

	select {
	case <-ctx.Done():
		s.sendEvent(protocol.TaskEvent{
			TaskID: taskID,
			Type:   protocol.EventState,
			State:  protocol.TaskCancelled,
		})
		return ctx.Err()
	default:
	}

	// 传输完成，记录时间
	elapsed := time.Since(s.startTime).Seconds()

	s.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskFinalizing,
	})

	var fileSHA256 string
	select {
	case fileSHA256 = <-finalHashCh:
	case <-time.After(tcfg.SHA256Timeout):
		fileSHA256 = ""
	}

	compFrame, err := protocol.EncodeComplete([]byte(fileSHA256))
	if err == nil {
		conns[0].SendFromMemory(compFrame)
	}

	time.Sleep(tcfg.CompleteSendWait)

	endFrame, _ := protocol.EncodeBatchEnd()
	conns[0].SendFromMemory(endFrame)

	speed := float64(totalSize) / elapsed / 1024 / 1024
	fmt.Printf("\n  总计 %s in %v (%.1f MB/s)\n",
		util.FormatSize(totalSize), time.Duration(elapsed*float64(time.Second)).Round(time.Millisecond), speed)

	var displayNames []string
	if isDir {
		displayNames = []string{rootName + "/"}
	} else {
		displayNames = topNames
	}

	s.Stats = TransferStats{
		TotalSize: totalSize,
		Duration:  elapsed,
		SpeedMBps: speed,
		FileCount: len(result.Files),
		FileNames: displayNames,
	}

	s.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskCompleted,
	})

	return nil
}

func precalculateChunks(files []FileInfo, totalSize int64, maxChunk int64) []chunkMeta {
	var chunks []chunkMeta
	var offset int64
	for offset < totalSize {
		maxReadable := maxReadableAt(files, offset)
		size := util.MinInt64(maxChunk, maxReadable)
		chunks = append(chunks, chunkMeta{offset: offset, size: size})
		offset += size
	}
	return chunks
}

func maxReadableAt(files []FileInfo, offset int64) int64 {
	var cumulative int64
	for _, f := range files {
		nextFileStart := cumulative + f.Size
		if offset < nextFileStart {
			return nextFileStart - offset
		}
		cumulative = nextFileStart
	}
	return 0
}