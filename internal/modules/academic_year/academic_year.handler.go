package academic_year

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
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
	ay := r.Group("/academic-year")
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
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Academic year created successfully")
}

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Academic year retrieved successfully")
}

// getActive is public — called by the UI on initial load to set
// the global academic year context before any auth takes place.
func (h *Handler) getActive(c *gin.Context) {
	resp, err := h.service.GetActive(c.Request.Context())
	if err != nil {
		response.NotFound(c, err.Error())
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
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Academic years listed successfully")
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
	response.Success(c, resp, "Academic year updated successfully")
}

// setActive is the endpoint the UI calls when the user switches
// the global academic year selector. Returns the newly active year
// so the UI can update its context in one round trip.
func (h *Handler) setActive(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.SetActive(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Academic year activated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Academic year deleted successfully")
}
