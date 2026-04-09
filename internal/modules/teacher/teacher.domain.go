package teacher

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type Teacher = database.Teacher

type Gender = database.Gender

// UpdateRequest covers all mutable profile fields.
// Creation is owned by the user module (RegisterTeacher).
type UpdateRequest struct {
	FirstName      *string    `json:"first_name" binding:"omitempty,min=2,max=50"`
	LastName       *string    `json:"last_name" binding:"omitempty,min=1,max=50"`
	Gender         *Gender    `json:"gender" binding:"omitempty,oneof=male female other"`
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
