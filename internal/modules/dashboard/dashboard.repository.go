package dashboard

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	// Institution overview
	CountStudentsByStatus(ctx context.Context) (active, total int, err error)
	CountTeachersByStatus(ctx context.Context) (active, total int, err error)
	CountSubjects(ctx context.Context) (int, error)
	CountClassSections(ctx context.Context, academicYearID string) (int, error)

	// Fee stats — all scoped to the active academic year
	CountFeeRecordsByStatus(ctx context.Context, academicYearID string) (paid, pending int, err error)
	SumFeeByStatus(ctx context.Context, academicYearID string, status database.FeeStatus) (decimal.Decimal, error)
	SumFeeCollectedBetween(ctx context.Context, academicYearID string, from, to time.Time) (decimal.Decimal, error)
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// ──────────────────────────────────────────────────────────────
// INSTITUTION
// ──────────────────────────────────────────────────────────────

func (r *repositoryImpl) CountStudentsByStatus(ctx context.Context) (int, int, error) {
	type result struct {
		Status database.StudentStatus
		Count  int
	}
	var rows []result
	err := r.db.WithContext(ctx).
		Model(&database.Student{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}

	var active, total int
	for _, row := range rows {
		total += row.Count
		if row.Status == database.StudentStatusActive {
			active = row.Count
		}
	}
	return active, total, nil
}

func (r *repositoryImpl) CountTeachersByStatus(ctx context.Context) (int, int, error) {
	type result struct {
		IsActive bool
		Count    int
	}
	var rows []result
	err := r.db.WithContext(ctx).
		Model(&database.Teacher{}).
		Select("is_active, COUNT(*) as count").
		Group("is_active").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}

	var active, total int
	for _, row := range rows {
		total += row.Count
		if row.IsActive {
			active = row.Count
		}
	}
	return active, total, nil
}

func (r *repositoryImpl) CountSubjects(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.Subject{}).Count(&count).Error
	return int(count), err
}

func (r *repositoryImpl) CountClassSections(ctx context.Context, academicYearID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ClassSection{}).
		Where("academic_year_id = ?", academicYearID).
		Count(&count).Error
	return int(count), err
}

// ──────────────────────────────────────────────────────────────
// FEES
// ──────────────────────────────────────────────────────────────

// CountFeeRecordsByStatus counts distinct student enrollments
// that have at least one PAID record vs at least one PENDING/OVERDUE record.
// This drives the "Students Paid" card.
func (r *repositoryImpl) CountFeeRecordsByStatus(ctx context.Context, academicYearID string) (int, int, error) {
	type row struct {
		Status database.FeeStatus
		Count  int
	}
	var rows []row

	err := r.db.WithContext(ctx).
		Model(&database.FeeRecord{}).
		Select("fee_record.status, COUNT(DISTINCT fee_record.student_enrollment_id) as count").
		Joins("JOIN student_enrollment ON student_enrollment.id = fee_record.student_enrollment_id").
		Joins("JOIN class_section ON class_section.id = student_enrollment.class_section_id").
		Where("class_section.academic_year_id = ?", academicYearID).
		Group("fee_record.status").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, err
	}

	var paid, pending int
	for _, r := range rows {
		switch r.Status {
		case database.FeeStatusPaid:
			paid += r.Count
		case database.FeeStatusPending, database.FeeStatusOverdue, database.FeeStatusPartial:
			pending += r.Count
		}
	}
	return paid, pending, nil
}

// SumFeeByStatus returns the total amount_due or amount_paid for a given status.
func (r *repositoryImpl) SumFeeByStatus(ctx context.Context, academicYearID string, status database.FeeStatus) (decimal.Decimal, error) {
	type result struct {
		Total decimal.Decimal
	}
	var res result

	// For PAID status sum amount_paid; for others sum amount_due - amount_paid (balance)
	selectClause := "COALESCE(SUM(fee_record.amount_paid), 0) as total"
	if status != database.FeeStatusPaid {
		selectClause = "COALESCE(SUM(fee_record.amount_due - fee_record.amount_paid), 0) as total"
	}

	err := r.db.WithContext(ctx).
		Model(&database.FeeRecord{}).
		Select(selectClause).
		Joins("JOIN student_enrollment ON student_enrollment.id = fee_record.student_enrollment_id").
		Joins("JOIN class_section ON class_section.id = student_enrollment.class_section_id").
		Where("class_section.academic_year_id = ? AND fee_record.status = ?", academicYearID, status).
		Scan(&res).Error

	return res.Total, err
}

// SumFeeCollectedBetween returns amount_paid for records paid within a date range.
func (r *repositoryImpl) SumFeeCollectedBetween(ctx context.Context, academicYearID string, from, to time.Time) (decimal.Decimal, error) {
	type result struct {
		Total decimal.Decimal
	}
	var res result

	err := r.db.WithContext(ctx).
		Model(&database.FeeRecord{}).
		Select("COALESCE(SUM(amount_paid), 0) as total").
		Joins("JOIN student_enrollment ON student_enrollment.id = fee_record.student_enrollment_id").
		Joins("JOIN class_section ON class_section.id = student_enrollment.class_section_id").
		Where("class_section.academic_year_id = ? AND fee_record.paid_date >= ? AND fee_record.paid_date <= ?",
			academicYearID, from, to).
		Scan(&res).Error

	return res.Total, err
}
