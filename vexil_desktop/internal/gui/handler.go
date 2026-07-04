package gui

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"vexil/internal/app"
	"vexil/internal/discovery"
	"vexil/internal/network"
	"vexil/internal/protocol"
	"vexil/internal/i18n"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/gen2brain/beeep"
)

type Handler struct {
	ctx           context.Context
	app           *app.Application
	mu            sync.Mutex
	subscriptions map[string]context.CancelFunc
	udpDisc       *discovery.UDPDiscovery
	mdnsDisc      *discovery.MDNSDiscovery
}

func NewHandler(app *app.Application) *Handler {
	return &Handler{
		app:           app,
		subscriptions: make(map[string]context.CancelFunc),
	}
}

func (h *Handler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *Handler) StartSend(host string, port int, paths []string, peerName string) (string, error) {
	taskID, err := h.app.StartSend(host, port, paths, peerName)
	if err != nil {
		return "", err
	}
	h.ensureObserving(taskID)
	return taskID, nil
}

func (h *Handler) StartRecv(port int, saveDir string) (string, error) {
	if h.udpDisc != nil {
		h.udpDisc.Stop()
	}
	if h.mdnsDisc != nil {
		h.mdnsDisc.Stop()
	}

	cfg := h.app.Config()
	deviceName := cfg.DeviceName

	h.udpDisc = discovery.NewUDPDiscoveryWithName(deviceName)
	if err := h.udpDisc.Start(port); err != nil {
		fmt.Printf("警告: UDP 发现启动失败: %v\n", err)
	}

	h.mdnsDisc = discovery.NewMDNSDiscoveryWithName(deviceName)
	if err := h.mdnsDisc.Start(port); err != nil {
		fmt.Printf("警告: mDNS 发现启动失败: %v\n", err)
	}

	taskID, err := h.app.StartRecv(port, saveDir)
	if err != nil {
		return "", err
	}
	h.ensureObserving(taskID)
	return taskID, nil
}

func (h *Handler) CancelTransfer(taskID string) error {
	return h.app.CancelTask(taskID)
}

func (h *Handler) GetActiveTasks() []TaskSummary {
	tasks := h.app.ActiveTasks()
	var summaries []TaskSummary
	for _, t := range tasks {
		h.ensureObserving(t.ID())
		summaries = append(summaries, TaskSummary{
			TaskID: t.ID(),
			State:  string(t.State()),
		})
	}
	return summaries
}

func (h *Handler) DiscoverDevices(timeoutSec int) []DeviceInfo {
	if timeoutSec <= 0 {
		timeoutSec = 3
	}
	timeout := time.Duration(timeoutSec) * time.Second

	devices, err := h.app.DiscoverDevices(timeout)
	if err != nil {
		return nil
	}

	var result []DeviceInfo
	for _, d := range devices {
		result = append(result, DeviceInfo{
			Name: d.Name,
			IP:   d.IP,
			Port: d.Port,
		})
	}
	return result
}

func (h *Handler) GetHistory(limit int) []HistoryEntry {
	entries, err := h.app.RecentHistory(limit)
	if err != nil {
		return nil
	}
	var result []HistoryEntry
	for _, e := range entries {
		result = append(result, HistoryEntry{
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
	return result
}

func (h *Handler) ClearHistory() error {
	return h.app.ClearHistory()
}

func (h *Handler) DeleteHistory(index int) error {
	entries, err := h.app.RecentHistory(1000)
	if err != nil {
		return err
	}
	if index < 1 || index > len(entries) {
		return fmt.Errorf("无效索引")
	}
	entry := entries[index-1]

	if entry.Direction == "recv" && entry.SavePath != "" {
		for _, name := range entry.FileNames {
			path := filepath.Join(entry.SavePath, name)
			os.Remove(path)
		}
	}
	if entry.Direction == "send" && entry.SavePath != "" {
		paths := strings.Split(entry.SavePath, ", ")
		for _, p := range paths {
			os.Remove(p)
		}
	}

	return h.app.DeleteHistory(index)
}

func (h *Handler) GetConfig() ConfigInfo {
	cfg := h.app.Config()
	return ConfigInfo{
		NumConns:     cfg.NumConns,
		MaxChunkMB:   cfg.MaxChunkMB,
		WindowSizeMB: cfg.WindowSizeMB,
		TLSEnabled:   cfg.TLSEnabled,
	}
}

func (h *Handler) UpdateConfig(cfg ConfigInfo) error {
	appCfg := h.app.Config()
	appCfg.NumConns = cfg.NumConns
	appCfg.MaxChunkMB = cfg.MaxChunkMB
	appCfg.WindowSizeMB = cfg.WindowSizeMB
	appCfg.TLSEnabled = cfg.TLSEnabled
	return h.app.UpdateConfig(appCfg)
}

func (h *Handler) GetTLSFingerprint() TLSFingerprint {
	return TLSFingerprint{
		SHA256: network.GetAutoCertFingerprint(),
	}
}

func (h *Handler) ensureObserving(taskID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	lang := h.app.Config().Language

	if _, exists := h.subscriptions[taskID]; exists {
		return
	}

	_, cancel := context.WithCancel(h.ctx)
	h.subscriptions[taskID] = cancel

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.subscriptions, taskID)
			h.mu.Unlock()

			if r := recover(); r != nil {
				fmt.Printf("事件处理 panic: %v\n", r)
			}
		}()

		task := h.app.Task(taskID)
		if task == nil {
			return
		}

		for ev := range task.Events() {
			runtime.EventsEmit(h.ctx, "transfer:update", TransferView{
				TaskID:  ev.TaskID,
				State:   string(ev.State),
				Percent: ev.Percent,
				Speed:   formatSpeed(ev.SpeedMBps),
				Sent:    formatSize(ev.Sent),
				Total:   formatSize(ev.Total),
				ETA:     formatETA(ev.ETA),
				Error:   errorString(ev.Error),
			})

			if ev.Type == protocol.EventState {
				switch ev.State {
				case protocol.TaskCompleted:
					beeep.Notify("Vexil", i18n.T(lang, "transfer_complete"), "")
				case protocol.TaskFailed:
					beeep.Notify("Vexil", i18n.T(lang, "transfer_failed"), "")
				case protocol.TaskCancelled:
					beeep.Notify("Vexil", i18n.T(lang, "transfer_cancelled"), "")
				}
			}
		}

		runtime.EventsEmit(h.ctx, "transfer:closed", map[string]string{
			"task_id": taskID,
		})
	}()
}

func (h *Handler) SetLanguage(lang string) error {
	if lang != "zh" && lang != "en" {
		return fmt.Errorf("invalid language")
	}
	cfg := h.app.Config()
	cfg.Language = lang
	return h.app.UpdateConfig(cfg)
}

func (h *Handler) GetLanguage() string {
	return h.app.Config().Language
}

func (h *Handler) OpenFileDialog() ([]string, error) {
	result, err := runtime.OpenMultipleFilesDialog(h.ctx, runtime.OpenDialogOptions{
		Title: "选择要发送的文件",
		// macOS 下不设置 Filters，系统会默认允许所有文件
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *Handler) OpenDirDialog() (string, error) {
	result, err := runtime.OpenDirectoryDialog(h.ctx, runtime.OpenDialogOptions{
		Title: "选择要发送的目录",
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

func (h *Handler) GetDeviceInfo() DeviceInfo {
	cfg := h.app.Config()
	name := cfg.DeviceName
	if name == "" {
		name, _ = os.Hostname()
	}
	ip := getLocalIP()
	return DeviceInfo{
		Name: name,
		IP:   ip,
	}
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}
	return ""
}

func (h *Handler) SetDeviceName(name string) error {
	cfg := h.app.Config()
	cfg.DeviceName = name
	return h.app.UpdateConfig(cfg)
}