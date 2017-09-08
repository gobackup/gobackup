package helper

import (
	"bytes"
	"errors"
	"github.com/huacnlee/gobackup/logger"
	"os/exec"
	"strings"
)

// Exec cli commands
func Exec(command string, args ...string) (output string, err error) {
	commands := strings.Split(command, " ")
	command = commands[0]
	commandArgs := []string{}
	if len(commands) > 1 {
		commandArgs = commands[1:]
	}
	commandArgs = append(commandArgs, args...)
	cmd := exec.Command(command, commandArgs...)

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	out, err := cmd.Output()
	if err != nil {
		logger.Debug(command, " ", strings.Join(commandArgs, " "))
		err = errors.New(stdErr.String())
		return
	}

	output = string(out)
	return
}
