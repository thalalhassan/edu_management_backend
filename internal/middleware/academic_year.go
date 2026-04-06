package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/constants"
)

func AcademicYearMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ayID := c.GetHeader("X-Academic-Year-ID")
		if ayID != "" {
			c.Set(constants.AcademicYearIDContextKey, ayID)
		} else {
			ayID = c.Query("academic_year_id")
		}
		c.Next()
	}
}

// GetAcademicYearIDFromContext extracts academic year ID from context
func GetAcademicYearIDFromContext(c *gin.Context) (string, error) {
	academicYearID, err := getStringFromContext(c, constants.AcademicYearIDContextKey)
	if err != nil {
		return "", fmt.Errorf("%w — ensure X-Academic-Year-ID header or academic_year_id query param is set", err)
	}
	return academicYearID, nil
}
