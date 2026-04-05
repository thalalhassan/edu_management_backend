package standard

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, s *Standard) error
	GetByID(ctx context.Context, id string) (*Standard, error)
	FindAll(ctx context.Context) ([]*Standard, error)
	FindByDepartment(ctx context.Context, departmentID string) ([]*Standard, error)
	Update(ctx context.Context, id string, s *Standard) error
	Delete(ctx context.Context, id string) error

	// Subject assignments
	AssignSubject(ctx context.Context, link *StandardSubject) error
	RemoveSubject(ctx context.Context, standardID, subjectID string) error
	GetSubjects(ctx context.Context, standardID string) ([]*StandardSubject, error)
	IsSubjectAssigned(ctx context.Context, standardID, subjectID string) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.Standard{})}
}

func (r *repositoryImpl) Create(ctx context.Context, s *Standard) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*Standard, error) {
	var s Standard
	err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Subjects.Subject").
		First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context) ([]*Standard, error) {
	var standards []*Standard
	err := r.db.WithContext(ctx).
		Preload("Department").
		Order("order_index ASC, name ASC").
		Find(&standards).Error
	return standards, err
}

func (r *repositoryImpl) FindByDepartment(ctx context.Context, departmentID string) ([]*Standard, error) {
	var standards []*Standard
	err := r.db.WithContext(ctx).
		Preload("Subjects.Subject").
		Where("department_id = ?", departmentID).
		Order("order_index ASC").
		Find(&standards).Error
	return standards, err
}

func (r *repositoryImpl) Update(ctx context.Context, id string, s *Standard) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(s).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Standard{}, "id = ?", id).Error
}

func (r *repositoryImpl) AssignSubject(ctx context.Context, link *StandardSubject) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *repositoryImpl) RemoveSubject(ctx context.Context, standardID, subjectID string) error {
	return r.db.WithContext(ctx).
		Where("standard_id = ? AND subject_id = ?", standardID, subjectID).
		Delete(&database.StandardSubject{}).Error
}

func (r *repositoryImpl) GetSubjects(ctx context.Context, standardID string) ([]*StandardSubject, error) {
	var links []*StandardSubject
	err := r.db.WithContext(ctx).
		Preload("Subject").
		Where("standard_id = ?", standardID).
		Find(&links).Error
	return links, err
}

func (r *repositoryImpl) IsSubjectAssigned(ctx context.Context, standardID, subjectID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.StandardSubject{}).
		Where("standard_id = ? AND subject_id = ?", standardID, subjectID).
		Count(&count).Error
	return count > 0, err
}
