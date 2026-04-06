package attendance

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
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
	att := r.Group("/attendance")
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

	// ── Teacher Attendance ────────────────────────────────────────────────────
	ta := r.Group("/teacher-attendance")
	ta.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		ta.POST("/", h.markTeacherAttendance)
		ta.POST("/bulk", h.bulkMarkTeacherAttendance)
		ta.GET("/", h.listTeacherAttendance)
		ta.GET("/:id", h.getTeacherAttendance)
		ta.PUT("/:id", h.updateTeacherAttendance)
		ta.DELETE("/:id", h.deleteTeacherAttendance)
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
	id := c.Param("id")
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
	classSectionID := c.Param("class_section_id")
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
	id := c.Param("id")
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
	id := c.Param("id")
	if err := h.service.DeleteAttendance(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Attendance record deleted successfully")
}

// ─── Teacher Attendance handlers ──────────────────────────────────────────────

func (h *Handler) markTeacherAttendance(c *gin.Context) {
	var req MarkTeacherAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.MarkTeacherAttendance(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Teacher attendance marked successfully")
}

func (h *Handler) bulkMarkTeacherAttendance(c *gin.Context) {
	var req BulkMarkTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.BulkMarkTeacherAttendance(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Teacher attendance marked in bulk successfully")
}

func (h *Handler) getTeacherAttendance(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetTeacherAttendanceByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher attendance record retrieved successfully")
}

func (h *Handler) listTeacherAttendance(c *gin.Context) {
	var f TeacherFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	q := query_params.Query[TeacherFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedTeacherSortFields, "date"),
		Filter:     f,
	}

	records, total, err := h.service.ListTeacherAttendance(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, records, meta, "Teacher attendance listed successfully")
}

func (h *Handler) updateTeacherAttendance(c *gin.Context) {
	id := c.Param("id")
	var req UpdateTeacherAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateTeacherAttendance(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher attendance updated successfully")
}

func (h *Handler) deleteTeacherAttendance(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteTeacherAttendance(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Teacher attendance record deleted successfully")
}
