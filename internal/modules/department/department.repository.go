package department

import (
	"context"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, d *Department) error
	GetByID(ctx context.Context, id uuid.UUID) (*Department, error)
	GetByCode(ctx context.Context, code string) (*Department, error)
	FindAll(ctx context.Context) ([]*Department, error)
	Update(ctx context.Context, id uuid.UUID, d *Department) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.Department{})}
}

func (r *repositoryImpl) Create(ctx context.Context, d *Department) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Department, error) {
	var d Department
	err := r.db.WithContext(ctx).
		Preload("HeadTeacher").
		Preload("Standards").
		First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *repositoryImpl) GetByCode(ctx context.Context, code string) (*Department, error) {
	var d Department
	if err := r.db.WithContext(ctx).First(&d, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

// FindAll returns departments ordered by name — no pagination needed,
// total count is always small (< 20 departments per school).
func (r *repositoryImpl) FindAll(ctx context.Context) ([]*Department, error) {
	var departments []*Department
	err := r.db.WithContext(ctx).
		Preload("HeadTeacher").
		Order("name ASC").
		Find(&departments).Error
	return departments, err
}

func (r *repositoryImpl) Update(ctx context.Context, id uuid.UUID, d *Department) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(d).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Department{}).Error
}
