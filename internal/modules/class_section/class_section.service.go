package class_section

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*ClassSectionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ClassSectionResponse, error)
	ListByAcademicYear(ctx context.Context, academicYearID uuid.UUID) ([]*ClassSectionSummary, error)
	ListByStandard(ctx context.Context, standardID, academicYearID uuid.UUID) ([]*ClassSectionSummary, error)
	ListByEmployee(ctx context.Context, employeeID, academicYearID uuid.UUID) ([]*ClassSectionSummary, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*ClassSectionResponse, error)
	AssignEmployee(ctx context.Context, id uuid.UUID, req AssignEmployeeRequest) (*ClassSectionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*ClassSectionResponse, error) {
	// Guard: section name must be unique within the same standard + academic year
	isDup, err := s.repo.IsDuplicate(ctx, req.AcademicYearID, req.StandardID, req.SectionName)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.Create.IsDuplicate: %w", err)
	}
	if isDup {
		return nil, fmt.Errorf("classsection.Service.Create: section %q already exists for this standard and academic year", req.SectionName)
	}

	maxStrength := req.MaxStrength
	if maxStrength == 0 {
		maxStrength = 40 // default
	}

	cs := &ClassSection{
		AcademicYearID:  req.AcademicYearID,
		StandardID:      req.StandardID,
		SectionName:     req.SectionName,
		ClassEmployeeID: req.ClassEmployeeID,
		RoomID:          req.RoomID,
		MaxStrength:     maxStrength,
	}
	if err := s.repo.Create(ctx, cs); err != nil {
		return nil, fmt.Errorf("classsection.Service.Create: %w", err)
	}

	created, err := s.repo.GetByID(ctx, cs.ID)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.Create.GetByID: %w", err)
	}
	return ToClassSectionResponse(created), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*ClassSectionResponse, error) {
	cs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.GetByID: %w", err)
	}
	return ToClassSectionResponse(cs), nil
}

func (s *service) ListByAcademicYear(ctx context.Context, academicYearID uuid.UUID) ([]*ClassSectionSummary, error) {
	sections, err := s.repo.FindByAcademicYear(ctx, academicYearID)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.ListByAcademicYear: %w", err)
	}
	return s.toSummaries(ctx, sections)
}

func (s *service) ListByStandard(ctx context.Context, standardID, academicYearID uuid.UUID) ([]*ClassSectionSummary, error) {
	sections, err := s.repo.FindByStandard(ctx, standardID, academicYearID)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.ListByStandard: %w", err)
	}
	return s.toSummaries(ctx, sections)
}

func (s *service) ListByEmployee(ctx context.Context, employeeID, academicYearID uuid.UUID) ([]*ClassSectionSummary, error) {
	sections, err := s.repo.FindByEmployee(ctx, employeeID, academicYearID)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.ListByEmployee: %w", err)
	}
	return s.toSummaries(ctx, sections)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*ClassSectionResponse, error) {
	cs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.Update.GetByID: %w", err)
	}

	// Guard: if section name is changing, check for duplicate
	if req.SectionName != nil && *req.SectionName != cs.SectionName {
		isDup, err := s.repo.IsDuplicate(ctx, cs.AcademicYearID, cs.StandardID, *req.SectionName)
		if err != nil {
			return nil, fmt.Errorf("classsection.Service.Update.IsDuplicate: %w", err)
		}
		if isDup {
			return nil, fmt.Errorf("classsection.Service.Update: section %q already exists", *req.SectionName)
		}
		cs.SectionName = *req.SectionName
	}
	if req.ClassEmployeeID != nil {
		cs.ClassEmployeeID = req.ClassEmployeeID
	}
	if req.RoomID != nil {
		cs.RoomID = req.RoomID
	}
	if req.MaxStrength != nil {
		// Guard: cannot set max strength below current enrollment
		enrolled, err := s.repo.CountEnrolled(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("classsection.Service.Update.CountEnrolled: %w", err)
		}
		if int64(*req.MaxStrength) < enrolled {
			return nil, fmt.Errorf("classsection.Service.Update: max_strength (%d) cannot be less than current enrollment (%d)", *req.MaxStrength, enrolled)
		}
		cs.MaxStrength = *req.MaxStrength
	}

	if err := s.repo.Update(ctx, id, cs); err != nil {
		return nil, fmt.Errorf("classsection.Service.Update.Save: %w", err)
	}
	return ToClassSectionResponse(cs), nil
}

func (s *service) AssignEmployee(ctx context.Context, id uuid.UUID, req AssignEmployeeRequest) (*ClassSectionResponse, error) {
	cs, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.AssignEmployee.GetByID: %w", err)
	}
	cs.ClassEmployeeID = req.ClassEmployeeID
	if err := s.repo.Update(ctx, id, cs); err != nil {
		return nil, fmt.Errorf("classsection.Service.AssignEmployee.Save: %w", err)
	}
	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("classsection.Service.AssignEmployee.Reload: %w", err)
	}
	return ToClassSectionResponse(updated), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	enrolled, err := s.repo.CountEnrolled(ctx, id)
	if err != nil {
		return fmt.Errorf("classsection.Service.Delete.CountEnrolled: %w", err)
	}
	if enrolled > 0 {
		return errors.New("classsection.Service.Delete: cannot delete a class section with enrolled students")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("classsection.Service.Delete: %w", err)
	}
	return nil
}

// toSummaries enriches a list of class sections with live enrollment counts.
func (s *service) toSummaries(ctx context.Context, sections []*ClassSection) ([]*ClassSectionSummary, error) {
	summaries := make([]*ClassSectionSummary, len(sections))
	for i, cs := range sections {
		enrolled, err := s.repo.CountEnrolled(ctx, cs.ID)
		if err != nil {
			return nil, fmt.Errorf("classsection.Service.toSummaries.CountEnrolled(%s): %w", cs.ID, err)
		}
		summaries[i] = ToClassSectionSummary(cs, enrolled)
	}
	return summaries, nil
}
