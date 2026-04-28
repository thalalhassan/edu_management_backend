package subject

import "github.com/thalalhassan/edu_management/internal/database"

type Subject = database.Subject

// AllowedSortFields whitelists columns safe to ORDER BY.
var AllowedSortFields = map[string]bool{
	"created_at": true,
	"name":       true,
	"code":       true,
}

// FilterParams binds from query string via ShouldBindQuery.
type FilterParams struct {
	Search *string `form:"search"` // name ILIKE or code ILIKE
}

type CreateRequest struct {
	Code        string  `json:"code"         binding:"required"`
	Name        string  `json:"name"         binding:"required"`
	Description *string `json:"description,omitempty"`
}

type UpdateRequest struct {
	Code        *string `json:"code,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type SubjectResponse struct {
	Subject
}

func ToSubjectResponse(s *Subject) *SubjectResponse {
	return &SubjectResponse{Subject: *s}
}
