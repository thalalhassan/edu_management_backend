package staff

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/shared/pagination"
)

type Service interface {
	GetByID(ctx context.Context, id string) (*StaffResponse, error)
	List(ctx context.Context, p pagination.Params) ([]*StaffResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*StaffResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (*StaffResponse, error) {
	staff, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("staff.Service.GetByID: %w", err)
	}
	return ToStaffResponse(staff), nil
}

func (s *service) List(ctx context.Context, p pagination.Params) ([]*StaffResponse, int64, error) {
	staffs, total, err := s.repo.FindAll(ctx, p)
	if err != nil {
		return nil, 0, fmt.Errorf("staff.Service.List: %w", err)
	}

	responses := make([]*StaffResponse, len(staffs))
	for i, staff := range staffs {
		responses[i] = ToStaffResponse(staff)
	}

	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*StaffResponse, error) {
	staff, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("staff.Service.Update.GetByID: %w", err)
	}

	if req.FirstName != nil {
		staff.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		staff.LastName = *req.LastName
	}

	if err := s.repo.Update(ctx, id, staff); err != nil {
		return nil, fmt.Errorf("staff.Service.Update.Save: %w", err)
	}
	return ToStaffResponse(staff), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("staff.Service.Delete: %w", err)
	}
	return nil
}
