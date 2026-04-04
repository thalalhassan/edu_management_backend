package user

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases so the rest of the module never imports database directly.
type User = database.User
type UserRole = database.UserRole
type Student = database.Student
type Teacher = database.Teacher
type Parent = database.Parent
type Staff = database.Staff

// ──────────────────────────────────────────────────────────────
// REGISTER REQUESTS
// Each role gets its own CreateXxxRequest so validation and
// field sets remain clean and explicit.
// ──────────────────────────────────────────────────────────────

// CreateStudentUserRequest creates a Student profile + User account
// in a single atomic transaction.
type CreateStudentUserRequest struct {
	// --- User credentials ---
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// --- Student profile ---
	AdmissionNo   string                 `json:"admission_no"   binding:"required"`
	FirstName     string                 `json:"first_name"     binding:"required"`
	LastName      string                 `json:"last_name"      binding:"required"`
	DOB           time.Time              `json:"dob"            binding:"required"`
	Gender        database.Gender        `json:"gender"         binding:"required,oneof=MALE FEMALE OTHER"`
	Status        database.StudentStatus `json:"status,omitempty"`
	Phone         *string                `json:"phone,omitempty"`
	Address       *string                `json:"address,omitempty"`
	BloodGroup    *string                `json:"blood_group,omitempty"`
	PhotoURL      *string                `json:"photo_url,omitempty"`
	AdmissionDate *time.Time             `json:"admission_date,omitempty"`
}

// CreateTeacherUserRequest creates a Teacher profile + User account.
type CreateTeacherUserRequest struct {
	// --- User credentials ---
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// --- Teacher profile ---
	EmployeeID     string          `json:"employee_id"     binding:"required"`
	FirstName      string          `json:"first_name"      binding:"required"`
	LastName       string          `json:"last_name"       binding:"required"`
	Gender         database.Gender `json:"gender"          binding:"required,oneof=MALE FEMALE OTHER"`
	DOB            *time.Time      `json:"dob,omitempty"`
	Phone          *string         `json:"phone,omitempty"`
	Address        *string         `json:"address,omitempty"`
	Qualification  *string         `json:"qualification,omitempty"`
	Specialization *string         `json:"specialization,omitempty"`
	JoiningDate    time.Time       `json:"joining_date"    binding:"required"`
	PhotoURL       *string         `json:"photo_url,omitempty"`
}

// CreateParentUserRequest creates a Parent profile + User account.
type CreateParentUserRequest struct {
	// --- User credentials ---
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// --- Parent profile ---
	FirstName    string  `json:"first_name"    binding:"required"`
	LastName     string  `json:"last_name"     binding:"required"`
	Relationship string  `json:"relationship"  binding:"required"` // Father / Mother / Guardian
	Phone        string  `json:"phone"         binding:"required"`
	Address      *string `json:"address,omitempty"`
	Occupation   *string `json:"occupation,omitempty"`

	// Optional: link to existing students at creation time.
	StudentIDs []string `json:"student_ids,omitempty"`
}

// CreateStaffUserRequest creates a Staff profile + User account.
type CreateStaffUserRequest struct {
	// --- User credentials ---
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// --- Staff profile ---
	EmployeeID  string          `json:"employee_id"  binding:"required"`
	FirstName   string          `json:"first_name"   binding:"required"`
	LastName    string          `json:"last_name"    binding:"required"`
	Gender      database.Gender `json:"gender"       binding:"required,oneof=MALE FEMALE OTHER"`
	Designation string          `json:"designation"  binding:"required"`
	Phone       *string         `json:"phone,omitempty"`
	JoiningDate time.Time       `json:"joining_date" binding:"required"`
}

// CreateAdminUserRequest creates a bare User with ADMIN / SUPER_ADMIN role —
// no profile record is created.
type CreateAdminUserRequest struct {
	Email    string   `json:"email"    binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Role     UserRole `json:"role"     binding:"required,oneof=SUPER_ADMIN ADMIN PRINCIPAL"`
}

// ──────────────────────────────────────────────────────────────
// UPDATE / MISC
// ──────────────────────────────────────────────────────────────

// UpdateUserRequest allows toggling is_active or reassigning role.
// Profile fields are updated through their respective modules
// (student module, teacher module, etc.).
type UpdateUserRequest struct {
	IsActive *bool     `json:"is_active,omitempty"`
	Role     *UserRole `json:"role,omitempty"`
}

// ChangePasswordRequest is used by an authenticated user changing their
// own password, or an admin resetting someone else's.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"` // ignored for admin reset
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}

// ──────────────────────────────────────────────────────────────
// RESPONSES
// ──────────────────────────────────────────────────────────────

// UserResponse is the safe, public-facing representation of a User.
// PasswordHash is never included.
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Role        UserRole   `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	// Inline profile snapshot — only one will be non-nil.
	TeacherID *string          `json:"teacher_id,omitempty"`
	StudentID *string          `json:"student_id,omitempty"`
	ParentID  *string          `json:"parent_id,omitempty"`
	StaffID   *string          `json:"staff_id,omitempty"`
	Profile   *ProfileSnapshot `json:"profile,omitempty"`
}

// ProfileSnapshot is a minimal display-ready view of the linked profile.
type ProfileSnapshot struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone,omitempty"`
	PhotoURL  *string `json:"photo_url,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// MAPPERS
// ──────────────────────────────────────────────────────────────

func ToUserResponse(u *User) *UserResponse {
	r := &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Role:        u.Role,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		TeacherID:   u.TeacherID,
		StudentID:   u.StudentID,
		ParentID:    u.ParentID,
		StaffID:     u.StaffID,
	}

	// Populate profile snapshot from whatever relation is loaded.
	switch {
	case u.Teacher != nil:
		r.Profile = &ProfileSnapshot{
			FirstName: u.Teacher.FirstName,
			LastName:  u.Teacher.LastName,
			Phone:     u.Teacher.Phone,
			PhotoURL:  u.Teacher.PhotoURL,
		}
	case u.Student != nil:
		r.Profile = &ProfileSnapshot{
			FirstName: u.Student.FirstName,
			LastName:  u.Student.LastName,
			Phone:     u.Student.Phone,
			PhotoURL:  u.Student.PhotoURL,
		}
	case u.Parent != nil:
		r.Profile = &ProfileSnapshot{
			FirstName: u.Parent.FirstName,
			LastName:  u.Parent.LastName,
			Phone:     &u.Parent.Phone,
		}
	case u.Staff != nil:
		r.Profile = &ProfileSnapshot{
			FirstName: u.Staff.FirstName,
			LastName:  u.Staff.LastName,
			Phone:     u.Staff.Phone,
		}
	}

	return r
}
