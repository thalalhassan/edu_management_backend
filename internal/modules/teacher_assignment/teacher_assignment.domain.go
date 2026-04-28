package teacher_assignment

import "time"

// ==========================================
// FILTER
// ==========================================
type FilterParams struct {
	ClassSectionID *string `form:"class_section_id"`
	EmployeeID     *string `form:"employee_id"`
	SubjectID      *string `form:"subject_id"`
}

// ==========================================
// REQUEST
// ==========================================
type CreateRequest struct {
	ClassSectionID string `json:"class_section_id" binding:"required,uuid"`
	EmployeeID     string `json:"employee_id"     binding:"required,uuid"`
	SubjectID      string `json:"subject_id" binding:"required,uuid"`
}

type UpdateRequest struct {
	EmployeeID *string `json:"employee_id,omitempty" binding:"omitempty,uuid"`
}

// ==========================================
// RESPONSE
// ==========================================
type Response struct {
	ID             string    `json:"id"`
	ClassSectionID string    `json:"class_section_id"`
	EmployeeID     string    `json:"employee_id"`
	SubjectID      string    `json:"subject_id"`
	CreatedAt      time.Time `json:"created_at"`
}
