package exam

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
)

type CreateExamRequest struct {
	AcademicYearID uuid.UUID         `json:"academic_year_id" binding:"required,uuid"`
	Name           string            `json:"name"             binding:"required,min=1,max=255"`
	Description    *string           `json:"description,omitempty" binding:"omitempty,max=1000"`
	ExamType       database.ExamType `json:"exam_type"        binding:"required,oneof=UNIT_TEST MID_TERM FINAL MOCK INTERNAL PRACTICAL"`
	StartDate      time.Time         `json:"start_date"       binding:"required"`
	EndDate        time.Time         `json:"end_date"         binding:"required"`
}

func (r *CreateExamRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = normalizeStringPtr(r.Description)
}

type UpdateExamRequest struct {
	Name        *string            `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string            `json:"description,omitempty" binding:"omitempty,max=1000"`
	ExamType    *database.ExamType `json:"exam_type,omitempty" binding:"omitempty,oneof=UNIT_TEST MID_TERM FINAL MOCK INTERNAL PRACTICAL"`
	StartDate   *time.Time         `json:"start_date,omitempty"`
	EndDate     *time.Time         `json:"end_date,omitempty"`
}

func (r *UpdateExamRequest) Normalize() {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
	r.Description = normalizeStringPtr(r.Description)
}

type PublishExamRequest struct {
	// Use *bool so PATCH can distinguish between "missing" and explicit false.
	IsPublished *bool `json:"is_published" binding:"required"`
}

func (r *PublishExamRequest) Normalize() {}

type CreateScheduleRequest struct {
	ClassSectionID uuid.UUID       `json:"class_section_id" binding:"required,uuid"`
	SubjectID      uuid.UUID       `json:"subject_id"       binding:"required,uuid"`
	ExamDate       time.Time       `json:"exam_date"        binding:"required"`
	StartTime      *time.Time      `json:"start_time,omitempty"`
	EndTime        *time.Time      `json:"end_time,omitempty"`
	MaxMarks       decimal.Decimal `json:"max_marks"        binding:"required,gt=0"`
	PassingMarks   decimal.Decimal `json:"passing_marks"    binding:"required,gt=0"`
	RoomID         *uuid.UUID      `json:"room_id,omitempty" binding:"omitempty,uuid"`
}

type UpdateScheduleRequest struct {
	ExamDate     *time.Time       `json:"exam_date,omitempty"`
	StartTime    *time.Time       `json:"start_time,omitempty"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	MaxMarks     *decimal.Decimal `json:"max_marks,omitempty" binding:"omitempty,gt=0"`
	PassingMarks *decimal.Decimal `json:"passing_marks,omitempty" binding:"omitempty,gt=0"`
	RoomID       *uuid.UUID       `json:"room_id,omitempty" binding:"omitempty,uuid"`
}

type CreateResultRequest struct {
	StudentEnrollmentID uuid.UUID        `json:"student_enrollment_id" binding:"required,uuid"`
	MarksObtained       *decimal.Decimal `json:"marks_obtained"        binding:"required,gt=0"`
	Grade               *string          `json:"grade,omitempty" binding:"omitempty,max=10"`
	Remarks             *string          `json:"remarks,omitempty" binding:"omitempty,max=500"`
}

func (r *CreateResultRequest) Normalize() {
	r.Grade = normalizeStringPtr(r.Grade)
	r.Remarks = normalizeStringPtr(r.Remarks)
}

type BulkCreateResultRequest struct {
	Results []CreateResultRequest `json:"results" binding:"required,min=1,max=500,dive"`
}

func (r *BulkCreateResultRequest) Normalize() {
	for i := range r.Results {
		r.Results[i].Normalize()
	}
}

type UpdateResultRequest struct {
	MarksObtained *decimal.Decimal `json:"marks_obtained,omitempty" binding:"omitempty,gt=0"`
	Grade         *string          `json:"grade,omitempty" binding:"omitempty,max=10"`
	Remarks       *string          `json:"remarks,omitempty" binding:"omitempty,max=500"`
}

func (r *UpdateResultRequest) Normalize() {
	r.Grade = normalizeStringPtr(r.Grade)
	r.Remarks = normalizeStringPtr(r.Remarks)
}

type ExamResponse struct {
	ID             uuid.UUID         `json:"id"`
	AcademicYearID uuid.UUID         `json:"academic_year_id"`
	Name           string            `json:"name"`
	Description    *string           `json:"description,omitempty"`
	ExamType       database.ExamType `json:"exam_type"`
	StartDate      time.Time         `json:"start_date"`
	EndDate        time.Time         `json:"end_date"`
	IsPublished    bool              `json:"is_published"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type ExamScheduleResponse struct {
	ID             uuid.UUID       `json:"id"`
	ExamID         uuid.UUID       `json:"exam_id"`
	ClassSectionID uuid.UUID       `json:"class_section_id"`
	SubjectID      uuid.UUID       `json:"subject_id"`
	ExamDate       time.Time       `json:"exam_date"`
	StartTime      *time.Time      `json:"start_time,omitempty"`
	EndTime        *time.Time      `json:"end_time,omitempty"`
	MaxMarks       decimal.Decimal `json:"max_marks"`
	PassingMarks   decimal.Decimal `json:"passing_marks"`
	RoomID         *uuid.UUID      `json:"room_id,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ExamResultResponse struct {
	ID                  uuid.UUID                 `json:"id"`
	ExamScheduleID      uuid.UUID                 `json:"exam_schedule_id"`
	StudentEnrollmentID uuid.UUID                 `json:"student_enrollment_id"`
	MarksObtained       *decimal.Decimal          `json:"marks_obtained"`
	Grade               *string                   `json:"grade,omitempty"`
	Status              database.ExamResultStatus `json:"status"`
	Remarks             *string                   `json:"remarks,omitempty"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
}

type FilterParams struct {
	AcademicYearID *uuid.UUID         `form:"academic_year_id" binding:"omitempty,uuid"`
	ExamType       *database.ExamType `form:"exam_type" binding:"omitempty,oneof=UNIT_TEST MID_TERM FINAL MOCK INTERNAL PRACTICAL"`
	IsPublished    *bool              `form:"is_published"`
}

var allowedSortFields = map[string]bool{
	"name":       true,
	"start_date": true,
	"end_date":   true,
	"exam_type":  true,
	"created_at": true,
}

type PaginatedResponse[T any] struct {
	Data []T             `json:"data"`
	Meta pagination.Meta `json:"meta"`
}

func ToExamResponse(e *database.Exam) *ExamResponse {
	return &ExamResponse{
		ID:             e.ID,
		AcademicYearID: e.AcademicYearID,
		Name:           e.Name,
		Description:    e.Description,
		ExamType:       e.ExamType,
		StartDate:      e.StartDate,
		EndDate:        e.EndDate,
		IsPublished:    e.IsPublished,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func ToScheduleResponse(s *database.ExamSchedule) *ExamScheduleResponse {
	return &ExamScheduleResponse{
		ID:             s.ID,
		ExamID:         s.ExamID,
		ClassSectionID: s.ClassSectionID,
		SubjectID:      s.SubjectID,
		ExamDate:       s.ExamDate,
		StartTime:      s.StartTime,
		EndTime:        s.EndTime,
		MaxMarks:       s.MaxMarks,
		PassingMarks:   s.PassingMarks,
		RoomID:         s.RoomID,
		CreatedAt:      s.CreatedAt,
	}
}

func ToResultResponse(r *database.ExamResult) *ExamResultResponse {
	return &ExamResultResponse{
		ID:                  r.ID,
		ExamScheduleID:      r.ExamScheduleID,
		StudentEnrollmentID: r.StudentEnrollmentID,
		MarksObtained:       r.MarksObtained,
		Grade:               r.Grade,
		Status:              r.Status,
		Remarks:             r.Remarks,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}
