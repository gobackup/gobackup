package helper

import (
	"os/exec"
)

func Run(command string, args ...string) (output string, err error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	output = string(out)
	return
}
