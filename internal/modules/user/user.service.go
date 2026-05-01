package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/pkg/crypto"
)

// Service is the public contract for the user module.
type Service interface {
	// Registration — each creates a profile + user atomically
	RegisterStudent(ctx context.Context, req CreateStudentUserRequest) (*UserResponse, error)
	RegisterEmployee(ctx context.Context, req CreateEmployeeUserRequest) (*UserResponse, error)
	RegisterParent(ctx context.Context, req CreateParentUserRequest) (*UserResponse, error)
	RegisterAdmin(ctx context.Context, req CreateAdminUserRequest) (*UserResponse, error)

	// User management
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	List(ctx context.Context, p pagination.Params) ([]*UserResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ──────────────────────────────────────────────────────────────
// REGISTRATION
//
// Workflow for every role:
//  1. Hash the password.
//  2. Begin a DB transaction.
//  3. Create the profile record (Student / Teacher / …).
//  4. Create the User record, setting the FK to the new profile.
//  5. Commit — if anything fails the whole thing rolls back.
//
// This ensures you never end up with an orphaned profile (no user)
// or an orphaned user (no profile).
// ──────────────────────────────────────────────────────────────

func (s *service) RegisterStudent(ctx context.Context, req CreateStudentUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterStudent.Hash: %w", err)
	}

	admissionDate := time.Now()
	if req.AdmissionDate != nil {
		admissionDate = *req.AdmissionDate
	}
	status := database.StudentStatusActive
	if req.Status != "" {
		status = req.Status
	}

	tx := s.repo.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("user.Service.RegisterStudent.Begin: %w", tx.Error)
	}

	student := &Student{
		AdmissionNo:   req.AdmissionNo,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DOB:           req.DOB,
		Gender:        req.Gender,
		Status:        status,
		Phone:         req.Phone,
		Address:       req.Address,
		PhotoURL:      req.PhotoURL,
		AdmissionDate: admissionDate,
	}
	if err := s.repo.CreateStudent(ctx, tx, student); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterStudent.CreateStudent: %w", err)
	}

	// Get or create role
	role := &database.Role{}
	if err := tx.FirstOrCreate(role, database.Role{Slug: string(database.SystemRoleStudent)}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterStudent.GetRole: %w", err)
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		RoleID:       role.ID,
		IsActive:     true,
		StudentID:    &student.ID,
	}
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterStudent.CreateUser: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("user.Service.RegisterStudent.Commit: %w", err)
	}

	u.Student = student
	u.Role = *role
	return ToUserResponse(u), nil
}

func (s *service) RegisterEmployee(ctx context.Context, req CreateEmployeeUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterEmployee.Hash: %w", err)
	}

	tx := s.repo.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("user.Service.RegisterEmployee.Begin: %w", tx.Error)
	}

	employee := &Employee{
		EmployeeCode:   req.EmployeeCode,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Gender:         req.Gender,
		Category:       req.Category,
		DOB:            req.DOB,
		Phone:          req.Phone,
		Address:        req.Address,
		Qualification:  req.Qualification,
		Specialization: req.Specialization,
		JoiningDate:    req.JoiningDate,
		IsActive:       true,
		PhotoURL:       req.PhotoURL,
	}
	if err := s.repo.CreateEmployee(ctx, tx, employee); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterEmployee.CreateEmployee: %w", err)
	}

	// Get or create role
	roleSlug := string(database.SystemRoleTeacher)
	if req.Category == database.EmployeeCategoryStaff {
		roleSlug = string(database.SystemRoleStaff)
	}
	role := &database.Role{}
	if err := tx.FirstOrCreate(role, database.Role{Slug: roleSlug}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterEmployee.GetRole: %w", err)
	}

	email := req.Email
	u := &User{
		Email:        email,
		PasswordHash: hash,
		RoleID:       role.ID,
		IsActive:     true,
		EmployeeID:   &employee.ID,
	}
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterEmployee.CreateUser: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("user.Service.RegisterEmployee.Commit: %w", err)
	}

	u.Employee = employee
	u.Role = *role
	return ToUserResponse(u), nil
}

func (s *service) RegisterParent(ctx context.Context, req CreateParentUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterParent.Hash: %w", err)
	}

	tx := s.repo.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("user.Service.RegisterParent.Begin: %w", tx.Error)
	}

	parent := &Parent{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Relationship: req.Relationship,
		Phone:        req.Phone,
		Email:        &req.Email,
		Address:      req.Address,
		Occupation:   req.Occupation,
	}
	if err := s.repo.CreateParent(ctx, tx, parent); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterParent.CreateParent: %w", err)
	}

	// Link students if provided
	for _, sid := range req.StudentIDs {
		sid := sid
		link := &database.StudentParent{
			StudentID: sid,
			ParentID:  parent.ID,
			IsPrimary: false,
		}
		if err := tx.Create(link).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("user.Service.RegisterParent.LinkStudent(%s): %w", sid, err)
		}
	}

	// Get or create role
	role := &database.Role{}
	if err := tx.FirstOrCreate(role, database.Role{Slug: string(database.SystemRoleParent)}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterParent.GetRole: %w", err)
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		RoleID:       role.ID,
		IsActive:     true,
		ParentID:     &parent.ID,
	}
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterParent.CreateUser: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("user.Service.RegisterParent.Commit: %w", err)
	}

	u.Parent = parent
	u.Role = *role
	return ToUserResponse(u), nil
}

func (s *service) RegisterAdmin(ctx context.Context, req CreateAdminUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterAdmin.Hash: %w", err)
	}

	// Get or create role
	roleSlug := string(req.Role)
	role := &database.Role{}
	if err := s.repo.DB().FirstOrCreate(role, database.Role{Slug: roleSlug}).Error; err != nil {
		return nil, fmt.Errorf("user.Service.RegisterAdmin.GetRole: %w", err)
	}

	// Admins have no profile record — just a User row.
	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		RoleID:       role.ID,
		IsActive:     true,
	}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, fmt.Errorf("user.Service.RegisterAdmin.CreateUser: %w", err)
	}

	u.Role = *role
	return ToUserResponse(u), nil
}

// ──────────────────────────────────────────────────────────────
// USER MANAGEMENT
// ──────────────────────────────────────────────────────────────

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.Service.GetByID: %w", err)
	}
	return ToUserResponse(u), nil
}

func (s *service) List(ctx context.Context, p pagination.Params) ([]*UserResponse, int64, error) {
	users, total, err := s.repo.FindAllUsers(ctx, p)
	if err != nil {
		return nil, 0, fmt.Errorf("user.Service.List: %w", err)
	}
	responses := make([]*UserResponse, len(users))
	for i, u := range users {
		responses[i] = ToUserResponse(u)
	}
	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*UserResponse, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.Service.Update.GetByID: %w", err)
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}
	if req.Role != nil {
		// Find role by slug
		role := &database.Role{}
		if err := s.repo.DB().Where("slug = ?", string(*req.Role)).First(role).Error; err != nil {
			return nil, fmt.Errorf("user.Service.Update.FindRole: %w", err)
		}
		u.RoleID = role.ID
		u.Role = *role
	}
	if err := s.repo.UpdateUser(ctx, id, u); err != nil {
		return nil, fmt.Errorf("user.Service.Update.Save: %w", err)
	}
	return ToUserResponse(u), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	// Revoke all sessions before hard-deleting
	// _ = s.repo.RevokeAllRefreshTokensForUser(ctx, id)
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("user.Service.Delete: %w", err)
	}
	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user.Service.ChangePassword.GetByID: %w", err)
	}
	if !crypto.CheckHash(req.CurrentPassword, u.PasswordHash) {
		return errors.New("user.Service.ChangePassword: current password is incorrect")
	}

	hash, err := crypto.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("user.Service.ChangePassword.Hash: %w", err)
	}
	u.PasswordHash = hash

	if err := s.repo.UpdateUser(ctx, userID, u); err != nil {
		return fmt.Errorf("user.Service.ChangePassword.Update: %w", err)
	}
	// Invalidate all sessions after a password change
	// _ = s.repo.RevokeAllRefreshTokensForUser(ctx, userID)
	return nil
}
