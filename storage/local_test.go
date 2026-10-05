package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestLocal_open(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-open-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storagePath := filepath.Join(tmpDir, "backup_storage")
	v := viper.New()
	v.Set("path", storagePath)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{
				Name:    "my_backup",
				WorkDir: tmpDir,
			},
			viper: v,
		},
	}

	err = local.open()
	assert.NoError(t, err)
	assert.True(t, fileExists(storagePath))
}

func TestLocal_open_relative_path(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-rel-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	v := viper.New()
	v.Set("path", "relative_backups")

	local := &Local{
		Base: Base{
			model: config.ModelConfig{
				Name:    "my_backup",
				WorkDir: tmpDir,
			},
			viper: v,
		},
	}

	err = local.open()
	assert.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "relative_backups")
	assert.Equal(t, expectedPath, local.path)
	assert.True(t, fileExists(expectedPath))
}

func TestLocal_upload_file(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-upload-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storagePath := filepath.Join(tmpDir, "dest")
	sourceFile := filepath.Join(tmpDir, "source.tar.gz")
	err = os.WriteFile(sourceFile, []byte("archive-content"), 0644)
	assert.NoError(t, err)

	v := viper.New()
	v.Set("path", storagePath)

	local := &Local{
		Base: Base{
			model:       config.ModelConfig{Name: "my_backup", WorkDir: tmpDir},
			archivePath: sourceFile,
			viper:       v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	err = local.upload("source.tar.gz")
	assert.NoError(t, err)

	targetFile := filepath.Join(storagePath, "source.tar.gz")
	assert.True(t, fileExists(targetFile))
	content, err := os.ReadFile(targetFile)
	assert.NoError(t, err)
	assert.Equal(t, "archive-content", string(content))
}

func TestLocal_upload_chunks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-chunks-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storagePath := filepath.Join(tmpDir, "dest")
	chunkDir := filepath.Join(tmpDir, "split_backup")
	err = os.MkdirAll(chunkDir, 0755)
	assert.NoError(t, err)

	chunk1 := filepath.Join(chunkDir, "chunk.001")
	chunk2 := filepath.Join(chunkDir, "chunk.002")
	err = os.WriteFile(chunk1, []byte("chunk-one"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(chunk2, []byte("chunk-two"), 0644)
	assert.NoError(t, err)

	v := viper.New()
	v.Set("path", storagePath)

	local := &Local{
		Base: Base{
			model:       config.ModelConfig{Name: "my_backup", WorkDir: tmpDir},
			archivePath: chunkDir,
			fileKeys: []string{
				filepath.Join("split_backup", "chunk.001"),
				filepath.Join("split_backup", "chunk.002"),
			},
			viper: v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	err = local.upload("split_backup")
	assert.NoError(t, err)

	targetChunk1 := filepath.Join(storagePath, "split_backup", "chunk.001")
	targetChunk2 := filepath.Join(storagePath, "split_backup", "chunk.002")
	assert.True(t, fileExists(targetChunk1))
	assert.True(t, fileExists(targetChunk2))
}

func TestLocal_upload_cycler_state_with_dummy_archive_path(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-state-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storagePath := filepath.Join(tmpDir, "storage")
	v := viper.New()
	v.Set("path", storagePath)

	stateDir := filepath.Join(tmpDir, ".gobackup-state")
	err = os.MkdirAll(stateDir, 0755)
	assert.NoError(t, err)

	stateFile := filepath.Join(stateDir, "model.json")
	err = os.WriteFile(stateFile, []byte(`[{"file_key":"archive.tar.gz"}]`), 0644)
	assert.NoError(t, err)

	local := &Local{
		Base: Base{
			model:       config.ModelConfig{Name: "model", WorkDir: tmpDir},
			archivePath: filepath.Join(tmpDir, "dummy"),
			viper:       v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	err = local.upload(".gobackup-state/model.json")
	assert.NoError(t, err)

	persistedState := filepath.Join(storagePath, ".gobackup-state", "model.json")
	assert.True(t, fileExists(persistedState))
	content, err := os.ReadFile(persistedState)
	assert.NoError(t, err)
	assert.Equal(t, `[{"file_key":"archive.tar.gz"}]`, string(content))
}

func TestLocal_delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-delete-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.tar.gz")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	assert.NoError(t, err)

	dirPath := filepath.Join(tmpDir, "test_dir")
	err = os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dirPath, "sub.txt"), []byte("sub"), 0644)
	assert.NoError(t, err)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{Name: "model", WorkDir: tmpDir},
		},
		path: tmpDir,
	}

	err = local.delete("test.tar.gz")
	assert.NoError(t, err)
	assert.False(t, fileExists(filePath))

	err = local.delete("test_dir")
	assert.NoError(t, err)
	assert.False(t, fileExists(dirPath))
}

func TestLocal_list(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "local-test-list-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("22"), 0644)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tmpDir, "subfolder"), 0755)
	assert.NoError(t, err)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{Name: "model", WorkDir: tmpDir},
		},
		path: tmpDir,
	}

	items, err := local.list("")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(items))
}

func TestCycler_loadRemote_Local_found(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-local-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = filepath.Join(tmpDir, "ephemeral_cycler")
	defer func() { cyclerPath = originalCyclerPath }()

	localStoragePath := filepath.Join(tmpDir, "persistent_storage")
	v := viper.New()
	v.Set("path", localStoragePath)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{Name: "demo", WorkDir: tmpDir},
			viper: v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	// Pre-create remote state file in local storage
	remoteStateDir := filepath.Join(localStoragePath, ".gobackup-state")
	err = os.MkdirAll(remoteStateDir, 0755)
	assert.NoError(t, err)

	expectedPackages := PackageList{
		Package{FileKey: "backup-001.tar.gz", CreatedAt: time.Now().Add(-1 * time.Hour)},
		Package{FileKey: "backup-002.tar.gz", CreatedAt: time.Now()},
	}
	stateJSON, err := json.Marshal(expectedPackages)
	assert.NoError(t, err)

	stateFilePath := filepath.Join(remoteStateDir, "demo.json")
	err = os.WriteFile(stateFilePath, stateJSON, 0644)
	assert.NoError(t, err)

	cycler := Cycler{name: "demo"}
	cyclerFileName := filepath.Join(cyclerPath, "demo.json")
	remoteStateKey := filepath.Join(".gobackup-state", "demo.json")

	cycler.loadRemote(local, cyclerFileName, remoteStateKey)

	assert.True(t, cycler.isLoaded)
	assert.Equal(t, 2, len(cycler.packages))
	assert.Equal(t, "backup-001.tar.gz", cycler.packages[0].FileKey)
	assert.Equal(t, "backup-002.tar.gz", cycler.packages[1].FileKey)
	// Local cache should also have been updated
	assert.True(t, fileExists(cyclerFileName))
}

func TestCycler_loadRemote_Local_not_found(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-local-nf-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = filepath.Join(tmpDir, "ephemeral_cycler")
	defer func() { cyclerPath = originalCyclerPath }()

	localStoragePath := filepath.Join(tmpDir, "persistent_storage")
	v := viper.New()
	v.Set("path", localStoragePath)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{Name: "demo", WorkDir: tmpDir},
			viper: v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	cycler := Cycler{name: "demo"}
	cyclerFileName := filepath.Join(cyclerPath, "demo.json")
	remoteStateKey := filepath.Join(".gobackup-state", "demo.json")

	cycler.loadRemote(local, cyclerFileName, remoteStateKey)

	assert.True(t, cycler.isLoaded)
	assert.Equal(t, 0, len(cycler.packages))
	assert.True(t, fileExists(cyclerFileName))
}

func TestCycler_loadRemote_Local_corrupted(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-local-corr-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = filepath.Join(tmpDir, "ephemeral_cycler")
	defer func() { cyclerPath = originalCyclerPath }()

	localStoragePath := filepath.Join(tmpDir, "persistent_storage")
	v := viper.New()
	v.Set("path", localStoragePath)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{Name: "demo", WorkDir: tmpDir},
			viper: v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	remoteStateDir := filepath.Join(localStoragePath, ".gobackup-state")
	err = os.MkdirAll(remoteStateDir, 0755)
	assert.NoError(t, err)

	stateFilePath := filepath.Join(remoteStateDir, "demo.json")
	err = os.WriteFile(stateFilePath, []byte("{invalid-json}"), 0644)
	assert.NoError(t, err)

	cycler := Cycler{name: "demo"}
	cyclerFileName := filepath.Join(cyclerPath, "demo.json")
	remoteStateKey := filepath.Join(".gobackup-state", "demo.json")

	cycler.loadRemote(local, cyclerFileName, remoteStateKey)

	// Should fall back to local cleanly
	assert.True(t, cycler.isLoaded)
	assert.Equal(t, 0, len(cycler.packages))
}

func TestCycler_saveRemote_Local(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-local-save-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = filepath.Join(tmpDir, "ephemeral_cycler")
	defer func() { cyclerPath = originalCyclerPath }()

	localStoragePath := filepath.Join(tmpDir, "persistent_storage")
	v := viper.New()
	v.Set("path", localStoragePath)

	local := &Local{
		Base: Base{
			model: config.ModelConfig{Name: "demo", WorkDir: tmpDir},
			viper: v,
		},
	}
	err = local.open()
	assert.NoError(t, err)

	cycler := Cycler{name: "demo"}
	cycler.add("backup1.tar.gz", nil)
	cycler.isLoaded = true

	cyclerFileName := filepath.Join(cyclerPath, "demo.json")
	remoteStateKey := filepath.Join(".gobackup-state", "demo.json")

	cycler.saveRemote(local, cyclerFileName, remoteStateKey)

	persistedState := filepath.Join(localStoragePath, ".gobackup-state", "demo.json")
	assert.True(t, fileExists(persistedState))

	content, err := os.ReadFile(persistedState)
	assert.NoError(t, err)

	var packages PackageList
	err = json.Unmarshal(content, &packages)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(packages))
	assert.Equal(t, "backup1.tar.gz", packages[0].FileKey)
}

func TestCycler_Local_retention_across_container_restart(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-restart-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	ephemeralCyclerDir := filepath.Join(tmpDir, "container_root", ".gobackup", "cycler")
	cyclerPath = ephemeralCyclerDir
	defer func() { cyclerPath = originalCyclerPath }()

	persistentStorageDir := filepath.Join(tmpDir, "mounted_storage")
	v := viper.New()
	v.Set("path", persistentStorageDir)

	deletedFiles := []string{}
	deletePackage := func(fileKey string) error {
		deletedFiles = append(deletedFiles, fileKey)
		return nil
	}

	createLocal := func() *Local {
		l := &Local{
			Base: Base{
				model: config.ModelConfig{Name: "mysql_daily", WorkDir: tmpDir},
				viper: v,
			},
		}
		_ = l.open()
		return l
	}

	// Run 1: first backup
	local1 := createLocal()
	cycler1 := Cycler{name: "mysql_daily"}
	cycler1.run(local1, "backup_1.tar.gz", nil, 2, deletePackage)
	assert.Equal(t, 0, len(deletedFiles))
	assert.Equal(t, 1, len(cycler1.packages))

	// Run 2: second backup
	local2 := createLocal()
	cycler2 := Cycler{name: "mysql_daily"}
	cycler2.run(local2, "backup_2.tar.gz", nil, 2, deletePackage)
	assert.Equal(t, 0, len(deletedFiles))
	assert.Equal(t, 2, len(cycler2.packages))

	// Simulate container restart by wiping ephemeral cycler dir (~/.gobackup/cycler)
	err = os.RemoveAll(ephemeralCyclerDir)
	assert.NoError(t, err)
	assert.False(t, fileExists(ephemeralCyclerDir))

	// Run 3: third backup on a fresh container.
	// Retention keep=2 should prune backup_1.tar.gz based on persistent state in Local storage.
	local3 := createLocal()
	cycler3 := Cycler{name: "mysql_daily"}
	cycler3.run(local3, "backup_3.tar.gz", nil, 2, deletePackage)

	assert.Equal(t, 1, len(deletedFiles))
	assert.Equal(t, "backup_1.tar.gz", deletedFiles[0])
	assert.Equal(t, 2, len(cycler3.packages))
	assert.Equal(t, "backup_2.tar.gz", cycler3.packages[0].FileKey)
	assert.Equal(t, "backup_3.tar.gz", cycler3.packages[1].FileKey)

	// Verify the persistent state file on local storage is updated
	stateFile := filepath.Join(persistentStorageDir, ".gobackup-state", "mysql_daily.json")
	assert.True(t, fileExists(stateFile))
	stateData, err := os.ReadFile(stateFile)
	assert.NoError(t, err)

	var persistedPackages PackageList
	err = json.Unmarshal(stateData, &persistedPackages)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(persistedPackages))
	assert.Equal(t, "backup_2.tar.gz", persistedPackages[0].FileKey)
	assert.Equal(t, "backup_3.tar.gz", persistedPackages[1].FileKey)
}
