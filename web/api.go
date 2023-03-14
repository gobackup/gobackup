package web

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
	"github.com/stoicperlman/fls"
)

//go:embed dist
var staticFS embed.FS
var logFile *os.File

type embedFileSystem struct {
	http.FileSystem
	indexes bool
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	f, err := e.Open(path)
	if err != nil {
		return false
	}

	// check if indexing is allowed
	s, _ := f.Stat()
	if s.IsDir() && !e.indexes {
		return false
	}

	return true
}

// StartHTTP run API server
func StartHTTP(version string) (err error) {
	logger := logger.Tag("API")

	logFile, err = os.Open(config.LogFilePath)
	if err != nil {
		return err
	}

	fmt.Printf("\nStarting API server on port http://127.0.0.1:%s\n", config.Web.Port)

	r := setupRouter(version)

	// Enable baseAuth
	if len(config.Web.Username) > 0 && len(config.Web.Password) > 0 {
		r.Use(gin.BasicAuth(gin.Accounts{
			config.Web.Username: config.Web.Password,
		}))
	}

	fe, _ := fs.Sub(staticFS, "dist")
	r.Use(static.Serve("/", embedFileSystem{http.FS(fe), true}))

	if gin.Mode() != gin.ReleaseMode {
		go func() {
			for {
				time.Sleep(5 * time.Second)
				logger.Info("Ping", time.Now())
			}
		}()
	}

	return r.Run(":" + config.Web.Port)
}

func setupRouter(version string) *gin.Engine {
	r := gin.Default()

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GoBackup is running.",
			"version": version,
		})
	})

	group := r.Group("/api")

	group.GET("/config", getConfig)
	group.POST("/perform", perform)
	group.GET("/log", log)
	return r
}

func getConfig(c *gin.Context) {
	models := []string{}
	for _, m := range model.GetModels() {
		models = append(models, m.Config.Name)
	}
	sort.Strings(models)

	c.JSON(200, gin.H{
		"models": models,
	})
}

func perform(c *gin.Context) {
	type performParam struct {
		Model string `form:"model" json:"model" binding:"required"`
	}

	var param performParam
	c.Bind(&param)

	m := model.GetModelByName(param.Model)
	if m == nil {
		c.AbortWithStatusJSON(404, gin.H{"message": fmt.Sprintf("Model: \"%s\" not found", param.Model)})
		return
	}

	go m.Perform()
	c.JSON(200, gin.H{"message": fmt.Sprintf("Backup: %s performed in background.", param.Model)})
}

func log(c *gin.Context) {
	chanStream := tailFile()

	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-chanStream; ok {
			time.Sleep(5 * time.Millisecond)
			c.Writer.WriteString(msg + "\n")
			c.Writer.Flush()

			return true
		}
		return false
	})
}

// tailFile tail the log file and make a chain to stream output log
func tailFile() chan string {
	out_chan := make(chan string)

	file := fls.LineFile(logFile)
	file.SeekLine(-200, io.SeekEnd)
	bf := bufio.NewReader(file)

	go func() {
		for {
			line, _, _ := bf.ReadLine()

			if len(line) == 0 {
				time.Sleep(100 * time.Millisecond)
			} else {
				out_chan <- string(line)
			}
		}
	}()

	return out_chan
}
