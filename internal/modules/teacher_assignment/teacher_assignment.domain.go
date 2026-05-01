package teacher_assignment

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// FILTER
// ==========================================
type FilterParams struct {
	ClassSectionID *uuid.UUID `form:"class_section_id"`
	EmployeeID     *uuid.UUID `form:"employee_id"`
	SubjectID      *uuid.UUID `form:"subject_id"`
}

// ==========================================
// REQUEST
// ==========================================
type CreateRequest struct {
	ClassSectionID uuid.UUID `json:"class_section_id" binding:"required,uuid"`
	EmployeeID     uuid.UUID `json:"employee_id"     binding:"required,uuid"`
	SubjectID      uuid.UUID `json:"subject_id" binding:"required,uuid"`
}

type UpdateRequest struct {
	EmployeeID *uuid.UUID `json:"employee_id,omitempty" binding:"omitempty,uuid"`
}

// ==========================================
// RESPONSE
// ==========================================
type Response struct {
	ID             uuid.UUID `json:"id"`
	ClassSectionID uuid.UUID `json:"class_section_id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	SubjectID      uuid.UUID `json:"subject_id"`
	CreatedAt      time.Time `json:"created_at"`
}
