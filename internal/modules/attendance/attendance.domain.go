package attendance

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type Attendance = database.Attendance
type EmployeeAttendance = database.EmployeeAttendance
type AttendanceStatus = database.AttendanceStatus

// ─── Student Attendance Requests ─────────────────────────────────────────────

// MarkStudentAttendanceRequest marks attendance for a single student enrollment.
type MarkStudentAttendanceRequest struct {
	StudentEnrollmentID string           `json:"student_enrollment_id" binding:"required,uuid"`
	Date                time.Time        `json:"date"                  binding:"required"`
	Status              AttendanceStatus `json:"status"                binding:"required"`
	Remark              *string          `json:"remark,omitempty"`
	RecordedByID        *string          `json:"recorded_by_id,omitempty"`
}

// BulkMarkRequest marks attendance for all students in a class section on a given date.
type BulkMarkRequest struct {
	ClassSectionID string                 `json:"class_section_id" binding:"required,uuid"`
	Date           time.Time              `json:"date"             binding:"required"`
	RecordedByID   *string                `json:"recorded_by_id,omitempty"`
	Records        []StudentAttendanceRow `json:"records"         binding:"required,min=1,dive"`
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

type MarkEmployeeAttendanceRequest struct {
	EmployeeID string           `json:"employee_id" binding:"required,uuid"`
	Date       time.Time        `json:"date"        binding:"required"`
	Status     AttendanceStatus `json:"status"      binding:"required"`
	Remark     *string          `json:"remark,omitempty"`
}

type BulkMarkEmployeeRequest struct {
	Date    time.Time               `json:"date"    binding:"required"`
	Records []EmployeeAttendanceRow `json:"records" binding:"required,min=1,dive"`
}

type EmployeeAttendanceRow struct {
	EmployeeID string           `json:"employee_id" binding:"required,uuid"`
	Status     AttendanceStatus `json:"status"      binding:"required"`
	Remark     *string          `json:"remark,omitempty"`
}

type UpdateEmployeeAttendanceRequest struct {
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
	RecordedByID        *string          `json:"recorded_by_id,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type EmployeeAttendanceResponse struct {
	ID         string           `json:"id"`
	EmployeeID string           `json:"employee_id"`
	Date       time.Time        `json:"date"`
	Status     AttendanceStatus `json:"status"`
	Remark     *string          `json:"remark,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// ClassAttendanceSummary gives daily class-level overview.
type ClassAttendanceSummary struct {
	Date           time.Time             `json:"date"`
	ClassSectionID string                `json:"class_section_id"`
	TotalStudents  int                   `json:"total_students"`
	Present        int                   `json:"present"`
	Absent         int                   `json:"absent"`
	HalfDay        int                   `json:"half_day"`
	Late           int                   `json:"late"`
	OnLeave        int                   `json:"on_leave"`
	Records        []*AttendanceResponse `json:"records"`
}

// ─── Filter params ────────────────────────────────────────────────────────────

type StudentFilterParams struct {
	ClassSectionID      *string    `form:"class_section_id"`
	StudentEnrollmentID *string    `form:"student_enrollment_id"`
	DateFrom            *time.Time `form:"date_from"`
	DateTo              *time.Time `form:"date_to"`
	Status              *string    `form:"status"`
}

type EmployeeFilterParams struct {
	EmployeeID *string    `form:"employee_id"`
	DateFrom   *time.Time `form:"date_from"`
	DateTo     *time.Time `form:"date_to"`
	Status     *string    `form:"status"`
}

var allowedStudentSortFields = map[string]bool{
	"date":       true,
	"status":     true,
	"created_at": true,
}

var allowedEmployeeSortFields = map[string]bool{
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
		RecordedByID:        a.RecordedByID,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}

func ToEmployeeAttendanceResponse(a *EmployeeAttendance) *EmployeeAttendanceResponse {
	return &EmployeeAttendanceResponse{
		ID:         a.ID,
		EmployeeID: a.EmployeeID,
		Date:       a.Date,
		Status:     a.Status,
		Remark:     a.Remark,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
