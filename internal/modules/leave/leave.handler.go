package leave

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

// Routes registers all leave endpoints under /leave/.
//
// Intended access model (enforced at the application / role layer, not here):
//   - POST   /           → teacher applies for leave
//   - GET    /           → admin / principal lists all; teacher sees own (filter by teacher_id)
//   - GET    /:id        → any authenticated user (own record check done at service / app layer)
//   - PUT    /:id        → teacher edits own PENDING request
//   - PATCH  /:id/review → admin / principal reviews (APPROVED | REJECTED)
//   - PATCH  /:id/cancel → teacher cancels own PENDING request
//   - DELETE /:id        → admin deletes a PENDING / REJECTED request
func (h *Handler) Routes(r *gin.RouterGroup) {
	leave := r.Group(constants.ApiLeavePath)
	leave.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		leave.POST("/", h.apply)
		leave.GET("/", h.list)
		leave.GET("/:id", h.getByID)
		leave.PUT("/:id", h.update)
		leave.PATCH("/:id/review", h.review)
		leave.PATCH("/:id/cancel", h.cancel)
		leave.DELETE("/:id", h.delete)
	}
}

// ─── Handlers ─────────────────────────────────────────────────────────────────

func (h *Handler) apply(c *gin.Context) {
	var req ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Apply(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Leave request submitted successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Leave request retrieved successfully")
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

	leaves, total, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, leaves, meta, "Leave requests listed successfully")
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
	response.Success(c, resp, "Leave request updated successfully")
}

// review is called by an admin or principal.
// The reviewer's user ID is pulled from the JWT claims set by AuthCheckMiddleware.
func (h *Handler) review(c *gin.Context) {
	id := c.Param("id")

	reviewerID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "reviewer identity not found in token")
		return
	}

	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.service.Review(c.Request.Context(), id, reviewerID.(string), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	msg := "Leave request approved"
	if req.Status == "REJECTED" {
		msg = "Leave request rejected"
	}
	response.Success(c, resp, msg)
}

// cancel is called by the teacher themselves to withdraw a pending request.
func (h *Handler) cancel(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.Cancel(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Leave request cancelled successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Leave request deleted successfully")
}
