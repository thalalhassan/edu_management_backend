package attendance

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/apperrors"
	"github.com/thalalhassan/edu_management/internal/database"
)

const MaxAttendanceRemarkLength = 1000

var validAttendanceStatuses = map[database.AttendanceStatus]bool{
	database.AttendanceStatusPresent: true,
	database.AttendanceStatusAbsent:  true,
	database.AttendanceStatusHalfDay: true,
	database.AttendanceStatusLate:    true,
	database.AttendanceStatusLeave:   true,
}

func isValidAttendanceStatus(status database.AttendanceStatus) bool {
	return validAttendanceStatuses[status]
}

func isValidAttendanceStatusString(value string) bool {
	if value == "" {
		return false
	}
	for status := range validAttendanceStatuses {
		if string(status) == strings.ToUpper(value) {
			return true
		}
	}
	return false
}

// ─── Request DTOs ───────────────────────────────────────────────────────────

type MarkStudentAttendanceRequest struct {
	StudentEnrollmentID uuid.UUID        `json:"student_enrollment_id" binding:"required,uuid"`
	Date                time.Time        `json:"date"                  binding:"required"`
	Status              AttendanceStatus `json:"status"                binding:"required"`
	Remark              *string          `json:"remark,omitempty"`
	RecordedByID        *uuid.UUID       `json:"recorded_by_id,omitempty" binding:"omitempty,uuid"`
}

func (r MarkStudentAttendanceRequest) Validate() error {
	ve := apperrors.NewValidationError()
	if r.StudentEnrollmentID == uuid.Nil {
		ve.Add("student_enrollment_id", "student_enrollment_id is required")
	}
	if r.Date.IsZero() {
		ve.Add("date", "date is required")
	} else if NormalizeDate(r.Date).After(NormalizeDate(time.Now())) {
		ve.Add("date", "attendance date cannot be in the future")
	}
	if !isValidAttendanceStatus(r.Status) {
		ve.Add("status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
	}
	if r.Remark != nil && len(strings.TrimSpace(*r.Remark)) > MaxAttendanceRemarkLength {
		ve.Add("remark", "remark cannot exceed 1000 characters")
	}
	if r.RecordedByID != nil && *r.RecordedByID == uuid.Nil {
		ve.Add("recorded_by_id", "recorded_by_id must be a valid UUID")
	}
	return ve.OrNil()
}

type BulkMarkRequest struct {
	ClassSectionID uuid.UUID              `json:"class_section_id" binding:"required,uuid"`
	Date           time.Time              `json:"date"             binding:"required"`
	RecordedByID   *uuid.UUID             `json:"recorded_by_id,omitempty" binding:"omitempty,uuid"`
	Records        []StudentAttendanceRow `json:"records"         binding:"required,min=1,dive"`
}

func (r BulkMarkRequest) Validate() error {
	ve := apperrors.NewValidationError()
	if r.ClassSectionID == uuid.Nil {
		ve.Add("class_section_id", "class_section_id is required")
	}
	if r.Date.IsZero() {
		ve.Add("date", "date is required")
	} else if NormalizeDate(r.Date).After(NormalizeDate(time.Now())) {
		ve.Add("date", "attendance date cannot be in the future")
	}
	if len(r.Records) == 0 {
		ve.Add("records", "records must contain at least one attendance row")
	}

	seen := map[uuid.UUID]struct{}{}
	for _, row := range r.Records {
		if row.StudentEnrollmentID == uuid.Nil {
			ve.Add("records[]student_enrollment_id", "student_enrollment_id is required")
		}
		if !isValidAttendanceStatus(row.Status) {
			ve.Add("records[]status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
		}
		if row.Remark != nil && len(strings.TrimSpace(*row.Remark)) > MaxAttendanceRemarkLength {
			ve.Add("records[]remark", "remark cannot exceed 1000 characters")
		}
		if _, exists := seen[row.StudentEnrollmentID]; exists {
			ve.Add("records", "records contain duplicate student_enrollment_id values")
		}
		seen[row.StudentEnrollmentID] = struct{}{}
	}
	return ve.OrNil()
}

type StudentAttendanceRow struct {
	StudentEnrollmentID uuid.UUID        `json:"student_enrollment_id" binding:"required,uuid"`
	Status              AttendanceStatus `json:"status"                binding:"required"`
	Remark              *string          `json:"remark,omitempty"`
}

func (r StudentAttendanceRow) Validate(index int) []apperrors.FieldError {
	var errors []apperrors.FieldError
	fieldPrefix := fmt.Sprintf("records[%d]", index)
	if r.StudentEnrollmentID == uuid.Nil {
		errors = append(errors, apperrors.FieldError{Field: fieldPrefix + ".student_enrollment_id", Message: "student_enrollment_id is required"})
	}
	if !isValidAttendanceStatus(r.Status) {
		errors = append(errors, apperrors.FieldError{Field: fieldPrefix + ".status", Message: "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE"})
	}
	if r.Remark != nil && len(strings.TrimSpace(*r.Remark)) > MaxAttendanceRemarkLength {
		errors = append(errors, apperrors.FieldError{Field: fieldPrefix + ".remark", Message: "remark cannot exceed 1000 characters"})
	}
	return errors
}

type UpdateStudentAttendanceRequest struct {
	Status AttendanceStatus `json:"status"           binding:"required"`
	Remark *string          `json:"remark,omitempty"`
}

func (r UpdateStudentAttendanceRequest) Validate() error {
	ve := apperrors.NewValidationError()
	if !isValidAttendanceStatus(r.Status) {
		ve.Add("status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
	}
	if r.Remark != nil && len(strings.TrimSpace(*r.Remark)) > MaxAttendanceRemarkLength {
		ve.Add("remark", "remark cannot exceed 1000 characters")
	}
	return ve.OrNil()
}

// ─── Employee Request DTOs ───────────────────────────────────────────────────

type MarkEmployeeAttendanceRequest struct {
	EmployeeID uuid.UUID        `json:"employee_id" binding:"required,uuid"`
	Date       time.Time        `json:"date"        binding:"required"`
	Status     AttendanceStatus `json:"status"      binding:"required"`
	Remark     *string          `json:"remark,omitempty"`
}

func (r MarkEmployeeAttendanceRequest) Validate() error {
	ve := apperrors.NewValidationError()
	if r.EmployeeID == uuid.Nil {
		ve.Add("employee_id", "employee_id is required")
	}
	if r.Date.IsZero() {
		ve.Add("date", "date is required")
	} else if NormalizeDate(r.Date).After(NormalizeDate(time.Now())) {
		ve.Add("date", "attendance date cannot be in the future")
	}
	if !isValidAttendanceStatus(r.Status) {
		ve.Add("status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
	}
	if r.Remark != nil && len(strings.TrimSpace(*r.Remark)) > MaxAttendanceRemarkLength {
		ve.Add("remark", "remark cannot exceed 1000 characters")
	}
	return ve.OrNil()
}

type BulkMarkEmployeeRequest struct {
	Date    time.Time               `json:"date"    binding:"required"`
	Records []EmployeeAttendanceRow `json:"records" binding:"required,min=1,dive"`
}

func (r BulkMarkEmployeeRequest) Validate() error {
	ve := apperrors.NewValidationError()
	if r.Date.IsZero() {
		ve.Add("date", "date is required")
	} else if NormalizeDate(r.Date).After(NormalizeDate(time.Now())) {
		ve.Add("date", "attendance date cannot be in the future")
	}
	if len(r.Records) == 0 {
		ve.Add("records", "records must contain at least one attendance row")
	}

	seen := map[uuid.UUID]struct{}{}
	for _, row := range r.Records {
		if row.EmployeeID == uuid.Nil {
			ve.Add("records[]employee_id", "employee_id is required")
		}
		if !isValidAttendanceStatus(row.Status) {
			ve.Add("records[]status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
		}
		if row.Remark != nil && len(strings.TrimSpace(*row.Remark)) > MaxAttendanceRemarkLength {
			ve.Add("records[]remark", "remark cannot exceed 1000 characters")
		}
		if _, exists := seen[row.EmployeeID]; exists {
			ve.Add("records", "records contain duplicate employee_id values")
		}
		seen[row.EmployeeID] = struct{}{}
	}
	return ve.OrNil()
}

type EmployeeAttendanceRow struct {
	EmployeeID uuid.UUID        `json:"employee_id" binding:"required,uuid"`
	Status     AttendanceStatus `json:"status"      binding:"required"`
	Remark     *string          `json:"remark,omitempty"`
}

func (r EmployeeAttendanceRow) Validate(index int) []apperrors.FieldError {
	var errors []apperrors.FieldError
	fieldPrefix := fmt.Sprintf("records[%d]", index)
	if r.EmployeeID == uuid.Nil {
		errors = append(errors, apperrors.FieldError{Field: fieldPrefix + ".employee_id", Message: "employee_id is required"})
	}
	if !isValidAttendanceStatus(r.Status) {
		errors = append(errors, apperrors.FieldError{Field: fieldPrefix + ".status", Message: "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE"})
	}
	if r.Remark != nil && len(strings.TrimSpace(*r.Remark)) > MaxAttendanceRemarkLength {
		errors = append(errors, apperrors.FieldError{Field: fieldPrefix + ".remark", Message: "remark cannot exceed 1000 characters"})
	}
	return errors
}

type UpdateEmployeeAttendanceRequest struct {
	Status AttendanceStatus `json:"status"           binding:"required"`
	Remark *string          `json:"remark,omitempty"`
}

func (r UpdateEmployeeAttendanceRequest) Validate() error {
	ve := apperrors.NewValidationError()
	if !isValidAttendanceStatus(r.Status) {
		ve.Add("status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
	}
	if r.Remark != nil && len(strings.TrimSpace(*r.Remark)) > MaxAttendanceRemarkLength {
		ve.Add("remark", "remark cannot exceed 1000 characters")
	}
	return ve.OrNil()
}

// ─── Response DTOs ──────────────────────────────────────────────────────────

type AttendanceResponse struct {
	ID                  uuid.UUID        `json:"id"`
	StudentEnrollmentID uuid.UUID        `json:"student_enrollment_id"`
	Date                time.Time        `json:"date"`
	Status              AttendanceStatus `json:"status"`
	Remark              *string          `json:"remark,omitempty"`
	RecordedByID        *uuid.UUID       `json:"recorded_by_id,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type EmployeeAttendanceResponse struct {
	ID         uuid.UUID        `json:"id"`
	EmployeeID uuid.UUID        `json:"employee_id"`
	Date       time.Time        `json:"date"`
	Status     AttendanceStatus `json:"status"`
	Remark     *string          `json:"remark,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type ClassAttendanceSummary struct {
	Date           time.Time             `json:"date"`
	ClassSectionID uuid.UUID             `json:"class_section_id"`
	TotalStudents  int                   `json:"total_students"`
	Present        int                   `json:"present"`
	Absent         int                   `json:"absent"`
	HalfDay        int                   `json:"half_day"`
	Late           int                   `json:"late"`
	OnLeave        int                   `json:"on_leave"`
	Records        []*AttendanceResponse `json:"records"`
}

// ─── Filter DTOs ────────────────────────────────────────────────────────────

type StudentFilterParams struct {
	ClassSectionID      *uuid.UUID `form:"class_section_id"`
	StudentEnrollmentID *uuid.UUID `form:"student_enrollment_id"`
	DateFrom            *string    `form:"date_from"`
	DateTo              *string    `form:"date_to"`
	Status              *string    `form:"status"`
}

func (f StudentFilterParams) Validate() error {
	ve := apperrors.NewValidationError()
	if f.Status != nil && !isValidAttendanceStatusString(*f.Status) {
		ve.Add("status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
	}
	if f.DateFrom != nil && *f.DateFrom != "" {
		if _, err := time.Parse("2006-01-02", *f.DateFrom); err != nil {
			ve.Add("date_from", "date_from must be YYYY-MM-DD")
		}
	}
	if f.DateTo != nil && *f.DateTo != "" {
		if _, err := time.Parse("2006-01-02", *f.DateTo); err != nil {
			ve.Add("date_to", "date_to must be YYYY-MM-DD")
		}
	}
	return ve.OrNil()
}

type EmployeeFilterParams struct {
	EmployeeID *uuid.UUID `form:"employee_id"`
	DateFrom   *string    `form:"date_from"`
	DateTo     *string    `form:"date_to"`
	Status     *string    `form:"status"`
}

func (f EmployeeFilterParams) Validate() error {
	ve := apperrors.NewValidationError()
	if f.Status != nil && !isValidAttendanceStatusString(*f.Status) {
		ve.Add("status", "status must be one of PRESENT, ABSENT, HALF_DAY, LATE, LEAVE")
	}
	if f.DateFrom != nil && *f.DateFrom != "" {
		if _, err := time.Parse("2006-01-02", *f.DateFrom); err != nil {
			ve.Add("date_from", "date_from must be YYYY-MM-DD")
		}
	}
	if f.DateTo != nil && *f.DateTo != "" {
		if _, err := time.Parse("2006-01-02", *f.DateTo); err != nil {
			ve.Add("date_to", "date_to must be YYYY-MM-DD")
		}
	}
	return ve.OrNil()
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
