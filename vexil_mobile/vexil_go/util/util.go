package util

import (
	"fmt"
	"time"
)

// MinInt64 返回两个 int64 中的较小值
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func FormatSize(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/1024/1024/1024)
	}
}

func FormatSpeed(bytesPerSec float64) string {
	switch {
	case bytesPerSec < 1024:
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	case bytesPerSec < 1024*1024:
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	case bytesPerSec < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/1024/1024)
	default:
		return fmt.Sprintf("%.1f GB/s", bytesPerSec/1024/1024/1024)
	}
}

func FormatETA(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}