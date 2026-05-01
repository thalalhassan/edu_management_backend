package attendance

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

// ─── Handler ─────────────────────────────────────────────────────────────────

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
	// ── Student Attendance ────────────────────────────────────────────────────
	att := r.Group(constants.ApiAttendancePath)
	att.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		att.POST("/", h.markAttendance)
		att.POST("/bulk", h.bulkMarkAttendance)
		att.GET("/", h.listStudentAttendance)
		att.GET("/:id", h.getAttendance)
		att.PUT("/:id", h.updateAttendance)
		att.DELETE("/:id", h.deleteAttendance)

		// Class-level daily summary — ?date=YYYY-MM-DD
		att.GET("/class/:class_section_id/summary", h.getClassAttendanceSummary)
	}

	// ── Employee Attendance ────────────────────────────────────────────────────
	ta := r.Group(constants.ApiEmployeeAttendancePath)
	ta.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		ta.POST("/", h.markEmployeeAttendance)
		ta.POST("/bulk", h.bulkMarkEmployeeAttendance)
		ta.GET("/", h.listEmployeeAttendance)
		ta.GET("/:id", h.getEmployeeAttendance)
		ta.PUT("/:id", h.updateEmployeeAttendance)
		ta.DELETE("/:id", h.deleteEmployeeAttendance)
	}
}

// ─── Student Attendance handlers ──────────────────────────────────────────────

func (h *Handler) markAttendance(c *gin.Context) {
	var req MarkStudentAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.MarkAttendance(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Attendance marked successfully")
}

func (h *Handler) bulkMarkAttendance(c *gin.Context) {
	var req BulkMarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.BulkMarkAttendance(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Attendance marked for all students successfully")
}

func (h *Handler) getAttendance(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	resp, err := h.service.GetAttendanceByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Attendance record retrieved successfully")
}

func (h *Handler) listStudentAttendance(c *gin.Context) {
	var f StudentFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	q := query_params.Query[StudentFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedStudentSortFields, "date"),
		Filter:     f,
	}

	records, total, err := h.service.ListStudentAttendance(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, records, meta, "Attendance records listed successfully")
}

func (h *Handler) getClassAttendanceSummary(c *gin.Context) {
	classSectionIDStr := c.Param("class_section_id")
	classSectionID, err := uuid.Parse(classSectionIDStr)
	if err != nil {
		response.BadRequest(c, "invalid class section ID format")
		return
	}
	date := c.Query("date")
	if date == "" {
		response.BadRequest(c, "query param 'date' is required (format: YYYY-MM-DD)")
		return
	}

	resp, err := h.service.GetClassAttendanceSummary(c.Request.Context(), classSectionID, date)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Class attendance summary retrieved successfully")
}

func (h *Handler) updateAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}
	var req UpdateStudentAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateAttendance(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Attendance updated successfully")
}

func (h *Handler) deleteAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	if err := h.service.DeleteAttendance(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Attendance record deleted successfully")
}

// ─── Employee Attendance handlers ──────────────────────────────────────────────

func (h *Handler) markEmployeeAttendance(c *gin.Context) {
	var req MarkEmployeeAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.MarkEmployeeAttendance(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Employee attendance marked successfully")
}

func (h *Handler) bulkMarkEmployeeAttendance(c *gin.Context) {
	var req BulkMarkEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.BulkMarkEmployeeAttendance(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Employee attendance marked in bulk successfully")
}

func (h *Handler) getEmployeeAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}
	resp, err := h.service.GetEmployeeAttendanceByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Employee attendance record retrieved successfully")
}

func (h *Handler) listEmployeeAttendance(c *gin.Context) {
	var f EmployeeFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	q := query_params.Query[EmployeeFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedEmployeeSortFields, "date"),
		Filter:     f,
	}

	records, total, err := h.service.ListEmployeeAttendance(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, records, meta, "Employee attendance listed successfully")
}

func (h *Handler) updateEmployeeAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}
	var req UpdateEmployeeAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateEmployeeAttendance(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher attendance updated successfully")
}

func (h *Handler) deleteEmployeeAttendance(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid ID format")
		return
	}

	if err := h.service.DeleteEmployeeAttendance(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Employee attendance record deleted successfully")
}
