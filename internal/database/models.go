package database

import (
	"time"

	"github.com/shopspring/decimal"
)

// ==========================================
// ENUMS
// ==========================================

type StudentStatus string

const (
	StudentStatusActive   StudentStatus = "ACTIVE"
	StudentStatusAlumni   StudentStatus = "ALUMNI"
	StudentStatusInactive StudentStatus = "INACTIVE"
)

type EnrollmentStatus string

const (
	EnrollmentStatusEnrolled EnrollmentStatus = "ENROLLED"
	EnrollmentStatusPromoted EnrollmentStatus = "PROMOTED"
	EnrollmentStatusDetained EnrollmentStatus = "DETAINED"
)

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "PRESENT"
	AttendanceStatusAbsent  AttendanceStatus = "ABSENT"
	AttendanceStatusHalfDay AttendanceStatus = "HALF_DAY"
	AttendanceStatusLate    AttendanceStatus = "LATE"
)

// ==========================================
// BASE
// ==========================================
type Base struct {
	ID        string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time  `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"                                  json:"updated_at"`
	CreatedBy *string    `gorm:"column:created_by"                               json:"created_by,omitempty"`
	DeletedAt *time.Time `gorm:"index"                                           json:"deleted_at,omitempty"`
}

// ==========================================
// CORE MODELS
// ==========================================

type Student struct {
	Base
	AdmissionNo string        `gorm:"column:admission_no;uniqueIndex;not null" json:"admission_no"`
	FirstName   string        `gorm:"column:first_name;not null"               json:"first_name"`
	LastName    string        `gorm:"column:last_name;not null"                json:"last_name"`
	DOB         time.Time     `gorm:"column:dob;not null"                      json:"dob"`
	Status      StudentStatus `gorm:"type:text;default:'ACTIVE'"               json:"status"`

	Enrollments []StudentEnrollment `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
}

func (Student) TableName() string { return "student" }

type Teacher struct {
	Base
	EmployeeID string `gorm:"column:employee_id;uniqueIndex;not null" json:"employee_id"`
	FirstName  string `gorm:"column:first_name;not null"              json:"first_name"`
	LastName   string `gorm:"column:last_name;not null"               json:"last_name"`
	IsActive   bool   `gorm:"column:is_active;default:true"           json:"is_active"`

	ClassSections []ClassSection      `gorm:"foreignKey:ClassTeacherID"              json:"class_sections,omitempty"`
	Assignments   []TeacherAssignment `gorm:"foreignKey:TeacherID"                   json:"assignments,omitempty"`
	TimeTables    []TimeTable         `gorm:"foreignKey:TeacherID"                   json:"time_tables,omitempty"`
}

func (Teacher) TableName() string { return "teacher" }

type Standard struct {
	Base
	Name string `gorm:"uniqueIndex;not null" json:"name"`

	ClassSections []ClassSection `gorm:"foreignKey:StandardID" json:"class_sections,omitempty"`
}

func (Standard) TableName() string { return "standard" }

type Subject struct {
	Base
	Code string `gorm:"uniqueIndex;not null" json:"code"`
	Name string `gorm:"not null"             json:"name"`

	TeacherAssignments []TeacherAssignment `gorm:"foreignKey:SubjectID" json:"teacher_assignments,omitempty"`
	TimeTables         []TimeTable         `gorm:"foreignKey:SubjectID" json:"time_tables,omitempty"`
	ExamSchedules      []ExamSchedule      `gorm:"foreignKey:SubjectID" json:"exam_schedules,omitempty"`
}

func (Subject) TableName() string { return "subject" }

// ==========================================
// ACADEMIC YEAR & CLASS MANAGEMENT
// ==========================================

type AcademicYear struct {
	Base
	Name      string    `gorm:"uniqueIndex;not null"       json:"name"`
	StartDate time.Time `gorm:"column:start_date;not null" json:"start_date"`
	EndDate   time.Time `gorm:"column:end_date;not null"   json:"end_date"`
	IsActive  bool      `gorm:"column:is_active;default:false" json:"is_active"`

	ClassSections []ClassSection `gorm:"foreignKey:AcademicYearID" json:"class_sections,omitempty"`
	Exams         []Exam         `gorm:"foreignKey:AcademicYearID" json:"exams,omitempty"`
}

func (AcademicYear) TableName() string { return "academic_year" }

type ClassSection struct {
	Base
	AcademicYearID string  `gorm:"column:academic_year_id;not null;index" json:"academic_year_id"`
	StandardID     string  `gorm:"column:standard_id;not null;index"      json:"standard_id"`
	SectionName    string  `gorm:"column:section_name;not null"           json:"section_name"`
	ClassTeacherID *string `gorm:"column:class_teacher_id"                json:"class_teacher_id,omitempty"`

	AcademicYear ClassSectionAcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Standard     Standard                 `gorm:"foreignKey:StandardID"     json:"standard,omitempty"`
	ClassTeacher *Teacher                 `gorm:"foreignKey:ClassTeacherID" json:"class_teacher,omitempty"`

	Enrollments   []StudentEnrollment `gorm:"foreignKey:ClassSectionID" json:"enrollments,omitempty"`
	Assignments   []TeacherAssignment `gorm:"foreignKey:ClassSectionID" json:"assignments,omitempty"`
	TimeTables    []TimeTable         `gorm:"foreignKey:ClassSectionID" json:"time_tables,omitempty"`
	ExamSchedules []ExamSchedule      `gorm:"foreignKey:ClassSectionID" json:"exam_schedules,omitempty"`
}

func (ClassSection) TableName() string { return "class_section" }

// Alias to avoid circular embed issues while keeping JSON clean.
type ClassSectionAcademicYear = AcademicYear

// ==========================================
// ENROLLMENT & SCHEDULING
// ==========================================

type StudentEnrollment struct {
	Base
	StudentID      string           `gorm:"column:student_id;not null;index"       json:"student_id"`
	ClassSectionID string           `gorm:"column:class_section_id;not null;index" json:"class_section_id"`
	RollNumber     int              `gorm:"column:roll_number;not null"            json:"roll_number"`
	Status         EnrollmentStatus `gorm:"type:text;default:'ENROLLED'"           json:"status"`

	Student      Student      `gorm:"foreignKey:StudentID"      json:"student,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`

	Attendances []Attendance `gorm:"foreignKey:StudentEnrollmentID" json:"attendances,omitempty"`
	ExamResults []ExamResult `gorm:"foreignKey:StudentEnrollmentID" json:"exam_results,omitempty"`
}

func (StudentEnrollment) TableName() string { return "student_enrollment" }

type TeacherAssignment struct {
	Base
	TeacherID      string `gorm:"column:teacher_id;not null;index"       json:"teacher_id"`
	ClassSectionID string `gorm:"column:class_section_id;not null;index" json:"class_section_id"`
	SubjectID      string `gorm:"column:subject_id;not null"             json:"subject_id"`

	Teacher      Teacher      `gorm:"foreignKey:TeacherID"      json:"teacher,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
}

func (TeacherAssignment) TableName() string { return "teacher_assignment" }

type TimeTable struct {
	Base
	ClassSectionID string    `gorm:"column:class_section_id;not null;index" json:"class_section_id"`
	SubjectID      string    `gorm:"column:subject_id;not null"             json:"subject_id"`
	TeacherID      string    `gorm:"column:teacher_id;not null;index"       json:"teacher_id"`
	DayOfWeek      int       `gorm:"column:day_of_week;not null"            json:"day_of_week"`
	StartTime      time.Time `gorm:"column:start_time;not null"             json:"start_time"`
	EndTime        time.Time `gorm:"column:end_time;not null"               json:"end_time"`

	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	Teacher      Teacher      `gorm:"foreignKey:TeacherID"      json:"teacher,omitempty"`
}

func (TimeTable) TableName() string { return "time_table" }

// ==========================================
// TRANSACTIONAL MODELS
// ==========================================

type Attendance struct {
	Base
	StudentEnrollmentID string           `gorm:"column:student_enrollment_id;not null;index" json:"student_enrollment_id"`
	Date                time.Time        `gorm:"column:date;type:date;not null;index"         json:"date"`
	Status              AttendanceStatus `gorm:"type:text;not null"                           json:"status"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
}

func (Attendance) TableName() string { return "attendance" }

type Exam struct {
	Base
	AcademicYearID string `gorm:"column:academic_year_id;not null;index" json:"academic_year_id"`
	Name           string `gorm:"not null"                               json:"name"`

	AcademicYear AcademicYear   `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Schedules    []ExamSchedule `gorm:"foreignKey:ExamID"         json:"schedules,omitempty"`
}

func (Exam) TableName() string { return "exam" }

type ExamSchedule struct {
	Base
	ExamID         string          `gorm:"column:exam_id;not null;index"          json:"exam_id"`
	ClassSectionID string          `gorm:"column:class_section_id;not null;index" json:"class_section_id"`
	SubjectID      string          `gorm:"column:subject_id;not null"             json:"subject_id"`
	ExamDate       time.Time       `gorm:"column:exam_date;type:date;not null"    json:"exam_date"`
	MaxMarks       decimal.Decimal `gorm:"column:max_marks;type:decimal(5,2)"     json:"max_marks"`

	Exam         Exam         `gorm:"foreignKey:ExamID"         json:"exam,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	Results      []ExamResult `gorm:"foreignKey:ExamScheduleID" json:"results,omitempty"`
}

func (ExamSchedule) TableName() string { return "exam_schedule" }

type ExamResult struct {
	Base
	ExamScheduleID      string          `gorm:"column:exam_schedule_id;not null;index"      json:"exam_schedule_id"`
	StudentEnrollmentID string          `gorm:"column:student_enrollment_id;not null;index" json:"student_enrollment_id"`
	MarksObtained       decimal.Decimal `gorm:"column:marks_obtained;type:decimal(5,2)"     json:"marks_obtained"`

	ExamSchedule      ExamSchedule      `gorm:"foreignKey:ExamScheduleID"      json:"exam_schedule,omitempty"`
	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
}

func (ExamResult) TableName() string { return "exam_result" }

// AllModels returns all GORM models for AutoMigrate.
func AllModels() []any {
	return []any{
		&Student{},
		&Teacher{},
		&Standard{},
		&Subject{},
		&AcademicYear{},
		&ClassSection{},
		&StudentEnrollment{},
		&TeacherAssignment{},
		&TimeTable{},
		&Attendance{},
		&Exam{},
		&ExamSchedule{},
		&ExamResult{},
	}
}
