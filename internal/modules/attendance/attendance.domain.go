package attendance

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type Attendance = database.Attendance
type TeacherAttendance = database.TeacherAttendance
type AttendanceStatus = database.AttendanceStatus

// ─── Student Attendance Requests ─────────────────────────────────────────────

// MarkStudentAttendanceRequest marks attendance for a single student enrollment.
type MarkStudentAttendanceRequest struct {
	StudentEnrollmentID string           `json:"student_enrollment_id" binding:"required,uuid"`
	Date                time.Time        `json:"date"                  binding:"required"`
	Status              AttendanceStatus `json:"status"                binding:"required"`
	Remark              *string          `json:"remark,omitempty"`
	RecordedByTeacherID *string          `json:"recorded_by_teacher_id,omitempty"`
}

// BulkMarkRequest marks attendance for all students in a class section on a given date.
type BulkMarkRequest struct {
	ClassSectionID      string                 `json:"class_section_id"       binding:"required,uuid"`
	Date                time.Time              `json:"date"                   binding:"required"`
	RecordedByTeacherID *string                `json:"recorded_by_teacher_id,omitempty"`
	Records             []StudentAttendanceRow `json:"records"                binding:"required,min=1,dive"`
}

type StudentAttendanceRow struct {
	StudentEnrollmentID string           `json:"student_enrollment_id" binding:"required,uuid"`
	Status              AttendanceStatus `json:"status"                binding:"required"`
	Remark              *string          `json:"remark,omitempty"`
}

type UpdateStudentAttendanceRequest struct {
	Status AttendanceStatus `json:"status"           binding:"required"`
	Remark *string          `json:"remark,omitempty"`
}

// ─── Teacher Attendance Requests ─────────────────────────────────────────────

type MarkTeacherAttendanceRequest struct {
	TeacherID string           `json:"teacher_id" binding:"required,uuid"`
	Date      time.Time        `json:"date"       binding:"required"`
	Status    AttendanceStatus `json:"status"     binding:"required"`
	Remark    *string          `json:"remark,omitempty"`
}

type BulkMarkTeacherRequest struct {
	Date    time.Time              `json:"date"    binding:"required"`
	Records []TeacherAttendanceRow `json:"records" binding:"required,min=1,dive"`
}

type TeacherAttendanceRow struct {
	TeacherID string           `json:"teacher_id" binding:"required,uuid"`
	Status    AttendanceStatus `json:"status"     binding:"required"`
	Remark    *string          `json:"remark,omitempty"`
}

type UpdateTeacherAttendanceRequest struct {
	Status AttendanceStatus `json:"status"           binding:"required"`
	Remark *string          `json:"remark,omitempty"`
}

// ─── Response shapes ──────────────────────────────────────────────────────────

type AttendanceResponse struct {
	ID                  string           `json:"id"`
	StudentEnrollmentID string           `json:"student_enrollment_id"`
	Date                time.Time        `json:"date"`
	Status              AttendanceStatus `json:"status"`
	Remark              *string          `json:"remark,omitempty"`
	RecordedByTeacherID *string          `json:"recorded_by_teacher_id,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type TeacherAttendanceResponse struct {
	ID        string           `json:"id"`
	TeacherID string           `json:"teacher_id"`
	Date      time.Time        `json:"date"`
	Status    AttendanceStatus `json:"status"`
	Remark    *string          `json:"remark,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// ClassAttendanceSummary gives daily class-level overview.
type ClassAttendanceSummary struct {
	Date           time.Time              `json:"date"`
	ClassSectionID string                 `json:"class_section_id"`
	TotalStudents  int                    `json:"total_students"`
	Present        int                    `json:"present"`
	Absent         int                    `json:"absent"`
	HalfDay        int                    `json:"half_day"`
	Late           int                    `json:"late"`
	OnLeave        int                    `json:"on_leave"`
	Records        []*AttendanceResponse  `json:"records"`
}

// ─── Filter params ────────────────────────────────────────────────────────────

type StudentFilterParams struct {
	ClassSectionID      *string    `form:"class_section_id"`
	StudentEnrollmentID *string    `form:"student_enrollment_id"`
	DateFrom            *time.Time `form:"date_from"`
	DateTo              *time.Time `form:"date_to"`
	Status              *string    `form:"status"`
}

type TeacherFilterParams struct {
	TeacherID *string    `form:"teacher_id"`
	DateFrom  *time.Time `form:"date_from"`
	DateTo    *time.Time `form:"date_to"`
	Status    *string    `form:"status"`
}

var allowedStudentSortFields = map[string]bool{
	"date":       true,
	"status":     true,
	"created_at": true,
}

var allowedTeacherSortFields = map[string]bool{
	"date":       true,
	"status":     true,
	"created_at": true,
}

// ─── Mappers ─────────────────────────────────────────────────────────────────

func ToAttendanceResponse(a *Attendance) *AttendanceResponse {
	return &AttendanceResponse{
		ID:                  a.ID,
		StudentEnrollmentID: a.StudentEnrollmentID,
		Date:                a.Date,
		Status:              a.Status,
		Remark:              a.Remark,
		RecordedByTeacherID: a.RecordedByTeacherID,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}

func ToTeacherAttendanceResponse(a *TeacherAttendance) *TeacherAttendanceResponse {
	return &TeacherAttendanceResponse{
		ID:        a.ID,
		TeacherID: a.TeacherID,
		Date:      a.Date,
		Status:    a.Status,
		Remark:    a.Remark,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
