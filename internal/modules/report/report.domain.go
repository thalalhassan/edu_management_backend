package report

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ──────────────────────────────────────────────────────────────
// REPORT CARD
// ──────────────────────────────────────────────────────────────

// ReportCardRequest scopes the report card to one student + one exam.
// Passing exam_id = "" returns results across all exams in the AY.
type ReportCardRequest struct {
	StudentEnrollmentID uuid.UUID `form:"student_enrollment_id" binding:"required,uuid"`
	ExamID              uuid.UUID `form:"exam_id"               binding:"omitempty,uuid"`
}

type ReportCard struct {
	StudentEnrollmentID uuid.UUID           `json:"student_enrollment_id"`
	AdmissionNo         string              `json:"admission_no"`
	StudentName         string              `json:"student_name"`
	ClassSection        string              `json:"class_section"` // "Grade 6 - A"
	AcademicYear        string              `json:"academic_year"`
	Exams               []ExamReportSection `json:"exams"`
	OverallPercentage   decimal.Decimal     `json:"overall_percentage"`
	OverallGrade        string              `json:"overall_grade"`
}

type ExamReportSection struct {
	ExamID   uuid.UUID       `json:"exam_id"`
	ExamName string          `json:"exam_name"`
	ExamType string          `json:"exam_type"`
	Subjects []SubjectResult `json:"subjects"`
	Total    ExamTotals      `json:"total"`
}

type SubjectResult struct {
	SubjectCode   string          `json:"subject_code"`
	SubjectName   string          `json:"subject_name"`
	MaxMarks      decimal.Decimal `json:"max_marks"`
	PassingMarks  decimal.Decimal `json:"passing_marks"`
	MarksObtained decimal.Decimal `json:"marks_obtained"`
	Percentage    decimal.Decimal `json:"percentage"`
	Grade         string          `json:"grade"`
	Status        string          `json:"status"` // PASS | FAIL | ABSENT | GRACE
}

type ExamTotals struct {
	MaxMarks      decimal.Decimal `json:"max_marks"`
	MarksObtained decimal.Decimal `json:"marks_obtained"`
	Percentage    decimal.Decimal `json:"percentage"`
	Grade         string          `json:"grade"`
}

// ──────────────────────────────────────────────────────────────
// STUDENT ATTENDANCE SUMMARY
// ──────────────────────────────────────────────────────────────

type StudentAttendanceRequest struct {
	StudentEnrollmentID uuid.UUID  `form:"student_enrollment_id" binding:"required,uuid"`
	FromDate            *time.Time `form:"from_date"`
	ToDate              *time.Time `form:"to_date"`
}

type StudentAttendanceSummary struct {
	StudentEnrollmentID uuid.UUID       `json:"student_enrollment_id"`
	StudentName         string          `json:"student_name"`
	AdmissionNo         string          `json:"admission_no"`
	ClassSection        string          `json:"class_section"`
	FromDate            *time.Time      `json:"from_date,omitempty"`
	ToDate              *time.Time      `json:"to_date,omitempty"`
	TotalDays           int             `json:"total_days"`
	Present             int             `json:"present"`
	Absent              int             `json:"absent"`
	HalfDay             int             `json:"half_day"`
	Late                int             `json:"late"`
	Leave               int             `json:"leave"`
	AttendancePercent   decimal.Decimal `json:"attendance_percent"`
}

// ──────────────────────────────────────────────────────────────
// CLASS ATTENDANCE SUMMARY
// ──────────────────────────────────────────────────────────────

type ClassAttendanceRequest struct {
	ClassSectionID uuid.UUID  `form:"class_section_id" binding:"required,uuid"`
	Date           *time.Time `form:"date"`  // single day — if nil returns monthly summary
	Month          *int       `form:"month"` // 1–12
	Year           *int       `form:"year"`
}

type ClassAttendanceSummary struct {
	ClassSectionID uuid.UUID              `json:"class_section_id"`
	ClassSection   string                 `json:"class_section"`
	Date           *time.Time             `json:"date,omitempty"`
	Month          *int                   `json:"month,omitempty"`
	Year           *int                   `json:"year,omitempty"`
	TotalStudents  int                    `json:"total_students"`
	Students       []StudentAttendanceRow `json:"students"`
}

type StudentAttendanceRow struct {
	EnrollmentID      uuid.UUID       `json:"enrollment_id"`
	RollNumber        int             `json:"roll_number"`
	StudentName       string          `json:"student_name"`
	AdmissionNo       string          `json:"admission_no"`
	TotalDays         int             `json:"total_days"`
	Present           int             `json:"present"`
	Absent            int             `json:"absent"`
	AttendancePercent decimal.Decimal `json:"attendance_percent"`
}

// ──────────────────────────────────────────────────────────────
// CLASS PERFORMANCE REPORT
// ──────────────────────────────────────────────────────────────

type ClassPerformanceRequest struct {
	ClassSectionID uuid.UUID `form:"class_section_id" binding:"required,uuid"`
	ExamID         uuid.UUID `form:"exam_id"          binding:"required,uuid"`
}

type ClassPerformanceReport struct {
	ClassSectionID uuid.UUID            `json:"class_section_id"`
	ClassSection   string               `json:"class_section"`
	ExamName       string               `json:"exam_name"`
	TotalStudents  int                  `json:"total_students"`
	Subjects       []SubjectPerformance `json:"subjects"`
	TopStudents    []StudentRank        `json:"top_students"` // top 5
}

type SubjectPerformance struct {
	SubjectCode    string          `json:"subject_code"`
	SubjectName    string          `json:"subject_name"`
	MaxMarks       decimal.Decimal `json:"max_marks"`
	ClassAverage   decimal.Decimal `json:"class_average"`
	HighestMarks   decimal.Decimal `json:"highest_marks"`
	LowestMarks    decimal.Decimal `json:"lowest_marks"`
	PassCount      int             `json:"pass_count"`
	FailCount      int             `json:"fail_count"`
	AbsentCount    int             `json:"absent_count"`
	PassPercentage decimal.Decimal `json:"pass_percentage"`
}

type StudentRank struct {
	Rank        int             `json:"rank"`
	StudentName string          `json:"student_name"`
	AdmissionNo string          `json:"admission_no"`
	TotalMarks  decimal.Decimal `json:"total_marks"`
	MaxMarks    decimal.Decimal `json:"max_marks"`
	Percentage  decimal.Decimal `json:"percentage"`
	Grade       string          `json:"grade"`
}

// ──────────────────────────────────────────────────────────────
// FEE COLLECTION REPORT
// ──────────────────────────────────────────────────────────────

type FeeCollectionRequest struct {
	AcademicYearID uuid.UUID  `form:"academic_year_id" binding:"required,uuid"`
	StandardID     *uuid.UUID `form:"standard_id"`      // nil = all standards
	ClassSectionID *uuid.UUID `form:"class_section_id"` // nil = all sections
}

type FeeCollectionReport struct {
	AcademicYear   string             `json:"academic_year"`
	TotalDue       decimal.Decimal    `json:"total_due"`
	TotalCollected decimal.Decimal    `json:"total_collected"`
	TotalBalance   decimal.Decimal    `json:"total_balance"`
	TotalWaived    decimal.Decimal    `json:"total_waived"`
	Rows           []FeeCollectionRow `json:"rows"`
}

type FeeCollectionRow struct {
	ClassSection   string          `json:"class_section"`
	FeeComponent   string          `json:"fee_component"`
	TotalStudents  int             `json:"total_students"`
	PaidCount      int             `json:"paid_count"`
	PendingCount   int             `json:"pending_count"`
	OverdueCount   int             `json:"overdue_count"`
	WaivedCount    int             `json:"waived_count"`
	TotalDue       decimal.Decimal `json:"total_due"`
	TotalCollected decimal.Decimal `json:"total_collected"`
	TotalBalance   decimal.Decimal `json:"total_balance"`
}

// ──────────────────────────────────────────────────────────────
// TEACHER ATTENDANCE SUMMARY
// ──────────────────────────────────────────────────────────────

type TeacherAttendanceRequest struct {
	TeacherID *uuid.UUID `form:"teacher_id"` // nil = all teachers
	FromDate  *time.Time `form:"from_date"`
	ToDate    *time.Time `form:"to_date"`
	Month     *int       `form:"month"`
	Year      *int       `form:"year"`
}

type TeacherAttendanceSummary struct {
	TeacherID         uuid.UUID       `json:"teacher_id"`
	EmployeeID        uuid.UUID       `json:"employee_id"`
	TeacherName       string          `json:"teacher_name"`
	TotalDays         int             `json:"total_days"`
	Present           int             `json:"present"`
	Absent            int             `json:"absent"`
	HalfDay           int             `json:"half_day"`
	Late              int             `json:"late"`
	Leave             int             `json:"leave"`
	AttendancePercent decimal.Decimal `json:"attendance_percent"`
}

// ──────────────────────────────────────────────────────────────
// HELPERS
// ──────────────────────────────────────────────────────────────

// gradeFromPercentage returns a letter grade based on percentage.
func gradeFromPercentage(pct decimal.Decimal) string {
	f, _ := pct.Float64()
	switch {
	case f >= 90:
		return "A+"
	case f >= 80:
		return "A"
	case f >= 70:
		return "B+"
	case f >= 60:
		return "B"
	case f >= 50:
		return "C"
	case f >= 40:
		return "D"
	default:
		return "F"
	}
}

func safeDiv(numerator, denominator decimal.Decimal) decimal.Decimal {
	if denominator.IsZero() {
		return decimal.Zero
	}
	return numerator.Div(denominator).Mul(decimal.NewFromInt(100)).Round(2)
}
