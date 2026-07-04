package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"vexil/internal/config"
	"vexil/internal/discovery"
	"vexil/internal/history"
	"vexil/internal/protocol"
	"vexil/internal/receiver"
	"vexil/internal/sender"
)

// RunningTask 运行中的传输任务
type RunningTask struct {
	id      string
	state   protocol.TaskState
	eventCh chan protocol.TaskEvent
	cancel  context.CancelFunc
	done    bool
	peer    string
	stats   interface{}
}

func (t *RunningTask) Events() <-chan protocol.TaskEvent { return t.eventCh }
func (t *RunningTask) ID() string                         { return t.id }
func (t *RunningTask) State() protocol.TaskState          { return t.state }

type Application struct {
	mu    sync.Mutex
	tasks map[string]*RunningTask
	cfg   *config.AppConfig
}

func New(cfg *config.AppConfig) *Application {
	return &Application{
		tasks: make(map[string]*RunningTask),
		cfg:   cfg,
	}
}

func generateTaskID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// StartSend 启动发送任务
func (a *Application) StartSend(host string, port int, paths []string, peerName string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	peer := fmt.Sprintf("%s:%d", host, port)
	for _, t := range a.tasks {
		if t.peer == peer && !t.done {
			return "", fmt.Errorf("已有传输连接到 %s", peer)
		}
	}

	taskID := generateTaskID()
	eventCh := make(chan protocol.TaskEvent, 16)
	ctx, cancel := context.WithCancel(context.Background())

	task := &RunningTask{
		id:      taskID,
		state:   protocol.TaskPreparing,
		eventCh: eventCh,
		cancel:  cancel,
		peer:    peer,
	}
	a.tasks[taskID] = task

	tcfg := config.Snapshot()

	go func() {
		defer func() {
			a.mu.Lock()
			task.done = true
			a.mu.Unlock()

			time.Sleep(5 * time.Second)
			a.mu.Lock()
			if a.tasks[taskID] == task {
				delete(a.tasks, taskID)
			}
			a.mu.Unlock()
		}()

		defer func() {
			if r := recover(); r != nil {
				select {
				case eventCh <- protocol.TaskEvent{
					TaskID: taskID,
					Type:   protocol.EventError,
					Error:  fmt.Errorf("发送任务 panic: %v", r),
				}:
				default:
				}
			}
			close(eventCh)
		}()

		s := sender.New(sender.Config{
			Files:    paths,
			Host:     host,
			Port:     port,
			PeerName: peerName, // 传入发送者名称
		})

		err := s.Run(ctx, tcfg, eventCh, taskID)

		entry := history.Entry{
			Time:      time.Now(),
			Direction: "send",
			Peer:      peer,
			PeerName:  peerName,
			Files:     s.Stats.FileCount,
			FileNames: s.Stats.FileNames,
			Size:      s.Stats.TotalSize,
			Duration:  s.Stats.Duration,
			SpeedMBps: s.Stats.SpeedMBps,
			Success:   err == nil,
			SavePath:  strings.Join(paths, ", "),
		}
		history.Save(entry)

		task.stats = s.Stats
	}()

	return taskID, nil
}

// StartRecv 启动接收任务
func (a *Application) StartRecv(port int, saveDir string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	taskID := generateTaskID()
	eventCh := make(chan protocol.TaskEvent, 16)
	ctx, cancel := context.WithCancel(context.Background())

	task := &RunningTask{
		id:      taskID,
		state:   protocol.TaskPreparing,
		eventCh: eventCh,
		cancel:  cancel,
		peer:    fmt.Sprintf(":%d", port),
	}
	a.tasks[taskID] = task

	tcfg := config.Snapshot()

	go func() {
		defer func() {
			a.mu.Lock()
			task.done = true
			a.mu.Unlock()

			time.Sleep(5 * time.Second)
			a.mu.Lock()
			if a.tasks[taskID] == task {
				delete(a.tasks, taskID)
			}
			a.mu.Unlock()
		}()

		defer func() {
			if r := recover(); r != nil {
				select {
				case eventCh <- protocol.TaskEvent{
					TaskID: taskID,
					Type:   protocol.EventError,
					Error:  fmt.Errorf("接收任务 panic: %v", r),
				}:
				default:
				}
			}
			close(eventCh)
		}()

		r := receiver.New(receiver.Config{
			SaveDir: saveDir,
			Port:    port,
		})

		err := r.Run(ctx, tcfg, eventCh, taskID)

		peer := fmt.Sprintf("port:%d", port)
		if r.PeerIP != "" {
			peer = fmt.Sprintf("%s:%d", r.PeerIP, port)
		}

		entry := history.Entry{
			Time:      time.Now(),
			Direction: "recv",
			Peer:      peer,
			PeerName:  r.PeerName,
			Files:     r.Stats.FileCount,
			FileNames: r.Stats.FileNames,
			Size:      r.Stats.TotalSize,
			Duration:  r.Stats.Duration,
			SpeedMBps: r.Stats.SpeedMBps,
			Success:   err == nil,
			SavePath:  saveDir,
		}
		history.Save(entry)

		task.stats = r.Stats
	}()

	return taskID, nil
}

// CancelTask 取消任务
func (a *Application) CancelTask(taskID string) error {
	a.mu.Lock()
	task, ok := a.tasks[taskID]
	a.mu.Unlock()

	if !ok {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	task.cancel()
	return nil
}

// Task 获取任务
func (a *Application) Task(taskID string) *RunningTask {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tasks[taskID]
}

// ActiveTasks 获取所有活跃任务
func (a *Application) ActiveTasks() []*RunningTask {
	a.mu.Lock()
	defer a.mu.Unlock()

	var tasks []*RunningTask
	for _, t := range a.tasks {
		if !t.done {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// DiscoverDevices 发现设备（修复为调用 Discover）
func (a *Application) DiscoverDevices(timeout time.Duration) ([]discovery.Device, error) {
	udpDisc := discovery.NewUDPDiscovery()
	mdnsDisc := discovery.NewMDNSDiscovery()

	var channels []<-chan discovery.Device

	ch1, err := udpDisc.Discover(timeout)
	if err == nil {
		channels = append(channels, ch1)
	}
	defer udpDisc.Stop()

	ch2, err := mdnsDisc.Discover(timeout)
	if err == nil {
		channels = append(channels, ch2)
	}
	defer mdnsDisc.Stop()

	if len(channels) == 0 {
		return nil, fmt.Errorf("没有可用的发现方式")
	}

	return discovery.MergeDevices(channels...), nil
}

// RecentHistory 获取最近历史
func (a *Application) RecentHistory(limit int) ([]history.Entry, error) {
	return history.List(limit)
}

// ClearHistory 清空历史
func (a *Application) ClearHistory() error {
	return history.Clear()
}

// DeleteHistory 删除历史记录
func (a *Application) DeleteHistory(index int) error {
	return history.Delete(index)
}

// Config 获取配置
func (a *Application) Config() *config.AppConfig {
	return a.cfg
}

// UpdateConfig 更新配置
func (a *Application) UpdateConfig(cfg *config.AppConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg = cfg
	cfg.Apply()
	return cfg.Save()
}