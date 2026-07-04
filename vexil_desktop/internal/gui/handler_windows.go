//go:build windows

package gui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (h *Handler) OpenPath(path string) error {
	if strings.Contains(path, ", ") {
		path = strings.Split(path, ", ")[0]
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("路径不存在: %s", path)
	}

	if info.IsDir() {
		return exec.Command("explorer", path).Start()
	}
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
}