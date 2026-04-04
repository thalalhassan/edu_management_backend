package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
)

type Envelope[T any] struct {
	Success bool             `json:"success"`
	Message string           `json:"message,omitempty"`
	Data    T                `json:"data"`
	Error   string           `json:"error,omitempty"`
	Meta    *pagination.Meta `json:"meta,omitempty"`
}

func Success[T any](c *gin.Context, data T, msg string) {
	c.JSON(http.StatusOK, Envelope[T]{Success: true, Data: data, Message: msg})
}

func Created[T any](c *gin.Context, data T, msg string) {
	c.JSON(http.StatusCreated, Envelope[T]{Success: true, Data: data, Message: msg})
}

func SuccessPaginated[T any](c *gin.Context, data T, meta pagination.Meta, msg string) {
	c.JSON(http.StatusOK, Envelope[T]{Success: true, Data: data, Meta: &meta, Message: msg})
}

func BadRequest(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, Envelope[any]{Error: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusNotFound, Envelope[any]{Error: msg})
}

func InternalError(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, Envelope[any]{Error: msg})
}

func Conflict(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusConflict, Envelope[any]{Error: msg})
}

func Unprocessable(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, Envelope[any]{Error: msg})
}

func Unauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, Envelope[any]{Error: msg})
}

func AbortWithError(c *gin.Context, statusCode int, errorCode string, msg string) {
	c.AbortWithStatusJSON(statusCode, Envelope[any]{Error: msg})
}

func AccessDenied(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusForbidden, Envelope[any]{Error: msg})
}
