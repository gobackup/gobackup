package helper

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/gobackup/gobackup/logger"
	"github.com/google/uuid"
)

var (
	spaceRegexp = regexp.MustCompile(`[\s]+`)
)

// Exec executes a CLI command and returns its output or an error.
// It splits the command string into the executable and arguments, then runs the command
// without directing stdout to the console. Environment variables from the current process
// are inherited.
func Exec(command string, args ...string) (output string, err error) {
	return ExecWithStdio(command, false, args...)
}

// ExecWithStdio executes a CLI command with control over stdout redirection.
// If stdout is true, output is directed to the console; otherwise, it is captured and returned.
// The function splits the command string, looks up the executable path, and runs the command,
// capturing stderr for error reporting. Additional arguments can be appended.
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

// ExecShell executes a command string through a shell (/bin/sh -c) to enable shell processing,
// such as quote handling and variable expansion. It captures the output and returns it or an error.
// This is useful for commands that require shell interpretation, like those with quoted arguments.
func ExecShell(command string) (output string, err error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = os.Environ()

	var stdErr bytes.Buffer
	var stdOut bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut

	err = cmd.Run()
	if err != nil {
		logger.Debug("sh -c", " ", command)
		err = errors.New(stdErr.String())
	}
	output = strings.Trim(stdOut.String(), "\n")

	return
}

// ExecScriptWithStdio executes a multi-line script with control over stdout redirection.
// It creates a temporary shell script file, writes the script content, makes it executable,
// and runs it via 'sh'. The temporary file is removed after execution. If stdout is true,
// output is directed to the console; otherwise, it is captured.
func ExecScriptWithStdio(script string, stdout bool) (string, error) {
	tmpFileName, _ := uuid.NewUUID()
	tmpFile := path.Join(os.TempDir(), tmpFileName.String())

	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}
	f.WriteString(script)
	err = os.Chmod(tmpFile, 0755)
	if err != nil {
		return "", err
	}

	defer f.Close()
	defer os.Remove(tmpFile)

	return ExecWithStdio("sh", stdout, tmpFile)
}

// ExecScript executes a multi-line script and returns its output or an error.
// It uses ExecScriptWithStdio internally without directing stdout to the console.
func ExecScript(script string) (output string, err error) {
	return ExecScriptWithStdio(script, false)
}
