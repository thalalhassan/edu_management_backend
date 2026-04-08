package teacher

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type Teacher = database.Teacher

// UpdateRequest covers all mutable profile fields.
// Creation is owned by the user module (RegisterTeacher).
type UpdateRequest struct {
	FirstName      *string          `json:"first_name,omitempty"`
	LastName       *string          `json:"last_name,omitempty"`
	Gender         *database.Gender `json:"gender,omitempty"`
	DOB            *time.Time       `json:"dob,omitempty"`
	Phone          *string          `json:"phone,omitempty"`
	Address        *string          `json:"address,omitempty"`
	Qualification  *string          `json:"qualification,omitempty"`
	Specialization *string          `json:"specialization,omitempty"`
	JoiningDate    *time.Time       `json:"joining_date,omitempty"`
	PhotoURL       *string          `json:"photo_url,omitempty"`
	IsActive       *bool            `json:"is_active,omitempty"`
}

type TeacherResponse struct {
	Teacher
}

func ToTeacherResponse(t *Teacher) *TeacherResponse {
	return &TeacherResponse{Teacher: *t}
}
