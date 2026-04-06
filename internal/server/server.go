package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
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

	// CORS middleware - allow all origins for simplicity (TODO: update in production)
	ginRouter.Use(cors.Default())

	// API versioning
	api := ginRouter.Group(constants.ApiVersion)

	if appInstance.Config.App.Env != "production" {
		// Swagger documentation setup
		docs.SwaggerInfo.BasePath = constants.ApiVersion
		api.StaticFile("/swagger-doc/doc.yaml", "./docs/openapi.yaml")
		api.GET("/docs/*any", ginSwagger.WrapHandler(
			swaggerFiles.Handler,
			ginSwagger.URL(constants.ApiVersion+"/swagger-doc/doc.yaml"),
			ginSwagger.PersistAuthorization(true),
		))
		api.GET("/docs", func(c *gin.Context) {
			c.Redirect(302, constants.ApiVersion+"/docs/index.html")
		})
	}

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
