package helper

import (
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
