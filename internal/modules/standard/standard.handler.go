package standard

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
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
	s := r.Group(constants.ApiStandardPath)
	s.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		s.POST("/", h.create)
		s.GET("/", h.list)
		s.GET("/:id", h.getByID)
		s.GET("/department/:department_id", h.listByDepartment)
		s.PUT("/:id", h.update)
		s.DELETE("/:id", h.delete)

		// Subject assignments
		s.GET("/:id/subjects", h.getSubjects)
		s.POST("/:id/subjects", h.assignSubject)
		s.DELETE("/:id/subjects/:subject_id", h.removeSubject)
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
	response.Created(c, resp, "Standard created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Standard retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	resp, err := h.service.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Standards listed successfully")
}

func (h *Handler) listByDepartment(c *gin.Context) {
	departmentID := c.Param("department_id")
	resp, err := h.service.ListByDepartment(c.Request.Context(), departmentID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Standards listed successfully")
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
	response.Success(c, resp, "Standard updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Standard deleted successfully")
}

func (h *Handler) getSubjects(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetSubjects(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Subjects retrieved successfully")
}

func (h *Handler) assignSubject(c *gin.Context) {
	id := c.Param("id")
	var req AssignSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.AssignSubject(c.Request.Context(), id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Subject assigned successfully")
}

func (h *Handler) removeSubject(c *gin.Context) {
	id := c.Param("id")
	subjectID := c.Param("subject_id")
	if err := h.service.RemoveSubject(c.Request.Context(), id, subjectID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Subject removed successfully")
}
