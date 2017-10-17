package storage

import (
	"encoding/json"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"time"
)

var (
	packages       PackageList
	cyclerFileName = path.Join(config.HomeDir, ".gobackup", "cycler.json")
	isLoaded       = false
)

type PackageList []Package

type Package struct {
	FileKey   string    `json:"file_key"`
	CreatedAt time.Time `json:"created_at"`
}

func (s PackageList) Len() int {
	return len(s)
}
func (s PackageList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s PackageList) Less(i, j int) bool {
	return s[i].CreatedAt.Unix() < s[j].CreatedAt.Unix()
}

func addPackage(fileKey string) {
	packages = append(packages, Package{
		FileKey:   fileKey,
		CreatedAt: time.Now(),
	})
}

func runCycler(fileKey string) {
	loadCycler()

	addPackage(fileKey)

	logger.Info("Cycler run...")
	sort.Sort(&packages)

	dumpCycler()
}

func loadCycler() {
	if !helper.IsExistsPath(cyclerFileName) {
		helper.Exec("touch", cyclerFileName)
	}

	f, err := ioutil.ReadFile(cyclerFileName)
	if err != nil {
		logger.Error("Load cycler.json failed:", err)
		return
	}
	err = json.Unmarshal(f, &packages)
	if err != nil {
		logger.Error("Unmarshal cycler.json failed:", err)
	}
	isLoaded = true
}

func dumpCycler() {
	if !isLoaded {
		logger.Warn("Skip save cycler.json because it not loaded")
		return
	}

	logger.Info("Current packages: ", len(packages))

	data, err := json.Marshal(&packages)
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
