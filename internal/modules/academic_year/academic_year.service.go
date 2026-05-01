package academic_year

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*AcademicYearResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AcademicYearResponse, error)
	GetActive(ctx context.Context) (*AcademicYearResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYearResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*AcademicYearResponse, error)
	SetActive(ctx context.Context, id uuid.UUID) (*AcademicYearResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*AcademicYearResponse, error) {
	a, err := NewAcademicYear(req.Name, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	if overlap, err := s.repo.HasOverlappingDates(ctx, req.StartDate, req.EndDate); err != nil {
		return nil, fmt.Errorf("failed to check overlaps: %w", err)
	} else if overlap {
		return nil, NewBusinessError("academic year dates overlap with existing year")
	}
	if dup, err := s.repo.IsDuplicateName(ctx, req.Name); err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	} else if dup {
		return nil, NewBusinessError("academic year name already exists")
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to create: %w", err)
	}
	return toResponse(a), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*AcademicYearResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "academic year", ID: id}
	}
	return toResponse(a), nil
}

func (s *service) GetActive(ctx context.Context) (*AcademicYearResponse, error) {
	a, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, &NotFoundError{Resource: "active academic year"}
	}
	return toResponse(a), nil
}

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYearResponse, error) {
	years, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}
	responses := make([]*AcademicYearResponse, len(years))
	for i, a := range years {
		responses[i] = toResponse(a)
	}
	return responses, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*AcademicYearResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "academic year", ID: id}
	}
	if err := a.Update(req); err != nil {
		return nil, err
	}
	if req.Name != nil {
		if dup, err := s.repo.IsDuplicateName(ctx, *req.Name); err != nil {
			return nil, fmt.Errorf("failed to check duplicate: %w", err)
		} else if dup {
			return nil, NewBusinessError("academic year name already exists")
		}
	}
	if err := s.repo.Update(ctx, id, a); err != nil {
		return nil, fmt.Errorf("failed to update: %w", err)
	}
	return toResponse(a), nil
}

func (s *service) SetActive(ctx context.Context, id uuid.UUID) (*AcademicYearResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "academic year", ID: id}
	}
	if a.IsActive {
		return toResponse(a), nil
	}
	if err := s.repo.SetActive(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to set active: %w", err)
	}
	a.Activate()
	return toResponse(a), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return &NotFoundError{Resource: "academic year", ID: id}
	}
	if a.IsActive {
		return NewBusinessError("cannot delete active academic year")
	}
	if hasSections, err := s.repo.HasClassSections(ctx, id); err != nil {
		return fmt.Errorf("failed to check sections: %w", err)
	} else if hasSections {
		return NewBusinessError("cannot delete academic year with class sections")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	return nil
}

func toResponse(a *AcademicYear) *AcademicYearResponse {
	return &AcademicYearResponse{
		ID:        a.ID,
		Name:      a.Name,
		StartDate: a.StartDate,
		EndDate:   a.EndDate,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt,
	}
}
