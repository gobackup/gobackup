package hook

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/longbridgeapp/assert"

	"github.com/gobackup/gobackup/config"
)

// marker returns a hook that appends `name` to the marker file, optionally failing
func marker(t *testing.T, file, name, when string) config.HookConfig {
	return config.HookConfig{
		Name:   name,
		When:   when,
		Script: "echo " + name + " >> " + file,
	}
}

func readMarker(t *testing.T, file string) string {
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return ""
	}
	assert.Nil(t, err)
	return string(data)
}

func markerFile(t *testing.T) string {
	return filepath.Join(t.TempDir(), "marker")
}

func TestContextString(t *testing.T) {
	assert.Equal(t, "m/run", Context{Model: "m", Stage: StageRun}.String())
	assert.Equal(t, "m/database:pg", Context{Model: "m", Stage: StageDatabase, Step: "pg"}.String())
}

func TestResultStatus(t *testing.T) {
	assert.Equal(t, "success", Result{}.Status())
	assert.Equal(t, "failure", Result{Err: errors.New("x")}.Status())
}

func TestRunBeforeOrderAndSkipAfter(t *testing.T) {
	file := markerFile(t)
	hooks := []config.HookConfig{
		marker(t, file, "a", config.HookBefore),
		marker(t, file, "skip", config.HookAfter),
		marker(t, file, "b", config.HookBefore),
	}

	err := RunBefore(hooks, Context{Model: "m", Stage: StageRun})
	assert.Nil(t, err)
	assert.Equal(t, "a\nb\n", readMarker(t, file))
}

func TestRunBeforeFails(t *testing.T) {
	file := markerFile(t)
	hooks := []config.HookConfig{
		marker(t, file, "a", config.HookBefore),
		{Name: "boom", When: config.HookBefore, Script: "echo boom >&2; exit 1"},
		marker(t, file, "b", config.HookBefore),
	}

	err := RunBefore(hooks, Context{Model: "m", Stage: StageDatabase, Step: "pg"})
	assert.NotNil(t, err)
	assert.Equal(t, `before hook "boom" (m/database:pg) failed: boom`, err.Error())
	// hook after the failed one must not run
	assert.Equal(t, "a\n", readMarker(t, file))
}

func TestRunBeforeContinueOnError(t *testing.T) {
	file := markerFile(t)
	hooks := []config.HookConfig{
		{When: config.HookBefore, Script: "exit 1", ContinueOnError: true},
		marker(t, file, "b", config.HookBefore),
	}

	err := RunBefore(hooks, Context{Model: "m", Stage: StageRun})
	assert.Nil(t, err)
	assert.Equal(t, "b\n", readMarker(t, file))
}

func TestRunBeforeTimeout(t *testing.T) {
	hooks := []config.HookConfig{
		{Name: "slow", When: config.HookBefore, Script: "sleep 10", Timeout: 1},
	}

	start := time.Now()
	err := RunBefore(hooks, Context{Model: "m", Stage: StageRun})
	assert.NotNil(t, err)
	assert.Equal(t, `before hook "slow" (m/run) failed: timed out after 1s`, err.Error())
	assert.True(t, time.Since(start) < 5*time.Second)
}

func TestRunBeforeShell(t *testing.T) {
	file := markerFile(t)
	hooks := []config.HookConfig{
		{When: config.HookBefore, Shell: "bash", Script: "echo ${BASH_VERSION:-none} > " + file},
	}

	err := RunBefore(hooks, Context{Model: "m", Stage: StageRun})
	assert.Nil(t, err)
	assert.NotEqual(t, "none\n", readMarker(t, file))

	hooks[0].Shell = "not-found-shell"
	err = RunBefore(hooks, Context{Model: "m", Stage: StageRun})
	assert.NotNil(t, err)
	assert.Equal(t, `before hook "#1" (m/run) failed: not-found-shell cannot be found`, err.Error())
}

func TestRunAfterOnExit(t *testing.T) {
	file := markerFile(t)
	hooks := []config.HookConfig{
		marker(t, file, "before", config.HookBefore),
		marker(t, file, "always", config.HookAfter),
		marker(t, file, "default", config.HookAfter),
		marker(t, file, "success", config.HookAfter),
		marker(t, file, "failure", config.HookAfter),
	}
	hooks[1].OnExit = config.HookOnExitAlways
	hooks[3].OnExit = config.HookOnExitSuccess
	hooks[4].OnExit = config.HookOnExitFailure

	RunAfter(hooks, Context{Model: "m", Stage: StageStorage, Step: "s3"}, Result{})
	assert.Equal(t, "always\ndefault\nsuccess\n", readMarker(t, file))

	os.Remove(file)
	RunAfter(hooks, Context{Model: "m", Stage: StageStorage, Step: "s3"}, Result{Err: errors.New("upload failed")})
	assert.Equal(t, "always\ndefault\nfailure\n", readMarker(t, file))
}

func TestRunAfterFailureDoesNotStop(t *testing.T) {
	file := markerFile(t)
	hooks := []config.HookConfig{
		{When: config.HookAfter, Script: "exit 1"},
		{When: config.HookAfter, Script: "sleep 10", Timeout: 1},
		marker(t, file, "last", config.HookAfter),
	}

	RunAfter(hooks, Context{Model: "m", Stage: StageRun}, Result{})
	assert.Equal(t, "last\n", readMarker(t, file))
}
