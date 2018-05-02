package storage

import (
	"encoding/json"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type PackageList []Package

type Package struct {
	FileKey   string    `json:"file_key"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	cyclerPath = path.Join(config.HomeDir, ".gobackup/cycler")
)

type Cycler struct {
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

func (c *Cycler) run(model string, fileKey string, keep int, deletePackage func(fileKey string) error) {
	cyclerFileName := path.Join(cyclerPath, model + ".json")

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
		if err != nil {
			logger.Warn("remove failed: ", err)
		}
	}
}

func (c *Cycler) load(cyclerFileName string) {
	if !helper.IsExistsPath(cyclerPath) {
		helper.Exec("mkdir", "-p", cyclerPath)
	}

	if !helper.IsExistsPath(cyclerFileName) {
		helper.Exec("touch", cyclerFileName)
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
	if !c.isLoaded {
		logger.Warn("Skip save cycler.json because it not loaded")
		return
	}

	data, err := json.Marshal(&c.packages)
	if err != nil {
		logger.Error("Marshal packages to cycler.json failed: ", err)
		return
	}

	err = ioutil.WriteFile(cyclerFileName, data, os.ModePerm)
	if err != nil {
		logger.Error("Save cycler.json failed: ", err)
		return
	}
}
