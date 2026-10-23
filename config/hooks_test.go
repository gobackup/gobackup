package config

import (
	"os"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestHooksConfig(t *testing.T) {
	assert.Nil(t, Init(testConfigFile))

	wLock.Lock()
	defer wLock.Unlock()

	model := GetModelConfigByName("hooks_test")
	assert.NotNil(t, model)

	// model level
	assert.Len(t, model.Hooks, 3)

	hook := model.Hooks[0]
	assert.Equal(t, "lock", hook.Name)
	assert.Equal(t, "lock", hook.DisplayName(0))
	assert.Equal(t, HookBefore, hook.When)
	assert.Equal(t, "", hook.OnExit)
	assert.Equal(t, `echo "lock"`, hook.Script)
	assert.Equal(t, "", hook.Shell)
	assert.Equal(t, 0, hook.Timeout)
	assert.Equal(t, false, hook.ContinueOnError)

	hook = model.Hooks[1]
	assert.Equal(t, "unlock", hook.Name)
	assert.Equal(t, HookAfter, hook.When)
	assert.Equal(t, HookOnExitFailure, hook.OnExit)
	assert.Equal(t, "bash", hook.Shell)
	assert.Equal(t, 30, hook.Timeout)
	assert.Equal(t, "echo \"unlock\"\necho \"second line\"\n", hook.Script)

	hook = model.Hooks[2]
	assert.Equal(t, "#3", hook.Name)
	assert.Equal(t, "#3", hook.DisplayName(0))
	assert.Equal(t, true, hook.ContinueOnError)

	// database step
	db := model.GetDatabaseByName("pg")
	assert.Len(t, db.Hooks, 1)
	assert.Equal(t, HookBefore, db.Hooks[0].When)
	assert.Equal(t, `echo "db before"`, db.Hooks[0].Script)

	// stages
	assert.Len(t, model.ArchiveHooks, 1)
	assert.Equal(t, HookAfter, model.ArchiveHooks[0].When)
	assert.Equal(t, "", model.ArchiveHooks[0].OnExit)
	assert.Equal(t, "#1", model.ArchiveHooks[0].Name)

	assert.Len(t, model.CompressWith.Hooks, 1)
	assert.Equal(t, `echo "compress before"`, model.CompressWith.Hooks[0].Script)

	assert.Len(t, model.EncryptWith.Hooks, 1)
	assert.Equal(t, HookOnExitSuccess, model.EncryptWith.Hooks[0].OnExit)

	assert.Len(t, model.SplitterHooks, 1)
	assert.Equal(t, `echo "split before"`, model.SplitterHooks[0].Script)

	// storage step
	assert.Len(t, model.Storages["local"].Hooks, 1)
	assert.Equal(t, `echo "storage after"`, model.Storages["local"].Hooks[0].Script)

	// models without hooks
	model = GetModelConfigByName("base_test")
	assert.Len(t, model.Hooks, 0)
	assert.Len(t, model.ArchiveHooks, 0)
	assert.Len(t, model.SplitterHooks, 0)
	assert.Len(t, model.CompressWith.Hooks, 0)
	assert.Len(t, model.EncryptWith.Hooks, 0)
	assert.Len(t, model.GetDatabaseByName("dummy_test").Hooks, 0)
	assert.Len(t, model.Storages["local"].Hooks, 0)
}

func TestHooksValidation(t *testing.T) {
	base := `models:
  myjob:
    storages:
      local:
        type: local
        keep: 2
    archive:
      includes:
        - /etc/hosts
`

	testCases := []struct {
		name          string
		hooks         string
		full          string
		expectedError string
	}{
		{
			name: "when_missing",
			hooks: `    hooks:
      - script: echo 1
`,
			expectedError: "models.myjob.hooks[#1]: when is required (before or after)",
		},
		{
			name: "when_invalid",
			hooks: `    hooks:
      - name: x
        when: during
        script: echo 1
`,
			expectedError: "models.myjob.hooks[x]: when must be before or after",
		},
		{
			name: "script_missing",
			hooks: `    hooks:
      - when: before
`,
			expectedError: "models.myjob.hooks[#1]: script is required",
		},
		{
			name: "script_blank",
			hooks: `    hooks:
      - when: before
        script: "   "
`,
			expectedError: "models.myjob.hooks[#1]: script is required",
		},
		{
			name: "on_exit_on_before",
			hooks: `    hooks:
      - when: before
        on_exit: always
        script: echo 1
`,
			expectedError: "models.myjob.hooks[#1]: on_exit is only allowed for after hooks",
		},
		{
			name: "on_exit_invalid",
			hooks: `    hooks:
      - when: after
        on_exit: never
        script: echo 1
`,
			expectedError: "models.myjob.hooks[#1]: on_exit must be one of always, success, failure",
		},
		{
			name: "continue_on_error_on_after",
			hooks: `    hooks:
      - when: after
        continue_on_error: false
        script: echo 1
`,
			expectedError: "models.myjob.hooks[#1]: continue_on_error is only allowed for before hooks",
		},
		{
			name: "timeout_negative",
			hooks: `    hooks:
      - when: before
        timeout: -1
        script: echo 1
`,
			expectedError: "models.myjob.hooks[#1]: timeout must not be negative",
		},
		{
			name: "hooks_not_a_list",
			hooks: `    hooks:
      when: before
      script: echo 1
`,
			expectedError: "models.myjob.hooks must be a list",
		},
		{
			name: "storage_hook_scope",
			full: `models:
  myjob:
    storages:
      local:
        type: local
        keep: 2
        hooks:
          - when: before
    archive:
      includes:
        - /etc/hosts
`,
			expectedError: "models.myjob.storages.local.hooks[#1]: script is required",
		},
		{
			name: "archive_hook_scope",
			full: `models:
  myjob:
    storages:
      local:
        type: local
        keep: 2
    archive:
      includes:
        - /etc/hosts
      hooks:
        - when: nope
          script: echo 1
`,
			expectedError: "models.myjob.archive.hooks[#1]: when must be before or after",
		},
		{
			name: "database_hook_scope",
			full: `models:
  myjob:
    storages:
      local:
        type: local
        keep: 2
    databases:
      pg:
        type: postgresql
        hooks:
          - when: after
            on_exit: sometimes
            script: echo 1
`,
			expectedError: "models.myjob.databases.pg.hooks[#1]: on_exit must be one of always, success, failure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "gobackup-hooks-test-*.yml")
			assert.Nil(t, err)
			t.Cleanup(func() {
				_ = os.Remove(file.Name())
			})

			content := tc.full
			if len(content) == 0 {
				content = base + tc.hooks
			}
			err = os.WriteFile(file.Name(), []byte(content), 0644)
			assert.Nil(t, err)

			err = Init(file.Name())
			assert.NotNil(t, err)
			if err != nil {
				assert.Contains(t, err.Error(), tc.expectedError)
			}

			err = Init(testConfigFile)
			assert.Nil(t, err)
		})
	}
}
