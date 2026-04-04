package user

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func RegisterRouter(r *gin.RouterGroup, a *app.App) {
	svc := NewService(NewRepository(a.DB.Gorm))
	h := NewHandler(svc)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {

	// ── User management routes ──────────────────────────────
	users := r.Group("/users")
	{
		users.GET("/", h.list)
		users.GET("/:id", h.getByID)
		users.PUT("/:id", h.update)
		users.DELETE("/:id", h.delete)
		users.POST("/:id/change-password", h.changePassword)
	}

	// ── Registration routes (admin-only in production; guard with middleware) ──
	reg := users.Group("/register")
	{
		reg.POST("/student", h.registerStudent)
		reg.POST("/teacher", h.registerTeacher)
		reg.POST("/parent", h.registerParent)
		reg.POST("/staff", h.registerStaff)
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

func (h *Handler) registerTeacher(c *gin.Context) {
	var req CreateTeacherUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RegisterTeacher(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Teacher registered successfully")
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

func (h *Handler) registerStaff(c *gin.Context) {
	var req CreateStaffUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.RegisterStaff(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, resp, "Staff registered successfully")
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

// ──────────────────────────────────────────────────────────────
// USER MANAGEMENT
// ──────────────────────────────────────────────────────────────

func (h *Handler) getByID(c *gin.Context) {
	id := c.Param("id")
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
	resp, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "User updated successfully")
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success[any](c, nil, "User deleted successfully")
}

func (h *Handler) changePassword(c *gin.Context) {
	id := c.Param("id")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Password changed successfully")
}
