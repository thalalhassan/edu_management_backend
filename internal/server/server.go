package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/response"
)

func StartServer(appInstance *app.App) *gin.Engine {

	if appInstance.Config.App.Env != "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginRouter := gin.New()

	ginRouter.Use(gin.Logger())
	ginRouter.Use(gin.Recovery())

	// Add your routes here

	// Health check endpoint
	ginRouter.GET("/health", healthCheckHandler)

	return ginRouter

}

func healthCheckHandler(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil, "Server is healthy")
}
