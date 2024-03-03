package helper

import (
	"os/exec"
	"runtime"
	"strings"
)

var (
	// IsGnuTar show tar type
	IsGnuTar = false
)

func init() {
	checkIsGnuTar()
}

func checkIsGnuTar() {
	out, _ := Exec("tar", "--version")
	IsGnuTar = strings.Contains(out, "GNU")
}

// CleanHost clean host url ftp://foo.bar.com -> foo.bar.com
func CleanHost(host string) string {
	// ftp://ftp.your-host.com -> ftp.your-host.com
	if strings.Contains(host, "://") {
		return strings.Split(host, "://")[1]
	}

	return host
}

// FormatEndpoint to add `https://` prefix if not present
func FormatEndpoint(endpoint string) string {
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	return endpoint
}

// IsWindows check if the OS is Windows
func IsWindows() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}

// IsExistsBin search for binary in PATH and returns true if exists
func IsExistsBin(bin string) bool {
	if path, err := exec.LookPath(bin); err == nil {
		if len(path) > 0 {
			return true
		}
		return false
	}
	return false
}
