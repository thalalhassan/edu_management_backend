package subject

import (
	"context"
	"errors"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*SubjectResponse, error)
	GetByID(ctx context.Context, id string) (*SubjectResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*SubjectResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*SubjectResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*SubjectResponse, error) {
	taken, err := s.repo.IsCodeTaken(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("subject.Service.Create.IsCodeTaken: %w", err)
	}
	if taken {
		return nil, fmt.Errorf("subject.Service.Create: code %q is already in use", req.Code)
	}

	sub := &Subject{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("subject.Service.Create: %w", err)
	}
	return ToSubjectResponse(sub), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*SubjectResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("subject.Service.GetByID: %w", err)
	}
	return ToSubjectResponse(sub), nil
}

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*SubjectResponse, int64, error) {
	subjects, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("subject.Service.List: %w", err)
	}
	responses := make([]*SubjectResponse, len(subjects))
	for i, sub := range subjects {
		responses[i] = ToSubjectResponse(sub)
	}
	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*SubjectResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("subject.Service.Update.GetByID: %w", err)
	}

	// Guard: if code is changing ensure it is not already taken
	if req.Code != nil && *req.Code != sub.Code {
		taken, err := s.repo.IsCodeTaken(ctx, *req.Code)
		if err != nil {
			return nil, fmt.Errorf("subject.Service.Update.IsCodeTaken: %w", err)
		}
		if taken {
			return nil, fmt.Errorf("subject.Service.Update: code %q is already in use", *req.Code)
		}
		sub.Code = *req.Code
	}
	if req.Name != nil {
		sub.Name = *req.Name
	}
	if req.Description != nil {
		sub.Description = req.Description
	}

	if err := s.repo.Update(ctx, id, sub); err != nil {
		return nil, fmt.Errorf("subject.Service.Update.Save: %w", err)
	}
	return ToSubjectResponse(sub), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("subject.Service.Delete.GetByID: %w", err)
	}
	// Guard: cannot delete a subject that is assigned to standards or timetables
	if len(sub.StandardSubjects) > 0 {
		return errors.New("subject.Service.Delete: subject is assigned to one or more standards — remove assignments first")
	}
	if len(sub.TimeTables) > 0 {
		return errors.New("subject.Service.Delete: subject is used in timetable entries — remove those first")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("subject.Service.Delete: %w", err)
	}
	return nil
}
