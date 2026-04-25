package teacher_assignment

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, m *database.TeacherAssignment) error
	GetByID(ctx context.Context, id string) (*database.TeacherAssignment, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]database.TeacherAssignment, int64, error)
	Update(ctx context.Context, m *database.TeacherAssignment) error
	Delete(ctx context.Context, id string) error

	ExistsConflict(ctx context.Context, classID, subjectID string) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(ctx context.Context, m *database.TeacherAssignment) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*database.TeacherAssignment, error) {
	var m database.TeacherAssignment
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repositoryImpl) List(ctx context.Context, q query_params.Query[FilterParams]) ([]database.TeacherAssignment, int64, error) {
	var list []database.TeacherAssignment
	var count int64

	db := r.db.WithContext(ctx).Model(&database.TeacherAssignment{})

	// Filters
	if q.Filter.ClassSectionID != nil {
		db = db.Where("class_section_id = ?", *q.Filter.ClassSectionID)
	}
	if q.Filter.TeacherID != nil {
		db = db.Where("teacher_id = ?", *q.Filter.TeacherID)
	}
	if q.Filter.SubjectID != nil {
		db = db.Where("subject_id = ?", *q.Filter.SubjectID)
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	db = q.Sort.Apply(db)

	if err := db.
		Limit(q.Pagination.Limit).
		Offset(q.Pagination.Offset).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, count, nil
}

func (r *repositoryImpl) Update(ctx context.Context, m *database.TeacherAssignment) error {
	return r.db.WithContext(ctx).Where("id = ?", m.ID).Save(m).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.TeacherAssignment{}).Error
}

// Ensure 1 subject per class section (no duplicates)
func (r *repositoryImpl) ExistsConflict(ctx context.Context, classID, subjectID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.TeacherAssignment{}).
		Where("class_section_id = ? AND subject_id = ?", classID, subjectID).
		Count(&count).Error

	return count > 0, err
}
