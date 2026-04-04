package modules

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/modules/auth"
	"github.com/thalalhassan/edu_management/internal/modules/student"
)

func RegisterModules(api *gin.RouterGroup, app *app.App) {

	// Register auth module routes
	auth.RegisterRouter(api, app)

	// Register student module routes
	student.RegisterRouter(api, app)

}
