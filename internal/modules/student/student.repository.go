package student

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, student *Student) error
	GetByID(ctx context.Context, id string) (*Student, error)
	FindAll(ctx context.Context, p pagination.Params) ([]*Student, int64, error)
	Update(ctx context.Context, student *Student) error
	Delete(ctx context.Context, id string) error
}

type RepositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &RepositoryImpl{db: db}
}

func (r *RepositoryImpl) Create(ctx context.Context, student *Student) error {
	return r.db.WithContext(ctx).Create(student).Error
}

func (r *RepositoryImpl) GetByID(ctx context.Context, id string) (*Student, error) {
	var student Student
	if err := r.db.WithContext(ctx).First(&student, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &student, nil
}

func (r *RepositoryImpl) FindAll(ctx context.Context, p pagination.Params) ([]*Student, int64, error) {
	var students []*Student
	var total int64

	query := r.db.WithContext(ctx).Model(&Student{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((p.Page - 1) * p.Limit).Limit(p.Limit).Find(&students).Error; err != nil {
		return nil, 0, err
	}
	return students, total, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, student *Student) error {
	return r.db.WithContext(ctx).Save(student).Error
}

func (r *RepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Student{}, "id = ?", id).Error
}
