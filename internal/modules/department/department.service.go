package department

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*DepartmentResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*DepartmentResponse, error)
	List(ctx context.Context) ([]*DepartmentResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*DepartmentResponse, error)
	AssignHead(ctx context.Context, id uuid.UUID, req AssignHeadRequest) (*DepartmentResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*DepartmentResponse, error) {
	// Guard: code must be unique
	existing, _ := s.repo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("department.Service.Create: code %q is already in use", req.Code)
	}

	d := &Department{
		Name:           req.Name,
		Code:           req.Code,
		Description:    req.Description,
		HeadEmployeeID: req.HeadEmployeeID,
		IsActive:       true,
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("department.Service.Create: %w", err)
	}

	// Reload to get HeadTeacher preloaded in response
	created, err := s.repo.GetByID(ctx, d.ID)
	if err != nil {
		return nil, fmt.Errorf("department.Service.Create.GetByID: %w", err)
	}
	return ToDepartmentResponse(created), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*DepartmentResponse, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("department.Service.GetByID: %w", err)
	}
	return ToDepartmentResponse(d), nil
}

func (s *service) List(ctx context.Context) ([]*DepartmentResponse, error) {
	departments, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("department.Service.List: %w", err)
	}
	responses := make([]*DepartmentResponse, len(departments))
	for i, d := range departments {
		responses[i] = ToDepartmentResponse(d)
	}
	return responses, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*DepartmentResponse, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("department.Service.Update.GetByID: %w", err)
	}

	// Guard: if code is changing, ensure the new code is not taken
	if req.Code != nil && *req.Code != d.Code {
		existing, _ := s.repo.GetByCode(ctx, *req.Code)
		if existing != nil {
			return nil, fmt.Errorf("department.Service.Update: code %q is already in use", *req.Code)
		}
		d.Code = *req.Code
	}
	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Description != nil {
		d.Description = req.Description
	}
	if req.HeadEmployeeID != nil {
		d.HeadEmployeeID = req.HeadEmployeeID
	}
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, id, d); err != nil {
		return nil, fmt.Errorf("department.Service.Update.Save: %w", err)
	}
	return ToDepartmentResponse(d), nil
}

// AssignHead sets or removes the head employee of a department.
// Passing nil in the request removes the current head.
func (s *service) AssignHead(ctx context.Context, id uuid.UUID, req AssignHeadRequest) (*DepartmentResponse, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("department.Service.AssignHead.GetByID: %w", err)
	}
	d.HeadEmployeeID = req.HeadEmployeeID
	if err := s.repo.Update(ctx, id, d); err != nil {
		return nil, fmt.Errorf("department.Service.AssignHead.Save: %w", err)
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("department.Service.AssignHead.Reload: %w", err)
	}
	return ToDepartmentResponse(updated), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("department.Service.Delete.GetByID: %w", err)
	}
	if len(d.Standards) > 0 {
		return errors.New("department.Service.Delete: cannot delete department with existing standards — remove or reassign standards first")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("department.Service.Delete: %w", err)
	}
	return nil
}
