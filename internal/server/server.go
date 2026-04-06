package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/modules"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

func StartServer(appInstance *app.App) *gin.Engine {

	if appInstance.Config.App.Env != "production" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginRouter := gin.New()

	ginRouter.Use(gin.Logger())
	ginRouter.Use(middleware.RecoveryWithResponse())
	ginRouter.Use(middleware.ZapLogger(appInstance.Logger))
	ginRouter.Use(middleware.AcademicYearMiddleware())

	// API versioning
	api := ginRouter.Group(constants.ApiVersion)

	// Health check endpoint
	ginRouter.GET("/health", healthCheckHandler)

	// Register module routes
	modules.RegisterModules(api, appInstance)

	return ginRouter

}

func healthCheckHandler(c *gin.Context) {
	response.Success(c, gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
	}, "Server is healthy")
}
