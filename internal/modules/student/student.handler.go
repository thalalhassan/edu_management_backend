package student

import (
	"fmt"

	"github.com/gin-gonic/gin"
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

func NewHandler(s Service, cfg *config.Config) *Handler {
	return &Handler{service: s, config: cfg}
}

func RegisterRouter(r *gin.RouterGroup, app *app.App) {
	s := NewService(NewRepository(app.DB.Gorm))
	h := NewHandler(s, app.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	student := r.Group(constants.ApiStudentPath)
	student.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		student.GET("/:id", h.getByID)
		student.GET("/", h.list)
		student.PUT("/:id", h.update)
		student.DELETE("/:id", h.delete)
	}
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, resp, "Student retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	p := pagination.NewFromRequest(c)

	resp, total, err := h.service.List(c.Request.Context(), p)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	fmt.Printf("Students: %+v\n", resp)

	response.SuccessPaginated(c, resp, pagination.NewMeta(total, p), "Students listed successfully")
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
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Student updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Student deleted successfully")
}
