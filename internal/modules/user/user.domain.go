package user

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases so the rest of the module never imports database directly.
type User = database.User
type SystemRole = database.SystemRole
type Student = database.Student
type Employee = database.Employee
type Parent = database.Parent

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
	Gender        database.Gender        `json:"gender"         binding:"required,oneof=MALE FEMALE"`
	Status        database.StudentStatus `json:"status,omitempty"`
	Phone         *string                `json:"phone,omitempty"`
	Address       *string                `json:"address,omitempty"`
	PhotoURL      *string                `json:"photo_url,omitempty"`
	AdmissionDate *time.Time             `json:"admission_date,omitempty"`
}

// CreateEmployeeUserRequest creates an Employee profile + User account.
type CreateEmployeeUserRequest struct {
	// --- User credentials ---
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// --- Employee profile ---
	EmployeeID     string                    `json:"employee_id"     binding:"required"`
	FirstName      string                    `json:"first_name"      binding:"required"`
	LastName       string                    `json:"last_name"       binding:"required"`
	Gender         database.Gender           `json:"gender"          binding:"required,oneof=MALE FEMALE"`
	Category       database.EmployeeCategory `json:"category"       binding:"required,oneof=TEACHER PRINCIPAL VICE_PRINCIPAL STAFF COUNSELOR LIBRARIAN ACCOUNTANT DRIVER NURSE"`
	DOB            *time.Time                `json:"dob,omitempty"`
	Phone          *string                   `json:"phone,omitempty"`
	Address        *string                   `json:"address,omitempty"`
	Qualification  *string                   `json:"qualification,omitempty"`
	Specialization *string                   `json:"specialization,omitempty"`
	JoiningDate    time.Time                 `json:"joining_date"    binding:"required"`
	PhotoURL       *string                   `json:"photo_url,omitempty"`
}

// CreateParentUserRequest creates a Parent profile + User account.
type CreateParentUserRequest struct {
	// --- User credentials ---
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	// --- Parent profile ---
	FirstName    string                      `json:"first_name"    binding:"required"`
	LastName     string                      `json:"last_name"     binding:"required"`
	Relationship database.ParentRelationship `json:"relationship"  binding:"required,oneof=FATHER MOTHER GUARDIAN SIBLING OTHER"`
	Phone        string                      `json:"phone"         binding:"required"`
	Address      *string                     `json:"address,omitempty"`
	Occupation   *string                     `json:"occupation,omitempty"`

	// Optional: link to existing students at creation time.
	StudentIDs []string `json:"student_ids,omitempty"`
}

// CreateAdminUserRequest creates a bare User with ADMIN / SUPER_ADMIN role —
// no profile record is created.
type CreateAdminUserRequest struct {
	Email    string     `json:"email"    binding:"required,email"`
	Password string     `json:"password" binding:"required,min=8"`
	Role     SystemRole `json:"role"     binding:"required,oneof=SUPER_ADMIN ADMIN PRINCIPAL"`
}

// ──────────────────────────────────────────────────────────────
// UPDATE / MISC
// ──────────────────────────────────────────────────────────────

// UpdateUserRequest allows toggling is_active or reassigning role.
// Profile fields are updated through their respective modules
// (student module, teacher module, etc.).
type UpdateUserRequest struct {
	IsActive *bool       `json:"is_active,omitempty"`
	Role     *SystemRole `json:"role,omitempty"`
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
	RoleSlug    string     `json:"role_slug"`
	RoleName    string     `json:"role_name"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	// Inline profile snapshot — only one will be non-nil.
	EmployeeID *string          `json:"employee_id,omitempty"`
	StudentID  *string          `json:"student_id,omitempty"`
	ParentID   *string          `json:"parent_id,omitempty"`
	Profile    *ProfileSnapshot `json:"profile,omitempty"`
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
		RoleSlug:    u.Role.Slug,
		RoleName:    u.Role.Name,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		EmployeeID:  u.EmployeeID,
		StudentID:   u.StudentID,
		ParentID:    u.ParentID,
	}

	// Populate profile snapshot from whatever relation is loaded.
	switch {
	case u.Employee != nil:
		r.Profile = &ProfileSnapshot{
			FirstName: u.Employee.FirstName,
			LastName:  u.Employee.LastName,
			Phone:     u.Employee.Phone,
			PhotoURL:  u.Employee.PhotoURL,
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
	}

	return r
}
