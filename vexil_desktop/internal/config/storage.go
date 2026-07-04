package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".vexil")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json"), nil
}

func LoadAppConfig() *AppConfig {
	cfg := DefaultAppConfig()
	path, err := configFilePath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	var saved AppConfig
	if err := json.Unmarshal(data, &saved); err != nil {
		return cfg
	}
	merge(&saved, cfg)
	return cfg
}

func (c *AppConfig) Save() error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func merge(src, dst *AppConfig) {
	if src.NumConns > 0 {
		dst.NumConns = src.NumConns
	}
	if src.WindowSizeMB > 0 {
		dst.WindowSizeMB = src.WindowSizeMB
	}
	if src.MaxChunkMB > 0 {
		dst.MaxChunkMB = src.MaxChunkMB
	}
	if src.DialTimeoutSec > 0 {
		dst.DialTimeoutSec = src.DialTimeoutSec
	}
	if src.ReadTimeoutSec > 0 {
		dst.ReadTimeoutSec = src.ReadTimeoutSec
	}
	if src.AcceptTimeoutSec > 0 {
		dst.AcceptTimeoutSec = src.AcceptTimeoutSec
	}
	if src.CompleteTimeoutSec > 0 {
		dst.CompleteTimeoutSec = src.CompleteTimeoutSec
	}
	if src.SHA256TimeoutSec > 0 {
		dst.SHA256TimeoutSec = src.SHA256TimeoutSec
	}
	if src.StallInitialSec > 0 {
		dst.StallInitialSec = src.StallInitialSec
	}
	if src.StallIntervalSec > 0 {
		dst.StallIntervalSec = src.StallIntervalSec
	}
	if src.StallThreshold > 0 {
		dst.StallThreshold = src.StallThreshold
	}
	if src.ACKIntervalMs > 0 {
		dst.ACKIntervalMs = src.ACKIntervalMs
	}
	if src.ResumeSaveIntervalSec > 0 {
		dst.ResumeSaveIntervalSec = src.ResumeSaveIntervalSec
	}
	if src.DrainRetries > 0 {
		dst.DrainRetries = src.DrainRetries
	}
	if src.DrainRetryMs > 0 {
		dst.DrainRetryMs = src.DrainRetryMs
	}
	if src.DialRetries > 0 {
		dst.DialRetries = src.DialRetries
	}
	if src.DialRetryBaseSec > 0 {
		dst.DialRetryBaseSec = src.DialRetryBaseSec
	}
	if src.ConnEstablishGapMs > 0 {
		dst.ConnEstablishGapMs = src.ConnEstablishGapMs
	}
	if src.CompleteSendWaitSec > 0 {
		dst.CompleteSendWaitSec = src.CompleteSendWaitSec
	}
	if src.UDPBroadcastPort > 0 {
		dst.UDPBroadcastPort = src.UDPBroadcastPort
	}
	if src.DiscoverTimeoutSec > 0 {
		dst.DiscoverTimeoutSec = src.DiscoverTimeoutSec
	}
	if src.TCPReadBufMB > 0 {
		dst.TCPReadBufMB = src.TCPReadBufMB
	}
	if src.TCPWriteBufMB > 0 {
		dst.TCPWriteBufMB = src.TCPWriteBufMB
	}
	if src.HistoryMaxEntries > 0 {
		dst.HistoryMaxEntries = src.HistoryMaxEntries
	}
	if src.TrackerSamples > 0 {
		dst.TrackerSamples = src.TrackerSamples
	}
	if src.TrackerIntervalMs > 0 {
		dst.TrackerIntervalMs = src.TrackerIntervalMs
	}
	if src.DeviceName != "" {
		dst.DeviceName = src.DeviceName
	}
	if src.Language != "" {
		dst.Language = src.Language
	}
	dst.TLSEnabled = src.TLSEnabled
	dst.TLSInsecureSkipVerify = src.TLSInsecureSkipVerify
	if src.TLSSessionCacheSize > 0 {
		dst.TLSSessionCacheSize = src.TLSSessionCacheSize
	}
}