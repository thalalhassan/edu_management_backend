package enrollment

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Repository interface {
	// Core CRUD
	Create(ctx context.Context, e *Enrollment) error
	GetByID(ctx context.Context, id string) (*Enrollment, error)
	Update(ctx context.Context, id string, e *Enrollment) error
	Delete(ctx context.Context, id string) error

	// Domain queries
	GetByStudentID(ctx context.Context, studentID string, p pagination.Params) ([]*Enrollment, int64, error)
	GetRosterByClassSection(ctx context.Context, classSectionID string) ([]*Enrollment, error)
	GetActiveEnrollment(ctx context.Context, studentID, classSectionID string) (*Enrollment, error)
	IsStudentEnrolledInYear(ctx context.Context, studentID, academicYearID string) (bool, error)
	CountEnrolledInClassSection(ctx context.Context, classSectionID string) (int64, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.StudentEnrollment{})}
}

func (r *repositoryImpl) Create(ctx context.Context, e *Enrollment) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*Enrollment, error) {
	var e Enrollment
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("ClassSection.Standard.Department").
		Preload("ClassSection.AcademicYear").
		Preload("ClassSection.ClassTeacher").
		First(&e, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repositoryImpl) Update(ctx context.Context, id string, e *Enrollment) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(e).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&Enrollment{}, "id = ?", id).Error
}

// GetByStudentID returns all enrollments for a student across all academic years,
// ordered newest first — gives a full academic history.
func (r *repositoryImpl) GetByStudentID(ctx context.Context, studentID string, p pagination.Params) ([]*Enrollment, int64, error) {
	var enrollments []*Enrollment
	var total int64

	query := r.db.WithContext(ctx).
		Where("student_id = ?", studentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("ClassSection.Standard.Department").
		Preload("ClassSection.AcademicYear").
		Order("created_at DESC").
		Offset((p.Page - 1) * p.Limit).
		Limit(p.Limit).
		Find(&enrollments).Error
	if err != nil {
		return nil, 0, err
	}

	return enrollments, total, nil
}

// GetRosterByClassSection returns all enrollments for a class section
// with student data preloaded — used to build the class roster.
func (r *repositoryImpl) GetRosterByClassSection(ctx context.Context, classSectionID string) ([]*Enrollment, error) {
	var enrollments []*Enrollment
	err := r.db.WithContext(ctx).
		Where("class_section_id = ? AND status = ?", classSectionID, database.EnrollmentStatusEnrolled).
		Preload("Student").
		Order("roll_number ASC").
		Find(&enrollments).Error
	if err != nil {
		return nil, err
	}
	return enrollments, nil
}

// GetActiveEnrollment checks if a student is already enrolled in a specific class section.
// Used to prevent duplicate enrollments.
func (r *repositoryImpl) GetActiveEnrollment(ctx context.Context, studentID, classSectionID string) (*Enrollment, error) {
	var e Enrollment
	err := r.db.WithContext(ctx).
		Where("student_id = ? AND class_section_id = ? AND status = ?",
			studentID, classSectionID, database.EnrollmentStatusEnrolled).
		First(&e).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// IsStudentEnrolledInYear checks if a student has any active enrollment
// in the given academic year — prevents enrolling in two sections of the same year.
func (r *repositoryImpl) IsStudentEnrolledInYear(ctx context.Context, studentID, academicYearID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.StudentEnrollment{}).
		Joins("JOIN class_section ON class_section.id = student_enrollment.class_section_id").
		Where("student_enrollment.student_id = ? AND class_section.academic_year_id = ? AND student_enrollment.status = ?",
			studentID, academicYearID, database.EnrollmentStatusEnrolled).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountEnrolledInClassSection returns the number of currently enrolled students
// in a class section — used to enforce MaxStrength.
func (r *repositoryImpl) CountEnrolledInClassSection(ctx context.Context, classSectionID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.StudentEnrollment{}).
		Where("class_section_id = ? AND status = ?", classSectionID, database.EnrollmentStatusEnrolled).
		Count(&count).Error
	return count, err
}
