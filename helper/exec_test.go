package helper

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/longbridgeapp/assert"
)

func TestExec(t *testing.T) {
	out, err := Exec("head", "-n1", "./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper")

	out, err = Exec("head -n1 ./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper")

	out, err = Exec("head  -n1  ./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper")

	out, err = Exec("head -n1", "./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper")

	out, err = Exec("not-found-command", "foo")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "not-found-command cannot be found")
	assert.Empty(t, out)
}

func TestExecWithStdio(t *testing.T) {
	out, err := ExecWithStdio("head -n1", false, "./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper")

	out, err = ExecWithStdio("head -n1", true, "./exec_test.go")
	assert.Nil(t, err)
	assert.Empty(t, out)
}

func TestExecScriptWithStdio(t *testing.T) {
	out, err := ExecScriptWithStdio("head -n1 ./exec_test.go\n# This is a comment\nhead -n1 ./exec_test.go | wc -c | sed 's/^[[:space:]]*//'\necho hello world\necho foobar", false)
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper\n15\nhello world\nfoobar")

	out, err = ExecScriptWithStdio("head -n1", true)
	assert.Nil(t, err)
	assert.Empty(t, out)
}

func TestExecScript(t *testing.T) {
	out, err := ExecScript("head -n1 ./exec_test.go\n# This is a comment\necho hello world")
	assert.Nil(t, err)
	assert.Equal(t, out, "package helper\nhello world")
}

func TestExecScriptWithOptions(t *testing.T) {
	// default shell
	out, err := ExecScriptWithOptions("echo hello\necho world", ScriptOptions{})
	assert.Nil(t, err)
	assert.Equal(t, out, "hello\nworld")

	// explicit shell
	out, err = ExecScriptWithOptions("echo ${BASH_VERSION:-none}", ScriptOptions{Shell: "bash"})
	assert.Nil(t, err)
	assert.NotEqual(t, out, "none")
	assert.True(t, len(out) > 0)

	// stdout passthrough
	out, err = ExecScriptWithOptions("echo hello", ScriptOptions{Stdout: true})
	assert.Nil(t, err)
	assert.Empty(t, out)

	// shell not found
	_, err = ExecScriptWithOptions("echo hello", ScriptOptions{Shell: "not-found-shell"})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "not-found-shell cannot be found")

	// non-zero exit returns stderr
	_, err = ExecScriptWithOptions("echo oops >&2\nexit 3", ScriptOptions{})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "oops")

	// non-zero exit without stderr returns exit error
	_, err = ExecScriptWithOptions("exit 3", ScriptOptions{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "exit status 3")
}

func TestExecScriptWithOptionsTimeout(t *testing.T) {
	start := time.Now()
	_, err := ExecScriptWithOptions("sleep 10", ScriptOptions{Timeout: 300 * time.Millisecond})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "timed out after 300ms")
	assert.True(t, time.Since(start) < 3*time.Second)

	// background child holding stdout must be killed with the process group,
	// otherwise Run blocks until the child exits on its own.
	start = time.Now()
	_, err = ExecScriptWithOptions("sleep 10 &\nsleep 10", ScriptOptions{Timeout: 300 * time.Millisecond})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "timed out after 300ms")
	assert.True(t, time.Since(start) < 3*time.Second)

	// finishes within timeout
	out, err := ExecScriptWithOptions("echo fast", ScriptOptions{Timeout: 5 * time.Second})
	assert.Nil(t, err)
	assert.Equal(t, out, "fast")

	// a script that traps TERM gets a chance to clean up before SIGKILL
	cleaned := filepath.Join(t.TempDir(), "cleaned")
	start = time.Now()
	_, err = ExecScriptWithOptions("trap 'echo yes > "+cleaned+"; exit 0' TERM\nsleep 10 &\nwait", ScriptOptions{Timeout: 300 * time.Millisecond})
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "timed out after 300ms")
	assert.True(t, time.Since(start) < 3*time.Second)
	data, readErr := os.ReadFile(cleaned)
	assert.Nil(t, readErr)
	assert.Equal(t, "yes\n", string(data))
}
