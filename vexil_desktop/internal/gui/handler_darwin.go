//go:build darwin

package gui

import (
	"os/exec"
)

func (h *Handler) OpenPath(path string) error {
	return exec.Command("open", path).Start()
}