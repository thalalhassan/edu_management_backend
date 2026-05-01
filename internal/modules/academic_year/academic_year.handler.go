package academic_year

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
	ay := r.Group(constants.ApiAcademicYearPath)
	{
		// Public — UI needs this on load to populate the selector
		// and to set the default active year before the user is
		// fully authenticated (e.g. login page dropdown).
		ay.GET("/active", h.getActive)
		ay.GET("/", h.list)

		// Protected — mutations require auth
		protected := ay.Group("/")
		protected.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
		{
			protected.POST("/", h.create)
			protected.GET("/:id", h.getByID)
			protected.PUT("/:id", h.update)
			protected.PATCH("/:id/activate", h.setActive)
			protected.DELETE("/:id", h.delete)
		}
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
		h.handleError(c, err)
		return
	}
	response.Created(c, resp, "Academic year created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Academic year retrieved successfully")
}

func (h *Handler) getActive(c *gin.Context) {
	resp, err := h.service.GetActive(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Active academic year retrieved successfully")
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
	resp, err := h.service.List(c.Request.Context(), q)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Academic years listed successfully")
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
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Academic year updated successfully")
}

func (h *Handler) setActive(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}
	resp, err := h.service.SetActive(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Success(c, resp, "Academic year activated successfully")
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
	response.Success[any](c, nil, "Academic year deleted successfully")
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *ValidationError:
		response.BadRequest(c, e.Error())
	case *NotFoundError:
		response.NotFound(c, e.Error())
	case *BusinessError:
		response.Conflict(c, e.Error())
	default:
		response.InternalError(c, "An unexpected error occurred")
	}
}

var allowedSortFields = map[string]bool{
	"name":       true,
	"start_date": true,
	"end_date":   true,
	"created_at": true,
}
