package class_section

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, cs *ClassSection) error
	GetByID(ctx context.Context, id string) (*ClassSection, error)
	FindByAcademicYear(ctx context.Context, academicYearID string) ([]*ClassSection, error)
	FindByStandard(ctx context.Context, standardID, academicYearID string) ([]*ClassSection, error)
	FindByTeacher(ctx context.Context, teacherID, academicYearID string) ([]*ClassSection, error)
	Update(ctx context.Context, id string, cs *ClassSection) error
	Delete(ctx context.Context, id string) error
	CountEnrolled(ctx context.Context, classSectionID string) (int64, error)
	IsDuplicate(ctx context.Context, academicYearID, standardID, sectionName string) (bool, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db.Model(&database.ClassSection{})}
}

func (r *repositoryImpl) Create(ctx context.Context, cs *ClassSection) error {
	return r.db.WithContext(ctx).Create(cs).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*ClassSection, error) {
	var cs ClassSection
	err := r.db.WithContext(ctx).
		Preload("AcademicYear").
		Preload("Standard.Department").
		Preload("ClassTeacher").
		First(&cs, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// FindByAcademicYear is the primary list query — scoped to the selected AY.
func (r *repositoryImpl) FindByAcademicYear(ctx context.Context, academicYearID string) ([]*ClassSection, error) {
	var sections []*ClassSection
	err := r.db.WithContext(ctx).
		Preload("AcademicYear").
		Preload("Standard.Department").
		Preload("ClassTeacher").
		Where("academic_year_id = ?", academicYearID).
		Order("standard_id ASC, section_name ASC").
		Find(&sections).Error
	return sections, err
}

// FindByStandard returns sections for a specific standard in an academic year.
func (r *repositoryImpl) FindByStandard(ctx context.Context, standardID, academicYearID string) ([]*ClassSection, error) {
	var sections []*ClassSection
	err := r.db.WithContext(ctx).
		Preload("ClassTeacher").
		Where("standard_id = ? AND academic_year_id = ?", standardID, academicYearID).
		Order("section_name ASC").
		Find(&sections).Error
	return sections, err
}

// FindByTeacher returns all sections a teacher is the class teacher of in an academic year.
func (r *repositoryImpl) FindByTeacher(ctx context.Context, teacherID, academicYearID string) ([]*ClassSection, error) {
	var sections []*ClassSection
	err := r.db.WithContext(ctx).
		Preload("Standard.Department").
		Preload("AcademicYear").
		Where("class_teacher_id = ? AND academic_year_id = ?", teacherID, academicYearID).
		Find(&sections).Error
	return sections, err
}

func (r *repositoryImpl) Update(ctx context.Context, id string, cs *ClassSection) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(cs).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&ClassSection{}, "id = ?", id).Error
}

func (r *repositoryImpl) CountEnrolled(ctx context.Context, classSectionID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.StudentEnrollment{}).
		Where("class_section_id = ? AND status = ?", classSectionID, database.EnrollmentStatusEnrolled).
		Count(&count).Error
	return count, err
}

// IsDuplicate checks if a section with the same name already exists for this
// standard + academic year combination.
func (r *repositoryImpl) IsDuplicate(ctx context.Context, academicYearID, standardID, sectionName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ClassSection{}).
		Where("academic_year_id = ? AND standard_id = ? AND section_name = ?",
			academicYearID, standardID, sectionName).
		Count(&count).Error
	return count > 0, err
}
