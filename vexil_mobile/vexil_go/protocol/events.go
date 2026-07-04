package protocol

import "time"

// EventType 事件类型
type EventType string

const (
	EventState    EventType = "state"
	EventProgress EventType = "progress"
	EventError    EventType = "error"
)

// TaskState 任务状态
type TaskState string

const (
	TaskPreparing  TaskState = "preparing"
	TaskConnecting TaskState = "connecting"
	TaskRunning    TaskState = "running"
	TaskFinalizing TaskState = "finalizing"
	TaskCompleted  TaskState = "completed"
	TaskFailed     TaskState = "failed"
	TaskCancelled  TaskState = "cancelled"
)

// TaskEvent 统一的传输事件
type TaskEvent struct {
	TaskID string
	Type   EventType

	// EventState 时有效
	State TaskState

	// EventProgress 时有效（由 tracker 控制频率）
	Percent   float64
	SpeedMBps float64
	Sent      int64
	Total     int64
	ETA       time.Duration

	// EventError 时有效
	Error error
}

// TransferConfig 传输配置快照（纯数据，无依赖）
type TransferConfig struct {
	NumConns           int
	WindowSize         int64
	MaxChunk           int64
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	AcceptTimeout      time.Duration
	CompleteTimeout    time.Duration
	SHA256Timeout      time.Duration
	StallInitial       time.Duration
	StallInterval      time.Duration
	StallThreshold     int
	ACKInterval        time.Duration
	ResumeSaveInterval time.Duration
	DrainRetries       int
	DrainRetryTimeout  time.Duration
	DialRetries        int
	DialRetryBaseDelay time.Duration
	ConnEstablishGap   time.Duration
	CompleteSendWait   time.Duration
	TCPReadBufSize     int
	TCPWriteBufSize    int
	TrackerSamples     int
	TrackerInterval    time.Duration
	// TLS 配置
	TLSEnabled            bool
	TLSCertFile           string
	TLSKeyFile            string
	TLSInsecureSkipVerify bool
	TLSSessionCacheSize   int
}