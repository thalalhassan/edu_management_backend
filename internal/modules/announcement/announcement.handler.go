package announcement

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

// Routes registers all announcement endpoints under /announcements/.
//
// All routes are protected. Intended access model:
//   - POST   /             → admin / principal creates a draft
//   - GET    /             → any authenticated user lists (filter by audience)
//   - GET    /:id          → any authenticated user
//   - PUT    /:id          → admin / principal edits a DRAFT announcement
//   - PATCH  /:id/publish  → admin / principal publishes or unpublishes
//   - DELETE /:id          → admin / principal deletes an unpublished announcement
func (h *Handler) Routes(r *gin.RouterGroup) {
	announcement := r.Group(constants.ApiAnnouncementPath)
	announcement.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		announcement.POST("/", h.create)
		announcement.GET("/", h.list)
		announcement.GET("/:id", h.getByID)
		announcement.PUT("/:id", h.update)
		announcement.PATCH("/:id/publish", h.publish)
		announcement.DELETE("/:id", h.delete)
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
	response.Created(c, resp, "Announcement created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Announcement retrieved successfully")
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

	announcements, total, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, announcements, meta, "Announcements listed successfully")
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
	response.Success(c, resp, "Announcement updated successfully")
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
	response.Success(c, resp, "Announcement publish status updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Announcement deleted successfully")
}
