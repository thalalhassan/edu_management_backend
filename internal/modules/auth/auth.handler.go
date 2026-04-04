package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func RegisterRouter(r *gin.RouterGroup, app *app.App) {
	s := NewService(NewRepository(app.DB.Gorm), &app.Config.JWT)
	h := NewHandler(s)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.login)
		auth.POST("/refresh", h.refresh)
		auth.POST("/logout", h.logout)
		auth.POST("/logout-all", h.logoutAll)
	}
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, resp, "Login successful")
}

func (h *Handler) refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, resp, "Token refreshed")
}

func (h *Handler) logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Logged out successfully")
}

func (h *Handler) logoutAll(c *gin.Context) {
	userID := c.GetString("userID") // Assume userID is set in context by auth middleware
	if userID == "" {
		response.Unauthorized(c, "user ID not found in context")
		return
	}
	if err := h.service.LogoutAllSessions(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "All sessions logged out successfully")
}

func handleAuthError(c *gin.Context, err error) {
	switch err {
	case constants.Errors.ErrUserNotFound:
		response.NotFound(c, "user not found")
	case constants.Errors.ErrInvalidCredentials:
		response.Unauthorized(c, "invalid credentials")
	case constants.Errors.ErrUserExists:
		response.Conflict(c, "user already exists")
	default:
		response.InternalError(c, "internal error")
	}
}
