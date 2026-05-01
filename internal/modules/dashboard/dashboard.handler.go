package dashboard

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
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
	svc := NewService(NewRepository(a.DB.Gorm), a.DB.Gorm)
	h := NewHandler(svc, a.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	d := r.Group("/dashboard")
	d.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		d.GET("/", h.getDashboard)
	}
}

// getDashboard godoc
// GET /dashboard/
// Reads academic_year_id from X-Academic-Year-ID header, falls back to query param.
// Returns Institution Overview, Fees & Collections, Payroll & Salaries sections.
func (h *Handler) getDashboard(c *gin.Context) {
	academicYearID := c.GetHeader("X-Academic-Year-ID")
	if academicYearID == "" {
		academicYearID = c.Query("academic_year_id")
	}
	if academicYearID == "" {
		response.BadRequest(c, "academic_year_id is required — pass via X-Academic-Year-ID header or ?academic_year_id= query param")
		return
	}
	academicYearUUID, err := uuid.Parse(academicYearID)
	if err != nil {
		response.BadRequest(c, "Invalid academic_year_id format")
		return
	}

	resp, err := h.service.GetDashboard(c.Request.Context(), academicYearUUID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Dashboard retrieved successfully")
}
