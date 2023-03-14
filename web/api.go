package web

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
	"github.com/gobackup/gobackup/storage"
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

	if os.Getenv("GO_ENV") == "dev" {
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
	group.GET("/list", list)
	group.GET("/download", download)
	group.POST("/perform", perform)
	group.GET("/log", log)
	return r
}

// GET /api/config
func getConfig(c *gin.Context) {
	models := map[string]any{}
	for _, m := range model.GetModels() {
		models[m.Config.Name] = gin.H{
			"description":   m.Config.Description,
			"schedule":      m.Config.Schedule,
			"schedule_info": m.Config.Schedule.String(),
		}
	}

	c.JSON(200, gin.H{
		"models": models,
	})
}

// POST /api/perform
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

// GET /api/list?model=xxx&parent=
func list(c *gin.Context) {
	modelName := c.Query("model")
	m := model.GetModelByName(modelName)
	if m == nil {
		c.AbortWithStatusJSON(404, gin.H{"message": fmt.Sprintf("Model: \"%s\" not found", modelName)})
		return
	}

	parent := c.Query("parent")
	if parent == "" {
		parent = "/"
	}

	files, err := storage.List(m.Config, parent)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"files": files})
}

// GET /api/download?model=xxx&path=
func download(c *gin.Context) {
	modelName := c.Query("model")
	m := model.GetModelByName(modelName)
	if m == nil {
		c.AbortWithStatusJSON(404, gin.H{"message": fmt.Sprintf("Model: \"%s\" not found", modelName)})
		return
	}

	file := c.Query("path")
	if file == "" {
		c.AbortWithStatusJSON(404, gin.H{"message": "File not found"})
		return
	}

	downloadURL, err := storage.Download(m.Config, file)
	if err != nil || len(downloadURL) == 0 {
		c.AbortWithStatusJSON(500, gin.H{"message": err.Error()})
		return
	}

	c.Redirect(302, downloadURL)
}

// GET /api/log
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
