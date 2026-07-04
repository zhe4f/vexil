package receiver

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"vexil/internal/network"
	"vexil/internal/protocol"
	"vexil/internal/receiver/file"
	"vexil/internal/tracker"
	"vexil/internal/util"
)

type Config struct {
	SaveDir string
	Port    int
}

type TransferStats struct {
	TotalSize int64
	Duration  float64
	SpeedMBps float64
	FileCount int
	FileNames []string
}

type Receiver struct {
	cfg          Config
	tcfg         protocol.TransferConfig
	files        []*os.File
	fileManifest []FileOffsetInfo
	doneOnce     sync.Once
	compOnce     sync.Once
	writeErrCh   chan *WriteError
	Stats        TransferStats
	PeerIP       string
	PeerName     string

	totalBytes  atomic.Int64
	collected   []chunkEntry
	collectMu   sync.RWMutex
	receivedSet sync.Map

	startOnce sync.Once
    startTime time.Time

	eventCh chan<- protocol.TaskEvent
	taskID  string

	errOnce sync.Once
}

type FileOffsetInfo struct {
	File  *os.File
	Path  string
	Size  int64
	Start int64
}

func New(cfg Config) *Receiver {
	return &Receiver{
		cfg:        cfg,
		doneOnce:   sync.Once{},
		compOnce:   sync.Once{},
		writeErrCh: make(chan *WriteError, 16),
	}
}

func (r *Receiver) WriteErrors() <-chan *WriteError {
	return r.writeErrCh
}

func (r *Receiver) sendEvent(ev protocol.TaskEvent) {
	select {
	case r.eventCh <- ev:
	default:
	}
}

func (r *Receiver) sendError(err error) {
	r.errOnce.Do(func() {
		r.eventCh <- protocol.TaskEvent{
			TaskID: r.taskID,
			Type:   protocol.EventError,
			Error:  err,
		}
	})
}

func (r *Receiver) Run(ctx context.Context, tcfg protocol.TransferConfig, eventCh chan<- protocol.TaskEvent, taskID string) error {
	r.tcfg = tcfg
	r.eventCh = eventCh
	r.taskID = taskID

	var tkr *tracker.Tracker
	defer func() {
		if tkr != nil {
			tkr.Stop()
		}
	}()

	r.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskPreparing,
	})

	defer func() {
		for _, f := range r.files {
			f.Close()
		}
	}()

	var tlsCfg *tls.Config
	if tcfg.TLSEnabled {
		builder := network.NewServerTLSBuilder(tcfg.TLSCertFile, tcfg.TLSKeyFile)
		var err error
		tlsCfg, err = builder.Build()
		if err != nil {
			r.sendError(err)
			return err
		}
	}

	pool, err := NewConnPool(r.cfg.Port, tcfg.AcceptTimeout, tlsCfg)
	if err != nil {
		r.sendError(err)
		return err
	}
	defer pool.Close()

	go func() {
		<-ctx.Done()
		pool.Close()
	}()

	var controlConn *network.FramedConn
	var dataConns []*network.FramedConn

	for controlConn == nil {
		fc, role, err := pool.AcceptOne(tcfg.ReadTimeout)
		if err != nil {
			r.sendError(err)
			return err
		}
		if role == "control" {
			controlConn = fc
		} else if role == "data" {
			dataConns = append(dataConns, fc)
		} else {
			fc.Close()
		}
	}
	r.PeerIP = pool.PeerIP

	helloFrame, err := controlConn.ReadFrame(tcfg.ReadTimeout)
	if err != nil {
		r.sendError(err)
		return err
	}
	if helloFrame.Type != protocol.MSG_HELLO {
		r.sendError(fmt.Errorf("expected HELLO, got %s", protocol.MessageName(helloFrame.Type)))
		return fmt.Errorf("expected HELLO")
	}
	hello, err := protocol.DecodeHello(helloFrame)
	if err != nil {
		r.sendError(err)
		return err
	}
	numConns := hello.NumConns
	if numConns <= 0 || numConns > 32 {
		numConns = tcfg.NumConns
	}
	expectedData := numConns - 1

	readyFrame, _ := protocol.EncodeReady()
	controlConn.SendFromMemory(readyFrame)

	for len(dataConns) < expectedData {
		fc, role, err := pool.AcceptOne(tcfg.ReadTimeout)
		if err != nil {
			r.sendError(err)
			return err
		}
		if role == "data" {
			dataConns = append(dataConns, fc)
		} else {
			fc.Close()
		}
		fmt.Printf("  数据连接 %d/%d\n", len(dataConns), expectedData)
	}

	goFrame, _ := protocol.EncodeGo()
	controlConn.SendFromMemory(goFrame)

	frame, err := controlConn.ReadFrame(tcfg.ReadTimeout)
	if err != nil {
		r.sendError(err)
		return err
	}
	if frame.Type != protocol.MSG_BATCH_START {
		err := fmt.Errorf("expected BATCH_START, got %s", protocol.MessageName(frame.Type))
		r.sendError(err)
		return err
	}

	batchStart, err := protocol.DecodeBatchStart(frame)
	if err != nil {
		r.sendError(err)
		return err
	}

	frame, err = controlConn.ReadFrame(tcfg.ReadTimeout)
	if err != nil {
		r.sendError(err)
		return err
	}
	if frame.Type != protocol.MSG_FILE_REQUEST {
		err := fmt.Errorf("expected FILE_REQUEST, got %s", protocol.MessageName(frame.Type))
		r.sendError(err)
		return err
	}

	fileReq, err := protocol.DecodeFileRequest(frame)
	if err != nil {
		r.sendError(err)
		return err
	}
	
	r.PeerName = fileReq.SenderName

	conns := make([]*network.FramedConn, 0, len(dataConns)+1)
	conns = append(conns, controlConn)
	conns = append(conns, dataConns...)

	fmt.Printf("  批量传输: %d 个文件，%d 个连接\n", len(batchStart.Files), len(conns))

	r.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskRunning,
	})

	fileCount := len(batchStart.Files)

	var displayNames []string
	if fileCount > 0 {
		firstName := filepath.ToSlash(batchStart.Files[0].Name)
		if fileCount == 1 && !strings.Contains(firstName, "/") {
			displayNames = []string{filepath.Base(firstName)}
		} else if fileCount >= 1 {
			topSet := make(map[string]bool)
			for _, mf := range batchStart.Files {
				parts := strings.Split(filepath.ToSlash(mf.Name), "/")
				if len(parts) > 1 {
					topSet[parts[0]+"/"] = true
				} else if len(parts) == 1 {
					topSet[parts[0]] = true
				}
			}
			for name := range topSet {
				displayNames = append(displayNames, name)
			}
		}
	}

	actualConns := len(conns)
	for i, mf := range batchStart.Files {
		fmt.Printf("  [%d] %s (%s)\n", i+1, mf.Name, util.FormatSize(mf.Size))
	}

	var globalOffset int64
	for _, mf := range batchStart.Files {
		targetPath := filepath.Join(r.cfg.SaveDir, filepath.FromSlash(mf.Name))

		absSaveDir, err := filepath.Abs(r.cfg.SaveDir)
		if err != nil {
			r.sendError(err)
			return fmt.Errorf("无法解析保存目录: %w", err)
		}
		absTarget, err := filepath.Abs(targetPath)
		if err != nil {
			r.sendError(err)
			return fmt.Errorf("无法解析目标路径: %w", err)
		}
		if !strings.HasPrefix(absTarget, absSaveDir+string(filepath.Separator)) {
			err := fmt.Errorf("非法路径: %s", mf.Name)
			r.sendError(err)
			return err
		}

		os.MkdirAll(filepath.Dir(targetPath), 0755)

		f, err := file.OpenFile(targetPath)
		if err != nil {
			r.sendError(err)
			return fmt.Errorf("创建文件 %s 失败: %w", targetPath, err)
		}
		if err := file.Preallocate(f, mf.Size); err != nil {
			f.Close()
			r.sendError(err)
			return fmt.Errorf("预分配 %s 失败: %w", targetPath, err)
		}
		r.files = append(r.files, f)
		r.fileManifest = append(r.fileManifest, FileOffsetInfo{
			File:  f,
			Path:  targetPath,
			Size:  mf.Size,
			Start: globalOffset,
		})
		globalOffset += mf.Size
	}
	totalSize := globalOffset

	r.totalBytes.Store(0)

	resumePath := filepath.Join(r.cfg.SaveDir, "__stream__.vexil")
	r.resetCollected()

	resumeState, _ := LoadResume(resumePath)
	if resumeState != nil && resumeState.FileSize == totalSize && len(resumeState.ChunkOffsets) > 0 {
		r.restoreAllHashes(resumeState.ChunkOffsets, resumeState.ChunkHashes, resumeState.ChunkSizes)
		missing := r.findMissingRanges(resumeState.ChunkOffsets, resumeState.ChunkSizes, totalSize, tcfg.MaxChunk)

		var restoredBytes int64
		var existing []protocol.BlockHash
		for i, off := range resumeState.ChunkOffsets {
			sz := tcfg.MaxChunk
			if i < len(resumeState.ChunkSizes) && resumeState.ChunkSizes[i] > 0 {
				sz = resumeState.ChunkSizes[i]
			}
			if off+sz > totalSize {
				sz = totalSize - off
			}
			restoredBytes += sz
			existing = append(existing, protocol.BlockHash{
				Offset: off,
				Hash:   resumeState.ChunkHashes[i],
			})
		}

		r.totalBytes.Store(restoredBytes)

		if len(missing) > 0 {
			resumeFrame, _ := protocol.EncodeResumeRangeEx(missing, existing)
			conns[0].SendFromMemory(resumeFrame)
		}
	} else {
		r.resetCollected()
		r.totalBytes.Store(0)
	}

	doneCh := make(chan struct{})
	compCh := make(chan struct{})
	compOnce := &sync.Once{}
	stallCh := make(chan error, 1)

	fmt.Printf("  实际连接数: %d（发送端: %d）\n", actualConns, fileReq.NumConns)

	var wg sync.WaitGroup
	for i := 0; i < actualConns; i++ {
		wg.Add(1)
		go r.RecvWorker(ctx, i, conns[i], r.fileManifest,
			resumePath, totalSize,
			doneCh, compCh, compOnce, &wg, &r.doneOnce, r.writeErrCh)
	}

	wg.Add(1)
	go AckFlusher(ctx, conns[0], &r.totalBytes, totalSize, doneCh, compCh, &wg, tcfg.ACKInterval)

	go StallDetector(ctx, &r.totalBytes, doneCh, compCh, stallCh, tcfg.StallInitial, tcfg.StallInterval, tcfg.StallThreshold)

	tkr = tracker.New(totalSize,
		int(totalSize/(1024*1024))+1,
		1024*1024,
		func() int {
			return int(r.totalBytes.Load()) / (1024 * 1024)
		},
		func() int64 {
			return r.totalBytes.Load()
		},
		doneCh,
		func(p tracker.Progress) {
			select {
			case r.eventCh <- protocol.TaskEvent{
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

	select {
	case <-doneCh:
	case <-ctx.Done():
		r.sendEvent(protocol.TaskEvent{
			TaskID: taskID,
			Type:   protocol.EventState,
			State:  protocol.TaskCancelled,
		})
		for _, c := range conns {
			c.Close()
		}
		wg.Wait()
		return ctx.Err()
	case err := <-stallCh:
		r.sendError(err)
		if tkr != nil {
			tkr.Stop()
			tkr = nil
		}
		for _, c := range conns {
			c.Close()
		}
		r.doneOnce.Do(func() { close(doneCh) })
		r.compOnce.Do(func() { close(compCh) })
		wg.Wait()
		return err
	case writeErr := <-r.WriteErrors():
		err := fmt.Errorf("写入文件失败: %w", writeErr)
		r.sendError(err)
		for _, c := range conns {
			c.Close()
		}
		r.doneOnce.Do(func() { close(doneCh) })
		r.compOnce.Do(func() { close(compCh) })
		wg.Wait()
		return err
	}

	r.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskFinalizing,
	})

	timeout := tcfg.CompleteTimeout
	deadline := time.After(timeout)
	completeReceived := false
	for !completeReceived {
		select {
		case <-compCh:
			completeReceived = true
		case <-ctx.Done():
			r.sendEvent(protocol.TaskEvent{
				TaskID: taskID,
				Type:   protocol.EventState,
				State:  protocol.TaskCancelled,
			})
			for _, c := range conns {
				c.Close()
			}
			r.doneOnce.Do(func() { close(doneCh) })
			r.compOnce.Do(func() { close(compCh) })
			wg.Wait()
			return ctx.Err()
		case <-deadline:
			received := r.totalBytes.Load()
			complete := r.isComplete(totalSize, tcfg.MaxChunk)
			if complete && received >= totalSize {
				completeReceived = true
			} else {
				for _, c := range conns {
					c.Close()
				}
				wg.Wait()
				err := fmt.Errorf("等待 COMPLETE 超时，数据不完整")
				r.sendError(err)
				return err
			}
		}
	}

	for _, c := range conns {
		c.Close()
	}
	wg.Wait()

	if !r.isComplete(totalSize, tcfg.MaxChunk) {
		err := fmt.Errorf("incomplete transfer")
		r.sendError(err)
		return err
	}

	for _, f := range r.files {
		f.Close()
	}
	r.files = nil

	os.Remove(resumePath)

	elapsed := time.Since(r.startTime).Seconds()
	speed := float64(0)
	if elapsed > 0 {
		speed = float64(totalSize) / elapsed / 1024 / 1024
	}
	r.Stats = TransferStats{
		TotalSize: totalSize,
		Duration:  elapsed,
		SpeedMBps: speed,
		FileCount: fileCount,
		FileNames: displayNames,
	}

	r.sendEvent(protocol.TaskEvent{
		TaskID: taskID,
		Type:   protocol.EventState,
		State:  protocol.TaskCompleted,
	})

	return nil
}

func locateFileByOffset(manifest []FileOffsetInfo, compactOffset int64) (*os.File, int64, int64) {
	lo, hi := 0, len(manifest)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		m := manifest[mid]
		if compactOffset < m.Start {
			hi = mid - 1
		} else if compactOffset >= m.Start+m.Size {
			lo = mid + 1
		} else {
			return m.File, compactOffset - m.Start, m.Start + m.Size
		}
	}
	return nil, 0, 0
}