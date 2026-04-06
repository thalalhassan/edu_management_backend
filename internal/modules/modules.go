package modules

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/modules/academic_year"
	"github.com/thalalhassan/edu_management/internal/modules/auth"
	"github.com/thalalhassan/edu_management/internal/modules/class_section"
	"github.com/thalalhassan/edu_management/internal/modules/department"
	"github.com/thalalhassan/edu_management/internal/modules/enrollment"
	"github.com/thalalhassan/edu_management/internal/modules/fee"
	"github.com/thalalhassan/edu_management/internal/modules/leave"
	"github.com/thalalhassan/edu_management/internal/modules/parent"
	"github.com/thalalhassan/edu_management/internal/modules/staff"
	"github.com/thalalhassan/edu_management/internal/modules/standard"
	"github.com/thalalhassan/edu_management/internal/modules/student"
	"github.com/thalalhassan/edu_management/internal/modules/teacher"
	"github.com/thalalhassan/edu_management/internal/modules/user"
)

func RegisterModules(api *gin.RouterGroup, app *app.App) {

	// Register auth module routes
	auth.RegisterRouter(api, app)

	// Register user module routes
	user.RegisterRouter(api, app)

	// Register student module routes
	student.RegisterRouter(api, app)

	// Register teacher module routes
	teacher.RegisterRouter(api, app)

	// Register parent module routes
	parent.RegisterRouter(api, app)

	// Register staff module routes
	staff.RegisterRouter(api, app)

	// Register department module routes
	department.RegisterRouter(api, app)

	// Register class section module routes
	class_section.RegisterRouter(api, app)

	// Register standard module routes
	standard.RegisterRouter(api, app)

	// Register enrollment module routes
	enrollment.RegisterRouter(api, app)

	// Register academic year module routes
	academic_year.RegisterRouter(api, app)

	// Register fee module routes
	fee.RegisterRouter(api, app)

	// Register leave module routes
	leave.RegisterRouter(api, app)

}
