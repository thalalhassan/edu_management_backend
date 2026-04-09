package teacher

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
)

type Service interface {
	GetByID(ctx context.Context, id string) (*TeacherResponse, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*TeacherResponse, error)
	List(ctx context.Context, p pagination.Params) ([]*TeacherResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*TeacherResponse, error)
	SetActive(ctx context.Context, id string, active bool) (*TeacherResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (*TeacherResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("teacher.Service.GetByID: %w", err)
	}
	return ToTeacherResponse(t), nil
}

func (s *service) GetByEmployeeID(ctx context.Context, employeeID string) (*TeacherResponse, error) {
	t, err := s.repo.GetByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("teacher.Service.GetByEmployeeID: %w", err)
	}
	return ToTeacherResponse(t), nil
}

func (s *service) List(ctx context.Context, p pagination.Params) ([]*TeacherResponse, int64, error) {
	teachers, total, err := s.repo.FindAll(ctx, p)
	if err != nil {
		return nil, 0, fmt.Errorf("teacher.Service.List: %w", err)
	}
	responses := make([]*TeacherResponse, len(teachers))
	for i, t := range teachers {
		responses[i] = ToTeacherResponse(t)
	}
	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*TeacherResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("teacher.Service.Update.GetByID: %w", err)
	}

	if req.FirstName != nil {
		t.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		t.LastName = *req.LastName
	}
	if req.Gender != nil {
		t.Gender = *req.Gender
	}
	if req.DOB != nil {
		t.DOB = req.DOB
	}
	if req.Phone != nil {
		t.Phone = req.Phone
	}
	if req.Address != nil {
		t.Address = req.Address
	}
	if req.Qualification != nil {
		t.Qualification = req.Qualification
	}
	if req.Specialization != nil {
		t.Specialization = req.Specialization
	}
	if req.JoiningDate != nil {
		t.JoiningDate = *req.JoiningDate
	}
	if req.PhotoURL != nil {
		t.PhotoURL = req.PhotoURL
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, id, t); err != nil {
		return nil, fmt.Errorf("teacher.Service.Update.Save: %w", err)
	}
	return ToTeacherResponse(t), nil
}

// SetActive is a dedicated endpoint for activating/deactivating a teacher
// without requiring a full update payload.
func (s *service) SetActive(ctx context.Context, id string, active bool) (*TeacherResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("teacher.Service.SetActive.GetByID: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, id, active); err != nil {
		return nil, fmt.Errorf("teacher.Service.SetActive.Save: %w", err)
	}

	return ToTeacherResponse(t), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	// Soft-delete only — GORM respects DeletedAt.
	// Hard-deleting a teacher would orphan TeacherAssignment and TimeTable rows.
	// Prefer SetActive(false) for deactivation; Delete for permanent removal after cleanup.
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("teacher.Service.Delete: %w", err)
	}
	return nil
}

// ensure database import is used via the alias in domain.go
var _ = database.GenderMale
