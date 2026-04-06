package query_params

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Query[F any] struct {
	Pagination pagination.Params
	Sort       SortParams
	Filter     F
}

var allowedDirections = map[string]bool{"asc": true, "desc": true}

type SortParams struct {
	Field     string
	Direction string // "asc" | "desc"
}

// NewSortFromRequest parses ?sort_by=created_at&sort_dir=desc.
// allowedFields whitelist prevents SQL injection — each module
// passes its own allowed set.
func NewSortFromRequest(c *gin.Context, allowedFields map[string]bool, defaultField string) SortParams {
	field := c.DefaultQuery("sort_by", defaultField)
	dir := strings.ToLower(c.DefaultQuery("sort_dir", "desc"))

	if !allowedFields[field] {
		field = defaultField
	}
	if !allowedDirections[dir] {
		dir = "desc"
	}
	return SortParams{Field: field, Direction: dir}
}

// Apply adds ORDER BY to a gorm query.
func (s SortParams) Apply(db *gorm.DB) *gorm.DB {
	return db.Order(fmt.Sprintf("%s %s", s.Field, s.Direction))
}
