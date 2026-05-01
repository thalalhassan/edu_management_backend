package standard

import (
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

type Standard = database.Standard
type StandardSubject = database.StandardSubject

type CreateRequest struct {
	Name         string    `json:"name"          binding:"required"`
	DepartmentID uuid.UUID `json:"department_id" binding:"required,uuid"`
	OrderIndex   int       `json:"order_index"`
	Description  *string   `json:"description,omitempty"`
}

type UpdateRequest struct {
	Name         *string    `json:"name,omitempty"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty" binding:"omitempty,uuid"`
	OrderIndex   *int       `json:"order_index,omitempty"`
	Description  *string    `json:"description,omitempty"`
}

// AssignSubjectRequest links or unlinks a subject from a standard.
type AssignSubjectRequest struct {
	SubjectID uuid.UUID `json:"subject_id" binding:"required,uuid"`
	IsCore    bool      `json:"is_core"`
}

type StandardResponse struct {
	Standard
}

func ToStandardResponse(s *Standard) *StandardResponse {
	return &StandardResponse{Standard: *s}
}
