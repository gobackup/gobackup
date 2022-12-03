package storage

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

type PackageList []Package

type Package struct {
	FileKey   string    `json:"file_key"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	cyclerPath = filepath.Join(config.GobackupDir, "cycler")
)

type Cycler struct {
	name     string
	packages PackageList
	isLoaded bool
}

func (c *Cycler) add(fileKey string) {
	c.packages = append(c.packages, Package{
		FileKey:   fileKey,
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

func (c *Cycler) run(fileKey string, keep int, deletePackage func(fileKey string) error) {
	logger := logger.Tag("Cycler")

	cyclerFileName := filepath.Join(cyclerPath, c.name+".json")

	c.load(cyclerFileName)
	c.add(fileKey)
	defer c.save(cyclerFileName)

	if keep == 0 {
		return
	}

	for {
		pkg := c.shiftByKeep(keep)
		if pkg == nil {
			break
		}

		err := deletePackage(pkg.FileKey)
		logger.Info("Removed", pkg.FileKey)
		if err != nil {
			logger.Warn("remove failed: ", err)
		}
	}
}

func (c *Cycler) load(cyclerFileName string) {
	logger := logger.Tag("Cycler")

	if err := helper.MkdirP(cyclerPath); err != nil {
		logger.Errorf("Failed to mkdir cycler path %s: %v", cyclerPath, err)
		return
	}

	// write example JSON if not exist
	if !helper.IsExistsPath(cyclerFileName) {
		ioutil.WriteFile(cyclerFileName, []byte("[]"), 0660)
	}

	f, err := ioutil.ReadFile(cyclerFileName)
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

func (c *Cycler) save(cyclerFileName string) {
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

	err = ioutil.WriteFile(cyclerFileName, data, 0660)
	if err != nil {
		logger.Error("Save cycler.json failed: ", err)
		return
	}
}
