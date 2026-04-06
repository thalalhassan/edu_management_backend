package staff

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type Staff = database.Staff

// ==========================================
// FILTER
// ==========================================
type FilterParams struct {
	Search   *string `form:"search"`    // name / employee_id
	IsActive *bool   `form:"is_active"` // optional
}

// ==========================================
// REQUEST
// ==========================================

// NOTE: Creation via user module (RegisterStaff)

type UpdateRequest struct {
	FirstName  *string `json:"first_name" binding:"required"`
	LastName   *string `json:"last_name,omitempty"`
	Department *string `json:"department,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

// ==========================================
// RESPONSE
// ==========================================

type Response struct {
	ID         string    `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	EmployeeID string    `json:"employee_id"`
	Department *string   `json:"department,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type StaffResponse struct {
	Staff
}

func ToStaffResponse(s *Staff) *StaffResponse {
	return &StaffResponse{
		Staff: *s,
	}
}
