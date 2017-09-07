package helper

import (
	"fmt"
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
	logger.Debug(command, fmt.Sprint(commandArgs...))
	out, err := cmd.Output()
	if err != nil {
		return
	}

	output = string(out)
	return
}
