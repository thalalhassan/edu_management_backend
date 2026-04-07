package teacher_assignment

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type Handler struct {
	service Service
	config  *config.Config
}

func NewHandler(s Service, c *config.Config) *Handler {
	return &Handler{service: s, config: c}
}

func RegisterRouter(r *gin.RouterGroup, a *app.App) {
	svc := NewService(NewRepository(a.DB.Gorm))
	h := NewHandler(svc, a.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	group := r.Group(constants.ApiTeacherAssignmentPath)
	group.Use(middleware.AuthCheckMiddleware(&h.config.JWT))

	group.POST("/", h.Create)
	group.GET("/", h.List)
	group.GET("/:id", h.GetByID)
	group.PUT("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}

var allowedSort = map[string]bool{
	"created_at": true,
}

// ==========================================
// HANDLERS
// ==========================================
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.service.Create(c, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, res, "Teacher assigned successfully")
}

func (h *Handler) List(c *gin.Context) {
	var f FilterParams
	_ = c.ShouldBindQuery(&f)

	q := query_params.Query[FilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedSort, "created_at"),
		Filter:     f,
	}

	data, count, err := h.service.List(c, q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	meta := pagination.NewMeta(count, q.Pagination)
	response.SuccessPaginated(c, data, meta, "Teacher assignments fetched")
}

func (h *Handler) GetByID(c *gin.Context) {
	res, err := h.service.GetByID(c, c.Param("id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, res, "Teacher assignment fetched")
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.service.Update(c, c.Param("id"), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, res, "Teacher assignment updated")
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c, c.Param("id")); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success[any](c, nil, "Teacher assignment deleted")
}
