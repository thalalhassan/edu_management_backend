package academic_year

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, a *AcademicYear) error
	GetByID(ctx context.Context, id uuid.UUID) (*AcademicYear, error)
	GetActive(ctx context.Context) (*AcademicYear, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYear, error)
	Update(ctx context.Context, id uuid.UUID, a *AcademicYear) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID) error
	IsDuplicateName(ctx context.Context, name string) (bool, error)
	HasClassSections(ctx context.Context, id uuid.UUID) (bool, error)
	HasOverlappingDates(ctx context.Context, start, end time.Time) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(ctx context.Context, a *AcademicYear) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*AcademicYear, error) {
	var a AcademicYear
	if err := r.db.WithContext(ctx).Unscoped().First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) GetActive(ctx context.Context) (*AcademicYear, error) {
	var a AcademicYear
	if err := r.db.WithContext(ctx).First(&a, "is_active = true").Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYear, error) {
	var years []*AcademicYear
	query := r.db.WithContext(ctx)
	f := q.Filter
	if f.Search != nil {
		like := "%" + *f.Search + "%"
		query = query.Where("name ILIKE ?", like)
	}
	err := q.Sort.Apply(query).Order(q.Sort).Find(&years).Error
	return years, err
}

func (r *repositoryImpl) Update(ctx context.Context, id uuid.UUID, a *AcademicYear) error {
	return r.db.WithContext(ctx).Model(&AcademicYear{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":       a.Name,
		"start_date": a.StartDate,
		"end_date":   a.EndDate,
	}).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AcademicYear{}, "id = ?", id).Error
}

func (r *repositoryImpl) SetActive(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&AcademicYear{}).Where("is_active = true").Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Model(&AcademicYear{}).Where("id = ?", id).Update("is_active", true).Error
	})
}

func (r *repositoryImpl) IsDuplicateName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AcademicYear{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasClassSections(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.ClassSection{}).Where("academic_year_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasOverlappingDates(ctx context.Context, start, end time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AcademicYear{}).Where("start_date < ? AND end_date > ?", end, start).Count(&count).Error
	return count > 0, err
}
