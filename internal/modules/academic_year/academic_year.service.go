package academic_year

import (
	"context"
	"errors"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*AcademicYearResponse, error)
	GetByID(ctx context.Context, id string) (*AcademicYearResponse, error)
	GetActive(ctx context.Context) (*AcademicYearResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYearResponse, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*AcademicYearResponse, error)
	SetActive(ctx context.Context, id string) (*AcademicYearResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*AcademicYearResponse, error) {
	// Guard: end date must be after start date
	if !req.EndDate.After(req.StartDate) {
		return nil, errors.New("academic_year.Service.Create: end_date must be after start_date")
	}

	// Guard: name must be unique
	isDup, err := s.repo.IsDuplicateName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("academic_year.Service.Create.IsDuplicateName: %w", err)
	}
	if isDup {
		return nil, fmt.Errorf("academic_year.Service.Create: academic year %q already exists", req.Name)
	}

	a := &AcademicYear{
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		IsActive:  false, // always inactive on creation — use SetActive explicitly
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("academic_year.Service.Create: %w", err)
	}
	return ToAcademicYearResponse(a), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*AcademicYearResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("academic_year.Service.GetByID: %w", err)
	}
	return ToAcademicYearResponse(a), nil
}

func (s *service) GetActive(ctx context.Context) (*AcademicYearResponse, error) {
	a, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("academic_year.Service.GetActive: no active academic year found")
	}
	return ToAcademicYearResponse(a), nil
}

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYearResponse, error) {
	years, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("academic_year.Service.List: %w", err)
	}
	responses := make([]*AcademicYearResponse, len(years))
	for i, a := range years {
		responses[i] = ToAcademicYearResponse(a)
	}
	return responses, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*AcademicYearResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("academic_year.Service.Update.GetByID: %w", err)
	}

	if req.Name != nil && *req.Name != a.Name {
		isDup, err := s.repo.IsDuplicateName(ctx, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("academic_year.Service.Update.IsDuplicateName: %w", err)
		}
		if isDup {
			return nil, fmt.Errorf("academic_year.Service.Update: name %q is already in use", *req.Name)
		}
		a.Name = *req.Name
	}
	if req.StartDate != nil {
		a.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		a.EndDate = *req.EndDate
	}

	// Re-validate date range after applying changes
	if !a.EndDate.After(a.StartDate) {
		return nil, errors.New("academic_year.Service.Update: end_date must be after start_date")
	}

	if err := s.repo.Update(ctx, id, a); err != nil {
		return nil, fmt.Errorf("academic_year.Service.Update.Save: %w", err)
	}
	return ToAcademicYearResponse(a), nil
}

// SetActive marks the given year as active and deactivates all others.
// This is the intended way to switch academic years — the UI calls this
// when the user changes the global academic year selector.
func (s *service) SetActive(ctx context.Context, id string) (*AcademicYearResponse, error) {
	// Confirm the year exists before touching anything
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("academic_year.Service.SetActive.GetByID: %w", err)
	}
	if a.IsActive {
		return ToAcademicYearResponse(a), nil // already active, no-op
	}

	if err := s.repo.SetActive(ctx, id); err != nil {
		return nil, fmt.Errorf("academic_year.Service.SetActive: %w", err)
	}

	a.IsActive = true
	return ToAcademicYearResponse(a), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("academic_year.Service.Delete.GetByID: %w", err)
	}
	if a.IsActive {
		return errors.New("academic_year.Service.Delete: cannot delete the active academic year — set another year active first")
	}

	hasSections, err := s.repo.HasClassSections(ctx, id)
	if err != nil {
		return fmt.Errorf("academic_year.Service.Delete.HasClassSections: %w", err)
	}
	if hasSections {
		return errors.New("academic_year.Service.Delete: cannot delete academic year with existing class sections")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("academic_year.Service.Delete: %w", err)
	}
	return nil
}
