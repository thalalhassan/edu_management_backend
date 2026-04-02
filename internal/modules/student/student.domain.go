package student

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type Student struct {
	AdmissionNo string                       `json:"admission_no"`
	FirstName   string                       `json:"first_name"`
	LastName    string                       `json:"last_name"`
	DOB         time.Time                    `json:"dob"`
	Status      database.StudentStatus       `json:"status"`
	Enrollments []database.StudentEnrollment `json:"enrollments,omitempty"`
}
type CreateRequest struct {
	AdmissionNo string                 `json:"admission_no" binding:"required"`
	FirstName   string                 `json:"first_name" binding:"required"`
	LastName    string                 `json:"last_name" binding:"required"`
	DOB         time.Time              `json:"dob" binding:"required"`
	Status      database.StudentStatus `json:"status,omitempty"`
}

type UpdateRequest struct {
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	DOB       *time.Time `json:"dob,omitempty"`
	Status    *string    `json:"status,omitempty"`
}

type StudentResponse struct {
	Student
}

func ToStudentResponse(s *Student) *StudentResponse {
	return &StudentResponse{
		Student: *s,
	}
}
