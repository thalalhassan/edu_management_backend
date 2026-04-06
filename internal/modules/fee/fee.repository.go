package fee

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────────────────────────
// STRUCTURE REPOSITORY
// ──────────────────────────────────────────────────────────────

type StructureRepository interface {
	Create(ctx context.Context, f *FeeStructure) error
	BulkCreate(ctx context.Context, structs []*FeeStructure) error
	GetByID(ctx context.Context, id string) (*FeeStructure, error)
	FindAll(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*FeeStructure, int64, error)
	FindByStandardAndYear(ctx context.Context, standardID, academicYearID string) ([]*FeeStructure, error)
	Update(ctx context.Context, id string, f *FeeStructure) error
	Delete(ctx context.Context, id string) error
	IsDuplicate(ctx context.Context, academicYearID, standardID, component string) (bool, error)
}

type structureRepo struct {
	db *gorm.DB
}

func NewStructureRepository(db *gorm.DB) StructureRepository {
	return &structureRepo{db: db.Model(&database.FeeStructure{})}
}

func (r *structureRepo) Create(ctx context.Context, f *FeeStructure) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *structureRepo) BulkCreate(ctx context.Context, structs []*FeeStructure) error {
	return r.db.WithContext(ctx).Create(&structs).Error
}

func (r *structureRepo) GetByID(ctx context.Context, id string) (*FeeStructure, error) {
	var f FeeStructure
	err := r.db.WithContext(ctx).
		Preload("AcademicYear").
		Preload("Standard").
		First(&f, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *structureRepo) FindAll(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*FeeStructure, int64, error) {
	var structs []*FeeStructure
	var total int64

	query := r.db.WithContext(ctx).Model(&database.FeeStructure{})

	f := q.Filter
	if f.AcademicYearID != nil {
		query = query.Where("academic_year_id = ?", *f.AcademicYearID)
	}
	if f.StandardID != nil {
		query = query.Where("standard_id = ?", *f.StandardID)
	}
	if f.FeeComponent != nil {
		query = query.Where("fee_component ILIKE ?", "%"+*f.FeeComponent+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Preload("AcademicYear").
		Preload("Standard").
		Offset((q.Pagination.Page - 1) * q.Pagination.Limit).
		Limit(q.Pagination.Limit).
		Find(&structs).Error

	return structs, total, err
}

func (r *structureRepo) FindByStandardAndYear(ctx context.Context, standardID, academicYearID string) ([]*FeeStructure, error) {
	var structs []*FeeStructure
	err := r.db.WithContext(ctx).
		Where("standard_id = ? AND academic_year_id = ?", standardID, academicYearID).
		Order("fee_component ASC").
		Find(&structs).Error
	return structs, err
}

func (r *structureRepo) Update(ctx context.Context, id string, f *FeeStructure) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(f).Error
}

func (r *structureRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&FeeStructure{}, "id = ?", id).Error
}

func (r *structureRepo) IsDuplicate(ctx context.Context, academicYearID, standardID, component string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.FeeStructure{}).
		Where("academic_year_id = ? AND standard_id = ? AND fee_component = ?",
			academicYearID, standardID, component).
		Count(&count).Error
	return count > 0, err
}

// ──────────────────────────────────────────────────────────────
// RECORD REPOSITORY
// ──────────────────────────────────────────────────────────────

type RecordRepository interface {
	Create(ctx context.Context, r *FeeRecord) error
	BulkCreate(ctx context.Context, records []*FeeRecord) error
	GetByID(ctx context.Context, id string) (*FeeRecord, error)
	FindAll(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*FeeRecord, int64, error)
	FindByEnrollment(ctx context.Context, studentEnrollmentID string) ([]*FeeRecord, error)
	Update(ctx context.Context, id string, r *FeeRecord) error
	Delete(ctx context.Context, id string) error
	SumByEnrollment(ctx context.Context, studentEnrollmentID string) (due decimal.Decimal, paid decimal.Decimal, err error)
}

type recordRepo struct {
	db *gorm.DB
}

func NewRecordRepository(db *gorm.DB) RecordRepository {
	return &recordRepo{db: db.Model(&database.FeeRecord{})}
}

func (r *recordRepo) Create(ctx context.Context, rec *FeeRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *recordRepo) BulkCreate(ctx context.Context, records []*FeeRecord) error {
	return r.db.WithContext(ctx).Create(&records).Error
}

func (r *recordRepo) GetByID(ctx context.Context, id string) (*FeeRecord, error) {
	var rec FeeRecord
	err := r.db.WithContext(ctx).
		Preload("StudentEnrollment.Student").
		Preload("StudentEnrollment.ClassSection.Standard").
		First(&rec, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recordRepo) FindAll(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*FeeRecord, int64, error) {
	var records []*FeeRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&database.FeeRecord{})

	f := q.Filter
	if f.StudentEnrollmentID != nil {
		query = query.Where("student_enrollment_id = ?", *f.StudentEnrollmentID)
	}
	if f.FeeComponent != nil {
		query = query.Where("fee_component ILIKE ?", "%"+*f.FeeComponent+"%")
	}
	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}
	if f.DueDateFrom != nil {
		query = query.Where("due_date >= ?", *f.DueDateFrom)
	}
	if f.DueDateTo != nil {
		query = query.Where("due_date <= ?", *f.DueDateTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Preload("StudentEnrollment.Student").
		Offset((q.Pagination.Page - 1) * q.Pagination.Limit).
		Limit(q.Pagination.Limit).
		Find(&records).Error

	return records, total, err
}

func (r *recordRepo) FindByEnrollment(ctx context.Context, studentEnrollmentID string) ([]*FeeRecord, error) {
	var records []*FeeRecord
	err := r.db.WithContext(ctx).
		Where("student_enrollment_id = ?", studentEnrollmentID).
		Order("due_date ASC, fee_component ASC").
		Find(&records).Error
	return records, err
}

func (r *recordRepo) Update(ctx context.Context, id string, rec *FeeRecord) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(rec).Error
}

func (r *recordRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&FeeRecord{}, "id = ?", id).Error
}

// SumByEnrollment returns the total due and total paid for a student enrollment.
func (r *recordRepo) SumByEnrollment(ctx context.Context, studentEnrollmentID string) (decimal.Decimal, decimal.Decimal, error) {
	type result struct {
		TotalDue  decimal.Decimal
		TotalPaid decimal.Decimal
	}
	var res result
	err := r.db.WithContext(ctx).
		Model(&database.FeeRecord{}).
		Select("COALESCE(SUM(amount_due), 0) as total_due, COALESCE(SUM(amount_paid), 0) as total_paid").
		Where("student_enrollment_id = ?", studentEnrollmentID).
		Scan(&res).Error
	return res.TotalDue, res.TotalPaid, err
}
