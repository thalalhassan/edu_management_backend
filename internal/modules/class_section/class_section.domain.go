package class_section

import "github.com/thalalhassan/edu_management/internal/database"

type ClassSection = database.ClassSection

type CreateRequest struct {
	AcademicYearID string  `json:"academic_year_id" binding:"required,uuid"`
	StandardID     string  `json:"standard_id"      binding:"required,uuid"`
	SectionName    string  `json:"section_name"     binding:"required"`
	ClassTeacherID *string `json:"class_teacher_id,omitempty" binding:"omitempty,uuid"`
	RoomNumber     *string `json:"room_number,omitempty"`
	MaxStrength    int     `json:"max_strength"     binding:"min=1"`
}

type UpdateRequest struct {
	SectionName    *string `json:"section_name,omitempty"`
	ClassTeacherID *string `json:"class_teacher_id,omitempty" binding:"omitempty,uuid"`
	RoomNumber     *string `json:"room_number,omitempty"`
	MaxStrength    *int    `json:"max_strength,omitempty"     binding:"omitempty,min=1"`
}

// AssignTeacherRequest is a dedicated request for assigning / removing the class teacher.
type AssignTeacherRequest struct {
	ClassTeacherID *string `json:"class_teacher_id" binding:"omitempty,uuid"` // nil = remove
}

// ClassSectionSummary is a lightweight response used in list views.
// Avoids loading full enrollments/assignments for every row.
type ClassSectionSummary struct {
	ID             string  `json:"id"`
	AcademicYearID string  `json:"academic_year_id"`
	AcademicYear   string  `json:"academic_year"`
	StandardID     string  `json:"standard_id"`
	Standard       string  `json:"standard"`
	Department     string  `json:"department"`
	SectionName    string  `json:"section_name"`
	ClassTeacher   *string `json:"class_teacher,omitempty"` // "FirstName LastName"
	RoomNumber     *string `json:"room_number,omitempty"`
	MaxStrength    int     `json:"max_strength"`
	Enrolled       int64   `json:"enrolled"` // current enrolled count
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
		RoomNumber:     cs.RoomNumber,
		MaxStrength:    cs.MaxStrength,
		Enrolled:       enrolled,
	}
	if cs.ClassTeacher != nil {
		full := cs.ClassTeacher.FirstName + " " + cs.ClassTeacher.LastName
		s.ClassTeacher = &full
	}
	return s
}
