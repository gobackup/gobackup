//go:build !windows

package helper

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// killGracePeriod is how long a timed out script gets after SIGTERM before SIGKILL
const killGracePeriod = 5 * time.Second

// setProcessGroup runs the command in its own process group so that on
// timeout the whole group (including child processes) can be terminated:
// SIGTERM first, SIGKILL after killGracePeriod.
//
// Only used when a timeout is configured. Without timeout the script stays in
// gobackup's process group, exactly as before, so terminal Ctrl-C reaches it.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		pgid := -cmd.Process.Pid
		if err := syscall.Kill(pgid, syscall.SIGTERM); err != nil {
			return err
		}
		time.AfterFunc(killGracePeriod, func() {
			_ = syscall.Kill(pgid, syscall.SIGKILL)
		})
		return nil
	}
	cmd.WaitDelay = killGracePeriod + time.Second
}

// forwardSignals forwards SIGINT / SIGTERM received by gobackup to the
// command's process group while it runs, so a script in its own group is
// still interrupted by Ctrl-C. The signal is re-raised on gobackup itself
// after the script exits, restoring the default behaviour.
// Returns a stop function to call once the command has finished.
func forwardSignals(cmd *exec.Cmd) (stop func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	var received os.Signal

	go func() {
		select {
		case sig := <-sigs:
			received = sig
			if cmd.Process != nil {
				_ = syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal))
			}
		case <-done:
		}
	}()

	return func() {
		close(done)
		signal.Stop(sigs)
		if received != nil {
			_ = syscall.Kill(os.Getpid(), received.(syscall.Signal))
		}
	}
}
