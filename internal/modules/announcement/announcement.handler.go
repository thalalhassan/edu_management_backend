package announcement

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
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	authorIDStr, _ := c.Get(constants.UserIDContextKey)
	authorID, err := uuid.Parse(authorIDStr.(string))
	if err != nil {
		response.BadRequest(c, "Invalid author ID format")
		return
	}
	resp, err := h.service.Create(c.Request.Context(), req, authorID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Announcement created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}

	userAudiences := h.getUserAudiences(c)
	resp, err := h.service.GetByID(c.Request.Context(), id, userAudiences)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Announcement retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	var f FilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, "invalid query: "+err.Error())
		return
	}

	q := query_params.Query[FilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedSortFields, "created_at"),
		Filter:     f,
	}

	userAudiences := h.getUserAudiences(c)
	announcements, total, err := h.service.List(c.Request.Context(), q, userAudiences)
	if err != nil {
		h.handleError(c, err)
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, announcements, meta, "Announcements listed successfully")
}

func (h *Handler) update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	resp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Announcement updated successfully")
}

func (h *Handler) publish(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}
	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}
	resp, err := h.service.Publish(c.Request.Context(), id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Announcement publish status updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}
	response.Success[any](c, nil, "Announcement deleted successfully")
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (h *Handler) handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case ValidationError:
		response.BadRequest(c, e.Error())
	case BusinessError:
		response.Unprocessable(c, e.Error())
	case NotFoundError:
		response.NotFound(c, e.Error())
	default:
		if err == ErrUnauthorized {
			response.AccessDenied(c, "access denied")
		} else {
			response.InternalError(c, "internal server error")
		}
	}
}

func (h *Handler) getUserAudiences(c *gin.Context) []AnnouncementAudience {
	role, _ := c.Get(constants.RoleContextKey)
	return GetUserAudience(role.(string))
}
