package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/constants"
)

func AcademicYearMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ayID := c.GetHeader("X-Academic-Year-ID")
		if ayID == "" {
			ayID = c.Query("academic_year_id")
		}
		c.Set(constants.AcademicYearIDContextKey, ayID)
		c.Next()
	}
}

// GetAcademicYearIDFromContext extracts academic year ID from context
func GetAcademicYearIDFromContext(c *gin.Context) (uuid.UUID, error) {
	academicYearID, err := getStringFromContext(c, constants.AcademicYearIDContextKey)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w — ensure X-Academic-Year-ID header or academic_year_id query param is set", err)
	}

	id, err := uuid.Parse(academicYearID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w — ensure academic_year_id is valid", err)
	}

	return id, nil
}
