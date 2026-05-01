package class_section

import (
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

type ClassSection = database.ClassSection

type CreateRequest struct {
	AcademicYearID  uuid.UUID  `json:"academic_year_id"  binding:"required,uuid"`
	StandardID      uuid.UUID  `json:"standard_id"       binding:"required,uuid"`
	SectionName     string     `json:"section_name"      binding:"required"`
	ClassEmployeeID *uuid.UUID `json:"class_employee_id,omitempty" binding:"omitempty,uuid"`
	RoomID          *uuid.UUID `json:"room_id,omitempty"`
	MaxStrength     int        `json:"max_strength"      binding:"min=1"`
}

type UpdateRequest struct {
	SectionName     *string    `json:"section_name,omitempty"`
	ClassEmployeeID *uuid.UUID `json:"class_employee_id,omitempty" binding:"omitempty,uuid"`
	RoomID          *uuid.UUID `json:"room_id,omitempty"`
	MaxStrength     *int       `json:"max_strength,omitempty"     binding:"omitempty,min=1"`
}

// AssignEmployeeRequest is a dedicated request for assigning / removing the class employee.
type AssignEmployeeRequest struct {
	ClassEmployeeID *uuid.UUID `json:"class_employee_id" binding:"omitempty,uuid"` // nil = remove
}

// ClassSectionSummary is a lightweight response used in list views.
// Avoids loading full enrollments/assignments for every row.
type ClassSectionSummary struct {
	ID             uuid.UUID  `json:"id"`
	AcademicYearID uuid.UUID  `json:"academic_year_id"`
	AcademicYear   string     `json:"academic_year"`
	StandardID     uuid.UUID  `json:"standard_id"`
	Standard       string     `json:"standard"`
	Department     string     `json:"department"`
	SectionName    string     `json:"section_name"`
	ClassEmployee  *string    `json:"class_employee,omitempty"` // "FirstName LastName"
	RoomID         *uuid.UUID `json:"room_id,omitempty"`
	MaxStrength    int        `json:"max_strength"`
	Enrolled       int64      `json:"enrolled"` // current enrolled count
}

type ClassSectionResponse struct {
	ClassSection
}

func ToClassSectionResponse(cs *ClassSection) *ClassSectionResponse {
	return &ClassSectionResponse{ClassSection: *cs}
}

func ToClassSectionSummary(cs *ClassSection, enrolled int64) *ClassSectionSummary {
	s := &ClassSectionSummary{
		ID:             cs.ID,
		AcademicYearID: cs.AcademicYearID,
		AcademicYear:   cs.AcademicYear.Name,
		StandardID:     cs.StandardID,
		Standard:       cs.Standard.Name,
		Department:     cs.Standard.Department.Name,
		SectionName:    cs.SectionName,
		RoomID:         cs.RoomID,
		MaxStrength:    cs.MaxStrength,
		Enrolled:       enrolled,
	}
	if cs.ClassEmployee != nil {
		full := cs.ClassEmployee.FirstName + " " + cs.ClassEmployee.LastName
		s.ClassEmployee = &full
	}
	return s
}
