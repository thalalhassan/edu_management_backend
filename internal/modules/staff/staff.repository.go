package staff

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Staff, error)
	FindAll(ctx context.Context, p pagination.Params) ([]*Staff, int64, error)
	Update(ctx context.Context, id string, staff *Staff) error
	Delete(ctx context.Context, id string) error
}

type RepositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &RepositoryImpl{db: db.Model(&database.Staff{})}
}

func (r *RepositoryImpl) GetByID(ctx context.Context, id string) (*Staff, error) {
	var staff Staff
	if err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *RepositoryImpl) FindAll(ctx context.Context, p pagination.Params) ([]*Staff, int64, error) {
	var staffs []*Staff
	var total int64

	query := r.db.WithContext(ctx)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((p.Page - 1) * p.Limit).Limit(p.Limit).Find(&staffs).Error; err != nil {
		return nil, 0, err
	}

	return staffs, total, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, id string, staff *Staff) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(staff).Error
}

func (r *RepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Staff{}).Error
}
