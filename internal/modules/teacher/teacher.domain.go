package teacher

import (
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

type Teacher = database.Employee

type Gender = database.Gender

type FilterParams struct {
	Search     *string    `form:"search"`
	IsActive   *bool      `form:"is_active"`
	EmployeeID *uuid.UUID `form:"employee_id"`
}

func (f *FilterParams) ToMap() map[string]interface{} {
	m := map[string]interface{}{}
	if f.IsActive != nil {
		m["is_active"] = *f.IsActive
	}
	if f.EmployeeID != nil {
		m["employee_id"] = *f.EmployeeID
	}
	return m
}

var AllowedTeacherSortFields = map[string]bool{
	"date":       true,
	"status":     true,
	"created_at": true,
}

// UpdateRequest covers all mutable profile fields.
// Creation is owned by the user module (RegisterTeacher).
type UpdateRequest struct {
	FirstName      *string    `json:"first_name" binding:"omitempty,min=2,max=50"`
	LastName       *string    `json:"last_name" binding:"omitempty,min=1,max=50"`
	Gender         *Gender    `json:"gender" binding:"omitempty,oneof=MALE FEMALE"`
	DOB            *time.Time `json:"dob" binding:"omitempty,lte"`
	Phone          *string    `json:"phone" binding:"omitempty,e164"`
	Address        *string    `json:"address" binding:"omitempty,max=255"`
	Qualification  *string    `json:"qualification" binding:"omitempty,max=100"`
	Specialization *string    `json:"specialization" binding:"omitempty,max=100"`
	JoiningDate    *time.Time `json:"joining_date" binding:"omitempty,lte"`
	PhotoURL       *string    `json:"photo_url" binding:"omitempty,url"`
	IsActive       *bool      `json:"is_active"`
}

type TeacherResponse struct {
	Teacher
}

func ToTeacherResponse(t *Teacher) *TeacherResponse {
	return &TeacherResponse{Teacher: *t}
}
