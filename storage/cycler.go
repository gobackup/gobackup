package storage

import (
	"encoding/json"
	"io"
	"net/http"
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
		// Use download method to get a presigned URL
		url, err := storage.download(remoteStateKey)
		if err == nil && url != "" {
			// Fetch the data from the URL
			resp, err := http.Get(url)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					remoteData, err := io.ReadAll(resp.Body)
					if err == nil && len(remoteData) > 0 {
						if err := json.Unmarshal(remoteData, &c.packages); err == nil {
							logger.Info("Loaded cycler state from remote storage")
							c.isLoaded = true
							// Also save to local for faster access next time
							c.save(cyclerFileName)
							return
						}
						logger.Warnf("Failed to unmarshal remote cycler state: %v", err)
					}
				} else {
					logger.Infof("Remote cycler state not found (HTTP %d), falling back to local", resp.StatusCode)
				}
			} else {
				logger.Infof("Failed to fetch remote cycler state: %v, falling back to local", err)
			}
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

	// Save to remote storage first using upload method
	if storage != nil {
		// Create a temporary directory for the state file
		tmpDir, err := os.MkdirTemp(cyclerPath, "cycler-state-*")
		if err != nil {
			logger.Warnf("Failed to create temp directory for state: %v", err)
		} else {
			defer os.RemoveAll(tmpDir) // Clean up temp directory

			// upload method expects file at filepath.Join(filepath.Dir(s.archivePath), fileKey)
			// and uploads to filepath.Join(s.path, fileKey)
			// So we need to create the file at the path that matches remoteStateKey structure
			tmpFilePath := filepath.Join(tmpDir, remoteStateKey)

			// Create parent directories if needed
			if err := os.MkdirAll(filepath.Dir(tmpFilePath), 0755); err != nil {
				logger.Warnf("Failed to create temp directory structure: %v", err)
			} else {
				// Write state data to temp file
				if err := os.WriteFile(tmpFilePath, data, 0660); err != nil {
					logger.Warnf("Failed to write state to temp file: %v", err)
				} else {
					// Get the storage's Base to temporarily modify archivePath
					// upload method uses filepath.Dir(s.archivePath) as the base directory
					base := getBaseFromStorage(storage)
					if base != nil {
						originalArchivePath := base.archivePath
						// Set archivePath to a file in tmpDir so filepath.Dir(archivePath) = tmpDir
						// This way filepath.Join(filepath.Dir(archivePath), remoteStateKey) = tmpFilePath
						base.archivePath = filepath.Join(tmpDir, "dummy")

						// Upload using remoteStateKey as the fileKey
						// This will read from tmpDir/remoteStateKey and upload to s.path/remoteStateKey
						if err := storage.upload(remoteStateKey); err != nil {
							logger.Warnf("Failed to save cycler state to remote storage: %v", err)
						} else {
							logger.Info("Saved cycler state to remote storage")
						}

						// Restore original archivePath
						base.archivePath = originalArchivePath
					} else {
						logger.Warn("Failed to get Base from storage for state upload")
					}
				}
			}
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

// getBaseFromStorage extracts the Base struct from a Storage implementation
// This is a helper to access the Base struct which is embedded in storage implementations
func getBaseFromStorage(s Storage) *Base {
	switch v := s.(type) {
	case *S3:
		return &v.Base
	case *GCS:
		return &v.Base
	case *Azure:
		return &v.Base
	case *Local:
		return &v.Base
	case *WebDAV:
		return &v.Base
	case *FTP:
		return &v.Base
	case *SCP:
		return &v.Base
	case *SFTP:
		return &v.Base
	default:
		return nil
	}
}
