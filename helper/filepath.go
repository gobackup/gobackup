package helper

import (
	"os"
	"path"
	"path/filepath"

	"github.com/gobackup/gobackup/logger"
)

// IsExistsPath check path exist
func IsExistsPath(p string) bool {
	_, err := os.Stat(p)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// MkdirP like mkdir -p
func MkdirP(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0750)
	}
	return nil
}

// ExplandHome ~/foo -> /home/jason/foo
func ExplandHome(filePath string) string {
	if len(filePath) < 2 {
		return filePath
	}

	if filePath[:2] != "~/" {
		return filePath
	}

	return path.Join(os.Getenv("HOME"), filePath[2:])
}

// Convert a file path into an absolute path
func AbsolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	path = ExplandHome(path)

	path, err := filepath.Abs(path)
	if err != nil {
		logger.Error("Convert config file path to absolute path failed: ", err)
		return path
	}

	return path
}
