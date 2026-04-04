package student

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type Student = database.Student

type CreateRequest struct {
	AdmissionNo   string                 `json:"admission_no" binding:"required"`
	FirstName     string                 `json:"first_name" binding:"required"`
	LastName      string                 `json:"last_name" binding:"required"`
	DOB           time.Time              `json:"dob" binding:"required"`
	Status        database.StudentStatus `json:"status,omitempty"`
	Phone         *string                `json:"phone,omitempty"`
	Address       *string                `json:"address,omitempty"`
	BloodGroup    *string                `json:"blood_group,omitempty"`
	PhotoURL      *string                `json:"photo_url,omitempty"`
	AdmissionDate *time.Time             `json:"admission_date,omitempty"`
}

type UpdateRequest struct {
	FirstName     *string    `json:"first_name,omitempty"`
	LastName      *string    `json:"last_name,omitempty"`
	DOB           *time.Time `json:"dob,omitempty"`
	Status        *string    `json:"status,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Address       *string    `json:"address,omitempty"`
	BloodGroup    *string    `json:"blood_group,omitempty"`
	PhotoURL      *string    `json:"photo_url,omitempty"`
	AdmissionDate *time.Time `json:"admission_date,omitempty"`
}

type StudentResponse struct {
	Student
}

func ToStudentResponse(s *Student) *StudentResponse {
	return &StudentResponse{
		Student: *s,
	}
}
