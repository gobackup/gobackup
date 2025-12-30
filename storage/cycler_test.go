package storage

import (
	"encoding/json"
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

// mockStorage is a mock storage implementation for testing remote state
type mockStorage struct {
	state map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		state: make(map[string][]byte),
	}
}

func (m *mockStorage) open() error                           { return nil }
func (m *mockStorage) close()                                {}
func (m *mockStorage) upload(fileKey string) error           { return nil }
func (m *mockStorage) delete(fileKey string) error           { return nil }
func (m *mockStorage) list(parent string) ([]FileItem, error) { return nil, nil }
func (m *mockStorage) download(fileKey string) (string, error) { return "", nil }


func TestCycler_loadWithRemote_fromRemote(t *testing.T) {
	// Setup mock storage with existing state
	mockStore := newMockStorage()
	existingPackages := PackageList{
		{FileKey: "remote_backup_1", CreatedAt: time.Now().Add(-24 * time.Hour)},
		{FileKey: "remote_backup_2", CreatedAt: time.Now().Add(-12 * time.Hour)},
	}
	stateData, _ := json.Marshal(existingPackages)
	mockStore.state[".gobackup-state/test_model.json"] = stateData

	cycler := Cycler{name: "test_model"}
	cycler.loadWithRemote("/tmp/test_cycler.json", ".gobackup-state/test_model.json", mockStore)

	assert.True(t, cycler.isLoaded)
	assert.Equal(t, len(cycler.packages), 2)
	assert.Equal(t, cycler.packages[0].FileKey, "remote_backup_1")
	assert.Equal(t, cycler.packages[1].FileKey, "remote_backup_2")
}

func TestCycler_saveWithRemote(t *testing.T) {
	mockStore := newMockStorage()
	
	cycler := Cycler{
		name:     "test_model",
		isLoaded: true,
		packages: PackageList{
			{FileKey: "backup_1", CreatedAt: time.Now()},
			{FileKey: "backup_2", CreatedAt: time.Now()},
		},
	}

	cycler.saveWithRemote("/tmp/test_cycler.json", ".gobackup-state/test_model.json", mockStore)

	// Verify remote state was saved
	savedData, ok := mockStore.state[".gobackup-state/test_model.json"]
	assert.True(t, ok)
	
	var savedPackages PackageList
	err := json.Unmarshal(savedData, &savedPackages)
	assert.Nil(t, err)
	assert.Equal(t, len(savedPackages), 2)
	assert.Equal(t, savedPackages[0].FileKey, "backup_1")
}

func TestCycler_remoteStatePath(t *testing.T) {
	// Verify the remote state path constant
	assert.Equal(t, remoteStatePath, ".gobackup-state")
}
