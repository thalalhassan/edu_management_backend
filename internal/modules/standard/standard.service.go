package standard

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*StandardResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*StandardResponse, error)
	List(ctx context.Context) ([]*StandardResponse, error)
	ListByDepartment(ctx context.Context, departmentID uuid.UUID) ([]*StandardResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*StandardResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// Subject management
	AssignSubject(ctx context.Context, standardID uuid.UUID, req AssignSubjectRequest) error
	RemoveSubject(ctx context.Context, standardID, subjectID uuid.UUID) error
	GetSubjects(ctx context.Context, standardID uuid.UUID) ([]*StandardSubject, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*StandardResponse, error) {
	st := &Standard{
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
		OrderIndex:   req.OrderIndex,
		Description:  req.Description,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		return nil, fmt.Errorf("standard.Service.Create: %w", err)
	}
	created, err := s.repo.GetByID(ctx, st.ID)
	if err != nil {
		return nil, fmt.Errorf("standard.Service.Create.GetByID: %w", err)
	}
	return ToStandardResponse(created), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*StandardResponse, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("standard.Service.GetByID: %w", err)
	}
	return ToStandardResponse(st), nil
}

func (s *service) List(ctx context.Context) ([]*StandardResponse, error) {
	standards, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("standard.Service.List: %w", err)
	}
	responses := make([]*StandardResponse, len(standards))
	for i, st := range standards {
		responses[i] = ToStandardResponse(st)
	}
	return responses, nil
}

func (s *service) ListByDepartment(ctx context.Context, departmentID uuid.UUID) ([]*StandardResponse, error) {
	standards, err := s.repo.FindByDepartment(ctx, departmentID)
	if err != nil {
		return nil, fmt.Errorf("standard.Service.ListByDepartment: %w", err)
	}
	responses := make([]*StandardResponse, len(standards))
	for i, st := range standards {
		responses[i] = ToStandardResponse(st)
	}
	return responses, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*StandardResponse, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("standard.Service.Update.GetByID: %w", err)
	}
	if req.Name != nil {
		st.Name = *req.Name
	}
	if req.DepartmentID != nil {
		st.DepartmentID = *req.DepartmentID
	}
	if req.OrderIndex != nil {
		st.OrderIndex = *req.OrderIndex
	}
	if req.Description != nil {
		st.Description = req.Description
	}
	if err := s.repo.Update(ctx, id, st); err != nil {
		return nil, fmt.Errorf("standard.Service.Update.Save: %w", err)
	}
	return ToStandardResponse(st), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("standard.Service.Delete.GetByID: %w", err)
	}
	if len(st.ClassSections) > 0 {
		return errors.New("standard.Service.Delete: cannot delete standard with existing class sections")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("standard.Service.Delete: %w", err)
	}
	return nil
}

func (s *service) AssignSubject(ctx context.Context, standardID uuid.UUID, req AssignSubjectRequest) error {
	// Guard: prevent duplicate assignment
	exists, err := s.repo.IsSubjectAssigned(ctx, standardID, req.SubjectID)
	if err != nil {
		return fmt.Errorf("standard.Service.AssignSubject.Check: %w", err)
	}
	if exists {
		return errors.New("standard.Service.AssignSubject: subject is already assigned to this standard")
	}

	link := &database.StandardSubject{
		StandardID:  standardID,
		SubjectID:   req.SubjectID,
		SubjectType: database.SubjectTypeCore,
	}
	if err := s.repo.AssignSubject(ctx, link); err != nil {
		return fmt.Errorf("standard.Service.AssignSubject: %w", err)
	}
	return nil
}

func (s *service) RemoveSubject(ctx context.Context, standardID, subjectID uuid.UUID) error {
	exists, err := s.repo.IsSubjectAssigned(ctx, standardID, subjectID)
	if err != nil {
		return fmt.Errorf("standard.Service.RemoveSubject.Check: %w", err)
	}
	if !exists {
		return errors.New("standard.Service.RemoveSubject: subject is not assigned to this standard")
	}
	if err := s.repo.RemoveSubject(ctx, standardID, subjectID); err != nil {
		return fmt.Errorf("standard.Service.RemoveSubject: %w", err)
	}
	return nil
}

func (s *service) GetSubjects(ctx context.Context, standardID uuid.UUID) ([]*StandardSubject, error) {
	subjects, err := s.repo.GetSubjects(ctx, standardID)
	if err != nil {
		return nil, fmt.Errorf("standard.Service.GetSubjects: %w", err)
	}
	return subjects, nil
}
