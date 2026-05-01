package database

// ============================================================
// model_v7.go — Education Management System
// Production-grade GORM model for PostgreSQL 15+.
//
// CHANGES FROM v6
// ───────────────
// BREAKING FIXES
//   • user table → "users", role table → "roles".
//     Both are PostgreSQL reserved identifiers. Any unquoted raw SQL
//     against the old names fails or targets system objects silently.
//     All FK references update automatically via GORM's TableName().
//   • GradeScale: idx_grade_scale_name was UNIQUE(grade) globally —
//     now correctly UNIQUE(academic_year_id, grade) via
//     migration 0012_grade_scale_fix.sql.
//   • UserScope: GORM composite uniqueIndex over nullable scope_id
//     cannot prevent duplicate global-scope rows (PostgreSQL treats
//     NULL ≠ NULL in unique indexes). GORM tags removed; replaced by
//     COALESCE-based partial unique index in migration 0013.
//
// DATA INTEGRITY FIXES
//   • TimeTable: second uniqueIndex (employee_id, day_of_week, start_time)
//     prevents teacher double-booking. Third index (room_id, day_of_week,
//     start_time) prevents room double-booking.
//   • StudentEnrollment: uniqueIndex (class_section_id, roll_number)
//     prevents duplicate roll numbers within a class section.
//   • SalaryStructure: uniqueIndex (employee_id, effective_from)
//     eliminates ambiguity in the "latest slab" lookup.
//   • TransportRoute: uniqueIndex on vehicle_number.
//   • Standard: uniqueIndex (department_id, name).
//   • LeaveBalance: DB trigger syncs used_days when EmployeeLeave status
//     changes to/from APPROVED (migration 0019).
//   • LibraryBook: Status column removed (always stale for multi-copy
//     books). Replaced by AvailableCopies counter maintained by trigger
//     (migration 0020).
//   • LibraryFineRate: new table for configurable per-day fine rates.
//
// TYPE SAFETY
//   • ExamType: converted from type alias (= string, unconstrained)
//     to a defined type with constants. DB CHECK added in migration 0022.
//   • ParentRelationship: new enum type replaces free-text string.
//   • EmployeeCategory, SubjectType, all status enums:
//     DB CHECK constraints added in migration 0022.
//
// REMOVED 3NF VIOLATIONS
//   • PromotionRecord.FromStandardID removed — derivable from
//     StudentEnrollment → ClassSection → Standard (transitive dependency).
//   • ExamResult.GPA removed — derivable from Grade → GradeScale.GPA
//     for the relevant academic year. Compute at query time.
//   • EmployeeLeave.TotalDays retained but marked as computed snapshot;
//     a DB-generated column approach is noted but not enforced here to
//     keep GORM compatibility.
//
// NEW MODELS
//   • Announcement — broadcast messages with audience targeting.
//     Addresses PermAnnounceSend which had no backing table.
//   • AnnouncementRead — per-user read receipt for announcements.
//   • LibraryFineRate — configurable fine rate per leave type.
//
// MISSING INDEXES ADDED
//   • user_refresh_token (expires_at) — cleanup job no longer full-scans.
//   • audit_log (resource_type, resource_id) — resource-scoped queries.
//   • employee_leave (employee_id, status) — pending leave list per employee.
//   • assignment_submission (assignment_id, status) — ungraded submissions.
//
// NEW MIGRATIONS (run after AutoMigrate, in order)
//   0011_rename_reserved_tables.sql    — user→users, role→roles
//   0012_grade_scale_fix.sql           — fix composite unique on grade_scale
//   0013_userscope_null_fix.sql        — COALESCE-based unique for scope_id
//   0014_timetable_conflicts.sql       — teacher + room double-booking indexes
//   0015_enrollment_rollnumber.sql     — (class_section_id, roll_number) unique
//   0016_salary_structure_unique.sql   — (employee_id, effective_from) unique
//   0017_transport_vehicle_unique.sql  — vehicle_number unique
//   0018_check_constraints.sql        — date ordering, marks, discount CHECKs
//   0019_leave_balance_trigger.sql     — sync used_days on leave approval
//   0020_library_trigger.sql          — available_copies counter trigger
//   0021_missing_indexes.sql          — expires_at, resource_type, status indexes
//   0022_status_checks.sql            — CHECK constraints for enum columns
//   0023_standard_unique.sql          — (department_id, name) unique
//
// DESIGN DECISIONS (updated from v6)
// ────────────────────────────────────
//   1. Single school — no school_id.
//   2. Table names: "users", "roles" to avoid PostgreSQL reserved identifiers.
//   3. Auth: JWT embeds role.slug + []permission codes ("attendance:mark").
//      Permission check in middleware is O(1) map lookup — no DB round-trip.
//   4. EmployeeCategory remains a Go constant (STI discriminator).
//   5. Base for entities (soft delete). BaseJunction for many-to-many (hard delete).
//   6. Monetary fields use decimal(12,2). Timetable time fields use time.Time
//      normalised to 1970-01-01 at service layer.
//   7. Unique constraints at model level so AutoMigrate creates them.
//      Service checks before insert — never rely on catching DB errors.
//   8. LeaveBalance.PendingDays NOT stored — computed at query time.
//   9. GradeScale enforces grade boundaries. ExamResult.Grade set by service
//      from GradeScale, never free text.
//  10. SalaryRecord derived columns are intentional snapshots. CHECK constraints
//      enforced in migration.
//  11. UserScope GORM uniqueIndex tags removed — uniqueness is handled by
//      migration 0013 using a COALESCE partial index.
//  12. LibraryBook.Status removed — AvailableCopies (trigger-maintained) is
//      the canonical availability signal.
//  13. ExamResult.GPA removed — join GradeScale at query time.
//  14. PromotionRecord.FromStandardID removed — 3NF violation.
// ============================================================

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ============================================================
// SECTION 0 — ENUMS & CONSTANTS
// ============================================================

type SystemRole string

const (
	SystemRoleSuperAdmin SystemRole = "super_admin"
	SystemRoleAdmin      SystemRole = "admin"
	SystemRolePrincipal  SystemRole = "principal"
	SystemRoleTeacher    SystemRole = "teacher"
	SystemRoleStudent    SystemRole = "student"
	SystemRoleParent     SystemRole = "parent"
	SystemRoleStaff      SystemRole = "staff"
)

type PermissionCode string

const (
	PermAttendanceMark   PermissionCode = "attendance:mark"
	PermAttendanceView   PermissionCode = "attendance:view"
	PermAttendanceExport PermissionCode = "attendance:export"

	PermStudentCreate PermissionCode = "student:create"
	PermStudentView   PermissionCode = "student:view"
	PermStudentEdit   PermissionCode = "student:edit"
	PermStudentDelete PermissionCode = "student:delete"

	PermEmployeeCreate PermissionCode = "employee:create"
	PermEmployeeView   PermissionCode = "employee:view"
	PermEmployeeEdit   PermissionCode = "employee:edit"
	PermEmployeeDelete PermissionCode = "employee:delete"

	PermExamCreate    PermissionCode = "exam:create"
	PermExamView      PermissionCode = "exam:view"
	PermExamPublish   PermissionCode = "exam:publish"
	PermResultEnter   PermissionCode = "result:enter"
	PermResultView    PermissionCode = "result:view"
	PermResultPublish PermissionCode = "result:publish"

	PermAssignmentCreate PermissionCode = "assignment:create"
	PermAssignmentView   PermissionCode = "assignment:view"
	PermAssignmentGrade  PermissionCode = "assignment:grade"

	PermFeeView    PermissionCode = "fee:view"
	PermFeeCollect PermissionCode = "fee:collect"
	PermFeeWaive   PermissionCode = "fee:waive"
	PermSalaryView PermissionCode = "salary:view"
	PermSalaryPay  PermissionCode = "salary:pay"

	PermNoticeCreate    PermissionCode = "notice:create"
	PermNoticePublish   PermissionCode = "notice:publish"
	PermMessageSend     PermissionCode = "message:send"
	PermAnnounceSend    PermissionCode = "announcement:send"
	PermAnnouncePublish PermissionCode = "announcement:publish"

	PermTimetableView PermissionCode = "timetable:view"
	PermTimetableEdit PermissionCode = "timetable:edit"

	PermLeaveApply   PermissionCode = "leave:apply"
	PermLeaveApprove PermissionCode = "leave:approve"
	PermLeaveView    PermissionCode = "leave:view"

	PermLibraryIssue  PermissionCode = "library:issue"
	PermLibraryReturn PermissionCode = "library:return"
	PermLibraryManage PermissionCode = "library:manage"

	PermTransportView   PermissionCode = "transport:view"
	PermTransportManage PermissionCode = "transport:manage"

	PermRoleCreate PermissionCode = "role:create"
	PermRoleEdit   PermissionCode = "role:edit"
	PermRoleDelete PermissionCode = "role:delete"
	PermRoleView   PermissionCode = "role:view"
	PermRoleGrant  PermissionCode = "role:grant"
	PermRoleRevoke PermissionCode = "role:revoke"

	PermAuditView    PermissionCode = "audit:view"
	PermSystemConfig PermissionCode = "system:config"
)

type EmployeeCategory string

const (
	EmployeeCategoryTeacher       EmployeeCategory = "TEACHER"
	EmployeeCategoryPrincipal     EmployeeCategory = "PRINCIPAL"
	EmployeeCategoryVicePrincipal EmployeeCategory = "VICE_PRINCIPAL"
	EmployeeCategoryStaff         EmployeeCategory = "STAFF"
	EmployeeCategoryCounselor     EmployeeCategory = "COUNSELOR"
	EmployeeCategoryLibrarian     EmployeeCategory = "LIBRARIAN"
	EmployeeCategoryAccountant    EmployeeCategory = "ACCOUNTANT"
	EmployeeCategoryDriver        EmployeeCategory = "DRIVER"
	EmployeeCategoryNurse         EmployeeCategory = "NURSE"
	EmployeeCategorySecurity      EmployeeCategory = "SECURITY"
	EmployeeCategoryITSupport     EmployeeCategory = "IT_SUPPORT"
)

// AcademicEmployeeCategories — O(1) lookup for class/timetable eligibility.
var AcademicEmployeeCategories = map[EmployeeCategory]bool{
	EmployeeCategoryTeacher:       true,
	EmployeeCategoryPrincipal:     true,
	EmployeeCategoryVicePrincipal: true,
}

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

type StudentStatus string

const (
	StudentStatusActive      StudentStatus = "ACTIVE"
	StudentStatusAlumni      StudentStatus = "ALUMNI"
	StudentStatusInactive    StudentStatus = "INACTIVE"
	StudentStatusTransferred StudentStatus = "TRANSFERRED"
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

type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "PENDING"
	LeaveStatusApproved  LeaveStatus = "APPROVED"
	LeaveStatusRejected  LeaveStatus = "REJECTED"
	LeaveStatusCancelled LeaveStatus = "CANCELLED"
)

type ExamResultStatus string

const (
	ExamResultStatusPass     ExamResultStatus = "PASS"
	ExamResultStatusFail     ExamResultStatus = "FAIL"
	ExamResultStatusAbsent   ExamResultStatus = "ABSENT"
	ExamResultStatusGrace    ExamResultStatus = "GRACE"
	ExamResultStatusWithheld ExamResultStatus = "WITHHELD"
)

type FeeStatus string

const (
	FeeStatusPending FeeStatus = "PENDING"
	FeeStatusPaid    FeeStatus = "PAID"
	FeeStatusPartial FeeStatus = "PARTIAL"
	FeeStatusOverdue FeeStatus = "OVERDUE"
	FeeStatusWaived  FeeStatus = "WAIVED"
)

type SalaryStatus string

const (
	SalaryStatusPending SalaryStatus = "PENDING"
	SalaryStatusPaid    SalaryStatus = "PAID"
	SalaryStatusPartial SalaryStatus = "PARTIAL"
	SalaryStatusOnHold  SalaryStatus = "ON_HOLD"
)

type NoticeAudience string

const (
	NoticeAudienceAll      NoticeAudience = "ALL"
	NoticeAudienceTeachers NoticeAudience = "TEACHERS"
	NoticeAudienceStudents NoticeAudience = "STUDENTS"
	NoticeAudienceParents  NoticeAudience = "PARENTS"
	NoticeAudienceStaff    NoticeAudience = "STAFF"
	NoticeAudienceClass    NoticeAudience = "CLASS" // requires ClassSectionID
)

// AnnouncementAudience — same values as NoticeAudience but typed separately
// so service validation is explicit.
type AnnouncementAudience string

const (
	AnnouncementAudienceAll      AnnouncementAudience = "ALL"
	AnnouncementAudienceTeachers AnnouncementAudience = "TEACHERS"
	AnnouncementAudienceStudents AnnouncementAudience = "STUDENTS"
	AnnouncementAudienceParents  AnnouncementAudience = "PARENTS"
	AnnouncementAudienceStaff    AnnouncementAudience = "STAFF"
)

type SubjectType string

const (
	SubjectTypeCore     SubjectType = "CORE"
	SubjectTypeElective SubjectType = "ELECTIVE"
	SubjectTypeOptional SubjectType = "OPTIONAL"
)

type AssignmentStatus string

const (
	AssignmentStatusDraft     AssignmentStatus = "DRAFT"
	AssignmentStatusPublished AssignmentStatus = "PUBLISHED"
	AssignmentStatusClosed    AssignmentStatus = "CLOSED"
)

type SubmissionStatus string

const (
	SubmissionStatusMissing   SubmissionStatus = "MISSING"
	SubmissionStatusSubmitted SubmissionStatus = "SUBMITTED"
	SubmissionStatusLate      SubmissionStatus = "LATE"
	SubmissionStatusGraded    SubmissionStatus = "GRADED"
)

type PromotionStatus string

const (
	PromotionStatusPromoted  PromotionStatus = "PROMOTED"
	PromotionStatusDetained  PromotionStatus = "DETAINED"
	PromotionStatusGraduated PromotionStatus = "GRADUATED"
)

type DocumentType string

const (
	DocumentTypeBirthCertificate    DocumentType = "BIRTH_CERTIFICATE"
	DocumentTypeTransferCertificate DocumentType = "TRANSFER_CERTIFICATE"
	DocumentTypeMarksheet           DocumentType = "MARKSHEET"
	DocumentTypeIDCard              DocumentType = "ID_CARD"
	DocumentTypeOther               DocumentType = "OTHER"
)

type BloodGroup string

const (
	BloodGroupAPos  BloodGroup = "A+"
	BloodGroupANeg  BloodGroup = "A-"
	BloodGroupBPos  BloodGroup = "B+"
	BloodGroupBNeg  BloodGroup = "B-"
	BloodGroupOPos  BloodGroup = "O+"
	BloodGroupONeg  BloodGroup = "O-"
	BloodGroupABPos BloodGroup = "AB+"
	BloodGroupABNeg BloodGroup = "AB-"
)

type RoomType string

const (
	RoomTypeClassroom RoomType = "CLASSROOM"
	RoomTypeLab       RoomType = "LAB"
	RoomTypeHall      RoomType = "HALL"
	RoomTypeOffice    RoomType = "OFFICE"
	RoomTypeLibrary   RoomType = "LIBRARY"
)

type HolidayType string

const (
	HolidayTypePublic  HolidayType = "PUBLIC"
	HolidayTypeSchool  HolidayType = "SCHOOL"
	HolidayTypeHalfDay HolidayType = "HALF_DAY"
)

type AuditDecision string

const (
	AuditDecisionAllow AuditDecision = "ALLOW"
	AuditDecisionDeny  AuditDecision = "DENY"
)

type ScopeType string

const (
	ScopeTypeGlobal       ScopeType = ""
	ScopeTypeClassSection ScopeType = "class_section"
	ScopeTypeDepartment   ScopeType = "department"
	ScopeTypeStudent      ScopeType = "student"
	ScopeTypeSubject      ScopeType = "subject"
)

// ExamType — v7 fix: no longer a type alias (= string).
// Type aliases give no compile-time or DB-level safety.
// DB CHECK constraint added in migration 0022.
type ExamType string

const (
	ExamTypeUnitTest  ExamType = "UNIT_TEST"
	ExamTypeMidTerm   ExamType = "MID_TERM"
	ExamTypeFinal     ExamType = "FINAL"
	ExamTypeMock      ExamType = "MOCK"
	ExamTypeInternal  ExamType = "INTERNAL"
	ExamTypePractical ExamType = "PRACTICAL"
)

// ParentRelationship — v7 fix: replaces free-text string.
// DB CHECK constraint added in migration 0022.
type ParentRelationship string

const (
	ParentRelationshipFather   ParentRelationship = "FATHER"
	ParentRelationshipMother   ParentRelationship = "MOTHER"
	ParentRelationshipGuardian ParentRelationship = "GUARDIAN"
	ParentRelationshipSibling  ParentRelationship = "SIBLING"
	ParentRelationshipOther    ParentRelationship = "OTHER"
)

// ============================================================
// SECTION 1 — BASE MODELS
// ============================================================

// Base is embedded by all entity tables (soft-delete capable).
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v7()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	CreatedBy *string        `gorm:"column:created_by;type:uuid"                    json:"created_by,omitempty"`
	UpdatedBy *string        `gorm:"column:updated_by;type:uuid"                    json:"updated_by,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"deleted_at,omitempty"`
}

// BaseJunction is embedded by pure many-to-many tables (hard delete).
// Hard delete avoids unique-index conflicts on re-insert after removal.
type BaseJunction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v7()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	CreatedBy *string   `gorm:"column:created_by;type:uuid"                    json:"created_by,omitempty"`
}

// ============================================================
// SECTION 2 — RBAC
//
// Architecture (unchanged from v6, table names updated)
//  permission      — atomic capability, seeded on boot
//  roles           — named group of permissions (WAS: role)
//  role_permission — junction: which permissions a role has
//  users           — has one role via FK (WAS: user)
//  user_scope      — scoped per-user overrides
//
// IMPORTANT: UserScope.uniqueIndex tags are REMOVED from this model.
// The uniqueness constraint is handled by a COALESCE-based partial unique
// index in migration 0013, because PostgreSQL's standard unique index
// treats NULL ≠ NULL — two rows with scope_id=NULL would not collide.
// ============================================================

type Permission struct {
	Base
	Resource    string `gorm:"column:resource;not null;uniqueIndex:idx_perm_res_action" json:"resource"`
	Action      string `gorm:"column:action;not null;uniqueIndex:idx_perm_res_action"   json:"action"`
	Description string `gorm:"column:description;not null;default:''"                   json:"description"`
	IsSystem    bool   `gorm:"column:is_system;not null;default:false"                  json:"is_system"`
}

func (Permission) TableName() string { return "permission" }

// Role — renamed table: "roles" (v7). "role" is a PostgreSQL reserved word.
type Role struct {
	Base
	Slug        string `gorm:"column:slug;uniqueIndex;not null"        json:"slug"`
	Name        string `gorm:"column:name;not null"                    json:"name"`
	Description string `gorm:"column:description;not null;default:''" json:"description"`
	IsSystem    bool   `gorm:"column:is_system;not null;default:false" json:"is_system"`
	Priority    int    `gorm:"column:priority;not null;default:10"     json:"priority"`

	Permissions []Permission    `gorm:"many2many:role_permission;joinForeignKey:RoleID;joinReferences:PermissionID" json:"permissions,omitempty"`
	Users       []User          `gorm:"foreignKey:RoleID"                                                           json:"users,omitempty"`
	RoleChanges []RoleChangeLog `gorm:"foreignKey:RoleID"                                                           json:"role_changes,omitempty"`
}

// TableName returns "roles" to avoid the PostgreSQL "role" reserved word.
func (Role) TableName() string { return "roles" }

// RolePermission — explicit junction for audit trail (GrantedByID, GrantedAt).
type RolePermission struct {
	BaseJunction
	RoleID       uuid.UUID  `gorm:"column:role_id;type:uuid;not null;uniqueIndex:idx_role_perm"       json:"role_id"`
	PermissionID uuid.UUID  `gorm:"column:permission_id;type:uuid;not null;uniqueIndex:idx_role_perm" json:"permission_id"`
	GrantedByID  uuid.UUID  `gorm:"column:granted_by_id;type:uuid;not null"                           json:"granted_by_id"`
	Role         Role       `gorm:"foreignKey:RoleID"       json:"role,omitempty"`
	Permission   Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	GrantedBy    User       `gorm:"foreignKey:GrantedByID"  json:"granted_by,omitempty"`
}

func (RolePermission) TableName() string { return "role_permission" }

// RoleChangeLog — immutable append-only RBAC audit trail.
type RoleChangeLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v7()" json:"id"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;index"                           json:"created_at"`
	RoleID       *uuid.UUID `gorm:"column:role_id;type:uuid;index"                 json:"role_id,omitempty"`
	TargetUserID *uuid.UUID `gorm:"column:target_user_id;type:uuid;index"          json:"target_user_id,omitempty"`
	PermissionID *uuid.UUID `gorm:"column:permission_id;type:uuid"                 json:"permission_id,omitempty"`
	// Action values: GRANT_PERMISSION | REVOKE_PERMISSION | ASSIGN_ROLE | REVOKE_ROLE
	//                CREATE_ROLE | UPDATE_ROLE | DELETE_ROLE
	Action    string    `gorm:"column:action;not null"                   json:"action"`
	OldValue  *string   `gorm:"column:old_value;type:text"               json:"old_value,omitempty"`
	NewValue  *string   `gorm:"column:new_value;type:text"               json:"new_value,omitempty"`
	ActorID   uuid.UUID `gorm:"column:actor_id;type:uuid;not null;index" json:"actor_id"`
	IPAddress *string   `gorm:"column:ip_address"                        json:"ip_address,omitempty"`
	Role      *Role     `gorm:"foreignKey:RoleID"  json:"role,omitempty"`
	Actor     User      `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
}

func (RoleChangeLog) TableName() string { return "role_change_log" }

// User — renamed table: "users" (v7). "user" is a PostgreSQL reserved word.
//
// Persona exclusivity: at most one of (employee_id, student_id, parent_id)
// may be non-null. Enforced by DB CHECK constraint (migration 0008):
//
//	CHECK (
//	  ((employee_id IS NOT NULL)::int +
//	   (student_id  IS NOT NULL)::int +
//	   (parent_id   IS NOT NULL)::int) <= 1
//	)
type User struct {
	Base
	Email        string     `gorm:"column:email;uniqueIndex;not null"        json:"email"`
	PasswordHash string     `gorm:"column:password_hash;not null"            json:"-"`
	RoleID       uuid.UUID  `gorm:"column:role_id;type:uuid;not null;index"  json:"role_id"`
	IsActive     bool       `gorm:"column:is_active;default:true"            json:"is_active"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"                     json:"last_login_at,omitempty"`
	// Persona FKs — at most one non-null (DB CHECK enforced in migration 0008).
	EmployeeID *uuid.UUID `gorm:"column:employee_id;type:uuid;uniqueIndex" json:"employee_id,omitempty"`
	StudentID  *uuid.UUID `gorm:"column:student_id;type:uuid;uniqueIndex"  json:"student_id,omitempty"`
	ParentID   *uuid.UUID `gorm:"column:parent_id;type:uuid;uniqueIndex"   json:"parent_id,omitempty"`

	Role          Role               `gorm:"foreignKey:RoleID"     json:"role,omitempty"`
	Employee      *Employee          `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Student       *Student           `gorm:"foreignKey:StudentID"  json:"student,omitempty"`
	Parent        *Parent            `gorm:"foreignKey:ParentID"   json:"parent,omitempty"`
	Scopes        []UserScope        `gorm:"foreignKey:UserID"     json:"scopes,omitempty"`
	RefreshTokens []UserRefreshToken `gorm:"foreignKey:UserID"     json:"refresh_tokens,omitempty"`
}

// TableName returns "users" to avoid the PostgreSQL "user" reserved word.
func (User) TableName() string { return "users" }

// UserRefreshToken — one row per device/session.
// ExpiresAt is indexed (v7 addition) so cleanup queries do not full-scan.
type UserRefreshToken struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v7()" json:"id"`
	UserID           uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"        json:"user_id"`
	Token            string    `gorm:"column:token;uniqueIndex;not null"              json:"-"`
	ExpiresAt        time.Time `gorm:"column:expires_at;not null;index"               json:"expires_at"` // v7: indexed
	Revoked          bool      `gorm:"column:revoked;default:false"                   json:"is_revoked"`
	CreatedAt        time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	IPAddress        *string   `gorm:"column:ip_address"                              json:"ip_address,omitempty"`
	UserAgent        *string   `gorm:"column:user_agent"                              json:"user_agent,omitempty"`
	RoleSnapshotSlug string    `gorm:"column:role_snapshot_slug;not null"             json:"role_snapshot_slug"`
	User             User      `gorm:"foreignKey:UserID"                              json:"user,omitempty"`
}

func (UserRefreshToken) TableName() string { return "user_refresh_token" }

// UserScope — scoped permission override.
//
// v7 CHANGE: uniqueIndex GORM tags REMOVED. PostgreSQL unique indexes do not
// treat NULL = NULL, so (user_id, permission_id, scope_type, NULL) would not
// collide with a second row having the same non-null fields and scope_id=NULL.
// The uniqueness constraint is enforced by migration 0013 using:
//
//	CREATE UNIQUE INDEX idx_user_scope_unique ON user_scope (
//	  user_id, permission_id, scope_type,
//	  COALESCE(scope_id, '00000000-0000-0000-0000-000000000000')
//	) WHERE deleted_at IS NULL;
//
// Resolution order (highest wins):
//  1. UserScope deny (is_deny=true)
//  2. UserScope allow (is_deny=false)
//  3. Role permission grant
type UserScope struct {
	Base
	UserID       uuid.UUID  `gorm:"column:user_id;type:uuid;not null;index"       json:"user_id"`
	PermissionID uuid.UUID  `gorm:"column:permission_id;type:uuid;not null;index" json:"permission_id"`
	ScopeType    ScopeType  `gorm:"column:scope_type;not null"                    json:"scope_type"`
	ScopeID      *uuid.UUID `gorm:"column:scope_id;type:uuid"                     json:"scope_id,omitempty"`
	IsDeny       bool       `gorm:"column:is_deny;not null;default:false"         json:"is_deny"`
	GrantedBy    string     `gorm:"column:granted_by;type:uuid;not null"          json:"granted_by"`
	ExpiresAt    *time.Time `gorm:"column:expires_at"                            json:"expires_at,omitempty"`
	User         User       `gorm:"foreignKey:UserID"       json:"user,omitempty"`
	Permission   Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
	Granter      User       `gorm:"foreignKey:GrantedBy"    json:"granter,omitempty"`
}

func (UserScope) TableName() string { return "user_scope" }

// AuditLog — append-only, no soft delete, no UpdatedAt.
// UserID is not a FK (intentional): audit records must survive user deletion.
// Composite index on (resource_type, resource_id) added in migration 0021
// to support resource-scoped audit queries without full table scan.
type AuditLog struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v7()" json:"id"`
	CreatedAt    time.Time     `gorm:"autoCreateTime;index"                           json:"created_at"`
	UserID       uuid.UUID     `gorm:"column:user_id;type:uuid;not null;index"        json:"user_id"`
	Action       string        `gorm:"column:action;not null"                         json:"action"`
	ResourceType *string       `gorm:"column:resource_type;index"                     json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID    `gorm:"column:resource_id"                             json:"resource_id,omitempty"`
	Decision     AuditDecision `gorm:"column:decision;not null"                       json:"decision"`
	Reason       *string       `gorm:"column:reason"                                  json:"reason,omitempty"`
	OldValue     *string       `gorm:"column:old_value;type:text"                     json:"old_value,omitempty"`
	NewValue     *string       `gorm:"column:new_value;type:text"                     json:"new_value,omitempty"`
	IPAddress    *string       `gorm:"column:ip_address"                              json:"ip_address,omitempty"`
	UserAgent    *string       `gorm:"column:user_agent"                              json:"user_agent,omitempty"`
}

func (AuditLog) TableName() string { return "audit_log" }

// ============================================================
// SECTION 3 — PEOPLE
// ============================================================

// Employee — STI: all staff in one table, Category discriminates behaviour.
// DB trigger enforces category = 'DRIVER' for TransportRoute.DriverID.
type Employee struct {
	Base
	EmployeeCode   string           `gorm:"column:employee_code;uniqueIndex;not null"  json:"employee_code"`
	FirstName      string           `gorm:"column:first_name;not null"                 json:"first_name"`
	LastName       string           `gorm:"column:last_name;not null"                  json:"last_name"`
	Gender         Gender           `gorm:"column:gender;type:text;not null"           json:"gender"`
	Category       EmployeeCategory `gorm:"column:category;type:text;not null;index"   json:"category"`
	Designation    string           `gorm:"column:designation;not null"                json:"designation"`
	DepartmentID   *uuid.UUID       `gorm:"column:department_id;type:uuid;index"       json:"department_id,omitempty"`
	DOB            *time.Time       `gorm:"column:dob"                                 json:"dob,omitempty"`
	Phone          *string          `gorm:"column:phone"                               json:"phone,omitempty"`
	Email          *string          `gorm:"column:email;uniqueIndex"                   json:"email,omitempty"`
	Address        *string          `gorm:"column:address;type:text"                   json:"address,omitempty"`
	Qualification  *string          `gorm:"column:qualification"                       json:"qualification,omitempty"`
	Specialization *string          `gorm:"column:specialization"                      json:"specialization,omitempty"`
	JoiningDate    time.Time        `gorm:"column:joining_date;not null"               json:"joining_date"`
	IsActive       bool             `gorm:"column:is_active;default:true"              json:"is_active"`
	PhotoURL       *string          `gorm:"column:photo_url"                           json:"photo_url,omitempty"`

	Department       *Department          `gorm:"foreignKey:DepartmentID"    json:"department,omitempty"`
	DepartmentHeadOf []Department         `gorm:"foreignKey:HeadEmployeeID"  json:"department_head_of,omitempty"`
	ClassSections    []ClassSection       `gorm:"foreignKey:ClassEmployeeID" json:"class_sections,omitempty"`
	Assignments      []TeacherAssignment  `gorm:"foreignKey:EmployeeID"      json:"assignments,omitempty"`
	TimeTables       []TimeTable          `gorm:"foreignKey:EmployeeID"      json:"time_tables,omitempty"`
	Attendances      []EmployeeAttendance `gorm:"foreignKey:EmployeeID"      json:"attendances,omitempty"`
	LeaveRequests    []EmployeeLeave      `gorm:"foreignKey:EmployeeID"      json:"leave_requests,omitempty"`
	SalaryStructures []SalaryStructure    `gorm:"foreignKey:EmployeeID"      json:"salary_structures,omitempty"`
	SalaryRecords    []SalaryRecord       `gorm:"foreignKey:EmployeeID"      json:"salary_records,omitempty"`
}

func (Employee) TableName() string { return "employee" }

// Student — BloodGroup is in StudentHealth (canonical).
type Student struct {
	Base
	AdmissionNo   string        `gorm:"column:admission_no;uniqueIndex;not null"   json:"admission_no"`
	FirstName     string        `gorm:"column:first_name;not null"                 json:"first_name"`
	LastName      string        `gorm:"column:last_name;not null"                  json:"last_name"`
	DOB           time.Time     `gorm:"column:dob;not null"                        json:"dob"`
	Gender        Gender        `gorm:"column:gender;type:text;not null"           json:"gender"`
	Status        StudentStatus `gorm:"column:status;type:text;default:'ACTIVE'"   json:"status"`
	Phone         *string       `gorm:"column:phone"                               json:"phone,omitempty"`
	Address       *string       `gorm:"column:address;type:text"                   json:"address,omitempty"`
	PhotoURL      *string       `gorm:"column:photo_url"                           json:"photo_url,omitempty"`
	AdmissionDate time.Time     `gorm:"column:admission_date;not null"             json:"admission_date"`
	Nationality   *string       `gorm:"column:nationality"                         json:"nationality,omitempty"`
	Religion      *string       `gorm:"column:religion"                            json:"religion,omitempty"`

	Enrollments []StudentEnrollment `gorm:"foreignKey:StudentID" json:"enrollments,omitempty"`
	Parents     []StudentParent     `gorm:"foreignKey:StudentID" json:"parents,omitempty"`
	Documents   []StudentDocument   `gorm:"foreignKey:StudentID" json:"documents,omitempty"`
	Health      *StudentHealth      `gorm:"foreignKey:StudentID" json:"health,omitempty"`
}

func (Student) TableName() string { return "student" }

// Parent — Relationship is now a typed enum (v7 fix).
// DB CHECK constraint added in migration 0022.
type Parent struct {
	Base
	FirstName    string             `gorm:"column:first_name;not null"              json:"first_name"`
	LastName     string             `gorm:"column:last_name;not null"               json:"last_name"`
	Relationship ParentRelationship `gorm:"column:relationship;type:text;not null"  json:"relationship"`
	Phone        string             `gorm:"column:phone;not null"                   json:"phone"`
	Email        *string            `gorm:"column:email;uniqueIndex"                json:"email,omitempty"`
	Address      *string            `gorm:"column:address;type:text"                json:"address,omitempty"`
	Occupation   *string            `gorm:"column:occupation"                       json:"occupation,omitempty"`
	Children     []StudentParent    `gorm:"foreignKey:ParentID"                     json:"children,omitempty"`
}

func (Parent) TableName() string { return "parent" }

// StudentParent — junction. IsPrimary flags the main contact parent.
type StudentParent struct {
	BaseJunction
	StudentID uuid.UUID `gorm:"column:student_id;type:uuid;not null;uniqueIndex:idx_student_parent" json:"student_id"`
	ParentID  uuid.UUID `gorm:"column:parent_id;type:uuid;not null;uniqueIndex:idx_student_parent"  json:"parent_id"`
	IsPrimary bool      `gorm:"column:is_primary;default:false"                                     json:"is_primary"`
	Student   Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Parent    Parent    `gorm:"foreignKey:ParentID"  json:"parent,omitempty"`
}

func (StudentParent) TableName() string { return "student_parent" }

type StudentDocument struct {
	Base
	StudentID    uuid.UUID    `gorm:"column:student_id;type:uuid;not null;index"    json:"student_id"`
	DocumentType DocumentType `gorm:"column:document_type;type:text;not null"       json:"document_type"`
	Title        string       `gorm:"column:title;not null"                         json:"title"`
	FileURL      string       `gorm:"column:file_url;not null"                      json:"file_url"`
	UploadedByID *uuid.UUID   `gorm:"column:uploaded_by_id;type:uuid;index"         json:"uploaded_by_id,omitempty"`
	Student      Student      `gorm:"foreignKey:StudentID"                          json:"student,omitempty"`
	UploadedBy   *User        `gorm:"foreignKey:UploadedByID"                       json:"uploaded_by,omitempty"`
}

func (StudentDocument) TableName() string { return "student_document" }

// StudentHealth — canonical source for medical data.
// Service layer must create this row when creating a Student.
type StudentHealth struct {
	Base
	StudentID      uuid.UUID   `gorm:"column:student_id;type:uuid;not null;uniqueIndex" json:"student_id"`
	BloodGroup     *BloodGroup `gorm:"column:blood_group;type:text"                     json:"blood_group,omitempty"`
	Allergies      *string     `gorm:"column:allergies;type:text"                       json:"allergies,omitempty"`
	MedicalHistory *string     `gorm:"column:medical_history;type:text"                 json:"medical_history,omitempty"`
	Disabilities   *string     `gorm:"column:disabilities;type:text"                    json:"disabilities,omitempty"`
	EmergencyPhone string      `gorm:"column:emergency_phone;not null"                  json:"emergency_phone"`
	Student        Student     `gorm:"foreignKey:StudentID"                             json:"student,omitempty"`
}

func (StudentHealth) TableName() string { return "student_health" }

// ============================================================
// SECTION 4 — ACADEMIC STRUCTURE
// ============================================================

type Department struct {
	Base
	Name           string     `gorm:"column:name;uniqueIndex;not null"  json:"name"`
	Code           string     `gorm:"column:code;uniqueIndex;not null"  json:"code"`
	Description    *string    `gorm:"column:description;type:text"     json:"description,omitempty"`
	HeadEmployeeID *uuid.UUID `gorm:"column:head_employee_id;type:uuid" json:"head_employee_id,omitempty"`
	IsActive       bool       `gorm:"column:is_active;default:true"    json:"is_active"`
	HeadEmployee   *Employee  `gorm:"foreignKey:HeadEmployeeID"        json:"head_employee,omitempty"`
	Standards      []Standard `gorm:"foreignKey:DepartmentID"          json:"standards,omitempty"`
}

func (Department) TableName() string { return "department" }

// Standard — v7: uniqueIndex on (department_id, name) added via migration 0023.
// Without this, a department can have two standards named "Class 10".
type Standard struct {
	Base
	DepartmentID uuid.UUID `gorm:"column:department_id;type:uuid;not null;index" json:"department_id"`
	Name         string    `gorm:"column:name;not null"                          json:"name"`
	// OrderIndex drives display sort; uniqueIndex on (department_id, order_index)
	// prevents two standards having the same display position within a department.
	OrderIndex        int               `gorm:"column:order_index;not null;default:0;uniqueIndex:idx_std_order" json:"order_index"`
	DepartmentIDOrder string            `gorm:"->;column:department_id;uniqueIndex:idx_std_order" json:"-"` // phantom for composite index
	Description       *string           `gorm:"column:description;type:text"                  json:"description,omitempty"`
	Department        Department        `gorm:"foreignKey:DepartmentID"  json:"department,omitempty"`
	ClassSections     []ClassSection    `gorm:"foreignKey:StandardID"    json:"class_sections,omitempty"`
	Subjects          []StandardSubject `gorm:"foreignKey:StandardID"    json:"subjects,omitempty"`
}

func (Standard) TableName() string { return "standard" }

type Subject struct {
	Base
	Code               string              `gorm:"column:code;uniqueIndex;not null" json:"code"`
	Name               string              `gorm:"column:name;not null"             json:"name"`
	Description        *string             `gorm:"column:description;type:text"     json:"description,omitempty"`
	StandardSubjects   []StandardSubject   `gorm:"foreignKey:SubjectID" json:"standard_subjects,omitempty"`
	TeacherAssignments []TeacherAssignment `gorm:"foreignKey:SubjectID" json:"teacher_assignments,omitempty"`
	TimeTables         []TimeTable         `gorm:"foreignKey:SubjectID" json:"time_tables,omitempty"`
	ExamSchedules      []ExamSchedule      `gorm:"foreignKey:SubjectID" json:"exam_schedules,omitempty"`
}

func (Subject) TableName() string { return "subject" }

type StandardSubject struct {
	BaseJunction
	StandardID  uuid.UUID   `gorm:"column:standard_id;type:uuid;not null;uniqueIndex:idx_std_subj" json:"standard_id"`
	SubjectID   uuid.UUID   `gorm:"column:subject_id;type:uuid;not null;uniqueIndex:idx_std_subj"  json:"subject_id"`
	SubjectType SubjectType `gorm:"column:subject_type;type:text;not null;default:'CORE'"          json:"subject_type"`
	Standard    Standard    `gorm:"foreignKey:StandardID"                                          json:"standard,omitempty"`
	Subject     Subject     `gorm:"foreignKey:SubjectID"                                           json:"subject,omitempty"`
}

func (StandardSubject) TableName() string { return "standard_subject" }

// AcademicYear — only one may have is_active=true at a time.
// Partial unique index in migration 0006.
// CHECK start_date < end_date in migration 0018.
type AcademicYear struct {
	Base
	Name          string          `gorm:"column:name;uniqueIndex;not null"       json:"name"`
	StartDate     time.Time       `gorm:"column:start_date;not null"             json:"start_date"`
	EndDate       time.Time       `gorm:"column:end_date;not null"               json:"end_date"`
	IsActive      bool            `gorm:"column:is_active;default:false"         json:"is_active"`
	ClassSections []ClassSection  `gorm:"foreignKey:AcademicYearID" json:"class_sections,omitempty"`
	Exams         []Exam          `gorm:"foreignKey:AcademicYearID" json:"exams,omitempty"`
	FeeStructures []FeeStructure  `gorm:"foreignKey:AcademicYearID" json:"fee_structures,omitempty"`
	Holidays      []SchoolHoliday `gorm:"foreignKey:AcademicYearID" json:"holidays,omitempty"`
	GradeScales   []GradeScale    `gorm:"foreignKey:AcademicYearID" json:"grade_scales,omitempty"`
}

func (AcademicYear) TableName() string { return "academic_year" }

type SchoolHoliday struct {
	Base
	AcademicYearID uuid.UUID    `gorm:"column:academic_year_id;type:uuid;not null;index"        json:"academic_year_id"`
	Date           time.Time    `gorm:"column:date;type:date;not null;uniqueIndex"              json:"date"`
	Name           string       `gorm:"column:name;not null"                                    json:"name"`
	HolidayType    HolidayType  `gorm:"column:holiday_type;type:text;not null;default:'PUBLIC'" json:"holiday_type"`
	Description    *string      `gorm:"column:description"                                      json:"description,omitempty"`
	AcademicYear   AcademicYear `gorm:"foreignKey:AcademicYearID"                               json:"academic_year,omitempty"`
}

func (SchoolHoliday) TableName() string { return "school_holiday" }

// GradeScale — v7 fix: unique index is now (academic_year_id, grade).
// Previous version had a bug: idx_grade_scale_name was only on (grade),
// making the grade label globally unique across all academic years.
// Migration 0012 drops the old index and creates the correct composite.
//
// CHECKs added in migration 0018:
//
//	min_percent >= 0
//	max_percent <= 100
//	min_percent < max_percent
type GradeScale struct {
	Base
	AcademicYearID uuid.UUID       `gorm:"column:academic_year_id;type:uuid;not null;uniqueIndex:idx_grade_year_pct;uniqueIndex:idx_grade_year_grade" json:"academic_year_id"`
	Grade          string          `gorm:"column:grade;not null;uniqueIndex:idx_grade_year_grade"                                                     json:"grade"`
	MinPercent     decimal.Decimal `gorm:"column:min_percent;type:decimal(5,2);not null;uniqueIndex:idx_grade_year_pct"                               json:"min_percent"`
	MaxPercent     decimal.Decimal `gorm:"column:max_percent;type:decimal(5,2);not null"                                                              json:"max_percent"`
	GPA            decimal.Decimal `gorm:"column:gpa;type:decimal(4,2);not null;default:0"                                                            json:"gpa"`
	Description    *string         `gorm:"column:description"                                                                                         json:"description,omitempty"`
	AcademicYear   AcademicYear    `gorm:"foreignKey:AcademicYearID"                                                                                  json:"academic_year,omitempty"`
}

func (GradeScale) TableName() string { return "grade_scale" }

// ============================================================
// SECTION 5 — CLASSES & SCHEDULING
// ============================================================

type Room struct {
	Base
	RoomNumber string   `gorm:"column:room_number;uniqueIndex;not null" json:"room_number"`
	RoomType   RoomType `gorm:"column:room_type;type:text;not null"     json:"room_type"`
	Capacity   int      `gorm:"column:capacity;default:40"              json:"capacity"`
	IsActive   bool     `gorm:"column:is_active;default:true"           json:"is_active"`
}

func (Room) TableName() string { return "room" }

// ClassSection — unique on (academic_year_id, standard_id, section_name).
// ClassEmployeeID must be in AcademicEmployeeCategories — service enforced.
type ClassSection struct {
	Base
	AcademicYearID  uuid.UUID  `gorm:"column:academic_year_id;type:uuid;not null;uniqueIndex:idx_cs" json:"academic_year_id"`
	StandardID      uuid.UUID  `gorm:"column:standard_id;type:uuid;not null;uniqueIndex:idx_cs"      json:"standard_id"`
	SectionName     string     `gorm:"column:section_name;not null;uniqueIndex:idx_cs"               json:"section_name"`
	ClassEmployeeID *uuid.UUID `gorm:"column:class_employee_id;type:uuid"                            json:"class_employee_id,omitempty"`
	RoomID          *uuid.UUID `gorm:"column:room_id;type:uuid"                                      json:"room_id,omitempty"`
	MaxStrength     int        `gorm:"column:max_strength;default:40"                                json:"max_strength"`

	AcademicYear  AcademicYear               `gorm:"foreignKey:AcademicYearID"  json:"academic_year,omitempty"`
	Standard      Standard                   `gorm:"foreignKey:StandardID"      json:"standard,omitempty"`
	ClassEmployee *Employee                  `gorm:"foreignKey:ClassEmployeeID" json:"class_employee,omitempty"`
	Room          *Room                      `gorm:"foreignKey:RoomID"          json:"room,omitempty"`
	Enrollments   []StudentEnrollment        `gorm:"foreignKey:ClassSectionID"  json:"enrollments,omitempty"`
	Assignments   []TeacherAssignment        `gorm:"foreignKey:ClassSectionID"  json:"assignments,omitempty"`
	TimeTables    []TimeTable                `gorm:"foreignKey:ClassSectionID"  json:"time_tables,omitempty"`
	ExamSchedules []ExamSchedule             `gorm:"foreignKey:ClassSectionID"  json:"exam_schedules,omitempty"`
	ElectiveSlots []ClassSectionElectiveSlot `gorm:"foreignKey:ClassSectionID"  json:"elective_slots,omitempty"`
}

func (ClassSection) TableName() string { return "class_section" }

// ClassSectionElectiveSlot — capacity gate for one elective in one section.
// CurrentEnrollment is incremented/decremented by DB trigger on
// student_elective INSERT/DELETE (migration 0002).
type ClassSectionElectiveSlot struct {
	Base
	ClassSectionID    uuid.UUID `gorm:"column:class_section_id;type:uuid;not null;uniqueIndex:idx_elec_slot"    json:"class_section_id"`
	StandardSubjectID uuid.UUID `gorm:"column:standard_subject_id;type:uuid;not null;uniqueIndex:idx_elec_slot" json:"standard_subject_id"`
	MaxCapacity       int       `gorm:"column:max_capacity;not null;default:30"                                 json:"max_capacity"`
	// CurrentEnrollment maintained by DB trigger — do not update directly.
	CurrentEnrollment int `gorm:"column:current_enrollment;not null;default:0" json:"current_enrollment"`

	ClassSection    ClassSection    `gorm:"foreignKey:ClassSectionID"    json:"class_section,omitempty"`
	StandardSubject StandardSubject `gorm:"foreignKey:StandardSubjectID" json:"standard_subject,omitempty"`
}

func (ClassSectionElectiveSlot) TableName() string { return "class_section_elective_slot" }

// StudentEnrollment — unique on (student_id, class_section_id).
// v7: Additional unique on (class_section_id, roll_number) via migration 0015
// prevents two students sharing the same roll number in the same class section.
type StudentEnrollment struct {
	Base
	StudentID      uuid.UUID `gorm:"column:student_id;type:uuid;not null;uniqueIndex:idx_enroll"                          json:"student_id"`
	ClassSectionID uuid.UUID `gorm:"column:class_section_id;type:uuid;not null;uniqueIndex:idx_enroll;index:idx_roll_cs"  json:"class_section_id"`
	// RollNumber unique within class section enforced by migration 0015.
	RollNumber     int              `gorm:"column:roll_number;not null;index:idx_roll_cs"                                         json:"roll_number"`
	Status         EnrollmentStatus `gorm:"column:status;type:text;default:'ENROLLED'"                                            json:"status"`
	EnrollmentDate time.Time        `gorm:"column:enrollment_date;not null"                                                       json:"enrollment_date"`
	LeftDate       *time.Time       `gorm:"column:left_date"                                                                      json:"left_date,omitempty"`

	Student      Student           `gorm:"foreignKey:StudentID"           json:"student,omitempty"`
	ClassSection ClassSection      `gorm:"foreignKey:ClassSectionID"      json:"class_section,omitempty"`
	Attendances  []Attendance      `gorm:"foreignKey:StudentEnrollmentID" json:"attendances,omitempty"`
	ExamResults  []ExamResult      `gorm:"foreignKey:StudentEnrollmentID" json:"exam_results,omitempty"`
	FeeRecords   []FeeRecord       `gorm:"foreignKey:StudentEnrollmentID" json:"fee_records,omitempty"`
	Electives    []StudentElective `gorm:"foreignKey:StudentEnrollmentID" json:"electives,omitempty"`
}

func (StudentEnrollment) TableName() string { return "student_enrollment" }

// StudentElective — gated by DB trigger (migration 0002) that checks capacity.
type StudentElective struct {
	BaseJunction
	StudentEnrollmentID uuid.UUID         `gorm:"column:student_enrollment_id;type:uuid;not null;uniqueIndex:idx_elective" json:"student_enrollment_id"`
	StandardSubjectID   uuid.UUID         `gorm:"column:standard_subject_id;type:uuid;not null;uniqueIndex:idx_elective"   json:"standard_subject_id"`
	StudentEnrollment   StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
	StandardSubject     StandardSubject   `gorm:"foreignKey:StandardSubjectID"   json:"standard_subject,omitempty"`
}

func (StudentElective) TableName() string { return "student_elective" }

// TeacherAssignment — unique on (employee_id, class_section_id, subject_id).
// EmployeeID must be in AcademicEmployeeCategories — service enforced.
type TeacherAssignment struct {
	Base
	EmployeeID     uuid.UUID    `gorm:"column:employee_id;type:uuid;not null;uniqueIndex:idx_ta"      json:"employee_id"`
	ClassSectionID uuid.UUID    `gorm:"column:class_section_id;type:uuid;not null;uniqueIndex:idx_ta" json:"class_section_id"`
	SubjectID      uuid.UUID    `gorm:"column:subject_id;type:uuid;not null;uniqueIndex:idx_ta"       json:"subject_id"`
	Employee       Employee     `gorm:"foreignKey:EmployeeID"     json:"employee,omitempty"`
	ClassSection   ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject        Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
}

func (TeacherAssignment) TableName() string { return "teacher_assignment" }

// TimeTable — v7 additions:
//
//	idx_tt_clash:         (class_section_id, day_of_week, start_time) — class slot conflict
//	idx_tt_teacher_clash: (employee_id, day_of_week, start_time)       — teacher double-booking
//	idx_tt_room_clash:    (room_id, day_of_week, start_time)            — room double-booking
//
// Teacher and room clash indexes added in migration 0014.
// DayOfWeek: 0=Monday … 6=Sunday.
// StartTime/EndTime normalised to 1970-01-01 at service layer.
// CHECK end_time > start_time added in migration 0018.
type TimeTable struct {
	Base
	ClassSectionID uuid.UUID  `gorm:"column:class_section_id;type:uuid;not null;uniqueIndex:idx_tt_clash"         json:"class_section_id"`
	SubjectID      uuid.UUID  `gorm:"column:subject_id;type:uuid;not null"                                         json:"subject_id"`
	EmployeeID     uuid.UUID  `gorm:"column:employee_id;type:uuid;not null;index:idx_tt_teacher_clash"             json:"employee_id"`
	DayOfWeek      int        `gorm:"column:day_of_week;not null;uniqueIndex:idx_tt_clash;index:idx_tt_teacher_clash;index:idx_tt_room_clash" json:"day_of_week"`
	StartTime      time.Time  `gorm:"column:start_time;not null;uniqueIndex:idx_tt_clash;index:idx_tt_teacher_clash;index:idx_tt_room_clash"  json:"start_time"`
	EndTime        time.Time  `gorm:"column:end_time;not null"                                                     json:"end_time"`
	RoomID         *uuid.UUID `gorm:"column:room_id;type:uuid;index:idx_tt_room_clash"                             json:"room_id,omitempty"`

	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	Employee     Employee     `gorm:"foreignKey:EmployeeID"     json:"employee,omitempty"`
	Room         *Room        `gorm:"foreignKey:RoomID"         json:"room,omitempty"`
}

func (TimeTable) TableName() string { return "time_table" }

// ============================================================
// SECTION 6 — ATTENDANCE & LEAVE
// ============================================================

type Attendance struct {
	Base
	StudentEnrollmentID uuid.UUID        `gorm:"column:student_enrollment_id;type:uuid;not null;uniqueIndex:idx_att" json:"student_enrollment_id"`
	Date                time.Time        `gorm:"column:date;type:date;not null;uniqueIndex:idx_att"                  json:"date"`
	Status              AttendanceStatus `gorm:"column:status;type:text;not null"                                    json:"status"`
	Remark              *string          `gorm:"column:remark;type:text"                                             json:"remark,omitempty"`
	// RecordedByID points to User (not Employee) to support admin-level recording.
	RecordedByID *uuid.UUID `gorm:"column:recorded_by_id;type:uuid;index" json:"recorded_by_id,omitempty"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
	RecordedBy        *User             `gorm:"foreignKey:RecordedByID"        json:"recorded_by,omitempty"`
}

func (Attendance) TableName() string { return "attendance" }

type EmployeeAttendance struct {
	Base
	EmployeeID uuid.UUID        `gorm:"column:employee_id;type:uuid;not null;uniqueIndex:idx_emp_att" json:"employee_id"`
	Date       time.Time        `gorm:"column:date;type:date;not null;uniqueIndex:idx_emp_att"        json:"date"`
	Status     AttendanceStatus `gorm:"column:status;type:text;not null"                              json:"status"`
	CheckInAt  *time.Time       `gorm:"column:check_in_at"                                            json:"check_in_at,omitempty"`
	CheckOutAt *time.Time       `gorm:"column:check_out_at"                                           json:"check_out_at,omitempty"`
	Remark     *string          `gorm:"column:remark;type:text"                                       json:"remark,omitempty"`
	Employee   Employee         `gorm:"foreignKey:EmployeeID"                                         json:"employee,omitempty"`
}

func (EmployeeAttendance) TableName() string { return "employee_attendance" }

type LeaveType struct {
	Base
	Name        string `gorm:"column:name;uniqueIndex;not null"        json:"name"`
	MaxDaysYear int    `gorm:"column:max_days_year;not null;default:0" json:"max_days_year"`
	IsPaid      bool   `gorm:"column:is_paid;default:true"             json:"is_paid"`
	IsActive    bool   `gorm:"column:is_active;default:true"           json:"is_active"`
}

func (LeaveType) TableName() string { return "leave_type" }

// LeaveBalance — UsedDays is maintained by DB trigger (migration 0019).
// When an EmployeeLeave transitions to APPROVED, the trigger increments
// UsedDays. On CANCELLED or REJECTED (if previously APPROVED), it decrements.
// PendingDays is still computed at query time (not stored).
// CHECK used_days <= total_days added in migration 0018.
type LeaveBalance struct {
	Base
	EmployeeID     uuid.UUID `gorm:"column:employee_id;type:uuid;not null;uniqueIndex:idx_leave_bal"      json:"employee_id"`
	LeaveTypeID    uuid.UUID `gorm:"column:leave_type_id;type:uuid;not null;uniqueIndex:idx_leave_bal"    json:"leave_type_id"`
	AcademicYearID uuid.UUID `gorm:"column:academic_year_id;type:uuid;not null;uniqueIndex:idx_leave_bal" json:"academic_year_id"`
	TotalDays      int       `gorm:"column:total_days;not null;default:0"                                 json:"total_days"`
	// UsedDays is maintained by DB trigger in migration 0019. Do not update directly.
	UsedDays     int          `gorm:"column:used_days;not null;default:0"  json:"used_days"`
	Employee     Employee     `gorm:"foreignKey:EmployeeID"     json:"employee,omitempty"`
	LeaveType    LeaveType    `gorm:"foreignKey:LeaveTypeID"    json:"leave_type,omitempty"`
	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
}

func (LeaveBalance) TableName() string { return "leave_balance" }

// EmployeeLeave — TotalDays is a snapshot computed at creation time.
// If from_date or to_date is corrected, service must recompute TotalDays
// and re-validate against LeaveBalance. This field is kept for display
// performance (avoids joining SchoolHoliday on every list query).
// Composite index on (employee_id, status) added in migration 0021.
type EmployeeLeave struct {
	Base
	EmployeeID  uuid.UUID   `gorm:"column:employee_id;type:uuid;not null;index:idx_emp_leave_status" json:"employee_id"`
	LeaveTypeID uuid.UUID   `gorm:"column:leave_type_id;type:uuid;not null"                          json:"leave_type_id"`
	FromDate    time.Time   `gorm:"column:from_date;type:date;not null"                              json:"from_date"`
	ToDate      time.Time   `gorm:"column:to_date;type:date;not null"                                json:"to_date"`
	TotalDays   int         `gorm:"column:total_days;not null"                                       json:"total_days"`
	Reason      string      `gorm:"column:reason;type:text;not null"                                 json:"reason"`
	Status      LeaveStatus `gorm:"column:status;type:text;default:'PENDING';index:idx_emp_leave_status" json:"status"`
	ReviewedBy  *uuid.UUID  `gorm:"column:reviewed_by;type:uuid"                                     json:"reviewed_by,omitempty"`
	ReviewNote  *string     `gorm:"column:review_note;type:text"                                     json:"review_note,omitempty"`
	ReviewedAt  *time.Time  `gorm:"column:reviewed_at"                                               json:"reviewed_at,omitempty"`
	Employee    Employee    `gorm:"foreignKey:EmployeeID"  json:"employee,omitempty"`
	LeaveType   LeaveType   `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
	Reviewer    *Employee   `gorm:"foreignKey:ReviewedBy"  json:"reviewer,omitempty"`
}

func (EmployeeLeave) TableName() string { return "employee_leave" }

// ============================================================
// SECTION 7 — ASSESSMENT
// ============================================================

// Exam — unique on (academic_year_id, name) via migration 0007.
// CHECK end_date >= start_date added in migration 0018.
type Exam struct {
	Base
	AcademicYearID uuid.UUID      `gorm:"column:academic_year_id;type:uuid;not null;index" json:"academic_year_id"`
	Name           string         `gorm:"column:name;not null"                             json:"name"`
	Description    *string        `gorm:"column:description;type:text"                     json:"description,omitempty"`
	ExamType       ExamType       `gorm:"column:exam_type;type:text;not null"              json:"exam_type"` // v7: proper type, not alias
	StartDate      time.Time      `gorm:"column:start_date;not null"                       json:"start_date"`
	EndDate        time.Time      `gorm:"column:end_date;not null"                         json:"end_date"`
	IsPublished    bool           `gorm:"column:is_published;default:false"                json:"is_published"`
	AcademicYear   AcademicYear   `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Schedules      []ExamSchedule `gorm:"foreignKey:ExamID"       json:"schedules,omitempty"`
}

func (Exam) TableName() string { return "exam" }

// ExamSchedule — unique on (exam_id, class_section_id, subject_id).
// CHECK passing_marks <= max_marks added in migration 0018.
type ExamSchedule struct {
	Base
	ExamID         uuid.UUID       `gorm:"column:exam_id;type:uuid;not null;uniqueIndex:idx_examsched"          json:"exam_id"`
	ClassSectionID uuid.UUID       `gorm:"column:class_section_id;type:uuid;not null;uniqueIndex:idx_examsched" json:"class_section_id"`
	SubjectID      uuid.UUID       `gorm:"column:subject_id;type:uuid;not null;uniqueIndex:idx_examsched"       json:"subject_id"`
	ExamDate       time.Time       `gorm:"column:exam_date;type:date;not null"                                  json:"exam_date"`
	StartTime      *time.Time      `gorm:"column:start_time"                                                    json:"start_time,omitempty"`
	EndTime        *time.Time      `gorm:"column:end_time"                                                      json:"end_time,omitempty"`
	MaxMarks       decimal.Decimal `gorm:"column:max_marks;type:decimal(6,2);not null"                          json:"max_marks"`
	PassingMarks   decimal.Decimal `gorm:"column:passing_marks;type:decimal(6,2);not null"                      json:"passing_marks"`
	RoomID         *uuid.UUID      `gorm:"column:room_id;type:uuid"                                             json:"room_id,omitempty"`

	Exam         Exam         `gorm:"foreignKey:ExamID"         json:"exam,omitempty"`
	ClassSection ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject      `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	Room         *Room        `gorm:"foreignKey:RoomID"         json:"room,omitempty"`
	Results      []ExamResult `gorm:"foreignKey:ExamScheduleID" json:"results,omitempty"`
}

func (ExamSchedule) TableName() string { return "exam_schedule" }

// ExamResult — v7 change: GPA field removed.
// GPA is derivable by joining GradeScale on (academic_year_id, grade)
// where academic_year comes from ExamSchedule → Exam. Storing it here
// created a transitive dependency and required syncing if GradeScale changed.
// Grade is set by service from GradeScale, never free text.
type ExamResult struct {
	Base
	ExamScheduleID      uuid.UUID        `gorm:"column:exam_schedule_id;type:uuid;not null;uniqueIndex:idx_examresult"      json:"exam_schedule_id"`
	StudentEnrollmentID uuid.UUID        `gorm:"column:student_enrollment_id;type:uuid;not null;uniqueIndex:idx_examresult" json:"student_enrollment_id"`
	MarksObtained       *decimal.Decimal `gorm:"column:marks_obtained;type:decimal(6,2)"                                    json:"marks_obtained,omitempty"`
	Grade               *string          `gorm:"column:grade"                                                               json:"grade,omitempty"`
	// GPA removed — join GradeScale at query time.
	Status     ExamResultStatus `gorm:"column:status;type:text;not null"       json:"status"`
	Remarks    *string          `gorm:"column:remarks;type:text"               json:"remarks,omitempty"`
	GradedByID *uuid.UUID       `gorm:"column:graded_by_id;type:uuid"          json:"graded_by_id,omitempty"`

	ExamSchedule      ExamSchedule      `gorm:"foreignKey:ExamScheduleID"      json:"exam_schedule,omitempty"`
	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
	GradedBy          *Employee         `gorm:"foreignKey:GradedByID"          json:"graded_by,omitempty"`
}

func (ExamResult) TableName() string { return "exam_result" }

type Assignment struct {
	Base
	ClassSectionID uuid.UUID        `gorm:"column:class_section_id;type:uuid;not null;index" json:"class_section_id"`
	SubjectID      uuid.UUID        `gorm:"column:subject_id;type:uuid;not null;index"       json:"subject_id"`
	AssignedByID   uuid.UUID        `gorm:"column:assigned_by_id;type:uuid;not null"         json:"assigned_by_id"`
	Title          string           `gorm:"column:title;not null"                            json:"title"`
	Description    *string          `gorm:"column:description;type:text"                     json:"description,omitempty"`
	DueDate        time.Time        `gorm:"column:due_date;not null"                         json:"due_date"`
	MaxMarks       *decimal.Decimal `gorm:"column:max_marks;type:decimal(6,2)"               json:"max_marks,omitempty"`
	Status         AssignmentStatus `gorm:"column:status;type:text;default:'DRAFT'"          json:"status"`

	ClassSection ClassSection           `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Subject      Subject                `gorm:"foreignKey:SubjectID"      json:"subject,omitempty"`
	AssignedBy   Employee               `gorm:"foreignKey:AssignedByID"   json:"assigned_by,omitempty"`
	Submissions  []AssignmentSubmission `gorm:"foreignKey:AssignmentID"   json:"submissions,omitempty"`
}

func (Assignment) TableName() string { return "assignment" }

// AssignmentSubmission — unique on (assignment_id, student_enrollment_id).
// Composite index on (assignment_id, status) added in migration 0021
// to support fast ungraded-submission queries per assignment.
type AssignmentSubmission struct {
	Base
	AssignmentID        uuid.UUID        `gorm:"column:assignment_id;type:uuid;not null;uniqueIndex:idx_submission;index:idx_sub_status"         json:"assignment_id"`
	StudentEnrollmentID uuid.UUID        `gorm:"column:student_enrollment_id;type:uuid;not null;uniqueIndex:idx_submission"                      json:"student_enrollment_id"`
	SubmittedAt         *time.Time       `gorm:"column:submitted_at"                                                                             json:"submitted_at,omitempty"`
	FileURL             *string          `gorm:"column:file_url"                                                                                 json:"file_url,omitempty"`
	Notes               *string          `gorm:"column:notes;type:text"                                                                          json:"notes,omitempty"`
	Status              SubmissionStatus `gorm:"column:status;type:text;not null;default:'MISSING';index:idx_sub_status"                         json:"status"`
	MarksAwarded        *decimal.Decimal `gorm:"column:marks_awarded;type:decimal(6,2)"                                                          json:"marks_awarded,omitempty"`
	Feedback            *string          `gorm:"column:feedback;type:text"                                                                       json:"feedback,omitempty"`

	Assignment        Assignment        `gorm:"foreignKey:AssignmentID"        json:"assignment,omitempty"`
	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
}

func (AssignmentSubmission) TableName() string { return "assignment_submission" }

// ============================================================
// SECTION 8 — FINANCE
// ============================================================

type FeeComponent struct {
	Base
	Name        string  `gorm:"column:name;uniqueIndex;not null" json:"name"`
	Description *string `gorm:"column:description"               json:"description,omitempty"`
	IsActive    bool    `gorm:"column:is_active;default:true"    json:"is_active"`
}

func (FeeComponent) TableName() string { return "fee_component" }

// FeeStructure — unique on (academic_year_id, standard_id, fee_component_id).
type FeeStructure struct {
	Base
	AcademicYearID uuid.UUID       `gorm:"column:academic_year_id;type:uuid;not null;uniqueIndex:idx_feestruct" json:"academic_year_id"`
	StandardID     uuid.UUID       `gorm:"column:standard_id;type:uuid;not null;uniqueIndex:idx_feestruct"      json:"standard_id"`
	FeeComponentID uuid.UUID       `gorm:"column:fee_component_id;type:uuid;not null;uniqueIndex:idx_feestruct" json:"fee_component_id"`
	Amount         decimal.Decimal `gorm:"column:amount;type:decimal(12,2);not null"                            json:"amount"`
	DueDate        *time.Time      `gorm:"column:due_date"                                                      json:"due_date,omitempty"`
	IsRecurring    bool            `gorm:"column:is_recurring;default:false"                                    json:"is_recurring"`

	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	Standard     Standard     `gorm:"foreignKey:StandardID"     json:"standard,omitempty"`
	FeeComponent FeeComponent `gorm:"foreignKey:FeeComponentID" json:"fee_component,omitempty"`
}

func (FeeStructure) TableName() string { return "fee_structure" }

// FeeRecord — AmountDue is a snapshot from FeeStructure at creation (documented).
// CHECKs added in migration 0018:
//
//	discount >= 0 AND discount <= amount_due
//	amount_paid >= 0 AND amount_paid <= (amount_due - discount)
type FeeRecord struct {
	Base
	StudentEnrollmentID uuid.UUID       `gorm:"column:student_enrollment_id;type:uuid;not null;index"   json:"student_enrollment_id"`
	FeeStructureID      *uuid.UUID      `gorm:"column:fee_structure_id;type:uuid"                       json:"fee_structure_id,omitempty"`
	FeeComponentID      uuid.UUID       `gorm:"column:fee_component_id;type:uuid;not null;index"        json:"fee_component_id"`
	AmountDue           decimal.Decimal `gorm:"column:amount_due;type:decimal(12,2);not null"           json:"amount_due"`
	Discount            decimal.Decimal `gorm:"column:discount;type:decimal(12,2);default:0"            json:"discount"`
	AmountPaid          decimal.Decimal `gorm:"column:amount_paid;type:decimal(12,2);default:0"         json:"amount_paid"`
	DueDate             time.Time       `gorm:"column:due_date;not null"                                json:"due_date"`
	PaidDate            *time.Time      `gorm:"column:paid_date"                                        json:"paid_date,omitempty"`
	Status              FeeStatus       `gorm:"column:status;type:text;default:'PENDING'"               json:"status"`
	TransactionRef      *string         `gorm:"column:transaction_ref"                                  json:"transaction_ref,omitempty"`
	CollectedByID       *uuid.UUID      `gorm:"column:collected_by_id;type:uuid"                        json:"collected_by_id,omitempty"`
	Remarks             *string         `gorm:"column:remarks;type:text"                                json:"remarks,omitempty"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
	FeeStructure      *FeeStructure     `gorm:"foreignKey:FeeStructureID"      json:"fee_structure,omitempty"`
	FeeComponent      FeeComponent      `gorm:"foreignKey:FeeComponentID"      json:"fee_component,omitempty"`
	CollectedBy       *Employee         `gorm:"foreignKey:CollectedByID"       json:"collected_by,omitempty"`
}

func (FeeRecord) TableName() string { return "fee_record" }

// SalaryStructure — append-only history.
// v7: unique on (employee_id, effective_from) prevents two slabs with the
// same start date for the same employee, which would make the service
// MAX(effective_from) query ambiguous.
type SalaryStructure struct {
	Base
	EmployeeID     uuid.UUID       `gorm:"column:employee_id;type:uuid;not null;uniqueIndex:idx_sal_struct_eff" json:"employee_id"`
	BasicSalary    decimal.Decimal `gorm:"column:basic_salary;type:decimal(12,2);not null"                      json:"basic_salary"`
	HRA            decimal.Decimal `gorm:"column:hra;type:decimal(12,2);default:0"                              json:"hra"`
	DA             decimal.Decimal `gorm:"column:da;type:decimal(12,2);default:0"                               json:"da"`
	OtherAllowance decimal.Decimal `gorm:"column:other_allowance;type:decimal(12,2);default:0"                  json:"other_allowance"`
	PF             decimal.Decimal `gorm:"column:pf;type:decimal(12,2);default:0"                               json:"pf"`
	ESI            decimal.Decimal `gorm:"column:esi;type:decimal(12,2);default:0"                              json:"esi"`
	TDS            decimal.Decimal `gorm:"column:tds;type:decimal(12,2);default:0"                              json:"tds"`
	OtherDeduction decimal.Decimal `gorm:"column:other_deduction;type:decimal(12,2);default:0"                  json:"other_deduction"`
	// v7: uniqueIndex:idx_sal_struct_eff ensures only one slab per employee per date.
	EffectiveFrom time.Time `gorm:"column:effective_from;not null;uniqueIndex:idx_sal_struct_eff" json:"effective_from"`
	Remarks       *string   `gorm:"column:remarks;type:text"                                      json:"remarks,omitempty"`
	Employee      Employee  `gorm:"foreignKey:EmployeeID"                                         json:"employee,omitempty"`
}

func (SalaryStructure) TableName() string { return "salary_structure" }

// SalaryRecord — all components snapshotted at pay time.
// Unique on (employee_id, month, year).
// GrossSalary = BasicSalary + HRA + DA + OtherAllowance (CHECK in migration 0001)
// TotalDeduction = PF + ESI + TDS + OtherDeduction (CHECK in migration 0001)
// NetSalary = GrossSalary - TotalDeduction (CHECK in migration 0001)
type SalaryRecord struct {
	Base
	EmployeeID     uuid.UUID       `gorm:"column:employee_id;type:uuid;not null;uniqueIndex:idx_salary_emp_month" json:"employee_id"`
	AcademicYearID uuid.UUID       `gorm:"column:academic_year_id;type:uuid;not null;index"                       json:"academic_year_id"`
	Month          int             `gorm:"column:month;not null;uniqueIndex:idx_salary_emp_month"                 json:"month"`
	Year           int             `gorm:"column:year;not null;uniqueIndex:idx_salary_emp_month"                  json:"year"`
	WorkingDays    int             `gorm:"column:working_days;not null;default:0"                                 json:"working_days"`
	PresentDays    int             `gorm:"column:present_days;not null;default:0"                                 json:"present_days"`
	BasicSalary    decimal.Decimal `gorm:"column:basic_salary;type:decimal(12,2);not null"                        json:"basic_salary"`
	HRA            decimal.Decimal `gorm:"column:hra;type:decimal(12,2);default:0"                                json:"hra"`
	DA             decimal.Decimal `gorm:"column:da;type:decimal(12,2);default:0"                                 json:"da"`
	OtherAllowance decimal.Decimal `gorm:"column:other_allowance;type:decimal(12,2);default:0"                    json:"other_allowance"`
	GrossSalary    decimal.Decimal `gorm:"column:gross_salary;type:decimal(12,2);not null"                        json:"gross_salary"`
	PF             decimal.Decimal `gorm:"column:pf;type:decimal(12,2);default:0"                                 json:"pf"`
	ESI            decimal.Decimal `gorm:"column:esi;type:decimal(12,2);default:0"                                json:"esi"`
	TDS            decimal.Decimal `gorm:"column:tds;type:decimal(12,2);default:0"                                json:"tds"`
	OtherDeduction decimal.Decimal `gorm:"column:other_deduction;type:decimal(12,2);default:0"                    json:"other_deduction"`
	TotalDeduction decimal.Decimal `gorm:"column:total_deduction;type:decimal(12,2);default:0"                    json:"total_deduction"`
	NetSalary      decimal.Decimal `gorm:"column:net_salary;type:decimal(12,2);not null"                          json:"net_salary"`
	PaidAmount     decimal.Decimal `gorm:"column:paid_amount;type:decimal(12,2);default:0"                        json:"paid_amount"`
	PaidDate       *time.Time      `gorm:"column:paid_date"                                                       json:"paid_date,omitempty"`
	Status         SalaryStatus    `gorm:"column:status;type:text;default:'PENDING'"                              json:"status"`
	TransactionRef *string         `gorm:"column:transaction_ref"                                                 json:"transaction_ref,omitempty"`
	Remarks        *string         `gorm:"column:remarks;type:text"                                               json:"remarks,omitempty"`

	Employee     Employee     `gorm:"foreignKey:EmployeeID"     json:"employee,omitempty"`
	AcademicYear AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
}

func (SalaryRecord) TableName() string { return "salary_record" }

// ============================================================
// SECTION 9 — COMMUNICATION & EVENTS
// ============================================================

// Notice — if ClassSectionID is set, Audience MUST be CLASS.
// DB CHECK: (audience = 'CLASS') = (class_section_id IS NOT NULL) added in migration 0018.
type Notice struct {
	Base
	Title          string         `gorm:"column:title;not null"                   json:"title"`
	Content        string         `gorm:"column:content;type:text;not null"        json:"content"`
	Audience       NoticeAudience `gorm:"column:audience;type:text;not null"       json:"audience"`
	ClassSectionID *uuid.UUID     `gorm:"column:class_section_id;type:uuid;index"  json:"class_section_id,omitempty"`
	PublishedAt    *time.Time     `gorm:"column:published_at"                      json:"published_at,omitempty"`
	ExpiresAt      *time.Time     `gorm:"column:expires_at"                        json:"expires_at,omitempty"`
	IsPublished    bool           `gorm:"column:is_published;default:false"        json:"is_published"`
	AuthorID       uuid.UUID      `gorm:"column:author_id;type:uuid;not null"      json:"author_id"`

	ClassSection *ClassSection `gorm:"foreignKey:ClassSectionID" json:"class_section,omitempty"`
	Author       User          `gorm:"foreignKey:AuthorID"       json:"author,omitempty"`
}

func (Notice) TableName() string { return "notice" }

// Announcement — v7 new model. Addresses the PermAnnounceSend permission
// which had no backing table in v6. Announcements are one-way broadcasts
// (no reply). Read receipts are tracked in AnnouncementRead.
type Announcement struct {
	Base
	Title       string               `gorm:"column:title;not null"                json:"title"`
	Body        string               `gorm:"column:body;type:text;not null"       json:"body"`
	Audience    AnnouncementAudience `gorm:"column:audience;type:text;not null"   json:"audience"`
	PublishedAt *time.Time           `gorm:"column:published_at"                  json:"published_at,omitempty"`
	ExpiresAt   *time.Time           `gorm:"column:expires_at"                    json:"expires_at,omitempty"`
	IsPublished bool                 `gorm:"column:is_published;default:false"    json:"is_published"`
	AuthorID    uuid.UUID            `gorm:"column:author_id;type:uuid;not null"  json:"author_id"`
	// AttachmentURL for optional attached document/image.
	AttachmentURL *string `gorm:"column:attachment_url" json:"attachment_url,omitempty"`

	Author User               `gorm:"foreignKey:AuthorID"     json:"author,omitempty"`
	Reads  []AnnouncementRead `gorm:"foreignKey:AnnouncementID" json:"reads,omitempty"`
}

func (Announcement) TableName() string { return "announcement" }

// AnnouncementRead — v7 new model. Tracks per-user read receipts for announcements.
// Hard delete (BaseJunction) since there's no meaningful soft-delete of a read event.
type AnnouncementRead struct {
	BaseJunction
	AnnouncementID uuid.UUID    `gorm:"column:announcement_id;type:uuid;not null;uniqueIndex:idx_ann_read" json:"announcement_id"`
	UserID         uuid.UUID    `gorm:"column:user_id;type:uuid;not null;uniqueIndex:idx_ann_read"         json:"user_id"`
	Announcement   Announcement `gorm:"foreignKey:AnnouncementID" json:"announcement,omitempty"`
	User           User         `gorm:"foreignKey:UserID"         json:"user,omitempty"`
}

func (AnnouncementRead) TableName() string { return "announcement_read" }

// MessageThread — two-party thread.
// Limitation: group messaging (e.g., teacher to class) requires N threads.
// If group messaging is required in future, introduce MessageParticipant junction
// and replace InitiatorID/RecipientID with a participants relationship.
type MessageThread struct {
	Base
	Subject           string    `gorm:"column:subject;not null"                  json:"subject"`
	InitiatorID       uuid.UUID `gorm:"column:initiator_id;type:uuid;not null"   json:"initiator_id"`
	RecipientID       uuid.UUID `gorm:"column:recipient_id;type:uuid;not null"   json:"recipient_id"`
	InitiatorArchived bool      `gorm:"column:initiator_archived;default:false"  json:"initiator_archived"`
	RecipientArchived bool      `gorm:"column:recipient_archived;default:false"  json:"recipient_archived"`

	Initiator User      `gorm:"foreignKey:InitiatorID" json:"initiator,omitempty"`
	Recipient User      `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Messages  []Message `gorm:"foreignKey:ThreadID"    json:"messages,omitempty"`
}

func (MessageThread) TableName() string { return "message_thread" }

// Message — composite index on (thread_id, created_at DESC) in migration 0005.
// IsRead semantics: in a two-party thread, each message has exactly one
// recipient (the party who did not send it). IsRead = true means that
// recipient has read the message. ReadAt records when.
type Message struct {
	Base
	ThreadID uuid.UUID  `gorm:"column:thread_id;type:uuid;not null;index" json:"thread_id"`
	SenderID uuid.UUID  `gorm:"column:sender_id;type:uuid;not null"        json:"sender_id"`
	Body     string     `gorm:"column:body;type:text;not null"             json:"body"`
	IsRead   bool       `gorm:"column:is_read;default:false"               json:"is_read"`
	ReadAt   *time.Time `gorm:"column:read_at"                             json:"read_at,omitempty"`
	FileURL  *string    `gorm:"column:file_url"                            json:"file_url,omitempty"`

	Thread MessageThread `gorm:"foreignKey:ThreadID" json:"thread,omitempty"`
	Sender User          `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

func (Message) TableName() string { return "message" }

// PromotionRecord — v7: FromStandardID removed (3NF violation).
// It was derivable from StudentEnrollment → ClassSection → Standard.
// Queries requiring FromStandard should join through StudentEnrollment.
type PromotionRecord struct {
	Base
	StudentEnrollmentID uuid.UUID  `gorm:"column:student_enrollment_id;type:uuid;not null;uniqueIndex" json:"student_enrollment_id"`
	FromAcademicYearID  uuid.UUID  `gorm:"column:from_academic_year_id;type:uuid;not null"             json:"from_academic_year_id"`
	ToAcademicYearID    *uuid.UUID `gorm:"column:to_academic_year_id;type:uuid"                        json:"to_academic_year_id,omitempty"`
	// FromStandardID removed — use StudentEnrollment → ClassSection → Standard.
	ToStandardID  *uuid.UUID      `gorm:"column:to_standard_id;type:uuid"                             json:"to_standard_id,omitempty"`
	Status        PromotionStatus `gorm:"column:status;type:text;not null"                            json:"status"`
	Remarks       *string         `gorm:"column:remarks;type:text"                                    json:"remarks,omitempty"`
	ProcessedByID uuid.UUID       `gorm:"column:processed_by_id;type:uuid;not null"                   json:"processed_by_id"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
	FromAcademicYear  AcademicYear      `gorm:"foreignKey:FromAcademicYearID"  json:"from_academic_year,omitempty"`
	ToAcademicYear    *AcademicYear     `gorm:"foreignKey:ToAcademicYearID"    json:"to_academic_year,omitempty"`
	ToStandard        *Standard         `gorm:"foreignKey:ToStandardID"        json:"to_standard,omitempty"`
	ProcessedBy       User              `gorm:"foreignKey:ProcessedByID"       json:"processed_by,omitempty"`
}

func (PromotionRecord) TableName() string { return "promotion_record" }

type Event struct {
	Base
	AcademicYearID *uuid.UUID `gorm:"column:academic_year_id;type:uuid;index"       json:"academic_year_id,omitempty"`
	Title          string     `gorm:"column:title;not null"                         json:"title"`
	Description    *string    `gorm:"column:description;type:text"                  json:"description,omitempty"`
	StartAt        time.Time  `gorm:"column:start_at;not null"                      json:"start_at"`
	EndAt          time.Time  `gorm:"column:end_at;not null"                        json:"end_at"`
	Location       *string    `gorm:"column:location"                               json:"location,omitempty"`
	IsPublic       bool       `gorm:"column:is_public;default:true"                 json:"is_public"`
	CreatedByID    uuid.UUID  `gorm:"column:created_by_id;type:uuid;not null"       json:"created_by_id"`

	AcademicYear *AcademicYear `gorm:"foreignKey:AcademicYearID" json:"academic_year,omitempty"`
	CreatedBy    User          `gorm:"foreignKey:CreatedByID"    json:"created_by,omitempty"`
}

func (Event) TableName() string { return "event" }

// ============================================================
// SECTION 10 — INFRASTRUCTURE
// ============================================================

// LibraryBook — v7 change: Status column removed.
// Status was a manually maintained enum that inevitably drifted from the
// actual state of LibraryIssue records. For multi-copy books, a single
// status field cannot represent partial availability.
// AvailableCopies is maintained by the DB trigger in migration 0020.
// Query: a book is available when available_copies > 0.
type LibraryBook struct {
	Base
	ISBN        *string `gorm:"column:isbn;uniqueIndex"                         json:"isbn,omitempty"`
	Title       string  `gorm:"column:title;not null"                           json:"title"`
	Author      string  `gorm:"column:author;not null"                          json:"author"`
	Publisher   *string `gorm:"column:publisher"                                json:"publisher,omitempty"`
	Category    string  `gorm:"column:category;not null;index"                  json:"category"`
	TotalCopies int     `gorm:"column:total_copies;not null;default:1"          json:"total_copies"`
	// AvailableCopies maintained by trigger in migration 0020. Do not update directly.
	AvailableCopies int            `gorm:"column:available_copies;not null;default:1"      json:"available_copies"`
	Issues          []LibraryIssue `gorm:"foreignKey:BookID"                               json:"issues,omitempty"`
}

func (LibraryBook) TableName() string { return "library_book" }

// LibraryFineRate — v7 new model. Provides the configurable per-day fine rate
// used when computing LibraryIssue.FineAmount. Without this, FineAmount
// cannot be independently verified or recalculated if the rate changes.
// Only one row may be active at a time (partial unique via migration).
type LibraryFineRate struct {
	Base
	RatePerDay    decimal.Decimal `gorm:"column:rate_per_day;type:decimal(8,2);not null" json:"rate_per_day"`
	EffectiveFrom time.Time       `gorm:"column:effective_from;not null;uniqueIndex"     json:"effective_from"`
	Remarks       *string         `gorm:"column:remarks"                                 json:"remarks,omitempty"`
}

func (LibraryFineRate) TableName() string { return "library_fine_rate" }

// LibraryIssue — FineAmount computed from LibraryFineRate at return time.
// due_date > issued_at CHECK added in migration 0018.
type LibraryIssue struct {
	Base
	BookID     uuid.UUID       `gorm:"column:book_id;type:uuid;not null;index"        json:"book_id"`
	IssuedToID uuid.UUID       `gorm:"column:issued_to_id;type:uuid;not null;index"   json:"issued_to_id"`
	IssuedAt   time.Time       `gorm:"column:issued_at;not null"                      json:"issued_at"`
	DueDate    time.Time       `gorm:"column:due_date;not null"                       json:"due_date"`
	ReturnedAt *time.Time      `gorm:"column:returned_at"                             json:"returned_at,omitempty"`
	FinePaid   bool            `gorm:"column:fine_paid;default:false"                 json:"fine_paid"`
	FineAmount decimal.Decimal `gorm:"column:fine_amount;type:decimal(8,2);default:0" json:"fine_amount"`

	Book     LibraryBook `gorm:"foreignKey:BookID"     json:"book,omitempty"`
	IssuedTo User        `gorm:"foreignKey:IssuedToID" json:"issued_to,omitempty"`
}

func (LibraryIssue) TableName() string { return "library_issue" }

// TransportRoute — v7: vehicle_number now has uniqueIndex.
// A vehicle can only serve one route at a time.
// DB trigger enforces category = 'DRIVER' for DriverID (migration 0003).
type TransportRoute struct {
	Base
	RouteName string `gorm:"column:route_name;uniqueIndex;not null"         json:"route_name"`
	// VehicleNumber uniqueIndex (v7): prevents same vehicle on two routes.
	VehicleNumber string     `gorm:"column:vehicle_number;not null;uniqueIndex"      json:"vehicle_number"`
	DriverID      *uuid.UUID `gorm:"column:driver_id;type:uuid"                     json:"driver_id,omitempty"`
	IsActive      bool       `gorm:"column:is_active;default:true"                  json:"is_active"`

	Driver   *Employee          `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Students []StudentTransport `gorm:"foreignKey:RouteID"  json:"students,omitempty"`
}

func (TransportRoute) TableName() string { return "transport_route" }

type StudentTransport struct {
	BaseJunction
	StudentEnrollmentID uuid.UUID `gorm:"column:student_enrollment_id;type:uuid;not null;uniqueIndex" json:"student_enrollment_id"`
	RouteID             uuid.UUID `gorm:"column:route_id;type:uuid;not null;index"                    json:"route_id"`
	PickupPoint         *string   `gorm:"column:pickup_point"                                         json:"pickup_point,omitempty"`

	StudentEnrollment StudentEnrollment `gorm:"foreignKey:StudentEnrollmentID" json:"student_enrollment,omitempty"`
	Route             TransportRoute    `gorm:"foreignKey:RouteID"             json:"route,omitempty"`
}

func (StudentTransport) TableName() string { return "student_transport" }

// ============================================================
// SECTION 11 — AUTO-MIGRATE ORDER
//
// AutoMigrate creates tables, columns, and indexes declared in GORM tags.
// It does NOT create:
//   • Partial / conditional unique indexes  → raw migrations required
//   • CHECK constraints                     → raw migrations required
//   • DB triggers                           → raw migrations required
//
// INHERITED FROM v6 (run first):
//   0001_salary_record_checks.sql       — SalaryRecord arithmetic CHECKs
//   0002_elective_capacity_trigger.sql  — Elective seat capacity trigger
//   0003_driver_category_trigger.sql    — Driver category trigger
//   0004_timetable_time_check.sql       — TimeTable epoch-date CHECKs
//   0005_message_thread_index.sql       — (thread_id, created_at DESC) index
//   0006_active_academic_year.sql       — Partial unique on is_active=true
//   0007_exam_name_unique.sql           — Partial unique on (academic_year_id, name)
//   0008_user_persona_check.sql         — User persona exclusivity CHECK
//   0009_fee_record_indexes.sql         — fee_component_id + status indexes
//   0010_grade_scale_name_unique.sql    — SUPERSEDED by migration 0012
//
// NEW IN v7:
//   0011_rename_reserved_tables.sql     — user→users, role→roles (BREAKING)
//   0012_grade_scale_fix.sql            — fix (academic_year_id, grade) unique
//   0013_userscope_null_fix.sql         — COALESCE-based unique for scope_id
//   0014_timetable_conflicts.sql        — teacher + room clash indexes
//   0015_enrollment_rollnumber.sql      — (class_section_id, roll_number) unique
//   0016_salary_structure_unique.sql    — (employee_id, effective_from) unique
//   0017_transport_vehicle_unique.sql   — vehicle_number unique
//   0018_check_constraints.sql         — date, marks, discount, leave CHECKs
//   0019_leave_balance_trigger.sql      — sync used_days on leave approval
//   0020_library_trigger.sql           — available_copies counter trigger
//   0021_missing_indexes.sql           — expires_at, resource_type+id, etc.
//   0022_status_checks.sql             — CHECK constraints for all enum columns
//   0023_standard_unique.sql           — (department_id, name) unique
// ============================================================

func AllModels() []any {
	return []any{
		// Auth & RBAC (Permission and Role before RolePermission and User)
		&Permission{}, &Role{}, &RolePermission{}, &RoleChangeLog{},
		&User{}, &UserRefreshToken{}, &UserScope{}, &AuditLog{},

		// People
		&Employee{}, &Student{}, &Parent{}, &StudentParent{},
		&StudentDocument{}, &StudentHealth{},

		// Academic structure
		&Department{}, &Standard{}, &Subject{}, &StandardSubject{},
		&AcademicYear{}, &SchoolHoliday{}, &GradeScale{},

		// Classes & scheduling
		&Room{}, &ClassSection{}, &ClassSectionElectiveSlot{},
		&StudentEnrollment{}, &StudentElective{},
		&TeacherAssignment{}, &TimeTable{},

		// Attendance & leave
		&Attendance{}, &EmployeeAttendance{},
		&LeaveType{}, &LeaveBalance{}, &EmployeeLeave{},

		// Assessment
		&Exam{}, &ExamSchedule{}, &ExamResult{},
		&Assignment{}, &AssignmentSubmission{},

		// Finance
		&FeeComponent{}, &FeeStructure{}, &FeeRecord{},
		&SalaryStructure{}, &SalaryRecord{},

		// Communication & events
		&Notice{}, &Announcement{}, &AnnouncementRead{},
		&MessageThread{}, &Message{},
		&PromotionRecord{}, &Event{},

		// Infrastructure
		&LibraryBook{}, &LibraryFineRate{}, &LibraryIssue{},
		&TransportRoute{}, &StudentTransport{},
	}
}
