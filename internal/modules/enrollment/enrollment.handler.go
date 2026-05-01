package enrollment

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
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
	svc := NewService(NewRepository(a.DB.Gorm), a.DB.Gorm)
	h := NewHandler(svc, a.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	e := r.Group(constants.ApiEnrollmentPath)
	e.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		// Enroll a student
		e.POST("/", h.enroll)

		// Get single enrollment with full details
		e.GET("/:id", h.getByID)

		// Academic history for a student — paginated
		e.GET("/student/:student_id", h.getByStudentID)

		// Class roster — all currently enrolled students in a section
		e.GET("/class/:class_section_id/roster", h.getRoster)

		// Promote / detain / withdraw
		e.PATCH("/:id/status", h.updateStatus)

		// Hard delete — only if no attendance/exam records exist
		e.DELETE("/:id", h.delete)
	}
}

func (h *Handler) enroll(c *gin.Context) {
	var req EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Enroll(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Student enrolled successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid enrollment ID")
		return
	}
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Enrollment retrieved successfully")
}

func (h *Handler) getByStudentID(c *gin.Context) {
	studentIDStr := c.Param("student_id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid student ID")
		return
	}
	p := pagination.NewFromRequest(c)
	resp, total, err := h.service.GetByStudentID(c.Request.Context(), studentID, p)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, p), "Enrollments retrieved successfully")
}

func (h *Handler) getRoster(c *gin.Context) {
	classSectionIDStr := c.Param("class_section_id")
	classSectionID, err := uuid.Parse(classSectionIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid class section ID")
		return
	}
	resp, err := h.service.GetRoster(c.Request.Context(), classSectionID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Roster retrieved successfully")
}

func (h *Handler) updateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid enrollment ID")
		return
	}
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Enrollment status updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid enrollment ID")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Enrollment deleted successfully")
}
