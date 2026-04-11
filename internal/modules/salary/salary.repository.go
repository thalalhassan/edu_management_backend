package salary

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
	Create(ctx context.Context, s *SalaryStructure) error
	GetByID(ctx context.Context, id string) (*SalaryStructure, error)
	// GetActiveForTeacher returns the most recent structure for a teacher
	// (highest effective_from date on or before today).
	GetActiveForTeacher(ctx context.Context, teacherID string) (*SalaryStructure, error)
	FindAll(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*SalaryStructure, int64, error)
	FindByTeacher(ctx context.Context, teacherID string) ([]*SalaryStructure, error)
	Update(ctx context.Context, id string, s *SalaryStructure) error
	Delete(ctx context.Context, id string) error
}

type structureRepo struct {
	db *gorm.DB
}

func NewStructureRepository(db *gorm.DB) StructureRepository {
	return &structureRepo{db: db.Model(&database.SalaryStructure{})}
}

func (r *structureRepo) Create(ctx context.Context, s *SalaryStructure) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *structureRepo) GetByID(ctx context.Context, id string) (*SalaryStructure, error) {
	var s SalaryStructure
	err := r.db.WithContext(ctx).
		Preload("Teacher").
		First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *structureRepo) GetActiveForTeacher(ctx context.Context, teacherID string) (*SalaryStructure, error) {
	var s SalaryStructure
	err := r.db.WithContext(ctx).
		Where("teacher_id = ? AND effective_from <= NOW()", teacherID).
		Order("effective_from DESC").
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *structureRepo) FindAll(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*SalaryStructure, int64, error) {
	var structs []*SalaryStructure
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SalaryStructure{})

	if q.Filter.TeacherID != nil {
		query = query.Where("teacher_id = ?", *q.Filter.TeacherID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Preload("Teacher").
		Offset((q.Pagination.Page - 1) * q.Pagination.Limit).
		Limit(q.Pagination.Limit).
		Find(&structs).Error

	return structs, total, err
}

func (r *structureRepo) FindByTeacher(ctx context.Context, teacherID string) ([]*SalaryStructure, error) {
	var structs []*SalaryStructure
	err := r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Order("effective_from DESC").
		Find(&structs).Error
	return structs, err
}

func (r *structureRepo) Update(ctx context.Context, id string, s *SalaryStructure) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(s).Error
}

func (r *structureRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&SalaryStructure{}, "id = ?", id).Error
}

// ──────────────────────────────────────────────────────────────
// RECORD REPOSITORY
// ──────────────────────────────────────────────────────────────

type RecordRepository interface {
	Create(ctx context.Context, r *SalaryRecord) error
	BulkCreate(ctx context.Context, records []*SalaryRecord) error
	GetByID(ctx context.Context, id string) (*SalaryRecord, error)
	FindAll(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*SalaryRecord, int64, error)
	FindByTeacher(ctx context.Context, teacherID string) ([]*SalaryRecord, error)
	Update(ctx context.Context, id string, r *SalaryRecord) error
	Delete(ctx context.Context, id string) error

	// IsDuplicate checks if a record already exists for teacher + month + year.
	IsDuplicate(ctx context.Context, teacherID string, month, year int) (bool, error)

	// GetMonthlySummary aggregates all records for a given month + year.
	GetMonthlySummary(ctx context.Context, academicYearID string, month, year int) (*MonthlySummary, error)

	// SumByStatus used by dashboard payroll section.
	SumByStatus(ctx context.Context, academicYearID string, status database.SalaryStatus) (decimal.Decimal, error)
	CountByStatus(ctx context.Context, month, year int) (paid, pending int, err error)
}

type recordRepo struct {
	db *gorm.DB
}

func NewRecordRepository(db *gorm.DB) RecordRepository {
	return &recordRepo{db: db.Model(&database.SalaryRecord{})}
}

func (r *recordRepo) Create(ctx context.Context, rec *SalaryRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *recordRepo) BulkCreate(ctx context.Context, records []*SalaryRecord) error {
	return r.db.WithContext(ctx).Create(&records).Error
}

func (r *recordRepo) GetByID(ctx context.Context, id string) (*SalaryRecord, error) {
	var rec SalaryRecord
	err := r.db.WithContext(ctx).
		Preload("Teacher").
		Preload("AcademicYear").
		First(&rec, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recordRepo) FindAll(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*SalaryRecord, int64, error) {
	var records []*SalaryRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&database.SalaryRecord{})

	f := q.Filter
	if f.TeacherID != nil {
		query = query.Where("teacher_id = ?", *f.TeacherID)
	}
	if f.AcademicYearID != nil {
		query = query.Where("academic_year_id = ?", *f.AcademicYearID)
	}
	if f.Month != nil {
		query = query.Where("month = ?", *f.Month)
	}
	if f.Year != nil {
		query = query.Where("year = ?", *f.Year)
	}
	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Preload("Teacher").
		Preload("AcademicYear").
		Offset((q.Pagination.Page - 1) * q.Pagination.Limit).
		Limit(q.Pagination.Limit).
		Find(&records).Error

	return records, total, err
}

func (r *recordRepo) FindByTeacher(ctx context.Context, teacherID string) ([]*SalaryRecord, error) {
	var records []*SalaryRecord
	err := r.db.WithContext(ctx).
		Where("teacher_id = ?", teacherID).
		Order("year DESC, month DESC").
		Find(&records).Error
	return records, err
}

func (r *recordRepo) Update(ctx context.Context, id string, rec *SalaryRecord) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(rec).Error
}

func (r *recordRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&SalaryRecord{}, "id = ?", id).Error
}

func (r *recordRepo) IsDuplicate(ctx context.Context, teacherID string, month, year int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.SalaryRecord{}).
		Where("teacher_id = ? AND month = ? AND year = ?", teacherID, month, year).
		Count(&count).Error
	return count > 0, err
}

func (r *recordRepo) GetMonthlySummary(ctx context.Context, academicYearID string, month, year int) (*MonthlySummary, error) {
	type row struct {
		Status      database.SalaryStatus
		Count       int
		TotalGross  decimal.Decimal
		TotalDeduct decimal.Decimal
		TotalNet    decimal.Decimal
		TotalPaid   decimal.Decimal
	}
	var rows []row

	err := r.db.WithContext(ctx).
		Model(&database.SalaryRecord{}).
		Select(`status,
			COUNT(*) as count,
			COALESCE(SUM(gross_salary), 0)    as total_gross,
			COALESCE(SUM(total_deduction), 0) as total_deduct,
			COALESCE(SUM(net_salary), 0)      as total_net,
			COALESCE(SUM(paid_amount), 0)     as total_paid`).
		Where("academic_year_id = ? AND month = ? AND year = ?", academicYearID, month, year).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Fetch AY name
	var ay database.AcademicYear
	r.db.WithContext(ctx).Select("name").First(&ay, "id = ?", academicYearID)

	summary := &MonthlySummary{
		Month:        month,
		Year:         year,
		AcademicYear: ay.Name,
	}
	for _, row := range rows {
		summary.TotalTeachers += row.Count
		summary.TotalGross = summary.TotalGross.Add(row.TotalGross)
		summary.TotalDeductions = summary.TotalDeductions.Add(row.TotalDeduct)
		summary.TotalNet = summary.TotalNet.Add(row.TotalNet)
		summary.TotalPaid = summary.TotalPaid.Add(row.TotalPaid)
		if row.Status == database.SalaryStatusPaid {
			summary.PaidCount += row.Count
		} else {
			summary.PendingCount += row.Count
		}
	}
	summary.TotalPending = summary.TotalNet.Sub(summary.TotalPaid)
	return summary, nil
}

func (r *recordRepo) SumByStatus(ctx context.Context, academicYearID string, status database.SalaryStatus) (decimal.Decimal, error) {
	type result struct{ Total decimal.Decimal }
	var res result
	err := r.db.WithContext(ctx).
		Model(&database.SalaryRecord{}).
		Select("COALESCE(SUM(paid_amount), 0) as total").
		Where("academic_year_id = ? AND status = ?", academicYearID, status).
		Scan(&res).Error
	return res.Total, err
}

func (r *recordRepo) CountByStatus(ctx context.Context, month, year int) (int, int, error) {
	type row struct {
		Status database.SalaryStatus
		Count  int
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&database.SalaryRecord{}).
		Select("status, COUNT(*) as count").
		Where("month = ? AND year = ?", month, year).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}
	var paid, pending int
	for _, row := range rows {
		if row.Status == database.SalaryStatusPaid {
			paid += row.Count
		} else {
			pending += row.Count
		}
	}
	return paid, pending, nil
}
