package student

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
)

type Service interface {
	GetByID(ctx context.Context, id string) (*StudentResponse, error)
	List(ctx context.Context, p pagination.Params) ([]*StudentResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*StudentResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id string) (*StudentResponse, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("student.Service.GetByID: %w", err)
	}
	return ToStudentResponse(student), nil
}

func (s *service) List(ctx context.Context, p pagination.Params) ([]*StudentResponse, int64, error) {
	students, total, err := s.repo.FindAll(ctx, p)
	if err != nil {
		return nil, 0, fmt.Errorf("student.Service.List: %w", err)
	}

	responses := make([]*StudentResponse, len(students))
	for i, student := range students {
		responses[i] = ToStudentResponse(student)
	}

	return responses, total, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*StudentResponse, error) {
	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("student.Service.Update.GetByID: %w", err)
	}

	if req.FirstName != nil {
		student.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		student.LastName = *req.LastName
	}
	if req.DOB != nil {
		student.DOB = *req.DOB
	}
	if req.Status != nil {
		student.Status = database.StudentStatus(*req.Status)
	}

	if err := s.repo.Update(ctx, id, student); err != nil {
		return nil, fmt.Errorf("student.Service.Update.Save: %w", err)
	}
	return ToStudentResponse(student), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("student.Service.Delete: %w", err)
	}
	return nil
}
