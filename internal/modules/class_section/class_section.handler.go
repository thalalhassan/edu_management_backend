package class_section

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/middleware"
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
	cs := r.Group("/class-section")
	cs.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		cs.POST("/", h.create)
		cs.GET("/:id", h.getByID)

		// Primary list — scoped to academic year (from header or query)
		cs.GET("/", h.listByAcademicYear)

		// Scoped lists
		cs.GET("/standard/:standard_id", h.listByStandard)
		cs.GET("/teacher/:teacher_id", h.listByTeacher)

		cs.PUT("/:id", h.update)
		cs.PATCH("/:id/teacher", h.assignTeacher)
		cs.DELETE("/:id", h.delete)
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
	response.Created(c, resp, "Class section created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Class section retrieved successfully")
}

// listByAcademicYear reads academic_year_id from the X-Academic-Year-ID header first,
// falling back to a query param for cases where the header is not set.
func (h *Handler) listByAcademicYear(c *gin.Context) {
	academicYearID := c.GetHeader("X-Academic-Year-ID")
	if academicYearID == "" {
		academicYearID = c.Query("academic_year_id")
	}
	if academicYearID == "" {
		response.BadRequest(c, "academic_year_id is required — pass via X-Academic-Year-ID header or query param")
		return
	}

	resp, err := h.service.ListByAcademicYear(c.Request.Context(), academicYearID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Class sections listed successfully")
}

func (h *Handler) listByStandard(c *gin.Context) {
	standardID := c.Param("standard_id")
	academicYearID := c.GetHeader("X-Academic-Year-ID")
	if academicYearID == "" {
		academicYearID = c.Query("academic_year_id")
	}
	if academicYearID == "" {
		response.BadRequest(c, "academic_year_id is required")
		return
	}

	resp, err := h.service.ListByStandard(c.Request.Context(), standardID, academicYearID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Class sections listed successfully")
}

func (h *Handler) listByTeacher(c *gin.Context) {
	teacherID := c.Param("teacher_id")
	academicYearID := c.GetHeader("X-Academic-Year-ID")
	if academicYearID == "" {
		academicYearID = c.Query("academic_year_id")
	}
	if academicYearID == "" {
		response.BadRequest(c, "academic_year_id is required")
		return
	}

	resp, err := h.service.ListByTeacher(c.Request.Context(), teacherID, academicYearID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Class sections listed successfully")
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")
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
	response.Success(c, resp, "Class section updated successfully")
}

func (h *Handler) assignTeacher(c *gin.Context) {
	id := c.Param("id")
	var req AssignTeacherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.AssignTeacher(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Class teacher updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Class section deleted successfully")
}
