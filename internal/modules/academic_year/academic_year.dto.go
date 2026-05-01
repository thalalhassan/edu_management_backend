package academic_year

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Name      string    `json:"name"      binding:"required,min=1,max=100" validate:"required,min=1,max=100,alphanumspace"`
	StartDate time.Time `json:"start_date" binding:"required" validate:"required,future"`
	EndDate   time.Time `json:"end_date"   binding:"required" validate:"required,gtfield=StartDate"`
}

type UpdateRequest struct {
	Name      *string    `json:"name,omitempty"      validate:"omitempty,min=1,max=100,alphanumspace"`
	StartDate *time.Time `json:"start_date,omitempty" validate:"omitempty,future"`
	EndDate   *time.Time `json:"end_date,omitempty"   validate:"omitempty"`
}

type AcademicYearResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type FilterParams struct {
	Search *string `form:"search" validate:"omitempty,min=1,max=50,alphanum"`
}
