package gui

// TransferView 前端事件数据结构
type TransferView struct {
	TaskID  string  `json:"task_id"`
	State   string  `json:"state"`
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed"`
	Sent    string  `json:"sent"`
	Total   string  `json:"total"`
	ETA     string  `json:"eta"`
	Error   string  `json:"error,omitempty"`
}

// TaskSummary 任务摘要
type TaskSummary struct {
	TaskID string `json:"task_id"`
	State  string `json:"state"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// HistoryEntry 历史记录
type HistoryEntry struct {
    Time      string   `json:"time"`
    Direction string   `json:"direction"`
    Peer      string   `json:"peer"`
	PeerName string    `json:"peer_name,omitempty"`
    Files     int      `json:"files"`
    FileNames []string `json:"file_names"`
    Size      int64    `json:"size"`
    Duration  float64  `json:"duration_sec"`
    SpeedMBps float64  `json:"speed_mbps"`
    Success   bool     `json:"success"`
    SavePath  string   `json:"save_path,omitempty"`
}

// ConfigInfo 配置信息
type ConfigInfo struct {
	NumConns     int  `json:"num_conns"`
	MaxChunkMB   int  `json:"max_chunk_mb"`
	WindowSizeMB int  `json:"window_size_mb"`
	TLSEnabled   bool `json:"tls_enabled"`
}

// TLSFingerprint 证书指纹信息
type TLSFingerprint struct {
	SHA256 string `json:"sha256"`
}