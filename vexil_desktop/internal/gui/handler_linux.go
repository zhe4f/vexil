//go:build linux

package gui

import (
	"os/exec"
)

func (h *Handler) OpenPath(path string) error {
	return exec.Command("xdg-open", path).Start()
}