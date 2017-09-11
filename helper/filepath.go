package helper

import (
	"os"
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
func MkdirP(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0777)
	}
}
