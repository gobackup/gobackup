package helper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gobackup/gobackup/logger"
	"github.com/google/uuid"
)

var (
	spaceRegexp = regexp.MustCompile(`[\s]+`)
)

// Exec cli commands
func Exec(command string, args ...string) (output string, err error) {
	return ExecWithStdio(command, false, args...)
}

func ExecWithStdio(command string, stdout bool, args ...string) (output string, err error) {
	commands := spaceRegexp.Split(command, -1)
	command = commands[0]
	commandArgs := []string{}
	if len(commands) > 1 {
		commandArgs = commands[1:]
	}
	if len(args) > 0 {
		commandArgs = append(commandArgs, args...)
	}

	fullCommand, err := exec.LookPath(command)
	if err != nil {
		return "", fmt.Errorf("%s cannot be found", command)
	}

	cmd := exec.Command(fullCommand, commandArgs...)
	cmd.Env = os.Environ()

	var stdErr bytes.Buffer
	var stdOut bytes.Buffer
	cmd.Stderr = &stdErr

	if stdout {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = &stdOut
	}

	err = cmd.Run()
	if err != nil {
		logger.Debug(fullCommand, " ", strings.Join(commandArgs, " "))
		err = errors.New(stdErr.String())
	}
	output = strings.Trim(stdOut.String(), "\n")

	return
}

// ScriptOptions for ExecScriptWithOptions
type ScriptOptions struct {
	// Shell is the interpreter the script file is passed to, `sh` by default
	Shell string
	// Timeout kills the script (and its process group) when exceeded, 0 means no limit
	Timeout time.Duration
	// Stdout true to pass script output to os.Stdout instead of returning it
	Stdout bool
}

// ExecScriptWithOptions writes the script into a temp file and executes it as `<shell> <tmpfile>`
func ExecScriptWithOptions(script string, opts ScriptOptions) (output string, err error) {
	shell := opts.Shell
	if len(shell) == 0 {
		shell = "sh"
	}

	fullCommand, err := exec.LookPath(shell)
	if err != nil {
		return "", fmt.Errorf("%s cannot be found", shell)
	}

	tmpFileName, _ := uuid.NewUUID()
	tmpFile := path.Join(os.TempDir(), tmpFileName.String())

	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)

	if _, err = f.WriteString(script); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = os.Chmod(tmpFile, 0755); err != nil {
		return "", err
	}

	ctx := context.Background()
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, fullCommand, tmpFile)
	cmd.Env = os.Environ()

	var stdErr bytes.Buffer
	var stdOut bytes.Buffer
	cmd.Stderr = &stdErr
	if opts.Stdout {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = &stdOut
	}

	if opts.Timeout > 0 {
		// own process group so the whole script tree can be terminated on timeout
		setProcessGroup(cmd)
		if err = cmd.Start(); err != nil {
			return "", err
		}
		stop := forwardSignals(cmd)
		err = cmd.Wait()
		stop()
	} else {
		err = cmd.Run()
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("timed out after %s", opts.Timeout)
	}
	if err != nil {
		logger.Debug(fullCommand, " ", tmpFile)
		if msg := strings.TrimSpace(stdErr.String()); len(msg) > 0 {
			return "", errors.New(msg)
		}
		return "", err
	}

	output = strings.Trim(stdOut.String(), "\n")
	return
}

// Execute multiple line script with stdio
func ExecScriptWithStdio(script string, stdout bool) (string, error) {
	return ExecScriptWithOptions(script, ScriptOptions{Stdout: stdout})
}

// Execute multiple line script
func ExecScript(script string) (output string, err error) {
	return ExecScriptWithStdio(script, false)
}
