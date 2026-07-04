package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// ========== 网络配置 ==========

var NumConns = getEnvInt("VEXIL_NUM_CONNS", 4)
var WindowSize = int64(getEnvInt("VEXIL_WINDOW_SIZE_MB", 128)) * 1024 * 1024
var MaxChunk = int64(getEnvInt("VEXIL_MAX_CHUNK_MB", 16)) * 1024 * 1024

// ========== 超时配置 ==========

var DialTimeout = getEnvDuration("VEXIL_DIAL_TIMEOUT_SEC", 10*time.Second)
var ReadTimeout = getEnvDuration("VEXIL_READ_TIMEOUT_SEC", 10*time.Second)
var AcceptTimeout = getEnvDuration("VEXIL_ACCEPT_TIMEOUT_SEC", 30*time.Second)
var CompleteTimeout = getEnvDuration("VEXIL_COMPLETE_TIMEOUT_SEC", 60*time.Second)
var SHA256Timeout = getEnvDuration("VEXIL_SHA256_TIMEOUT_SEC", 120*time.Second)
var StallInitialTolerance = getEnvDuration("VEXIL_STALL_INITIAL_SEC", 15*time.Second)
var StallCheckInterval = getEnvDuration("VEXIL_STALL_INTERVAL_SEC", 10*time.Second)
var StallThreshold = getEnvInt("VEXIL_STALL_THRESHOLD", 4)

// ========== 接收端配置 ==========

var ACKInterval = getEnvDuration("VEXIL_ACK_INTERVAL_MS", 200*time.Millisecond)
var ResumeSaveInterval = getEnvDuration("VEXIL_RESUME_SAVE_INTERVAL_SEC", 10*time.Second)
var DrainRetries = getEnvInt("VEXIL_DRAIN_RETRIES", 3)
var DrainRetryTimeout = getEnvDuration("VEXIL_DRAIN_RETRY_MS", 200*time.Millisecond)

// ========== 发送端配置 ==========

var DialRetries = getEnvInt("VEXIL_DIAL_RETRIES", 3)
var DialRetryBaseDelay = getEnvDuration("VEXIL_DIAL_RETRY_BASE_SEC", 2*time.Second)
var ConnEstablishGap = getEnvDuration("VEXIL_CONN_ESTABLISH_GAP_MS", 100*time.Millisecond)
var CompleteSendWait = getEnvDuration("VEXIL_COMPLETE_SEND_WAIT_SEC", 3*time.Second)

// ========== 发现配置 ==========

var UDPBroadcastPort = getEnvInt("VEXIL_UDP_BROADCAST_PORT", 54321)
var DiscoverTimeout = getEnvDuration("VEXIL_DISCOVER_TIMEOUT_SEC", 3*time.Second)

// ========== TLS 配置 ==========

var TLSCertFile = os.Getenv("VEXIL_TLS_CERT_FILE")
var TLSKeyFile = os.Getenv("VEXIL_TLS_KEY_FILE")
var TLSEnabled = getEnvBool("VEXIL_TLS_ENABLED", true)
var TLSInsecureSkipVerify = getEnvBool("VEXIL_TLS_INSECURE_SKIP", false)
var TLSSessionCacheSize = getEnvInt("VEXIL_TLS_SESSION_CACHE", 32)

// ========== 其他配置 ==========

var TCPReadBufSize = getEnvInt("VEXIL_TCP_READ_BUF_MB", 4) * 1024 * 1024
var TCPWriteBufSize = getEnvInt("VEXIL_TCP_WRITE_BUF_MB", 4) * 1024 * 1024
var HistoryMaxEntries = getEnvInt("VEXIL_HISTORY_MAX", 20)
var TrackerSampleSize = getEnvInt("VEXIL_TRACKER_SAMPLES", 6)
var TrackerUpdateInterval = getEnvDuration("VEXIL_TRACKER_INTERVAL_MS", 500*time.Millisecond)

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return n
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil && d > 0 {
			return d
		}
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		lower := strings.ToLower(val)
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return defaultVal
}

type AppConfig struct {
	NumConns              int    `json:"num_conns"`
	WindowSizeMB          int    `json:"window_size_mb"`
	MaxChunkMB            int    `json:"max_chunk_mb"`
	DialTimeoutSec        int    `json:"dial_timeout_sec"`
	ReadTimeoutSec        int    `json:"read_timeout_sec"`
	AcceptTimeoutSec      int    `json:"accept_timeout_sec"`
	CompleteTimeoutSec    int    `json:"complete_timeout_sec"`
	SHA256TimeoutSec      int    `json:"sha256_timeout_sec"`
	StallInitialSec       int    `json:"stall_initial_sec"`
	StallIntervalSec      int    `json:"stall_interval_sec"`
	StallThreshold        int    `json:"stall_threshold"`
	ACKIntervalMs         int    `json:"ack_interval_ms"`
	ResumeSaveIntervalSec int    `json:"resume_save_interval_sec"`
	DrainRetries          int    `json:"drain_retries"`
	DrainRetryMs          int    `json:"drain_retry_ms"`
	DialRetries           int    `json:"dial_retries"`
	DialRetryBaseSec      int    `json:"dial_retry_base_sec"`
	ConnEstablishGapMs    int    `json:"conn_establish_gap_ms"`
	CompleteSendWaitSec   int    `json:"complete_send_wait_sec"`
	UDPBroadcastPort      int    `json:"udp_broadcast_port"`
	DiscoverTimeoutSec    int    `json:"discover_timeout_sec"`
	TCPReadBufMB          int    `json:"tcp_read_buf_mb"`
	TCPWriteBufMB         int    `json:"tcp_write_buf_mb"`
	HistoryMaxEntries     int    `json:"history_max_entries"`
	TrackerSamples        int    `json:"tracker_samples"`
	TrackerIntervalMs     int    `json:"tracker_interval_ms"`
	TLSEnabled            bool   `json:"tls_enabled"`
	TLSCertFile           string `json:"tls_cert_file"`
	TLSKeyFile            string `json:"tls_key_file"`
	TLSInsecureSkipVerify bool   `json:"tls_insecure_skip_verify"`
	TLSSessionCacheSize   int    `json:"tls_session_cache_size"`
	DeviceName            string `json:"device_name"`
	Language              string `json:"language"` // zh / en
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		NumConns:              NumConns,
		WindowSizeMB:          int(WindowSize / 1024 / 1024),
		MaxChunkMB:            int(MaxChunk / 1024 / 1024),
		DialTimeoutSec:        int(DialTimeout.Seconds()),
		ReadTimeoutSec:        int(ReadTimeout.Seconds()),
		AcceptTimeoutSec:      int(AcceptTimeout.Seconds()),
		CompleteTimeoutSec:    int(CompleteTimeout.Seconds()),
		SHA256TimeoutSec:      int(SHA256Timeout.Seconds()),
		StallInitialSec:       int(StallInitialTolerance.Seconds()),
		StallIntervalSec:      int(StallCheckInterval.Seconds()),
		StallThreshold:        StallThreshold,
		ACKIntervalMs:         int(ACKInterval.Milliseconds()),
		ResumeSaveIntervalSec: int(ResumeSaveInterval.Seconds()),
		DrainRetries:          DrainRetries,
		DrainRetryMs:          int(DrainRetryTimeout.Milliseconds()),
		DialRetries:           DialRetries,
		DialRetryBaseSec:      int(DialRetryBaseDelay.Seconds()),
		ConnEstablishGapMs:    int(ConnEstablishGap.Milliseconds()),
		CompleteSendWaitSec:   int(CompleteSendWait.Seconds()),
		UDPBroadcastPort:      UDPBroadcastPort,
		DiscoverTimeoutSec:    int(DiscoverTimeout.Seconds()),
		TCPReadBufMB:          TCPReadBufSize / 1024 / 1024,
		TCPWriteBufMB:         TCPWriteBufSize / 1024 / 1024,
		HistoryMaxEntries:     HistoryMaxEntries,
		TrackerSamples:        TrackerSampleSize,
		TrackerIntervalMs:     int(TrackerUpdateInterval.Milliseconds()),
		TLSEnabled:            TLSEnabled,
		TLSCertFile:           TLSCertFile,
		TLSKeyFile:            TLSKeyFile,
		TLSInsecureSkipVerify: TLSInsecureSkipVerify,
		TLSSessionCacheSize:   TLSSessionCacheSize,
		Language:              "zh",
	}
}

func (c *AppConfig) Apply() {
	NumConns = c.NumConns
	WindowSize = int64(c.WindowSizeMB) * 1024 * 1024
	MaxChunk = int64(c.MaxChunkMB) * 1024 * 1024
	DialTimeout = time.Duration(c.DialTimeoutSec) * time.Second
	ReadTimeout = time.Duration(c.ReadTimeoutSec) * time.Second
	AcceptTimeout = time.Duration(c.AcceptTimeoutSec) * time.Second
	CompleteTimeout = time.Duration(c.CompleteTimeoutSec) * time.Second
	SHA256Timeout = time.Duration(c.SHA256TimeoutSec) * time.Second
	StallInitialTolerance = time.Duration(c.StallInitialSec) * time.Second
	StallCheckInterval = time.Duration(c.StallIntervalSec) * time.Second
	StallThreshold = c.StallThreshold
	ACKInterval = time.Duration(c.ACKIntervalMs) * time.Millisecond
	ResumeSaveInterval = time.Duration(c.ResumeSaveIntervalSec) * time.Second
	DrainRetries = c.DrainRetries
	DrainRetryTimeout = time.Duration(c.DrainRetryMs) * time.Millisecond
	DialRetries = c.DialRetries
	DialRetryBaseDelay = time.Duration(c.DialRetryBaseSec) * time.Second
	ConnEstablishGap = time.Duration(c.ConnEstablishGapMs) * time.Millisecond
	CompleteSendWait = time.Duration(c.CompleteSendWaitSec) * time.Second
	UDPBroadcastPort = c.UDPBroadcastPort
	DiscoverTimeout = time.Duration(c.DiscoverTimeoutSec) * time.Second
	TCPReadBufSize = c.TCPReadBufMB * 1024 * 1024
	TCPWriteBufSize = c.TCPWriteBufMB * 1024 * 1024
	HistoryMaxEntries = c.HistoryMaxEntries
	TrackerSampleSize = c.TrackerSamples
	TrackerUpdateInterval = time.Duration(c.TrackerIntervalMs) * time.Millisecond
	TLSEnabled = c.TLSEnabled
	TLSCertFile = c.TLSCertFile
	TLSKeyFile = c.TLSKeyFile
	TLSInsecureSkipVerify = c.TLSInsecureSkipVerify
	TLSSessionCacheSize = c.TLSSessionCacheSize
}