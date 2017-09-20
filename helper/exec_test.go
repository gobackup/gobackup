package helper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExec(t *testing.T) {
	res, err := Exec("head", "-n1", "./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, res, "package helper")

	res, err = Exec("head -n1 ./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, res, "package helper")

	res, err = Exec("head  -n1  ./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, res, "package helper")

	res, err = Exec("head -n1", "./exec_test.go")
	assert.Nil(t, err)
	assert.Equal(t, res, "package helper")

	res, err = Exec("not-found-command", "foo")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "not-found-command cannot be found")
	assert.Empty(t, res)
}
