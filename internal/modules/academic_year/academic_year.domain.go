package academic_year

import (
	"time"

	"github.com/google/uuid"
)

type AcademicYear struct {
	ID        uuid.UUID
	Name      string
	StartDate time.Time
	EndDate   time.Time
	IsActive  bool
	CreatedAt time.Time
}

func NewAcademicYear(name string, start, end time.Time) (*AcademicYear, error) {
	if !end.After(start) {
		return nil, NewValidationError("end_date must be after start_date")
	}
	if len(name) < 1 || len(name) > 100 {
		return nil, NewValidationError("name must be 1-100 characters")
	}
	return &AcademicYear{
		Name:      name,
		StartDate: start,
		EndDate:   end,
		IsActive:  false,
	}, nil
}

func (a *AcademicYear) Update(updates UpdateRequest) error {
	if updates.Name != nil {
		if len(*updates.Name) < 1 || len(*updates.Name) > 100 {
			return NewValidationError("name must be 1-100 characters")
		}
		a.Name = *updates.Name
	}
	if updates.StartDate != nil {
		a.StartDate = *updates.StartDate
	}
	if updates.EndDate != nil {
		a.EndDate = *updates.EndDate
	}
	if !a.EndDate.After(a.StartDate) {
		return NewValidationError("end_date must be after start_date")
	}
	return nil
}

func (a *AcademicYear) Activate() {
	a.IsActive = true
}

func (a *AcademicYear) Deactivate() {
	a.IsActive = false
}

func (AcademicYear) TableName() string { return "academic_year" }
