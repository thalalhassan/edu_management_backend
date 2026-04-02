package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Params defines the input for the database query.
type Params struct {
	Limit  int
	Offset int
	Page   int
}

// Meta defines the output metadata for the API response.
type Meta struct {
	CurrentPage  int   `json:"current_page"`
	ItemsPerPage int   `json:"items_per_page"`
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
}

// NewFromRequest extracts pagination parameters from HTTP query strings.
func NewFromRequest(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10 // Default limit
	}

	return Params{
		Limit:  limit,
		Offset: (page - 1) * limit,
		Page:   page,
	}
}

// NewMeta calculates the response metadata.
func NewMeta(totalItems int64, params Params) Meta {
	totalPages := int(totalItems) / params.Limit
	if int(totalItems)%params.Limit != 0 {
		totalPages++
	}

	return Meta{
		CurrentPage:  params.Page,
		ItemsPerPage: params.Limit,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
	}
}
