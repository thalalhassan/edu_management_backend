package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/pkg/crypto"
)

// Service is the public contract for the user module.
type Service interface {
	// Registration — each creates a profile + user atomically
	RegisterStudent(ctx context.Context, req CreateStudentUserRequest) (*UserResponse, error)
	RegisterTeacher(ctx context.Context, req CreateTeacherUserRequest) (*UserResponse, error)
	RegisterParent(ctx context.Context, req CreateParentUserRequest) (*UserResponse, error)
	RegisterStaff(ctx context.Context, req CreateStaffUserRequest) (*UserResponse, error)
	RegisterAdmin(ctx context.Context, req CreateAdminUserRequest) (*UserResponse, error)

	// User management
	GetByID(ctx context.Context, id string) (*UserResponse, error)
	List(ctx context.Context, p pagination.Params) ([]*UserResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id string) error
	ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error
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
		BloodGroup:    req.BloodGroup,
		PhotoURL:      req.PhotoURL,
		AdmissionDate: admissionDate,
	}
	if err := s.repo.CreateStudent(ctx, tx, student); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterStudent.CreateStudent: %w", err)
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         database.UserRoleStudent,
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
	return ToUserResponse(u), nil
}

func (s *service) RegisterTeacher(ctx context.Context, req CreateTeacherUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterTeacher.Hash: %w", err)
	}

	tx := s.repo.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("user.Service.RegisterTeacher.Begin: %w", tx.Error)
	}

	teacher := &Teacher{
		EmployeeID:     req.EmployeeID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Gender:         req.Gender,
		DOB:            req.DOB,
		Phone:          req.Phone,
		Address:        req.Address,
		Qualification:  req.Qualification,
		Specialization: req.Specialization,
		JoiningDate:    req.JoiningDate,
		IsActive:       true,
		PhotoURL:       req.PhotoURL,
	}
	if err := s.repo.CreateTeacher(ctx, tx, teacher); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterTeacher.CreateTeacher: %w", err)
	}

	email := req.Email
	u := &User{
		Email:        email,
		PasswordHash: hash,
		Role:         database.UserRoleTeacher,
		IsActive:     true,
		TeacherID:    &teacher.ID,
	}
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterTeacher.CreateUser: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("user.Service.RegisterTeacher.Commit: %w", err)
	}

	u.Teacher = teacher
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

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         database.UserRoleParent,
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
	return ToUserResponse(u), nil
}

func (s *service) RegisterStaff(ctx context.Context, req CreateStaffUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterStaff.Hash: %w", err)
	}

	tx := s.repo.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("user.Service.RegisterStaff.Begin: %w", tx.Error)
	}

	staff := &Staff{
		EmployeeID:  req.EmployeeID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Gender:      req.Gender,
		Designation: req.Designation,
		Phone:       req.Phone,
		JoiningDate: req.JoiningDate,
		IsActive:    true,
	}
	if err := s.repo.CreateStaff(ctx, tx, staff); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterStaff.CreateStaff: %w", err)
	}

	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         database.UserRoleStaff,
		IsActive:     true,
		StaffID:      &staff.ID,
	}
	if err := tx.Create(u).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user.Service.RegisterStaff.CreateUser: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("user.Service.RegisterStaff.Commit: %w", err)
	}

	u.Staff = staff
	return ToUserResponse(u), nil
}

func (s *service) RegisterAdmin(ctx context.Context, req CreateAdminUserRequest) (*UserResponse, error) {
	hash, err := crypto.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RegisterAdmin.Hash: %w", err)
	}

	// Admins have no profile record — just a User row.
	u := &User{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		IsActive:     true,
	}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, fmt.Errorf("user.Service.RegisterAdmin.CreateUser: %w", err)
	}
	return ToUserResponse(u), nil
}

// ──────────────────────────────────────────────────────────────
// USER MANAGEMENT
// ──────────────────────────────────────────────────────────────

func (s *service) GetByID(ctx context.Context, id string) (*UserResponse, error) {
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

func (s *service) Update(ctx context.Context, id string, req UpdateUserRequest) (*UserResponse, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user.Service.Update.GetByID: %w", err)
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if err := s.repo.UpdateUser(ctx, id, u); err != nil {
		return nil, fmt.Errorf("user.Service.Update.Save: %w", err)
	}
	return ToUserResponse(u), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	// Revoke all sessions before hard-deleting
	// _ = s.repo.RevokeAllRefreshTokensForUser(ctx, id)
	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("user.Service.Delete: %w", err)
	}
	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
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
