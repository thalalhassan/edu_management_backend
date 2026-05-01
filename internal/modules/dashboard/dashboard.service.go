package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Service interface {
	GetDashboard(ctx context.Context, academicYearID uuid.UUID) (*DashboardResponse, error)
}

type service struct {
	repo Repository
	db   *gorm.DB // raw db for AY name lookup
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{repo: repo, db: db}
}

func (s *service) GetDashboard(ctx context.Context, academicYearID uuid.UUID) (*DashboardResponse, error) {
	// Resolve academic year name
	var ay database.AcademicYear
	if err := s.db.WithContext(ctx).First(&ay, "id = ?", academicYearID).Error; err != nil {
		return nil, fmt.Errorf("dashboard.Service.GetDashboard: academic year not found: %w", err)
	}

	institution, err := s.buildInstitutionOverview(ctx, academicYearID)
	if err != nil {
		return nil, fmt.Errorf("dashboard.Service.GetDashboard.Institution: %w", err)
	}

	fees, err := s.buildFeesSummary(ctx, academicYearID)
	if err != nil {
		return nil, fmt.Errorf("dashboard.Service.GetDashboard.Fees: %w", err)
	}

	// Payroll module is not yet built — return zeroed stubs.
	payroll := s.buildPayrollStub()

	return &DashboardResponse{
		AcademicYear: ay.Name,
		Institution:  institution,
		Fees:         fees,
		Payroll:      payroll,
	}, nil
}

// ──────────────────────────────────────────────────────────────
// INSTITUTION OVERVIEW
// ──────────────────────────────────────────────────────────────

func (s *service) buildInstitutionOverview(ctx context.Context, academicYearID uuid.UUID) (InstitutionOverview, error) {
	activeStudents, totalStudents, err := s.repo.CountStudentsByStatus(ctx)
	if err != nil {
		return InstitutionOverview{}, fmt.Errorf("CountStudentsByStatus: %w", err)
	}

	activeTeachers, totalTeachers, err := s.repo.CountEmployeesByStatus(ctx)
	if err != nil {
		return InstitutionOverview{}, fmt.Errorf("CountEmployeesByStatus: %w", err)
	}

	totalSubjects, err := s.repo.CountSubjects(ctx)
	if err != nil {
		return InstitutionOverview{}, fmt.Errorf("CountSubjects: %w", err)
	}

	totalClasses, err := s.repo.CountClassSections(ctx, academicYearID)
	if err != nil {
		return InstitutionOverview{}, fmt.Errorf("CountClassSections: %w", err)
	}

	return InstitutionOverview{
		ActiveStudents: CountStat{Active: activeStudents, Total: totalStudents},
		ActiveTeachers: CountStat{Active: activeTeachers, Total: totalTeachers},
		TotalSubjects:  totalSubjects,
		TotalClasses:   totalClasses,
	}, nil
}

// ──────────────────────────────────────────────────────────────
// FEES
// ──────────────────────────────────────────────────────────────

func (s *service) buildFeesSummary(ctx context.Context, academicYearID uuid.UUID) (FeesSummary, error) {
	paid, pending, err := s.repo.CountFeeRecordsByStatus(ctx, academicYearID)
	if err != nil {
		return FeesSummary{}, fmt.Errorf("CountFeeRecordsByStatus: %w", err)
	}

	totalCollected, err := s.repo.SumFeeByStatus(ctx, academicYearID, database.FeeStatusPaid)
	if err != nil {
		return FeesSummary{}, fmt.Errorf("SumFeeCollected: %w", err)
	}

	totalPending, err := s.repo.SumFeeByStatus(ctx, academicYearID, database.FeeStatusPending)
	if err != nil {
		return FeesSummary{}, fmt.Errorf("SumFeePending: %w", err)
	}

	// This month
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	collectedThisMonth, err := s.repo.SumFeeCollectedBetween(ctx, academicYearID, monthStart, monthEnd)
	if err != nil {
		return FeesSummary{}, fmt.Errorf("SumFeeThisMonth: %w", err)
	}

	// This year (calendar year)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(now.Year(), 12, 31, 23, 59, 59, 0, time.UTC)

	collectedThisYear, err := s.repo.SumFeeCollectedBetween(ctx, academicYearID, yearStart, yearEnd)
	if err != nil {
		return FeesSummary{}, fmt.Errorf("SumFeeThisYear: %w", err)
	}

	return FeesSummary{
		StudentsPaid:       PaidStat{Paid: paid, Pending: pending},
		TotalCollected:     totalCollected,
		TotalPending:       totalPending,
		CollectedThisMonth: collectedThisMonth,
		CollectedThisYear:  collectedThisYear,
	}, nil
}

// ──────────────────────────────────────────────────────────────
// PAYROLL — stub until payroll module is built
// ──────────────────────────────────────────────────────────────

func (s *service) buildPayrollStub() PayrollSummary {
	return PayrollSummary{
		TeachersPaid:         PaidStat{Paid: 0, Pending: 0},
		TotalSalariesPaid:    decimal.Zero,
		TotalSalariesPending: decimal.Zero,
		ToPayThisMonth:       decimal.Zero,
		InstitutionBalance:   decimal.Zero,
		BalanceStatus:        "Positive",
	}
}
