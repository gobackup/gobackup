// Package hook executes user scripts configured in `hooks` lists around a
// unit of work: the whole model run, a stage (archive, compress, encrypt,
// split) or a single step (one database dump, one storage upload).
package hook

import (
	"fmt"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Stage names, used in logs
const (
	StageRun      = "run"
	StageDatabase = "database"
	StageArchive  = "archive"
	StageCompress = "compress"
	StageEncrypt  = "encrypt"
	StageSplit    = "split"
	StageStorage  = "storage"
)

// Context describes the unit of work the hooks wrap, for logs and error messages
type Context struct {
	Model string
	Stage string
	// Step is the entry name for databases.<db> / storages.<st>, empty otherwise
	Step string
}

func (c Context) String() string {
	if len(c.Step) > 0 {
		return fmt.Sprintf("%s/%s:%s", c.Model, c.Stage, c.Step)
	}
	return fmt.Sprintf("%s/%s", c.Model, c.Stage)
}

// Result of the wrapped unit of work, passed to after hooks
type Result struct {
	Err      error
	Duration time.Duration
}

// Status returns `success` or `failure`
func (r Result) Status() string {
	if r.Err != nil {
		return config.HookOnExitFailure
	}
	return config.HookOnExitSuccess
}

// RunBefore executes before hooks in declaration order.
// Returns the first error that must abort the wrapped work, i.e. a failed hook
// without continue_on_error. Hooks after a failed continue_on_error hook still run.
func RunBefore(hooks []config.HookConfig, ctx Context) error {
	logger := logger.Tag("Hook")

	for i, hook := range hooks {
		if hook.When != config.HookBefore {
			continue
		}

		name := hook.DisplayName(i)
		logger.Infof("Run before hook %q (%s)", name, ctx)

		if err := run(hook); err != nil {
			if hook.ContinueOnError {
				logger.Warnf("Before hook %q (%s) failed, continue: %v", name, ctx, err)
				continue
			}
			return fmt.Errorf("before hook %q (%s) failed: %v", name, ctx, err)
		}

		logger.Infof("Run before hook %q (%s) succeeded", name, ctx)
	}

	return nil
}

// RunAfter executes after hooks in declaration order, filtered by on_exit
// against the result of the wrapped work. Hook failures are logged and never
// change the outcome of the backup.
func RunAfter(hooks []config.HookConfig, ctx Context, result Result) {
	logger := logger.Tag("Hook")
	status := result.Status()

	for i, hook := range hooks {
		if hook.When != config.HookAfter {
			continue
		}

		name := hook.DisplayName(i)

		onExit := hook.OnExit
		if len(onExit) == 0 {
			onExit = config.HookOnExitAlways
		}
		if onExit != config.HookOnExitAlways && onExit != status {
			logger.Infof("Skip after hook %q (%s): on_exit is %s, status is %s", name, ctx, onExit, status)
			continue
		}

		logger.Infof("Run after hook %q (%s)", name, ctx)
		if err := run(hook); err != nil {
			logger.Errorf("After hook %q (%s) failed: %v", name, ctx, err)
			continue
		}
		logger.Infof("Run after hook %q (%s) succeeded", name, ctx)
	}
}

func run(hook config.HookConfig) error {
	_, err := helper.ExecScriptWithOptions(hook.Script, helper.ScriptOptions{
		Shell:   hook.Shell,
		Timeout: time.Duration(hook.Timeout) * time.Second,
		Stdout:  true,
	})
	return err
}
