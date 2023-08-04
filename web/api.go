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

	if len(config.Web.Password) == 0 {
		logger.Warn("You are running with insecure API server. Please don't forget setup `web.password` in config file for more safety.")
	}

	logFile, err = os.Open(config.LogFilePath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	logger.Infof("Starting API server on port http://%s:%s", config.Web.Host, config.Web.Port)

	if os.Getenv("GO_ENV") == "dev" {
		go func() {
			for {
				time.Sleep(5 * time.Second)
				logger.Info("Ping", time.Now())
			}
		}()
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := setupRouter(version)

	// Enable baseAuth
	if len(config.Web.Username) > 0 && len(config.Web.Password) > 0 {
		r.Use(gin.BasicAuth(gin.Accounts{
			config.Web.Username: config.Web.Password,
		}))
	}

	fe, _ := fs.Sub(staticFS, "dist")
	embedFs := embedFileSystem{http.FS(fe), true}
	r.Use(static.Serve("/", embedFs))
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", embedFs)
	})

	return r.Run(config.Web.Host + ":" + config.Web.Port)
}

func setupRouter(version string) *gin.Engine {
	r := gin.Default()

	r.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GoBackup is running.",
			"version": version,
		})
	})

	r.Use(func(c *gin.Context) {
		c.Next()

		// Skip if no errors
		if len(c.Errors) == 0 {
			return
		}

		c.AbortWithStatusJSON(c.Writer.Status(), gin.H{
			"message": c.Errors.String(),
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
	if err := c.Bind(&param); err != nil {
		logger.Errorf("Bind error: %v", err)
	}

	m := model.GetModelByName(param.Model)
	if m == nil {
		c.AbortWithError(404, fmt.Errorf("Model: \"%s\" not found", param.Model))
		return
	}

	go func() {
		if err := m.Perform(); err != nil {
			logger.Errorf("Perform error: %v", err)
		}
	}()
	c.JSON(200, gin.H{"message": fmt.Sprintf("Backup: %s performed in background.", param.Model)})
}

// GET /api/list?model=xxx&parent=
func list(c *gin.Context) {
	modelName := c.Query("model")
	m := model.GetModelByName(modelName)
	if m == nil {
		c.AbortWithError(404, fmt.Errorf("Model: \"%s\" not found", modelName))
		return
	}

	parent := c.Query("parent")
	if parent == "" {
		parent = "/"
	}

	files, err := storage.List(m.Config, parent)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{"files": files})
}

// GET /api/download?model=xxx&path=
func download(c *gin.Context) {
	modelName := c.Query("model")
	m := model.GetModelByName(modelName)
	if m == nil {
		c.AbortWithError(404, fmt.Errorf("Model: \"%s\" not found", modelName))
		return
	}

	file := c.Query("path")
	if file == "" {
		c.AbortWithError(404, fmt.Errorf("File not found"))
		return
	}

	downloadURL, err := storage.Download(m.Config, file)
	if err != nil || len(downloadURL) == 0 {
		c.AbortWithError(500, err)
		return
	}

	c.Redirect(302, downloadURL)
}

// GET /api/log
func log(c *gin.Context) {
	// https://github.com/gin-gonic/examples/blob/master/realtime-chat/main.go#L27
	chanStream := tailFile()
	clientGone := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			println("Client gone, close stream.")
			return false
		case msg := <-chanStream:
			if os.Getenv("GO_ENV") == "dev" {
				println(msg)
			}

			if _, err := c.Writer.WriteString(msg + "\n"); err != nil {
				logger.Errorf("Failed to write to stream: %v", err)
			}
			c.Writer.Flush()
			return true
		}
	})
}

// tailFile tail the log file and make a chain to stream output log
func tailFile() chan string {
	out_chan := make(chan string)

	file := fls.LineFile(logFile)
	if _, err := file.SeekLine(-50, io.SeekEnd); err != nil {
		logger.Errorf("Failed to seek log file: %v", err)
	}
	bf := bufio.NewReader(file)

	go func() {
		for {
			line, _, _ := bf.ReadLine()

			if len(line) == 0 {
				time.Sleep(50 * time.Millisecond)
			} else {
				out_chan <- string(line)
			}
		}
	}()

	return out_chan
}
