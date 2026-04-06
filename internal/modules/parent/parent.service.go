package parent

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/shared/pagination"
)

type Service interface {
	GetByID(ctx context.Context, id string) (*ParentResponse, error)
	List(ctx context.Context, p pagination.Params) ([]*ParentResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*ParentResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (*ParentResponse, error) {
	parent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("parent.Service.GetByID: %w", err)
	}
	return ToParentResponse(parent), nil
}

func (s *service) List(ctx context.Context, p pagination.Params) ([]*ParentResponse, int64, error) {
	parents, total, err := s.repo.FindAll(ctx, p)
	if err != nil {
		return nil, 0, fmt.Errorf("parent.Service.List: %w", err)
	}

	responses := make([]*ParentResponse, len(parents))
	for i, parent := range parents {
		responses[i] = ToParentResponse(parent)
	}

	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*ParentResponse, error) {
	parent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("parent.Service.Update.GetByID: %w", err)
	}

	if req.FirstName != nil {
		parent.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		parent.LastName = *req.LastName
	}

	if err := s.repo.Update(ctx, id, parent); err != nil {
		return nil, fmt.Errorf("parent.Service.Update.Save: %w", err)
	}
	return ToParentResponse(parent), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("parent.Service.Delete: %w", err)
	}
	return nil
}
