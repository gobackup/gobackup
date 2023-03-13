package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
	"github.com/spf13/viper"
)

func AuthRequired(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")

		if auth != token {
			c.AbortWithStatusJSON(403, gin.H{"error": "Access denied"})
			return
		}
	}
}

// StartHTTP run API server
func StartHTTP(version string, apiToken string) error {
	logger := logger.Tag("API")

	viper.SetDefault("api.port", 7023)

	port := viper.GetString("api.port")
	logger.Infof("Starting API server on port http://127.0.0.1:%s", port)
	logger.Infof("API Token: %s", apiToken)

	r := setupRouter(version, apiToken)

	return r.Run(":" + port)
}

func setupRouter(version string, apiToken string) *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GoBackup is running.",
			"version": version,
		})
	})

	group := r.Group("/api")
	group.Use(AuthRequired(apiToken))

	group.POST("/perform", perform)
	group.GET("/status", status)

	return r
}

func status(c *gin.Context) {
	infos := gin.H{
		"models": len(model.GetModels()),
	}

	c.JSON(200, infos)
}

func perform(c *gin.Context) {
	m := model.GetModelByName(c.Params.ByName("model"))
	if m == nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Model not found"})
		return
	}

	if err := m.Perform(); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
	} else {
		c.JSON(200, gin.H{"message": "Success"})
	}
}
