package exam

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type Exam = database.Exam
type ExamSchedule = database.ExamSchedule
type ExamResult = database.ExamResult
type ExamResultStatus = database.ExamResultStatus

// ─── Exam Requests ─────────────────────────────────────────────────────────

type CreateExamRequest struct {
	AcademicYearID string    `json:"academic_year_id" binding:"required,uuid"`
	Name           string    `json:"name"             binding:"required"`
	Description    *string   `json:"description,omitempty"`
	ExamType       string    `json:"exam_type"        binding:"required"` // UNIT_TEST | MIDTERM | FINAL | INTERNAL
	StartDate      time.Time `json:"start_date"       binding:"required"`
	EndDate        time.Time `json:"end_date"         binding:"required"`
}

type UpdateExamRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	ExamType    *string    `json:"exam_type,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type PublishExamRequest struct {
	IsPublished bool `json:"is_published"`
}

// ─── ExamSchedule Requests ─────────────────────────────────────────────────

type CreateScheduleRequest struct {
	ClassSectionID string          `json:"class_section_id" binding:"required,uuid"`
	SubjectID      string          `json:"subject_id"       binding:"required,uuid"`
	ExamDate       time.Time       `json:"exam_date"        binding:"required"`
	StartTime      *time.Time      `json:"start_time,omitempty"`
	EndTime        *time.Time      `json:"end_time,omitempty"`
	MaxMarks       decimal.Decimal `json:"max_marks"        binding:"required"`
	PassingMarks   decimal.Decimal `json:"passing_marks"    binding:"required"`
	RoomNumber     *string         `json:"room_number,omitempty"`
}

type UpdateScheduleRequest struct {
	ExamDate     *time.Time       `json:"exam_date,omitempty"`
	StartTime    *time.Time       `json:"start_time,omitempty"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	MaxMarks     *decimal.Decimal `json:"max_marks,omitempty"`
	PassingMarks *decimal.Decimal `json:"passing_marks,omitempty"`
	RoomNumber   *string          `json:"room_number,omitempty"`
}

// ─── ExamResult Requests ────────────────────────────────────────────────────

type CreateResultRequest struct {
	StudentEnrollmentID string          `json:"student_enrollment_id" binding:"required,uuid"`
	MarksObtained       decimal.Decimal `json:"marks_obtained"        binding:"required"`
	Grade               *string         `json:"grade,omitempty"`
	Remarks             *string         `json:"remarks,omitempty"`
}

// BulkCreateResultRequest allows creating results for all students in a schedule at once.
type BulkCreateResultRequest struct {
	Results []CreateResultRequest `json:"results" binding:"required,min=1,dive"`
}

type UpdateResultRequest struct {
	MarksObtained *decimal.Decimal `json:"marks_obtained,omitempty"`
	Grade         *string          `json:"grade,omitempty"`
	Remarks       *string          `json:"remarks,omitempty"`
}

// ─── Response shapes ────────────────────────────────────────────────────────

type ExamResponse struct {
	ID             string     `json:"id"`
	AcademicYearID string     `json:"academic_year_id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	ExamType       string     `json:"exam_type"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	IsPublished    bool       `json:"is_published"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ExamScheduleResponse struct {
	ID             string          `json:"id"`
	ExamID         string          `json:"exam_id"`
	ClassSectionID string          `json:"class_section_id"`
	SubjectID      string          `json:"subject_id"`
	ExamDate       time.Time       `json:"exam_date"`
	StartTime      *time.Time      `json:"start_time,omitempty"`
	EndTime        *time.Time      `json:"end_time,omitempty"`
	MaxMarks       decimal.Decimal `json:"max_marks"`
	PassingMarks   decimal.Decimal `json:"passing_marks"`
	RoomNumber     *string         `json:"room_number,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ExamResultResponse struct {
	ID                  string           `json:"id"`
	ExamScheduleID      string           `json:"exam_schedule_id"`
	StudentEnrollmentID string           `json:"student_enrollment_id"`
	MarksObtained       decimal.Decimal  `json:"marks_obtained"`
	Grade               *string          `json:"grade,omitempty"`
	Status              ExamResultStatus `json:"status"`
	Remarks             *string          `json:"remarks,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

// ─── Filter params ───────────────────────────────────────────────────────────

type FilterParams struct {
	AcademicYearID *string `form:"academic_year_id"`
	ExamType       *string `form:"exam_type"`
	IsPublished    *bool   `form:"is_published"`
}

var allowedSortFields = map[string]bool{
	"name":       true,
	"start_date": true,
	"end_date":   true,
	"exam_type":  true,
	"created_at": true,
}

// ─── Mappers ────────────────────────────────────────────────────────────────

func ToExamResponse(e *Exam) *ExamResponse {
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

func ToScheduleResponse(s *ExamSchedule) *ExamScheduleResponse {
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
		RoomNumber:     s.RoomNumber,
		CreatedAt:      s.CreatedAt,
	}
}

func ToResultResponse(r *ExamResult) *ExamResultResponse {
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
