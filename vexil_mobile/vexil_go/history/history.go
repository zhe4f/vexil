package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vexil_go/config"
	"vexil_go/util"
)

type Entry struct {
	Time      time.Time `json:"time"`
	Direction string    `json:"direction"`
	Peer      string    `json:"peer"`
	PeerName  string    `json:"peer_name,omitempty"`
	Files     int       `json:"files"`
	FileNames []string  `json:"file_names"`
	Size      int64     `json:"size"`
	Duration  float64   `json:"duration_sec"`
	SpeedMBps float64   `json:"speed_mbps"`
	Success   bool      `json:"success"`
	SavePath  string    `json:"save_path,omitempty"`
}

var historyDir string

func SetDir(dir string) {
	historyDir = dir
}

func filePath() (string, error) {
	dir := filepath.Join(historyDir, ".vexil")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "history.json"), nil
}

func Save(entry Entry) error {
	path, err := filePath()
	if err != nil {
		fmt.Printf("[History] filePath error: %v\n", err)
		return err
	}
	fmt.Printf("[History] 保存到: %s\n", path)

	var entries []Entry
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &entries)
	}

	entries = append([]Entry{entry}, entries...)
	if len(entries) > config.HistoryMaxEntries {
		entries = entries[:config.HistoryMaxEntries]
	}

	fmt.Printf("[History] 写入 %d 条记录\n", len(entries))
	newData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, newData, 0644)
}

func Delete(index int) error {
	path, err := filePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	if index < 1 || index > len(entries) {
		return fmt.Errorf("无效的索引: %d (有效范围 1-%d)", index, len(entries))
	}

	entries = append(entries[:index-1], entries[index:]...)

	newData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, newData, 0644)
}

func List(limit int) ([]Entry, error) {
	path, err := filePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	if limit > 0 && limit < len(entries) {
		return entries[:limit], nil
	}
	return entries, nil
}

func Clear() error {
	path, err := filePath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte("[]"), 0644)
}

func FormatEntry(e Entry, index int) string {
	status := "✅"
	if !e.Success {
		status = "❌"
	}

	dir := ""
	if e.Direction == "send" {
		dir = "↑ 发送至"
	} else {
		dir = "↓ 接收自"
	}

	line := fmt.Sprintf("  [%d] %s  %s  %s %s\n",
		index+1,
		e.Time.Format("2006-01-02 15:04:05"),
		status,
		dir,
		e.Peer,
	)

	fileInfo := formatFileInfo(e)
	line += fmt.Sprintf("       %s\n", fileInfo)

	stats := fmt.Sprintf("       %s", util.FormatSize(e.Size))
	if e.Duration > 0 {
		stats += fmt.Sprintf("  %.1fs", e.Duration)
	}
	if e.SpeedMBps > 0 {
		stats += fmt.Sprintf("  %.1f MB/s", e.SpeedMBps)
	}
	line += stats

	if e.SavePath != "" {
		line += fmt.Sprintf("  → %s", e.SavePath)
	}

	return line
}

func formatFileInfo(e Entry) string {
	if len(e.FileNames) == 0 {
		return fmt.Sprintf("%d个文件", e.Files)
	}

	if len(e.FileNames) == 1 {
		name := e.FileNames[0]
		if strings.HasSuffix(name, "/") {
			dirName := strings.TrimSuffix(name, "/")
			if e.Files == 1 {
				return fmt.Sprintf("%s/ (1个文件)", dirName)
			}
			return fmt.Sprintf("%s/ (%d个文件)", dirName, e.Files)
		}
		return name
	}

	if len(e.FileNames) <= 3 {
		return strings.Join(e.FileNames, ", ")
	}
	return fmt.Sprintf("%s, %s, %s ...等%d项",
		e.FileNames[0], e.FileNames[1], e.FileNames[2], len(e.FileNames))
}