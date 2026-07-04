package bridge

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"vexil_go/app"
	"vexil_go/config"
	"vexil_go/discovery"
	"vexil_go/history"
	"vexil_go/protocol"
)

// Bridge 对外暴露的 API
type Bridge struct {
	application *app.Application
	listener    EventListener
	mu          sync.Mutex
	udpDisc     *discovery.UDPDiscovery
	mdnsDisc    *discovery.MDNSDiscovery
	historyDir  string
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// HistoryInfo 历史记录
type HistoryInfo struct {
	Time      string   `json:"time"`
	Direction string   `json:"direction"`
	Peer      string   `json:"peer"`
	PeerName  string   `json:"peer_name"`
	Files     int      `json:"files"`
	FileNames []string `json:"file_names"`
	Size      int64    `json:"size"`
	Duration  float64  `json:"duration_sec"`
	SpeedMBps float64  `json:"speed_mbps"`
	Success   bool     `json:"success"`
	SavePath  string   `json:"save_path"`
}

// EventListener is the callback interface for transfer events
type EventListener interface {
	OnProgress(taskID string, state string, percent float64, speedMBps float64, sent int64, total int64, eta int64)
	OnComplete(taskID string)
	OnError(taskID string, err string)
}

// NewBridge 创建 Bridge 实例
func NewBridge(filesDir string, tzOffset int) *Bridge {
	time.Local = time.FixedZone("Local", tzOffset)
	config.SetDir(filesDir)
	history.SetDir(filesDir)
	cfg := config.LoadAppConfig()
	cfg.Apply()
	return &Bridge{
		application: app.New(cfg),
		historyDir:  filesDir,
	}
}

// SetEventListener 设置事件回调
func (b *Bridge) SetEventListener(listener EventListener) {
	b.listener = listener
}

// DiscoverDevices 发现设备，返回 JSON 字符串
func (b *Bridge) DiscoverDevices(timeoutSec int) (string, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	devices, err := b.application.DiscoverDevices(timeout)
	if err != nil {
		return "", err
	}

	var result []DeviceInfo
	for _, d := range devices {
		result = append(result, DeviceInfo{
			Name: d.Name,
			IP:   d.IP,
			Port: d.Port,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// StartSend 启动发送，返回 taskID
func (b *Bridge) StartSend(host string, port int, filePathsJSON string, peerName string, numConns int, maxChunkMB int, tlsEnabled bool) (string, error) {
	cfg := config.Snapshot()
	cfg.NumConns = numConns
	cfg.MaxChunk = int64(maxChunkMB) * 1024 * 1024
	cfg.TLSEnabled = tlsEnabled
	config.ApplyConfig(cfg)

	var paths []string
	if err := json.Unmarshal([]byte(filePathsJSON), &paths); err != nil {
		return "", err
	}

	taskID, err := b.application.StartSend(host, port, paths, peerName)
	if err != nil {
		return "", err
	}

	go b.observeTask(taskID)
	return taskID, nil
}

// StartReceive 启动接收，返回 taskID
func (b *Bridge) StartReceive(port int, saveDir string, deviceName string, numConns int, maxChunkMB int, tlsEnabled bool) (string, error) {
	cfg := config.Snapshot()
	cfg.NumConns = numConns
	cfg.MaxChunk = int64(maxChunkMB) * 1024 * 1024
	cfg.TLSEnabled = tlsEnabled
	config.ApplyConfig(cfg)
	
	b.mu.Lock()
	if b.udpDisc != nil {
		b.udpDisc.Stop()
	}
	if b.mdnsDisc != nil {
		b.mdnsDisc.Stop()
	}

	udpDisc := discovery.NewUDPDiscoveryWithName(deviceName)
	if err := udpDisc.Start(port); err != nil {
		b.mu.Unlock()
		fmt.Printf("警告: UDP 发现启动失败: %v\n", err)
	}
	b.udpDisc = udpDisc

	mdnsDisc := discovery.NewMDNSDiscoveryWithName(deviceName)
	if err := mdnsDisc.Start(port); err != nil {
		b.mu.Unlock()
		fmt.Printf("警告: mDNS 发现启动失败: %v\n", err)
	}
	b.mdnsDisc = mdnsDisc
	b.mu.Unlock()

	taskID, err := b.application.StartRecv(port, saveDir)
	if err != nil {
		udpDisc.Stop()
		mdnsDisc.Stop()
		return "", err
	}

	go b.observeTaskWithBroadcast(taskID, udpDisc, mdnsDisc)

	return taskID, nil
}

func (b *Bridge) observeTaskWithBroadcast(taskID string, udpDisc *discovery.UDPDiscovery, mdnsDisc *discovery.MDNSDiscovery) {
    if b.listener == nil {
        return
    }

    task := b.application.Task(taskID)
    if task == nil {
        return
    }

    for ev := range task.Events() {
        fmt.Printf("[observeTask] event type=%s state=%s\n", ev.Type, ev.State)
        switch ev.Type {
        case protocol.EventProgress:
            fmt.Printf("[observeTask] progress: %.1f%% sent=%d total=%d\n", ev.Percent, ev.Sent, ev.Total)
            b.listener.OnProgress(
                ev.TaskID,
                string(ev.State),
                ev.Percent,
                ev.SpeedMBps,
                ev.Sent,
                ev.Total,
                int64(ev.ETA.Seconds()),
            )
        case protocol.EventState:
            fmt.Printf("[observeTask] state: %s\n", ev.State)
            switch ev.State {
            case protocol.TaskCompleted:
                b.listener.OnComplete(ev.TaskID)
                udpDisc.Stop()
                mdnsDisc.Stop()
                return
            case protocol.TaskFailed, protocol.TaskCancelled:
                errMsg := ""
                if ev.Error != nil {
                    errMsg = ev.Error.Error()
                }
                b.listener.OnError(ev.TaskID, errMsg)
                udpDisc.Stop()
                mdnsDisc.Stop()
                return
            default:
                b.listener.OnProgress(
                    ev.TaskID,
                    string(ev.State),
                    0, 0, 0, 0, 0,
                )
            }
        case protocol.EventError:
            fmt.Printf("[observeTask] error: %v\n", ev.Error)
            errMsg := ""
            if ev.Error != nil {
                errMsg = ev.Error.Error()
            }
            b.listener.OnError(ev.TaskID, errMsg)
        }
    }
}

// CancelTransfer 取消传输
func (b *Bridge) CancelTransfer(taskID string) error {
	return b.application.CancelTask(taskID)
}

// GetHistory 获取历史记录，返回 JSON
func (b *Bridge) GetHistory(limit int) (string, error) {
	entries, err := b.application.RecentHistory(limit)
	if err != nil {
		return "", err
	}

	var result []HistoryInfo
	for _, e := range entries {
		result = append(result, HistoryInfo{
			Time:      e.Time.Format("2006-01-02 15:04:05"),
			Direction: e.Direction,
			Peer:      e.Peer,
			PeerName:  e.PeerName,
			Files:     e.Files,
			FileNames: e.FileNames,
			Size:      e.Size,
			Duration:  e.Duration,
			SpeedMBps: e.SpeedMBps,
			Success:   e.Success,
			SavePath:  e.SavePath,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ClearHistory 清空历史
func (b *Bridge) ClearHistory() error {
	return b.application.ClearHistory()
}

// DeleteHistory 删除单条历史
func (b *Bridge) DeleteHistory(index int) error {
	return b.application.DeleteHistory(index)
}

// observeTask 监听传输任务事件
func (b *Bridge) observeTask(taskID string) {
	if b.listener == nil {
		return
	}

	task := b.application.Task(taskID)
	if task == nil {
		return
	}

	for ev := range task.Events() {
		fmt.Printf("[observeTask] event type=%s state=%s\n", ev.Type, ev.State)
		switch ev.Type {
		case protocol.EventProgress:
			fmt.Printf("[observeTask] progress: %.1f%% sent=%d total=%d\n", ev.Percent, ev.Sent, ev.Total)
			b.listener.OnProgress(
				ev.TaskID,
				string(ev.State),
				ev.Percent,
				ev.SpeedMBps,
				ev.Sent,
				ev.Total,
				int64(ev.ETA.Seconds()),
			)
		case protocol.EventState:
			fmt.Printf("[observeTask] state: %s\n", ev.State)
			switch ev.State {
			case protocol.TaskCompleted:
				b.listener.OnComplete(ev.TaskID)
			case protocol.TaskFailed, protocol.TaskCancelled:
				errMsg := ""
				if ev.Error != nil {
					errMsg = ev.Error.Error()
				}
				b.listener.OnError(ev.TaskID, errMsg)
			default:
				b.listener.OnProgress(
					ev.TaskID,
					string(ev.State),
					0, 0, 0, 0, 0,
				)
			}
		case protocol.EventError:
			fmt.Printf("[observeTask] error: %v\n", ev.Error)
			errMsg := ""
			if ev.Error != nil {
				errMsg = ev.Error.Error()
			}
			b.listener.OnError(ev.TaskID, errMsg)
		}
	}
}