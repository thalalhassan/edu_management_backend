package teacher

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/response"
	"github.com/thalalhassan/edu_management/internal/shared/validation"
)

type Handler struct {
	service Service
	config  *config.Config
}

func NewHandler(s Service, cfg *config.Config) *Handler {
	return &Handler{service: s, config: cfg}
}

func RegisterRouter(r *gin.RouterGroup, app *app.App) {
	s := NewService(NewRepository(app.DB.Gorm))
	h := NewHandler(s, app.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	teacher := r.Group(constants.ApiTeacherPath)
	teacher.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		teacher.GET("/", h.list)
		teacher.GET("/:id", h.getByID)
		teacher.GET("/employee/:employee_id", h.getByEmployeeID)
		teacher.PUT("/:id", h.update)
		teacher.PATCH("/:id/active", h.setActive)
		teacher.DELETE("/:id", h.delete)
	}
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher retrieved successfully")
}

func (h *Handler) getByEmployeeID(c *gin.Context) {
	empID := c.Param("employee_id")
	resp, err := h.service.GetByEmployeeID(c.Request.Context(), empID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	p := pagination.NewFromRequest(c)
	resp, total, err := h.service.List(c.Request.Context(), p)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, p), "Teachers listed successfully")
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		response.BadRequest(c, validation.FormatErrors(err))
		return
	}
	resp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher updated successfully")
}

// setActive handles PATCH /teacher/:id/active
// Body: { "is_active": true } or { "is_active": false }
func (h *Handler) setActive(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.service.SetActive(c.Request.Context(), id, body.IsActive)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, resp, "Teacher status updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Teacher deleted successfully")
}
