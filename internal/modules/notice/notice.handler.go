package notice

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

// Routes registers all notice endpoints under /notice/.
//
// All routes are protected. Intended access model:
//   - POST   /             → admin / principal creates a draft
//   - GET    /             → any authenticated user lists (filter by audience/class)
//   - GET    /:id          → any authenticated user
//   - PUT    /:id          → admin / principal edits a DRAFT notice
//   - PATCH  /:id/publish  → admin / principal publishes or unpublishes
//   - DELETE /:id          → admin / principal deletes an unpublished notice
func (h *Handler) Routes(r *gin.RouterGroup) {
	notice := r.Group(constants.ApiNoticePath)
	notice.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		notice.POST("/", h.create)
		notice.GET("/", h.list)
		notice.GET("/:id", h.getByID)
		notice.PUT("/:id", h.update)
		notice.PATCH("/:id/publish", h.publish)
		notice.DELETE("/:id", h.delete)
	}
}

// ─── Handlers ────────────────────────────────────────────────────────────────

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
	response.Created(c, resp, "Notice created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Notice retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	var f FilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	q := query_params.Query[FilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedSortFields, "created_at"),
		Filter:     f,
	}

	notices, total, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, notices, meta, "Notices listed successfully")
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
	response.Success(c, resp, "Notice updated successfully")
}

func (h *Handler) publish(c *gin.Context) {
	id := c.Param("id")
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Publish(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	msg := "Notice unpublished successfully"
	if req.IsPublished {
		msg = "Notice published successfully"
	}
	response.Success(c, resp, msg)
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Notice deleted successfully")
}
