package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

type PackageList []Package

// When `FileKeys` is not empty, `FileKey` is the directory
type Package struct {
	FileKey   string    `json:"file_key"`
	FileKeys  []string  `json:"file_keys,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	cyclerPath = filepath.Join(config.GoBackupDir, "cycler")
	// Remote state path prefix used for storing cycler state on remote storage
	remoteStatePath = ".gobackup-state"
)

type Cycler struct {
	name     string
	packages PackageList
	isLoaded bool
}

func (c *Cycler) add(fileKey string, fileKeys []string) {
	c.packages = append(c.packages, Package{
		FileKey:   fileKey,
		FileKeys:  fileKeys,
		CreatedAt: time.Now(),
	})
}

func (c *Cycler) shiftByKeep(keep int) (first *Package) {
	total := len(c.packages)
	if total <= keep {
		return nil
	}

	first, c.packages = &c.packages[0], c.packages[1:]
	return
}

func (c *Cycler) run(fileKey string, fileKeys []string, keep int, deletePackage func(fileKey string) error, storage Storage) {
	logger := logger.Tag("Cycler")

	cyclerFileName := filepath.Join(cyclerPath, c.name+".json")
	remoteStateKey := filepath.Join(remoteStatePath, c.name+".json")

	c.loadWithRemote(cyclerFileName, remoteStateKey, storage)
	c.add(fileKey, fileKeys)
	defer c.saveWithRemote(cyclerFileName, remoteStateKey, storage)

	if keep == 0 {
		return
	}

	for {
		pkg := c.shiftByKeep(keep)
		if pkg == nil {
			break
		}

		fk := pkg.FileKey
		if len(pkg.FileKeys) != 0 && !strings.HasSuffix(fk, "/") {
			fk += "/"
		}
		for _, k := range append(pkg.FileKeys, fk) {
			// deletePackage() should handle directory case which has `/` suffix
			err := deletePackage(k)
			if err != nil {
				logger.Warnf("Remove %s failed: %v", k, err)
			} else {
				logger.Info("Removed", k)
			}
		}
	}
}

// loadWithRemote tries to load cycler state from remote storage first,
// falls back to local state if remote is unavailable.
// This ensures retention policy works correctly in containerized environments
// where local filesystem is ephemeral.
func (c *Cycler) loadWithRemote(cyclerFileName string, remoteStateKey string, storage Storage) {
	logger := logger.Tag("Cycler")

	// Load from remote storage
	if storage != nil {
		remoteData, err := storage.downloadState(remoteStateKey)
		if err == nil && len(remoteData) > 0 {
			if err := json.Unmarshal(remoteData, &c.packages); err == nil {
				logger.Info("Loaded cycler state from remote storage")
				c.isLoaded = true
				// Also save to local for faster access next time
				c.save(cyclerFileName)
				return
			}
			logger.Warnf("Failed to unmarshal remote cycler state: %v", err)
		} else if err != nil {
			logger.Infof("Remote cycler state not found or unavailable: %v, falling back to local", err)
		}
	}

	// Fall back to local state
	c.load(cyclerFileName)
}

func (c *Cycler) load(cyclerFileName string) {
	logger := logger.Tag("Cycler")

	if err := helper.MkdirP(cyclerPath); err != nil {
		logger.Errorf("Failed to mkdir cycler path %s: %v", cyclerPath, err)
		return
	}

	// write example JSON if not exist
	if !helper.IsExistsPath(cyclerFileName) {
		if err := os.WriteFile(cyclerFileName, []byte("[]"), 0660); err != nil {
			logger.Errorf("Failed to write file %s: %v", cyclerFileName, err)
			return
		}
	}

	f, err := os.ReadFile(cyclerFileName)
	if err != nil {
		logger.Error("Load cycler.json failed:", err)
		return
	}
	err = json.Unmarshal(f, &c.packages)
	if err != nil {
		logger.Error("Unmarshal cycler.json failed:", err)
	}
	c.isLoaded = true
}

// saveWithRemote saves cycler state to both remote storage and local filesystem.
// Remote storage ensures persistence across container restarts.
// Local storage provides faster access and serves as a fallback.
func (c *Cycler) saveWithRemote(cyclerFileName string, remoteStateKey string, storage Storage) {
	logger := logger.Tag("Cycler")

	if !c.isLoaded {
		logger.Warn("Skip save cycler.json because it is not loaded")
		return
	}

	data, err := json.Marshal(&c.packages)
	if err != nil {
		logger.Error("Marshal packages to cycler.json failed: ", err)
		return
	}

	// Save to remote storage first
	if storage != nil {
		if err := storage.uploadState(remoteStateKey, data); err != nil {
			logger.Warnf("Failed to save cycler state to remote storage: %v", err)
		} else {
			logger.Info("Saved cycler state to remote storage")
		}
	}

	// Always save locally as well
	c.save(cyclerFileName)
}

func (c *Cycler) save(cyclerFileName string) {
	logger := logger.Tag("Cycler")

	if err := helper.MkdirP(cyclerPath); err != nil {
		logger.Errorf("Failed to mkdir cycler path %s: %v", cyclerPath, err)
		return
	}

	data, err := json.Marshal(&c.packages)
	if err != nil {
		logger.Error("Marshal packages to cycler.json failed: ", err)
		return
	}

	err = os.WriteFile(cyclerFileName, data, 0660)
	if err != nil {
		logger.Error("Save cycler.json failed: ", err)
		return
	}
}
