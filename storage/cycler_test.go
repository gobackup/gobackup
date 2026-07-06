package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/longbridgeapp/assert"
)

func TestCycler_add(t *testing.T) {
	cycler := Cycler{}
	cycler.add("foo", []string{})
	cycler.add("bar", []string{})

	assert.Equal(t, len(cycler.packages), 2)
}

func TestCycler_shiftByKeep(t *testing.T) {
	cycler := Cycler{
		packages: PackageList{
			Package{
				FileKey:   "p1",
				CreatedAt: time.Now(),
			},
			Package{
				FileKey:   "p2",
				CreatedAt: time.Now(),
			},
		},
	}
	cycler.add("p3", []string{})
	cycler.add("p4", []string{})
	cycler.add("p5", []string{})
	cycler.add("p6", []string{})

	pkg := cycler.shiftByKeep(2)
	assert.Equal(t, len(cycler.packages), 5)
	assert.Equal(t, pkg.FileKey, "p1")
	pkg = cycler.shiftByKeep(2)
	assert.Equal(t, len(cycler.packages), 4)
	assert.Equal(t, pkg.FileKey, "p2")
	pkg = cycler.shiftByKeep(4)
	assert.Equal(t, len(cycler.packages), 4)
	assert.Nil(t, pkg)
}

func TestCycler_save_and_load(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Override cyclerPath for testing
	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	// Create and populate a cycler
	cycler := Cycler{name: "test-model"}
	cycler.add("backup1.tar.gz", []string{})
	cycler.add("backup2.tar.gz", []string{})
	cycler.add("backup3.tar.gz", []string{"file1", "file2"})
	cycler.isLoaded = true

	// Save the cycler
	cyclerFileName := filepath.Join(tmpDir, "test-model.json")
	cycler.save(cyclerFileName)

	// Verify file was created
	assert.True(t, fileExists(cyclerFileName))

	// Load into a new cycler
	cycler2 := Cycler{name: "test-model"}
	cycler2.load(cyclerFileName)

	// Verify loaded data matches
	assert.True(t, cycler2.isLoaded)
	assert.Equal(t, 3, len(cycler2.packages))
	assert.Equal(t, "backup1.tar.gz", cycler2.packages[0].FileKey)
	assert.Equal(t, "backup2.tar.gz", cycler2.packages[1].FileKey)
	assert.Equal(t, "backup3.tar.gz", cycler2.packages[2].FileKey)
	assert.Equal(t, 2, len(cycler2.packages[2].FileKeys))
	assert.Equal(t, "file1", cycler2.packages[2].FileKeys[0])
}

func TestCycler_load_nonexistent_file(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Override cyclerPath for testing
	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	cycler := Cycler{name: "nonexistent"}
	cyclerFileName := filepath.Join(tmpDir, "nonexistent.json")

	// Load should create an empty file
	cycler.load(cyclerFileName)

	// File should be created with empty array
	assert.True(t, fileExists(cyclerFileName))
	data, err := os.ReadFile(cyclerFileName)
	assert.NoError(t, err)
	assert.Equal(t, "[]", string(data))
	assert.True(t, cycler.isLoaded)
	assert.Equal(t, 0, len(cycler.packages))
}

func TestCycler_save_without_load(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	cycler := Cycler{name: "test"}
	cycler.add("backup.tar.gz", []string{})
	// isLoaded is false, so saveRemote should skip

	cyclerFileName := filepath.Join(tmpDir, "test.json")
	// Direct save should still work
	cycler.isLoaded = true
	cycler.save(cyclerFileName)
	assert.True(t, fileExists(cyclerFileName))
}

func TestCycler_shiftByKeep_edge_cases(t *testing.T) {
	t.Run("empty packages", func(t *testing.T) {
		cycler := Cycler{packages: PackageList{}}
		pkg := cycler.shiftByKeep(5)
		assert.Nil(t, pkg)
		assert.Equal(t, 0, len(cycler.packages))
	})

	t.Run("keep zero", func(t *testing.T) {
		cycler := Cycler{
			packages: PackageList{
				Package{FileKey: "p1", CreatedAt: time.Now()},
			},
		}
		pkg := cycler.shiftByKeep(0)
		assert.NotNil(t, pkg)
		assert.Equal(t, "p1", pkg.FileKey)
	})

	t.Run("keep equals total", func(t *testing.T) {
		cycler := Cycler{
			packages: PackageList{
				Package{FileKey: "p1", CreatedAt: time.Now()},
				Package{FileKey: "p2", CreatedAt: time.Now()},
			},
		}
		pkg := cycler.shiftByKeep(2)
		assert.Nil(t, pkg)
		assert.Equal(t, 2, len(cycler.packages))
	})
}

func TestCycler_run(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	// Track deleted files
	deletedFiles := []string{}
	deletePackage := func(fileKey string) error {
		deletedFiles = append(deletedFiles, fileKey)
		return nil
	}

	// Create a mock storage (can be nil for this test since we'll mock the methods)
	cycler := Cycler{name: "test-run"}

	// Pre-populate with some packages
	cycler.packages = PackageList{
		Package{FileKey: "old1.tar.gz", CreatedAt: time.Now().Add(-48 * time.Hour)},
		Package{FileKey: "old2.tar.gz", CreatedAt: time.Now().Add(-24 * time.Hour)},
	}
	cycler.isLoaded = true

	// Run with keep=2, adding a new package should trigger deletion of old1.tar.gz
	cycler.run(nil, "new.tar.gz", []string{}, 2, deletePackage)

	// Should have deleted old1.tar.gz
	assert.Equal(t, 1, len(deletedFiles))
	assert.Equal(t, "old1.tar.gz", deletedFiles[0])

	// Should have 2 packages left (old2.tar.gz and new.tar.gz)
	assert.Equal(t, 2, len(cycler.packages))
}

func TestCycler_run_with_directory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	deletedFiles := []string{}
	deletePackage := func(fileKey string) error {
		deletedFiles = append(deletedFiles, fileKey)
		return nil
	}

	cycler := Cycler{name: "test-dir"}
	cycler.packages = PackageList{
		Package{
			FileKey:   "backup-dir",
			FileKeys:  []string{"file1.txt", "file2.txt"},
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}
	cycler.isLoaded = true

	// Adding new package with keep=1 should delete the directory and its files
	cycler.run(nil, "new.tar.gz", []string{}, 1, deletePackage)

	// Should have deleted: file1.txt, file2.txt, and backup-dir/
	assert.Equal(t, 3, len(deletedFiles))
	assert.Contains(t, deletedFiles, "file1.txt")
	assert.Contains(t, deletedFiles, "file2.txt")
	assert.Contains(t, deletedFiles, "backup-dir/")
}

func TestCycler_run_keep_zero(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	deletedFiles := []string{}
	deletePackage := func(fileKey string) error {
		deletedFiles = append(deletedFiles, fileKey)
		return nil
	}

	cycler := Cycler{name: "test-keep-zero"}
	cycler.isLoaded = true

	// Run with keep=0 should not delete anything
	cycler.run(nil, "new.tar.gz", []string{}, 0, deletePackage)

	assert.Equal(t, 0, len(deletedFiles))
	assert.Equal(t, 1, len(cycler.packages)) // Only new package added
}

func TestGetBaseFromStorage(t *testing.T) {
	t.Run("S3", func(t *testing.T) {
		s3 := &S3{Base: Base{}}
		base := getBaseFromStorage(s3)
		assert.NotNil(t, base)
	})

	t.Run("Local", func(t *testing.T) {
		local := &Local{Base: Base{}}
		base := getBaseFromStorage(local)
		assert.NotNil(t, base)
	})

	t.Run("GCS", func(t *testing.T) {
		gcs := &GCS{Base: Base{}}
		base := getBaseFromStorage(gcs)
		assert.NotNil(t, base)
	})

	t.Run("Azure", func(t *testing.T) {
		azure := &Azure{Base: Base{}}
		base := getBaseFromStorage(azure)
		assert.NotNil(t, base)
	})

	t.Run("WebDAV", func(t *testing.T) {
		webdav := &WebDAV{Base: Base{}}
		base := getBaseFromStorage(webdav)
		assert.NotNil(t, base)
	})

	t.Run("FTP", func(t *testing.T) {
		ftp := &FTP{Base: Base{}}
		base := getBaseFromStorage(ftp)
		assert.NotNil(t, base)
	})

	t.Run("SCP", func(t *testing.T) {
		scp := &SCP{Base: Base{}}
		base := getBaseFromStorage(scp)
		assert.NotNil(t, base)
	})

	t.Run("SFTP", func(t *testing.T) {
		sftp := &SFTP{Base: Base{}}
		base := getBaseFromStorage(sftp)
		assert.NotNil(t, base)
	})

	t.Run("Unknown type", func(t *testing.T) {
		// Create a mock storage that doesn't match any known type
		var unknownStorage Storage
		base := getBaseFromStorage(unknownStorage)
		assert.Nil(t, base)
	})
}

func TestPackage_JSON_marshaling(t *testing.T) {
	pkg := Package{
		FileKey:   "test.tar.gz",
		FileKeys:  []string{"file1", "file2"},
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Marshal to JSON
	data, err := json.Marshal(pkg)
	assert.NoError(t, err)

	// Unmarshal back
	var pkg2 Package
	err = json.Unmarshal(data, &pkg2)
	assert.NoError(t, err)

	assert.Equal(t, pkg.FileKey, pkg2.FileKey)
	assert.Equal(t, len(pkg.FileKeys), len(pkg2.FileKeys))
	assert.Equal(t, pkg.CreatedAt.Unix(), pkg2.CreatedAt.Unix())
}

func TestPackageList_JSON_marshaling(t *testing.T) {
	packages := PackageList{
		Package{
			FileKey:   "backup1.tar.gz",
			FileKeys:  []string{},
			CreatedAt: time.Now(),
		},
		Package{
			FileKey:   "backup2",
			FileKeys:  []string{"file1", "file2"},
			CreatedAt: time.Now(),
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(packages)
	assert.NoError(t, err)

	// Unmarshal back
	var packages2 PackageList
	err = json.Unmarshal(data, &packages2)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(packages2))
	assert.Equal(t, packages[0].FileKey, packages2[0].FileKey)
	assert.Equal(t, packages[1].FileKey, packages2[1].FileKey)
	assert.Equal(t, 2, len(packages2[1].FileKeys))
}

func TestCycler_add_with_FileKeys(t *testing.T) {
	cycler := Cycler{}
	fileKeys := []string{"file1.txt", "file2.txt", "file3.txt"}

	cycler.add("backup-dir", fileKeys)

	assert.Equal(t, 1, len(cycler.packages))
	assert.Equal(t, "backup-dir", cycler.packages[0].FileKey)
	assert.Equal(t, 3, len(cycler.packages[0].FileKeys))
	assert.Equal(t, fileKeys, cycler.packages[0].FileKeys)
	assert.True(t, time.Since(cycler.packages[0].CreatedAt) < time.Second)
}

func TestCycler_loadRemote_nil_storage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	cycler := Cycler{name: "test-nil-storage"}
	cyclerFileName := filepath.Join(tmpDir, "test.json")
	remoteStateKey := ".gobackup-state/test.json"

	// Should not panic with nil storage
	cycler.loadRemote(nil, cyclerFileName, remoteStateKey)

	// When storage is nil, loadRemote returns early without loading local state
	assert.False(t, cycler.isLoaded)
}

func TestCycler_saveRemote_nil_storage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cycler-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalCyclerPath := cyclerPath
	cyclerPath = tmpDir
	defer func() { cyclerPath = originalCyclerPath }()

	cycler := Cycler{name: "test-save-nil"}
	cycler.add("backup.tar.gz", []string{})
	cycler.isLoaded = true

	cyclerFileName := filepath.Join(tmpDir, "test.json")
	remoteStateKey := ".gobackup-state/test.json"

	// When storage is nil, saveRemote returns early without saving locally
	cycler.saveRemote(nil, cyclerFileName, remoteStateKey)

	// Verify local file was NOT created because storage is nil
	assert.False(t, fileExists(cyclerFileName))
}

func TestCycler_multiple_shifts(t *testing.T) {
	cycler := Cycler{}

	// Add 10 packages
	for i := 0; i < 10; i++ {
		cycler.add(filepath.Join("backup", string(rune('a'+i))+".tar.gz"), []string{})
	}

	assert.Equal(t, 10, len(cycler.packages))

	// Keep only 3, should remove 7
	removed := 0
	for {
		pkg := cycler.shiftByKeep(3)
		if pkg == nil {
			break
		}
		removed++
	}

	assert.Equal(t, 7, removed)
	assert.Equal(t, 3, len(cycler.packages))
}

// Helper function to check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
