package helper

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/huacnlee/gobackup/logger"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	spaceRegexp = regexp.MustCompile("[\\s]+")
)

// Exec cli commands
func Exec(command string, args ...string) (output string, err error) {
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
	cmd.Stderr = &stdErr

	// logger.Debug(fullCommand, " ", strings.Join(commandArgs, " "))

	out, err := cmd.Output()
	if err != nil {
		logger.Debug(fullCommand, " ", strings.Join(commandArgs, " "))
		err = errors.New(stdErr.String())
		return
	}

	output = strings.Trim(string(out), "\n")
	return
}
