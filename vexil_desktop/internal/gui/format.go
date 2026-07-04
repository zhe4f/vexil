package gui

import (
	"fmt"
	"time"

	"vexil/internal/util"
)

func formatSize(bytes int64) string {
	return util.FormatSize(bytes)
}

func formatSpeed(mbps float64) string {
	if mbps < 1 {
		return fmt.Sprintf("%.1f KB/s", mbps*1024)
	}
	return fmt.Sprintf("%.1f MB/s", mbps)
}

func formatETA(d time.Duration) string {
	return util.FormatETA(d)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}