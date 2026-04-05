package middleware

import "github.com/gin-gonic/gin"

// middleware/academic_year.go
func AcademicYearMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ayID := c.GetHeader("X-Academic-Year-ID")
		if ayID != "" {
			c.Set("academic_year_id", ayID)
		}
		c.Next()
	}
}
