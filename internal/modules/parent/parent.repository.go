package parent

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Parent, error)
	FindAll(ctx context.Context, p pagination.Params) ([]*Parent, int64, error)
	Update(ctx context.Context, id string, parent *Parent) error
	Delete(ctx context.Context, id string) error
}

type RepositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &RepositoryImpl{db: db.Model(&database.Parent{})}
}

func (r *RepositoryImpl) GetByID(ctx context.Context, id string) (*Parent, error) {
	var parent Parent
	if err := r.db.WithContext(ctx).First(&parent, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &parent, nil
}

func (r *RepositoryImpl) FindAll(ctx context.Context, p pagination.Params) ([]*Parent, int64, error) {
	var parents []*Parent
	var total int64

	query := r.db.WithContext(ctx)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((p.Page - 1) * p.Limit).Limit(p.Limit).Find(&parents).Error; err != nil {
		return nil, 0, err
	}

	return parents, total, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, id string, parent *Parent) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(parent).Error
}

func (r *RepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Parent{}).Error
}
