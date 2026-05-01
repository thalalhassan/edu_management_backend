package timetable

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/helper"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type Handler struct {
	service Service
	config  *config.Config
}

func NewHandler(s Service, config *config.Config) *Handler {
	return &Handler{service: s, config: config}
}

func RegisterRouter(r *gin.RouterGroup, a *app.App) {
	svc := NewService(NewRepository(a.DB.Gorm))
	h := NewHandler(svc, a.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	tt := r.Group(constants.ApiTimetablePath)
	tt.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		tt.POST("/", h.create)
		tt.GET("/", h.list)
		tt.GET("/:id", h.getByID)

		// Weekly schedule views — primary UI surfaces
		tt.GET("/class/:class_section_id/schedule", h.getClassSchedule)
		tt.GET("/employee/:employee_id/schedule", h.getEmployeeSchedule)

		tt.PUT("/:id", h.update)
		tt.DELETE("/:id", h.delete)
	}
}

func (h *Handler) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Timetable entry created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Timetable entry retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	var f FilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	q := query_params.Query[FilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, AllowedSortFields, "day_of_week"),
		Filter:     f,
	}

	resp, total, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, q.Pagination), "Timetable entries listed successfully")
}

// getClassSchedule returns the weekly timetable for a class section
// grouped by day — the view students and class teachers use most.
func (h *Handler) getClassSchedule(c *gin.Context) {
	classSectionID, valid := helper.ParseParamUUIDWithAbort(c, "class_section_id")
	if !valid {
		return
	}
	resp, err := h.service.GetClassSchedule(c.Request.Context(), classSectionID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Class schedule retrieved successfully")
}

// getEmployeeSchedule returns the full week schedule for an employee
// across all their assigned class sections.
func (h *Handler) getEmployeeSchedule(c *gin.Context) {
	employeeID, valid := helper.ParseParamUUIDWithAbort(c, "employee_id")
	if !valid {
		return
	}
	resp, err := h.service.GetEmployeeSchedule(c.Request.Context(), employeeID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Employee schedule retrieved successfully")
}

func (h *Handler) update(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Timetable entry updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Timetable entry deleted successfully")
}
