package database

import (
	"fmt"
	"os"
	"os/exec"
)

// Base interface
type Base interface {
	Perform() error
}

// New - initialize Database
func New(model string) (ctx Base, err error) {
	switch model {
	case "mysql":
		ctx = MySQL{}
	case "redis":
		ctx = newRedis()
	default:
		err = fmt.Errorf("%s model is not implement", model)
	}

	return
}

func isExistsPath(p string) bool {
	_, err := os.Stat(p)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func ensureDir(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0777)
	}
}

func run(command string, args ...string) (output string, err error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	output = string(out)
	return
}
