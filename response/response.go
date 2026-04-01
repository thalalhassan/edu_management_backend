package response

import (
	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, statusCode int, data interface{}, meta interface{}, message string) {
	c.JSON(statusCode, gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

func Error(c *gin.Context, statusCode int, err error, message string) {
	c.JSON(statusCode, gin.H{
		"status":  "error",
		"message": message,
		"error":   err.Error(),
	})
}
