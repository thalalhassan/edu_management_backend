package academic_year

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, a *AcademicYear) error
	GetByID(ctx context.Context, id string) (*AcademicYear, error)
	GetActive(ctx context.Context) (*AcademicYear, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYear, error)
	Update(ctx context.Context, id string, a *AcademicYear) error
	Delete(ctx context.Context, id string) error

	// SetActive deactivates all years then activates the given one —
	// done in a transaction to ensure only one active year at a time.
	SetActive(ctx context.Context, id string) error

	IsDuplicateName(ctx context.Context, name string) (bool, error)
	HasClassSections(ctx context.Context, id string) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.AcademicYear{})}
}

func (r *repositoryImpl) Create(ctx context.Context, a *AcademicYear) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*AcademicYear, error) {
	var a AcademicYear
	if err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
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

// FindAll returns all academic years ordered newest first.
func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*AcademicYear, error) {
	var years []*AcademicYear
	query := r.db.WithContext(ctx)

	f := q.Filter

	if f.Search != nil {
		like := "%" + *f.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ?", like, like)
	}

	var total int64
	query.Count(&total)

	err := q.Sort.Apply(query).
		Order("start_date DESC").
		Find(&years).Error
	return years, err
}

func (r *repositoryImpl) Update(ctx context.Context, id string, a *AcademicYear) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(a).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&AcademicYear{}, "id = ?", id).Error
}

// SetActive deactivates all academic years then marks the target one active.
// Wrapped in a transaction to guarantee only one active year at any time.
func (r *repositoryImpl) SetActive(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate all
		if err := tx.Model(&database.AcademicYear{}).
			Where("is_active = true").
			Update("is_active", false).Error; err != nil {
			return err
		}
		// Activate the target
		if err := tx.Model(&database.AcademicYear{}).
			Where("id = ?", id).
			Update("is_active", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *repositoryImpl) IsDuplicateName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.AcademicYear{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasClassSections(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ClassSection{}).
		Where("academic_year_id = ?", id).
		Count(&count).Error
	return count > 0, err
}
