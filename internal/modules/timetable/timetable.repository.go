package timetable

import (
	"context"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, t *TimeTable) error
	GetByID(ctx context.Context, id string) (*TimeTable, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*TimeTable, int64, error)
	FindByClassSection(ctx context.Context, classSectionID string) ([]*TimeTable, error)
	FindByTeacher(ctx context.Context, teacherID string) ([]*TimeTable, error)
	Update(ctx context.Context, id string, t *TimeTable) error
	Delete(ctx context.Context, id string) error

	// Conflict detection
	HasConflict(ctx context.Context, classSectionID string, dayOfWeek int, start, end time.Time, excludeID string) (bool, error)
	HasTeacherConflict(ctx context.Context, teacherID string, dayOfWeek int, start, end time.Time, excludeID string) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.TimeTable{})}
}

func (r *repositoryImpl) Create(ctx context.Context, t *TimeTable) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*TimeTable, error) {
	var t TimeTable
	err := r.db.WithContext(ctx).
		Preload("ClassSection.Standard.Department").
		Preload("Subject").
		Preload("Teacher").
		First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*TimeTable, int64, error) {
	var entries []*TimeTable
	var total int64

	query := r.db.WithContext(ctx).Model(&database.TimeTable{})

	f := q.Filter
	if f.ClassSectionID != nil {
		query = query.Where("class_section_id = ?", *f.ClassSectionID)
	}
	if f.TeacherID != nil {
		query = query.Where("teacher_id = ?", *f.TeacherID)
	}
	if f.SubjectID != nil {
		query = query.Where("subject_id = ?", *f.SubjectID)
	}
	if f.DayOfWeek != nil {
		query = query.Where("day_of_week = ?", *f.DayOfWeek)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Preload("ClassSection.Standard").
		Preload("Subject").
		Preload("Teacher").
		Offset((q.Pagination.Page - 1) * q.Pagination.Limit).
		Limit(q.Pagination.Limit).
		Find(&entries).Error

	return entries, total, err
}

// FindByClassSection returns the full week schedule for a class section,
// ordered by day then start time — used to render the weekly timetable.
func (r *repositoryImpl) FindByClassSection(ctx context.Context, classSectionID string) ([]*TimeTable, error) {
	var entries []*TimeTable
	err := r.db.WithContext(ctx).
		Preload("Subject").
		Preload("Teacher").
		Preload("ClassSection.Standard").
		Where("class_section_id = ?", classSectionID).
		Order("day_of_week ASC, start_time ASC").
		Find(&entries).Error
	return entries, err
}

// FindByTeacher returns the full week schedule for a teacher across all their class sections.
func (r *repositoryImpl) FindByTeacher(ctx context.Context, teacherID string) ([]*TimeTable, error) {
	var entries []*TimeTable
	err := r.db.WithContext(ctx).
		Preload("Subject").
		Preload("ClassSection.Standard").
		Where("teacher_id = ?", teacherID).
		Order("day_of_week ASC, start_time ASC").
		Find(&entries).Error
	return entries, err
}

func (r *repositoryImpl) Update(ctx context.Context, id string, t *TimeTable) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(t).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&TimeTable{}, "id = ?", id).Error
}

// HasConflict checks if a class section already has a period that overlaps
// the requested time slot on the same day.
// excludeID is the current entry's ID during updates — pass "" for creates.
func (r *repositoryImpl) HasConflict(ctx context.Context, classSectionID string, dayOfWeek int, start, end time.Time, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&database.TimeTable{}).
		Where("class_section_id = ? AND day_of_week = ? AND start_time < ? AND end_time > ?",
			classSectionID, dayOfWeek, end, start)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// HasTeacherConflict checks if the teacher is already assigned to another
// class at the same time on the same day.
func (r *repositoryImpl) HasTeacherConflict(ctx context.Context, teacherID string, dayOfWeek int, start, end time.Time, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&database.TimeTable{}).
		Where("teacher_id = ? AND day_of_week = ? AND start_time < ? AND end_time > ?",
			teacherID, dayOfWeek, end, start)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
