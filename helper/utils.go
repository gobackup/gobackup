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
