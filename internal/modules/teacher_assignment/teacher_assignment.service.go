package teacher_assignment

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Response, error)
	GetByID(ctx context.Context, id string) (*Response, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]Response, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*Response, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	exists, err := s.repo.ExistsConflict(ctx, req.ClassSectionID, req.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("teacherassignment.Service.Create: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("subject already assigned to this class section")
	}

	m := &database.TeacherAssignment{
		ClassSectionID: req.ClassSectionID,
		EmployeeID:     req.EmployeeID,
		SubjectID:      req.SubjectID,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("teacherassignment.Service.Create: %w", err)
	}

	return mapToResponse(m), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Response, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("teacherassignment.Service.GetByID: %w", err)
	}
	return mapToResponse(m), nil
}

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]Response, int64, error) {
	list, count, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("teacherassignment.Service.List: %w", err)
	}

	res := make([]Response, len(list))
	for i := range list {
		res[i] = *mapToResponse(&list[i])
	}

	return res, count, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*Response, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("teacherassignment.Service.Update: %w", err)
	}

	if req.EmployeeID != nil {
		m.EmployeeID = *req.EmployeeID
	}

	if err := s.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("teacherassignment.Service.Update: %w", err)
	}

	return mapToResponse(m), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("teacherassignment.Service.Delete: %w", err)
	}
	return nil
}

// mapper
func mapToResponse(m *database.TeacherAssignment) *Response {
	return &Response{
		ID:             m.ID,
		ClassSectionID: m.ClassSectionID,
		EmployeeID:     m.EmployeeID,
		SubjectID:      m.SubjectID,
		CreatedAt:      m.CreatedAt,
	}
}
