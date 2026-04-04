package database

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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
	EnrollmentStatusEnrolled  EnrollmentStatus = "ENROLLED"
	EnrollmentStatusPromoted  EnrollmentStatus = "PROMOTED"
	EnrollmentStatusDetained  EnrollmentStatus = "DETAINED"
	EnrollmentStatusWithdrawn EnrollmentStatus = "WITHDRAWN"
)

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "PRESENT"
	AttendanceStatusAbsent  AttendanceStatus = "ABSENT"
	AttendanceStatusHalfDay AttendanceStatus = "HALF_DAY"
	AttendanceStatusLate    AttendanceStatus = "LATE"
	AttendanceStatusLeave   AttendanceStatus = "LEAVE"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "SUPER_ADMIN" // Full system access
	UserRoleAdmin      UserRole = "ADMIN"       // School-level admin
	UserRolePrincipal  UserRole = "PRINCIPAL"   // Read-all + approve actions
	UserRoleTeacher    UserRole = "TEACHER"     // Own classes/subjects
	UserRoleStudent    UserRole = "STUDENT"     // Own data only
	UserRoleParent     UserRole = "PARENT"      // Child's data only
	UserRoleStaff      UserRole = "STAFF"       // Non-teaching staff
)

type PermissionAction string

const (
	PermissionActionCreate PermissionAction = "CREATE"
	PermissionActionRead   PermissionAction = "READ"
	PermissionActionUpdate PermissionAction = "UPDATE"
	PermissionActionDelete PermissionAction = "DELETE"
	PermissionActionManage PermissionAction = "MANAGE" // All of the above
)

type ExamResultStatus string

const (
	ExamResultStatusPass   ExamResultStatus = "PASS"
	ExamResultStatusFail   ExamResultStatus = "FAIL"
	ExamResultStatusAbsent ExamResultStatus = "ABSENT"
	ExamResultStatusGrace  ExamResultStatus = "GRACE"
)

type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "PENDING"
	LeaveStatusApproved LeaveStatus = "APPROVED"
	LeaveStatusRejected LeaveStatus = "REJECTED"
)

type FeeStatus string

const (
	FeeStatusPending FeeStatus = "PENDING"
	FeeStatusPaid    FeeStatus = "PAID"
	FeeStatusPartial FeeStatus = "PARTIAL"
	FeeStatusOverdue FeeStatus = "OVERDUE"
	FeeStatusWaived  FeeStatus = "WAIVED"
)

type NoticeAudience string

const (
	NoticeAudienceAll      NoticeAudience = "ALL"
	NoticeAudienceTeachers NoticeAudience = "TEACHERS"
	NoticeAudienceStudents NoticeAudience = "STUDENTS"
	NoticeAudienceParents  NoticeAudience = "PARENTS"
	NoticeAudienceStaff    NoticeAudience = "STAFF"
)

// ==========================================
// BASE
// ==========================================

type Base struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	CreatedBy *string        `gorm:"column:created_by;type:uuid"                    json:"created_by"`
	UpdatedBy *string        `gorm:"column:updated_by;type:uuid"                    json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"deleted_at,omitempty"`
}

// ==========================================
// AUTH & PERMISSIONS
// ==========================================

// User is the central identity for all actors in the system.
// Each teacher, student, parent, and staff member has exactly one User.
type User struct {
	Base
	Email        string   `gorm:"column:email;uniqueIndex;not null"   json:"email"`
	PasswordHash string   `gorm:"column:password_hash;not null"       json:"-"`
	Role         UserRole `gorm:"column:role;type:text;not null"      json:"role"`
	IsActive     bool     `gorm:"column:is_active;default:true"       json:"is_active"`
	// Profile link — only one of these will be non-nil depending on role
	TeacherID *string `gorm:"column:teacher_id;type:uuid;uniqueIndex" json:"teacher_id,omitempty"`
	StudentID *string `gorm:"column:student_id;type:uuid;uniqueIndex" json:"student_id,omitempty"`
	ParentID  *string `gorm:"column:parent_id;type:uuid;uniqueIndex"  json:"parent_id,omitempty"`
	StaffID   *string `gorm:"column:staff_id;type:uuid;uniqueIndex"   json:"staff_id,omitempty"`

	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`

	Teacher *Teacher `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Student *Student `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Parent  *Parent  `gorm:"foreignKey:ParentID"  json:"parent,omitempty"`
	Staff   *Staff   `gorm:"foreignKey:StaffID"   json:"staff,omitempty"`

	RolePermissions []RolePermission   `gorm:"foreignKey:UserID" json:"role_permissions,omitempty"`
	RefreshTokens   []UserRefreshToken `gorm:"foreignKey:UserID" json:"refresh_tokens,omitempty"`
}

func (User) TableName() string { return "user" }

// UserRefreshToken stores active refresh tokens for JWT rotation.
type UserRefreshToken struct {
	Base
	UserID    string    `gorm:"column:user_id;type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"column:token;uniqueIndex;not null"        json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"              json:"expires_at"`
	Revoked   bool      `gorm:"column:revoked;default:false"            json:"revoked"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserRefreshToken) TableName() string { return "user_refresh_token" }

// Permission defines a resource + action pair (e.g. "student" + "READ").
type Permission struct {
	Base
	Resource    string           `gorm:"column:resource;not null" json:"resource"` // e.g. "student", "exam"
	Action      PermissionAction `gorm:"column:action;type:text;not null" json:"action"`
	Description string           `gorm:"column:description"               json:"description"`
}

func (Permission) TableName() string { return "permission" }

// RolePermission is a grant: a User (or role archetype) holds a Permission,
// optionally scoped to a specific resource scope (e.g. a ClassSection).
type RolePermission struct {
	Base
	UserID       string `gorm:"column:user_id;type:uuid;not null;index"       json:"user_id"`
	PermissionID string `gorm:"column:permission_id;type:uuid;not null;index" json:"permission_id"`
	// Optional scope — if nil, the permission is global for the user
	ScopeType *string `gorm:"column:scope_type" json:"scope_type,omitempty"` // e.g. "class_section"
	ScopeID   *string `gorm:"column:scope_id;type:uuid" json:"scope_id,omitempty"`

	User       User       `gorm:"foreignKey:UserID"       json:"user,omitempty"`
	Permission Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

func (RolePermission) TableName() string { return "role_permission" }

// ==========================================
// CORE PEOPLE MODELS
// ==========================================

type Student struct {
	Base
	AdmissionNo   string        `gorm:"column:admission_no;uniqueIndex;not null" json:"admission_no"`
	FirstName     string        `gorm:"column:first_name;not null"               json:"first_name"`
	LastName      string        `gorm:"column:last_name;not null"                json:"last_name"`
	DOB           time.Time     `gorm:"column:dob;not null"                      json:"dob"`
	Gender        Gender        `gorm:"column:gender;type:text;not null"         json:"gender"`
	Status        StudentStatus `gorm:"column:status;type:text;default:'ACTIVE'" json:"status"`
	Phone         *string       `gorm:"column:phone"                             json:"phone,omitempty"`
	Address       *string       `gorm:"column:address;type:text"                 json:"address,omitempty"`
	BloodGroup    *string       `gorm:"column:blood_group"                       json:"blood_group,omitempty"`
	PhotoURL      *string       `gorm:"column:photo_url"                         json:"photo_url,omitempty"`
	AdmissionDate time.Time     `gorm:"column:admission_date;not null"           json:"admission_date"`

	Enrollments []StudentEnrollment `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
	Parents     []StudentParent     `gorm:"foreignKey:StudentID" json:"parents,omitempty"`
}

func (Student) TableName() string { return "student" }

type Parent struct {
	Base
	FirstName    string  `gorm:"column:first_name;not null" json:"first_name"`
	LastName     string  `gorm:"column:last_name;not null"  json:"last_name"`
	Relationship string  `gorm:"column:relationship;not null" json:"relationship"` // Father, Mother, Guardian
	Phone        string  `gorm:"column:phone;not null"      json:"phone"`
	Email        *string `gorm:"column:email;uniqueIndex"   json:"email,omitempty"`
	Address      *string `gorm:"column:address;type:text"   json:"address,omitempty"`
	Occupation   *string `gorm:"column:occupation"          json:"occupation,omitempty"`

	Children []StudentParent `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (Parent) TableName() string { return "parent" }

// StudentParent is the many-to-many join between students and parents.
type StudentParent struct {
	Base
	StudentID string `gorm:"column:student_id;type:uuid;not null;index"  json:"student_id"`
	ParentID  string `gorm:"column:parent_id;type:uuid;not null;index"   json:"parent_id"`
	IsPrimary bool   `gorm:"column:is_primary;default:false"             json:"is_primary"`

	Student Student `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Parent  Parent  `gorm:"foreignKey:ParentID"  json:"parent,omitempty"`
}

func (StudentParent) TableName() string { return "student_parent" }

type Teacher struct {
	Base
	EmployeeID     string     `gorm:"column:employee_id;uniqueIndex;not null"  json:"employee_id"`
	FirstName      string     `gorm:"column:first_name;not null"               json:"first_name"`
	LastName       string     `gorm:"column:last_name;not null"                json:"last_name"`
	Gender         Gender     `gorm:"column:gender;type:text;not null"         json:"gender"`
	DOB            *time.Time `gorm:"column:dob"                              json:"dob,omitempty"`
	Phone          *string    `gorm:"column:phone"                             json:"phone,omitempty"`
	Email          *string    `gorm:"column:email;uniqueIndex"                 json:"email,omitempty"`
	Address        *string    `gorm:"column:address;type:text"                 json:"address,omitempty"`
	Qualification  *string    `gorm:"column:qualification"                     json:"qualification,omitempty"`
	Specialization *string    `gorm:"column:specialization"                    json:"specialization,omitempty"`
	JoiningDate    time.Time  `gorm:"column:joining_date;not null"             json:"joining_date"`
	IsActive       bool       `gorm:"column:is_active;default:true"            json:"is_active"`
	PhotoURL       *string    `gorm:"column:photo_url"                         json:"photo_url,omitempty"`

	// Reverse relations
	DepartmentHeadOf []Department        `gorm:"foreignKey:HeadTeacherID"   json:"department_head_of,omitempty"`
	ClassSections    []ClassSection      `gorm:"foreignKey:ClassTeacherID"  json:"class_sections,omitempty"`
	Assignments      []TeacherAssignment `gorm:"foreignKey:TeacherID"  json:"assignments,omitempty"`
	TimeTables       []TimeTable         `gorm:"foreignKey:TeacherID"       json:"time_tables,omitempty"`
	LeaveRequests    []TeacherLeave      `gorm:"foreignKey:TeacherID"       json:"leave_requests,omitempty"`
}

func (Teacher) TableName() string { return "teacher" }

// Staff covers non-teaching personnel (accountants, librarians, etc.).
type Staff struct {
	Base
	EmployeeID  string    `gorm:"column:employee_id;uniqueIndex;not null" json:"employee_id"`
	FirstName   string    `gorm:"column:first_name;not null"              json:"first_name"`
	LastName    string    `gorm:"column:last_name;not null"               json:"last_name"`
	Gender      Gender    `gorm:"column:gender;type:text;not null"        json:"gender"`
	Designation string    `gorm:"column:designation;not null"             json:"designation"`
	Phone       *string   `gorm:"column:phone"                            json:"phone,omitempty"`
	Email       *string   `gorm:"column:email;uniqueIndex"                json:"email,omitempty"`
	JoiningDate time.Time `gorm:"column:joining_date;not null"            json:"joining_date"`
	IsActive    bool      `gorm:"column:is_active;default:true"           json:"is_active"`
}

func (Staff) TableName() string { return "staff" }

// ==========================================
// DEPARTMENT → CLASS → SECTION HIERARCHY
// ==========================================

// Department groups related standards together under a head teacher.
// Examples: Primary (Std 1–5), Middle (Std 6–8), Secondary (Std 9–10), Senior Secondary (Std 11–12).
type Department struct {
	Base
	Name          string  `gorm:"column:name;uniqueIndex;not null"                json:"name"`
	Code          string  `gorm:"column:code;uniqueIndex;not null"                json:"code"`
	Description   *string `gorm:"column:description;type:text"                   json:"description,omitempty"`
	HeadTeacherID *string `gorm:"column:head_teacher_id;type:uuid"               json:"head_teacher_id,omitempty"`
	IsActive      bool    `gorm:"column:is_active;default:true"                  json:"is_active"`

	HeadTeacher *Teacher   `gorm:"foreignKey:HeadTeacherID" json:"head_teacher,omitempty"`
	Standards   []Standard `gorm:"foreignKey:DepartmentID"  json:"standards,omitempty"`
}

func (Department) TableName() string { return "department" }

// Standard (Class) belongs to a Department.
// E.g. "Grade 6" belongs to "Middle School" department.
type Standard struct {
	Base
	Name         string  `gorm:"column:name;not null"                            json:"name"`
	DepartmentID string  `gorm:"column:department_id;type:uuid;not null;index"   json:"department_id"`
	OrderIndex   int     `gorm:"column:order_index;not null;default:0"           json:"order_index"` // for sorting
	Description  *string `gorm:"column:description;type:text"                   json:"description,omitempty"`

	Department    Department        `gorm:"foreignKey:DepartmentID"  json:"department,omitempty"`
	ClassSections []ClassSection    `gorm:"foreignKey:StandardID"    json:"class_sections,omitempty"`
	Subjects      []StandardSubject `gorm:"foreignKey:StandardID" json:"subjects,omitempty"`
}

func (Standard) TableName() string { return "standard" }

// Subject is a school-wide catalogue of subjects.
type Subject struct {
	Base
	Code        string  `gorm:"uniqueIndex;not null" json:"code"`
	Name        string  `gorm:"not null"             json:"name"`
	Description *string `gorm:"column:description;type:text" json:"description,omitempty"`
	IsElective  bool    `gorm:"column:is_elective;default:false" json:"is_elective"`

	StandardSubjects   []StandardSubject   `gorm:"foreignKey:SubjectID"   json:"standard_subjects,omitempty"`
	TeacherAssignments []TeacherAssignment `gorm:"foreignKey:SubjectID"   json:"teacher_assignments,omitempty"`
	TimeTables         []TimeTable         `gorm:"foreignKey:SubjectID"   json:"time_tables,omitempty"`
	ExamSchedules      []ExamSchedule      `gorm:"foreignKey:SubjectID"   json:"exam_schedules,omitempty"`
}

func (Subject) TableName() string { return "subject" }

// StandardSubject maps which subjects are taught in which standard.
type StandardSubject struct {
	Base
	StandardID string `gorm:"column:standard_id;type:uuid;not null;index" json:"standard_id"`
	SubjectID  string `gorm:"column:subject_id;type:uuid;not null;index"  json:"subject_id"`
	IsCore     bool   `gorm:"column:is_core;default:true"                 json:"is_core"`

	Standard Standard `gorm:"foreignKey:StandardID" json:"standard,omitempty"`
	Subject  Subject  `gorm:"foreignKey:SubjectID"  json:"subject,omitempty"`
}

func (StandardSubject) TableName() string { return "standard_subject" }

// ==========================================
// ACADEMIC YEAR & CLASS MANAGEMENT
// ==========================================

type AcademicYear struct {
	Base
	Name      string    `gorm:"uniqueIndex;not null"           json:"name"`
	StartDate time.Time `gorm:"column:start_date;not null"     json:"start_date"`
	EndDate   time.Time `gorm:"column:end_date;not null"       json:"end_date"`
	IsActive  bool      `gorm:"column:is_active;default:false" json:"is_active"`

	ClassSections []ClassSection `gorm:"foreignKey:AcademicYearID" json:"class_sections,omitempty"`
	Exams         []Exam         `gorm:"foreignKey:AcademicYearID" json:"exams,omitempty"`
	FeeStructures []FeeStructure `gorm:"foreignKey:AcademicYearID" json:"fee_structures,omitempty"`
}

func (AcademicYear) TableName() string { return "academic_year" }

// ClassSection is a concrete, academic-year-scoped instance of Standard + Section.
// E.g. "Grade 6 – Section A" for academic year "2024-25".
type ClassSection struct {
	Base
	AcademicYearID string  `gorm:"column:academic_year_id;type:uuid;not null;index" json:"academic_year_id"`
	StandardID     string  `gorm:"column:standard_id;type:uuid;not null;index"      json:"standard_id"`
	SectionName    string  `gorm:"column:section_name;not null"                     json:"section_name"`
	ClassTeacherID *string `gorm:"column:class_teacher_id;type:uuid"                json:"class_teacher_id,omitempty"`
	RoomNumber     *string `gorm:"column:room_number"                               json:"room_number,omitempty"`
	MaxStrength    int     `gorm:"column:max_strength;default:40"                   json:"max_strength"`

	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Standard     Standard     `gorm:"foreignKey:StandardID"     json:"standard,omitempty"`
	ClassTeacher *Teacher     `gorm:"foreignKey:ClassTeacherID" json:"class_teacher,omitempty"`

	Enrollments   []StudentEnrollment `gorm:"foreignKey:ClassSectionID" json:"enrollments,omitempty"`
	Assignments   []TeacherAssignment `gorm:"foreignKey:ClassSectionID" json:"assignments,omitempty"`
	TimeTables    []TimeTable         `gorm:"foreignKey:ClassSectionID" json:"time_tables,omitempty"`
	ExamSchedules []ExamSchedule      `gorm:"foreignKey:ClassSectionID" json:"exam_schedules,omitempty"`
}

func (ClassSection) TableName() string { return "class_section" }

// ==========================================
// ENROLLMENT & SCHEDULING
// ==========================================

type StudentEnrollment struct {
	Base
	StudentID      string           `gorm:"column:student_id;type:uuid;not null;index"       json:"student_id"`
	ClassSectionID string           `gorm:"column:class_section_id;type:uuid;not null;index" json:"class_section_id"`
	RollNumber     int              `gorm:"column:roll_number;not null"                      json:"roll_number"`
	Status         EnrollmentStatus `gorm:"column:status;type:text;default:'ENROLLED'"       json:"status"`
	EnrollmentDate time.Time        `gorm:"column:enrollment_date;not null"                  json:"enrollment_date"`
	LeftDate       *time.Time       `gorm:"column:left_date"                                 json:"left_date,omitempty"`

	Student      Student      `gorm:"foreignKey:StudentID"      json:"student,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`

	Attendances []Attendance `gorm:"foreignKey:StudentEnrollmentID" json:"attendances,omitempty"`
	ExamResults []ExamResult `gorm:"foreignKey:StudentEnrollmentID" json:"exam_results,omitempty"`
	FeeRecords  []FeeRecord  `gorm:"foreignKey:StudentEnrollmentID" json:"fee_records,omitempty"`
}

func (StudentEnrollment) TableName() string { return "student_enrollment" }

type TeacherAssignment struct {
	Base
	TeacherID      string `gorm:"column:teacher_id;type:uuid;not null;index"       json:"teacher_id"`
	ClassSectionID string `gorm:"column:class_section_id;type:uuid;not null;index" json:"class_section_id"`
	SubjectID      string `gorm:"column:subject_id;type:uuid;not null;index"       json:"subject_id"`

	Teacher      Teacher      `gorm:"foreignKey:TeacherID"      json:"teacher,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
}

func (TeacherAssignment) TableName() string { return "teacher_assignment" }

type TimeTable struct {
	Base
	ClassSectionID string    `gorm:"column:class_section_id;type:uuid;not null;index" json:"class_section_id"`
	SubjectID      string    `gorm:"column:subject_id;type:uuid;not null"             json:"subject_id"`
	TeacherID      string    `gorm:"column:teacher_id;type:uuid;not null;index"       json:"teacher_id"`
	DayOfWeek      int       `gorm:"column:day_of_week;not null"                      json:"day_of_week"` // 0=Sun … 6=Sat
	StartTime      time.Time `gorm:"column:start_time;not null"                       json:"start_time"`
	EndTime        time.Time `gorm:"column:end_time;not null"                         json:"end_time"`
	RoomNumber     *string   `gorm:"column:room_number"                               json:"room_number,omitempty"`

	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	Teacher      Teacher      `gorm:"foreignKey:TeacherID"      json:"teacher,omitempty"`
}

func (TimeTable) TableName() string { return "time_table" }

// ==========================================
// ATTENDANCE
// ==========================================

type Attendance struct {
	Base
	StudentEnrollmentID string           `gorm:"column:student_enrollment_id;type:uuid;not null;index" json:"student_enrollment_id"`
	Date                time.Time        `gorm:"column:date;type:date;not null;index"                  json:"date"`
	Status              AttendanceStatus `gorm:"column:status;type:text;not null"                      json:"status"`
	Remark              *string          `gorm:"column:remark;type:text"                               json:"remark,omitempty"`
	RecordedByTeacherID *string          `gorm:"column:recorded_by_teacher_id;type:uuid"               json:"recorded_by_teacher_id,omitempty"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID"        json:"student_enrollment,omitempty"`
	RecordedBy        *Teacher          `gorm:"foreignKey:RecordedByTeacherID"        json:"recorded_by,omitempty"`
}

func (Attendance) TableName() string { return "attendance" }

// TeacherAttendance tracks teacher presence independent of class/student flow.
type TeacherAttendance struct {
	Base
	TeacherID string           `gorm:"column:teacher_id;type:uuid;not null;index" json:"teacher_id"`
	Date      time.Time        `gorm:"column:date;type:date;not null;index"        json:"date"`
	Status    AttendanceStatus `gorm:"column:status;type:text;not null"            json:"status"`
	Remark    *string          `gorm:"column:remark;type:text"                     json:"remark,omitempty"`

	Teacher Teacher `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
}

func (TeacherAttendance) TableName() string { return "teacher_attendance" }

// ==========================================
// LEAVE MANAGEMENT
// ==========================================

type TeacherLeave struct {
	Base
	TeacherID  string      `gorm:"column:teacher_id;type:uuid;not null;index"   json:"teacher_id"`
	FromDate   time.Time   `gorm:"column:from_date;type:date;not null"           json:"from_date"`
	ToDate     time.Time   `gorm:"column:to_date;type:date;not null"             json:"to_date"`
	Reason     string      `gorm:"column:reason;type:text;not null"              json:"reason"`
	Status     LeaveStatus `gorm:"column:status;type:text;default:'PENDING'"     json:"status"`
	ReviewedBy *string     `gorm:"column:reviewed_by;type:uuid"                 json:"reviewed_by,omitempty"`
	ReviewNote *string     `gorm:"column:review_note;type:text"                 json:"review_note,omitempty"`
	ReviewedAt *time.Time  `gorm:"column:reviewed_at"                           json:"reviewed_at,omitempty"`

	Teacher  Teacher  `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
	Reviewer *Teacher `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (TeacherLeave) TableName() string { return "teacher_leave" }

// ==========================================
// EXAMS & RESULTS
// ==========================================

type Exam struct {
	Base
	AcademicYearID string  `gorm:"column:academic_year_id;type:uuid;not null;index" json:"academic_year_id"`
	Name           string  `gorm:"not null"                                         json:"name"`
	Description    *string `gorm:"column:description;type:text"                   json:"description,omitempty"`
	// e.g. "UNIT_TEST", "MIDTERM", "FINAL", "INTERNAL"
	ExamType    string    `gorm:"column:exam_type;not null"                        json:"exam_type"`
	StartDate   time.Time `gorm:"column:start_date;not null"                    json:"start_date"`
	EndDate     time.Time `gorm:"column:end_date;not null"                      json:"end_date"`
	IsPublished bool      `gorm:"column:is_published;default:false"                json:"is_published"`

	AcademicYear AcademicYear   `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Schedules    []ExamSchedule `gorm:"foreignKey:ExamID"         json:"schedules,omitempty"`
}

func (Exam) TableName() string { return "exam" }

type ExamSchedule struct {
	Base
	ExamID         string          `gorm:"column:exam_id;type:uuid;not null;index"          json:"exam_id"`
	ClassSectionID string          `gorm:"column:class_section_id;type:uuid;not null;index" json:"class_section_id"`
	SubjectID      string          `gorm:"column:subject_id;type:uuid;not null"             json:"subject_id"`
	ExamDate       time.Time       `gorm:"column:exam_date;type:date;not null"              json:"exam_date"`
	StartTime      *time.Time      `gorm:"column:start_time"                                json:"start_time,omitempty"`
	EndTime        *time.Time      `gorm:"column:end_time"                                  json:"end_time,omitempty"`
	MaxMarks       decimal.Decimal `gorm:"column:max_marks;type:decimal(5,2);not null"      json:"max_marks"`
	PassingMarks   decimal.Decimal `gorm:"column:passing_marks;type:decimal(5,2);not null"  json:"passing_marks"`
	RoomNumber     *string         `gorm:"column:room_number"                               json:"room_number,omitempty"`

	Exam         Exam         `gorm:"foreignKey:ExamID"         json:"exam,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	Results      []ExamResult `gorm:"foreignKey:ExamScheduleID" json:"results,omitempty"`
}

func (ExamSchedule) TableName() string { return "exam_schedule" }

type ExamResult struct {
	Base
	ExamScheduleID      string           `gorm:"column:exam_schedule_id;type:uuid;not null;index"      json:"exam_schedule_id"`
	StudentEnrollmentID string           `gorm:"column:student_enrollment_id;type:uuid;not null;index" json:"student_enrollment_id"`
	MarksObtained       decimal.Decimal  `gorm:"column:marks_obtained;type:decimal(5,2)"              json:"marks_obtained"`
	Grade               *string          `gorm:"column:grade"                                         json:"grade,omitempty"`
	Status              ExamResultStatus `gorm:"column:status;type:text;not null"                     json:"status"`
	Remarks             *string          `gorm:"column:remarks;type:text"                             json:"remarks,omitempty"`

	ExamSchedule      ExamSchedule      `gorm:"foreignKey:ExamScheduleID"      json:"exam_schedule,omitempty"`
	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
}

func (ExamResult) TableName() string { return "exam_result" }

// ==========================================
// FEE MANAGEMENT
// ==========================================

// FeeStructure defines the fee components for a standard in an academic year.
type FeeStructure struct {
	Base
	AcademicYearID string          `gorm:"column:academic_year_id;type:uuid;not null;index" json:"academic_year_id"`
	StandardID     string          `gorm:"column:standard_id;type:uuid;not null;index"      json:"standard_id"`
	FeeComponent   string          `gorm:"column:fee_component;not null"                    json:"fee_component"` // e.g. "Tuition", "Transport"
	Amount         decimal.Decimal `gorm:"column:amount;type:decimal(10,2);not null"        json:"amount"`
	DueDate        *time.Time      `gorm:"column:due_date"                                  json:"due_date,omitempty"`

	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Standard     Standard     `gorm:"foreignKey:StandardID"     json:"standard,omitempty"`
}

func (FeeStructure) TableName() string { return "fee_structure" }

// FeeRecord is an individual fee transaction tied to a student's enrollment.
type FeeRecord struct {
	Base
	StudentEnrollmentID string          `gorm:"column:student_enrollment_id;type:uuid;not null;index" json:"student_enrollment_id"`
	FeeComponent        string          `gorm:"column:fee_component;not null"                         json:"fee_component"`
	AmountDue           decimal.Decimal `gorm:"column:amount_due;type:decimal(10,2);not null"         json:"amount_due"`
	AmountPaid          decimal.Decimal `gorm:"column:amount_paid;type:decimal(10,2);default:0"       json:"amount_paid"`
	DueDate             time.Time       `gorm:"column:due_date;not null"                              json:"due_date"`
	PaidDate            *time.Time      `gorm:"column:paid_date"                                      json:"paid_date,omitempty"`
	Status              FeeStatus       `gorm:"column:status;type:text;default:'PENDING'"             json:"status"`
	TransactionRef      *string         `gorm:"column:transaction_ref"                                json:"transaction_ref,omitempty"`
	Remarks             *string         `gorm:"column:remarks;type:text"                              json:"remarks,omitempty"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
}

func (FeeRecord) TableName() string { return "fee_record" }

// ==========================================
// NOTICEBOARD / ANNOUNCEMENTS
// ==========================================

type Notice struct {
	Base
	Title       string         `gorm:"column:title;not null"                   json:"title"`
	Content     string         `gorm:"column:content;type:text;not null"       json:"content"`
	Audience    NoticeAudience `gorm:"column:audience;type:text;not null"      json:"audience"`
	PublishedAt *time.Time     `gorm:"column:published_at"                     json:"published_at,omitempty"`
	ExpiresAt   *time.Time     `gorm:"column:expires_at"                       json:"expires_at,omitempty"`
	IsPublished bool           `gorm:"column:is_published;default:false"       json:"is_published"`
	// Optional scope — if set, the notice targets only this class section
	ClassSectionID *string `gorm:"column:class_section_id;type:uuid;index" json:"class_section_id,omitempty"`

	ClassSection *ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
}

func (Notice) TableName() string { return "notice" }

// ==========================================
// ALL MODELS (for AutoMigrate)
// ==========================================

// AllModels returns all GORM models for AutoMigrate.
// Order matters: parent tables must come before child tables.
func AllModels() []any {
	return []any{
		// Auth & permissions
		&User{},
		&UserRefreshToken{},
		&Permission{},
		&RolePermission{},

		// Core people
		&Staff{},
		&Teacher{},
		&Student{},
		&Parent{},
		&StudentParent{},

		// Academic structure
		&Department{},
		&Standard{},
		&Subject{},
		&StandardSubject{},
		&AcademicYear{},
		&ClassSection{},

		// Enrollment & scheduling
		&StudentEnrollment{},
		&TeacherAssignment{},
		&TimeTable{},

		// Attendance
		&Attendance{},
		&TeacherAttendance{},

		// Leave
		&TeacherLeave{},

		// Exams
		&Exam{},
		&ExamSchedule{},
		&ExamResult{},

		// Fees
		&FeeStructure{},
		&FeeRecord{},

		// Communication
		&Notice{},
	}
}
