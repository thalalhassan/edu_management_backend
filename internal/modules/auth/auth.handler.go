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
		auth.POST("/signin", h.signIn)
	}
}

func (h *Handler) signIn(c *gin.Context) {
	var req SignInRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.service.SignIn(c.Request.Context(), &req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	response.Success(c, resp, "Sign-in successful")
}

func handleAuthError(c *gin.Context, err error) {
	switch err {
	case constants.Errors.ErrUserNotFound:
		response.NotFound(c, "user not found")
	case constants.Errors.ErrInvalidCredentials:
		response.UnAuthorized(c, "invalid credentials")
	case constants.Errors.ErrUserExists:
		response.Conflict(c, "user already exists")
	default:
		response.InternalError(c, "internal error")
	}
}
