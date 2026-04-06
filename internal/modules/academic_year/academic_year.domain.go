package academic_year

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type AcademicYear = database.AcademicYear

type CreateRequest struct {
	Name      string    `json:"name"       binding:"required"`
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date"   binding:"required"`
}

type UpdateRequest struct {
	Name      *string    `json:"name,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// AcademicYearResponse is the standard response shape.
// ClassSections, Exams, FeeStructures are intentionally omitted —
// those are fetched through their own modules scoped to this AY.
type AcademicYearResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func ToAcademicYearResponse(a *AcademicYear) *AcademicYearResponse {
	return &AcademicYearResponse{
		ID:        a.ID,
		Name:      a.Name,
		StartDate: a.StartDate,
		EndDate:   a.EndDate,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
	}
}

type FilterParams struct {
	Search *string `form:"search"`
}

var allowedSortFields = map[string]bool{
	"name":       true,
	"start_date": true,
	"end_date":   true,
	"created_at": true,
}
