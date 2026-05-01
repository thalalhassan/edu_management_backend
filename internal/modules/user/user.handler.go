package user

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func RegisterRouter(r *gin.RouterGroup, a *app.App) {
	svc := NewService(NewRepository(a.DB.Gorm))
	h := NewHandler(svc, a.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {

	// ── User management routes ──────────────────────────────
	users := r.Group(constants.ApiUserPath)
	users.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		users.GET("/", middleware.IsAdmin(), h.list)
		users.GET("/:id", middleware.IsAdmin(), h.getByID)
		users.PUT("/:id", middleware.IsAdmin(), h.update)
		users.DELETE("/:id", middleware.IsAdmin(), h.delete)
		users.GET("/my-profile", h.userProfile)
		users.PUT("/my-profile", h.userUpdateProfile)
		users.POST("/change-password", h.changePassword)
	}

	// ── Registration routes (admin-only in production; guard with middleware) ──
	reg := users.Group("/register")
	reg.Use(middleware.IsAdmin())
	{
		reg.POST("/student", h.registerStudent)
		reg.POST("/employee", h.registerEmployee)
		reg.POST("/parent", h.registerParent)
		reg.POST("/admin", h.registerAdmin)
	}
}

// ──────────────────────────────────────────────────────────────
// REGISTRATION
// ──────────────────────────────────────────────────────────────

func (h *Handler) registerStudent(c *gin.Context) {
	var req CreateStudentUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RegisterStudent(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Student registered successfully")
}

func (h *Handler) registerEmployee(c *gin.Context) {
	var req CreateEmployeeUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RegisterEmployee(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Employee registered successfully")
}

func (h *Handler) registerParent(c *gin.Context) {
	var req CreateParentUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RegisterParent(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Parent registered successfully")
}

func (h *Handler) registerAdmin(c *gin.Context) {
	var req CreateAdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RegisterAdmin(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Admin registered successfully")
}

func (h *Handler) userProfile(c *gin.Context) {
	userIDStr, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	resp, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "User profile retrieved successfully")
}

// ──────────────────────────────────────────────────────────────
// USER MANAGEMENT
// ──────────────────────────────────────────────────────────────

func (h *Handler) getByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}
	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "User retrieved successfully")
}

func (h *Handler) list(c *gin.Context) {
	p := pagination.NewFromRequest(c)
	resp, total, err := h.service.List(c.Request.Context(), p)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, p), "Users listed successfully")
}

func (h *Handler) update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	idUUID, err := uuid.Parse(id)
	if err != nil {
		response.BadRequest(c, "Invalid id format")
		return
	}
	resp, err := h.service.Update(c.Request.Context(), idUUID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "User updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	idUUID, err := uuid.Parse(id)
	if err != nil {
		response.BadRequest(c, "Invalid id format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), idUUID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "User deleted successfully")
}

func (h *Handler) changePassword(c *gin.Context) {
	userIDStr, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), userID, req); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Password changed successfully")
}

func (h *Handler) userUpdateProfile(c *gin.Context) {
	userIDStr, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID format")
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Update(c.Request.Context(), userID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "User updated successfully")
}
