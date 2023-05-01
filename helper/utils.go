package helper

import (
	"strings"
)

// IsGnuTar show tar type
var IsGnuTar = false

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
