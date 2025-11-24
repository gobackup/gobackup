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

// PasswordContainsSpecialCharacters verifies whether the provided password includes at least one special character.
//
// Note: This function is currently limited to detecting only the dollar sign ('$') as a special character.
// The dollar sign is considered special because, in certain contexts such as shell scripts (e.g., Bash),
// it denotes the start of a variable name. To prevent unintended variable expansion when using the password
// in such environments, it must be enclosed in single quotes (e.g., 'pa$$word'). Without quotes, the shell
// may interpret the sequence following the '$' as a variable, potentially leading to errors or security issues.
//
// Parameters:
//   - password: The string representing the password to be evaluated.
//
// Returns:
//   - A boolean value: true if the password contains at least one '$' character; false otherwise.
func PasswordContainsSpecialCharacters(password string) bool {
	specialChars := "$"
	for _, char := range password {
		if strings.ContainsRune(specialChars, char) {
			return true
		}
	}
	return false
}
