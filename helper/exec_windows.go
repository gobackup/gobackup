//go:build windows

package helper

import (
	"os/exec"
	"time"
)

// setProcessGroup is a no-op on Windows, exec.CommandContext kills the direct child only.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.WaitDelay = time.Second
}

// forwardSignals is a no-op on Windows.
func forwardSignals(cmd *exec.Cmd) (stop func()) {
	return func() {}
}
