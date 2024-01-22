package helper

import (
	"testing"

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
