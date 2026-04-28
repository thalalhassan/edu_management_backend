package subject

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, s *Subject) error
	GetByID(ctx context.Context, id string) (*Subject, error)
	GetByCode(ctx context.Context, code string) (*Subject, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Subject, int64, error)
	Update(ctx context.Context, id string, s *Subject) error
	Delete(ctx context.Context, id string) error
	IsCodeTaken(ctx context.Context, code string) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.Subject{})}
}

func (r *repositoryImpl) Create(ctx context.Context, s *Subject) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*Subject, error) {
	var s Subject
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repositoryImpl) GetByCode(ctx context.Context, code string) (*Subject, error) {
	var s Subject
	if err := r.db.WithContext(ctx).First(&s, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Subject, int64, error) {
	var subjects []*Subject
	var total int64

	query := r.db.WithContext(ctx).Model(&database.Subject{})

	// Apply filters
	f := q.Filter
	if f.Search != nil && *f.Search != "" {
		like := "%" + *f.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset((q.Pagination.Page - 1) * q.Pagination.Limit).
		Limit(q.Pagination.Limit).
		Find(&subjects).Error

	return subjects, total, err
}

func (r *repositoryImpl) Update(ctx context.Context, id string, s *Subject) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(s).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Subject{}).Error
}

func (r *repositoryImpl) IsCodeTaken(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.Subject{}).
		Where("code = ?", code).
		Count(&count).Error
	return count > 0, err
}
