//go:build !windows
// +build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

func startDetached(exe string, args ...string) error {
	cmd := exec.Command(exe, args...)

	// detach on Unix: set process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// disconnect stdio
	devNull, _ := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	if err := cmd.Start(); err != nil {
		return err
	}
	_ = cmd.Process.Release()
	return nil
}
