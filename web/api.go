package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
	"github.com/spf13/viper"
)

//go:embed dist
var staticFS embed.FS

// StartHTTP run API server
func StartHTTP(version string, apiToken string) error {
	logger := logger.Tag("API")

	viper.SetDefault("api.port", 2703)

	port := viper.GetString("api.port")
	logger.Infof("Starting API server on port http://127.0.0.1:%s", port)
	logger.Infof("API Token: %s", apiToken)

	r := setupRouter(version, apiToken)

	mutex := http.NewServeMux()

	fe, _ := fs.Sub(staticFS, "dist")
	mutex.Handle("/", http.FileServer(http.FS(fe)))
	mutex.Handle("/status", r)
	mutex.Handle("/api/", r)

	return http.ListenAndServe(":"+port, mutex)
}

func setupRouter(version string, apiToken string) *gin.Engine {
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
