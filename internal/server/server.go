package server

import (
	"embed"
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
	var openapiFS embed.FS

	if appInstance.Config.App.Env != "production" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginRouter := gin.New()

	ginRouter.Use(gin.Logger())
	ginRouter.Use(middleware.RecoveryWithResponse())

	// CORS middleware - allow all origins for simplicity (TODO: update in production)
	ginRouter.Use(cors.New(cors.Config{
		AllowOrigins:     appInstance.Config.Cors.AllowOrigins,
		AllowMethods:     appInstance.Config.Cors.AllowMethods,
		AllowHeaders:     appInstance.Config.Cors.AllowHeaders,
		ExposeHeaders:    appInstance.Config.Cors.ExposeHeaders,
		AllowCredentials: appInstance.Config.Cors.AllowCredentials,
		MaxAge:           appInstance.Config.Cors.MaxAge,
	}))

	ginRouter.Use(middleware.ZapLogger(appInstance.Logger))

	ginRouter.Use(middleware.AcademicYearMiddleware())

	// API versioning
	api := ginRouter.Group(constants.ApiVersion)

	if appInstance.Config.App.Env != "production" {
		// Swagger documentation setup
		docs.SwaggerInfo.BasePath = constants.ApiVersion

		data, _ := openapiFS.ReadFile("docs/openapi.yaml")
		api.GET("/swagger-doc/doc.yaml", func(c *gin.Context) {
			c.Data(200, "application/yaml", data)
		})

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
