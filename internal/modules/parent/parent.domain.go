package parent

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type Parent = database.Parent

func ToParentResponse(p *Parent) *ParentResponse {
	children := make([]StudentBrief, len(p.Children))
	for i, s := range p.Children {
		children[i] = StudentBrief{
			ID:        s.ID,
			FirstName: s.Student.FirstName,
			LastName:  s.Student.LastName,
		}
	}

	return &ParentResponse{
		ID:        p.ID,
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Phone:     p.Phone,
		Address:   p.Address,
		Children:  children,
		CreatedAt: p.CreatedAt,
	}
}

// ==========================================
// FILTER
// ==========================================
type FilterParams struct {
	Search   *string `form:"search"`    // name / email / phone
	IsActive *bool   `form:"is_active"` // optional
}

// ==========================================
// REQUEST
// ==========================================

// NOTE: Creation happens via user module (RegisterParent)

type UpdateRequest struct {
	FirstName *string `json:"first_name" binding:"required"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     string  `json:"phone" binding:"required"`
	Address   *string `json:"address,omitempty"`
}

// Link / unlink student
type LinkStudentRequest struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
	IsPrimary bool   `json:"is_primary"`
}

type UnlinkStudentRequest struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
}

// ==========================================
// RESPONSE
// ==========================================

type StudentBrief struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ParentResponse struct {
	ID        string         `json:"id"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Phone     string         `json:"phone"`
	Address   *string        `json:"address,omitempty"`
	Children  []StudentBrief `json:"children"`
	CreatedAt time.Time      `json:"created_at"`
}
