package teacher

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (*Teacher, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*Teacher, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Teacher, int64, error)
	Update(ctx context.Context, id string, teacher *Teacher) error
	UpdateStatus(ctx context.Context, id string, isActive bool) error
	Delete(ctx context.Context, id string) error
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.Employee{})}
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*Teacher, error) {
	var t Teacher
	err := r.db.WithContext(ctx).
		Preload("Assignments.ClassSection.Standard").
		Preload("Assignments.Subject").
		Preload("ClassSections.Standard").
		First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repositoryImpl) GetByEmployeeID(ctx context.Context, employeeID string) (*Teacher, error) {
	var t Teacher
	if err := r.db.WithContext(ctx).First(&t, "employee_id = ?", employeeID).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Teacher, int64, error) {
	var teachers []*Teacher
	var total int64

	query := r.db.WithContext(ctx)

	// Filter : Note change filter to scoped chain for standard
	f := q.Filter

	if f.Search != nil {
		like := "%" + *f.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ?", like, like)
	}

	query = query.Where(f.ToMap())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Order(q.Sort).
		Find(&teachers).Error
	if err != nil {
		return nil, 0, err
	}

	return teachers, total, nil
}

func (r *repositoryImpl) Update(ctx context.Context, id string, teacher *Teacher) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Updates(teacher).Error
}

func (r *repositoryImpl) UpdateStatus(ctx context.Context, id string, isActive bool) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Teacher{}).Error
}
