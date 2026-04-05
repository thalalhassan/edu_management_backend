package department

import "github.com/thalalhassan/edu_management/internal/database"

type Department = database.Department

type CreateRequest struct {
	Name          string  `json:"name"          binding:"required"`
	Code          string  `json:"code"          binding:"required"`
	Description   *string `json:"description,omitempty"`
	HeadTeacherID *string `json:"head_teacher_id,omitempty" binding:"omitempty,uuid"`
}

type UpdateRequest struct {
	Name          *string `json:"name,omitempty"`
	Code          *string `json:"code,omitempty"`
	Description   *string `json:"description,omitempty"`
	HeadTeacherID *string `json:"head_teacher_id,omitempty" binding:"omitempty,uuid"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// AssignHeadRequest is a dedicated request for assigning / removing the head teacher.
// Keeping it separate makes the intent explicit and avoids a full PUT for a single field change.
type AssignHeadRequest struct {
	HeadTeacherID *string `json:"head_teacher_id" binding:"omitempty,uuid"` // nil = remove head
}

type DepartmentResponse struct {
	Department
}

func ToDepartmentResponse(d *Department) *DepartmentResponse {
	return &DepartmentResponse{Department: *d}
}
