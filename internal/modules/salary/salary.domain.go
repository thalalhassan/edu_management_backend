package salary

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
)

type SalaryStructure = database.SalaryStructure
type SalaryRecord = database.SalaryRecord
type SalaryStatus = database.SalaryStatus

// ──────────────────────────────────────────────────────────────
// SORT / FILTER
// ──────────────────────────────────────────────────────────────

var AllowedStructureSortFields = map[string]bool{
	"created_at":     true,
	"effective_from": true,
	"basic_salary":   true,
}

var AllowedRecordSortFields = map[string]bool{
	"created_at": true,
	"year":       true,
	"month":      true,
	"net_salary": true,
	"paid_date":  true,
}

type StructureFilterParams struct {
	TeacherID *uuid.UUID `form:"teacher_id"`
}

type RecordFilterParams struct {
	TeacherID      *uuid.UUID    `form:"teacher_id"`
	AcademicYearID *uuid.UUID    `form:"academic_year_id"`
	Month          *int          `form:"month"`
	Year           *int          `form:"year"`
	Status         *SalaryStatus `form:"status"`
}

// ──────────────────────────────────────────────────────────────
// SALARY STRUCTURE REQUESTS
// ──────────────────────────────────────────────────────────────

type CreateStructureRequest struct {
	EmployeeID     uuid.UUID       `json:"employee_id"     binding:"required,uuid"`
	BasicSalary    decimal.Decimal `json:"basic_salary"    binding:"required"`
	HRA            decimal.Decimal `json:"hra"`
	DA             decimal.Decimal `json:"da"`
	OtherAllowance decimal.Decimal `json:"other_allowance"`
	PF             decimal.Decimal `json:"pf"`
	TDS            decimal.Decimal `json:"tds"`
	OtherDeduction decimal.Decimal `json:"other_deduction"`
	EffectiveFrom  time.Time       `json:"effective_from"  binding:"required"`
	Remarks        *string         `json:"remarks,omitempty"`
}

type UpdateStructureRequest struct {
	BasicSalary    *decimal.Decimal `json:"basic_salary,omitempty"`
	HRA            *decimal.Decimal `json:"hra,omitempty"`
	DA             *decimal.Decimal `json:"da,omitempty"`
	OtherAllowance *decimal.Decimal `json:"other_allowance,omitempty"`
	PF             *decimal.Decimal `json:"pf,omitempty"`
	TDS            *decimal.Decimal `json:"tds,omitempty"`
	OtherDeduction *decimal.Decimal `json:"other_deduction,omitempty"`
	EffectiveFrom  *time.Time       `json:"effective_from,omitempty"`
	Remarks        *string          `json:"remarks,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// SALARY RECORD REQUESTS
// ──────────────────────────────────────────────────────────────

// BulkGenerateRequest generates monthly salary records for all
// active teachers using their current salary structure.
type BulkGenerateRequest struct {
	AcademicYearID uuid.UUID `json:"academic_year_id" binding:"required,uuid"`
	Month          int       `json:"month"            binding:"required,min=1,max=12"`
	Year           int       `json:"year"             binding:"required,min=2000"`
	// If true, salary is prorated based on present_days / working_days.
	// If false, full net salary is paid regardless of attendance.
	Prorate bool `json:"prorate"`
}

// RecordPaymentRequest marks a salary record as paid.
type RecordPaymentRequest struct {
	PaidAmount     decimal.Decimal `json:"paid_amount"     binding:"required"`
	PaidDate       time.Time       `json:"paid_date"       binding:"required"`
	TransactionRef *string         `json:"transaction_ref,omitempty"`
	Remarks        *string         `json:"remarks,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// RESPONSES
// ──────────────────────────────────────────────────────────────

type SalaryStructureResponse struct {
	SalaryStructure
	GrossSalary    decimal.Decimal `json:"gross_salary"`    // computed
	TotalDeduction decimal.Decimal `json:"total_deduction"` // computed
	NetSalary      decimal.Decimal `json:"net_salary"`      // computed
}

type SalaryRecordResponse struct {
	SalaryRecord
	Balance decimal.Decimal `json:"balance"` // NetSalary - PaidAmount
}

// MonthlySummary is a rolled-up view of all salary records for a month.
// Used in the dashboard payroll section.
type MonthlySummary struct {
	Month           int             `json:"month"`
	Year            int             `json:"year"`
	AcademicYear    string          `json:"academic_year"`
	TotalTeachers   int             `json:"total_teachers"`
	PaidCount       int             `json:"paid_count"`
	PendingCount    int             `json:"pending_count"`
	TotalGross      decimal.Decimal `json:"total_gross"`
	TotalDeductions decimal.Decimal `json:"total_deductions"`
	TotalNet        decimal.Decimal `json:"total_net"`
	TotalPaid       decimal.Decimal `json:"total_paid"`
	TotalPending    decimal.Decimal `json:"total_pending"`
}

// ──────────────────────────────────────────────────────────────
// MAPPERS
// ──────────────────────────────────────────────────────────────

func ToStructureResponse(s *SalaryStructure) *SalaryStructureResponse {
	gross := s.BasicSalary.Add(s.HRA).Add(s.DA).Add(s.OtherAllowance)
	deduction := s.PF.Add(s.TDS).Add(s.OtherDeduction)
	return &SalaryStructureResponse{
		SalaryStructure: *s,
		GrossSalary:     gross,
		TotalDeduction:  deduction,
		NetSalary:       gross.Sub(deduction),
	}
}

func ToRecordResponse(r *SalaryRecord) *SalaryRecordResponse {
	return &SalaryRecordResponse{
		SalaryRecord: *r,
		Balance:      r.NetSalary.Sub(r.PaidAmount),
	}
}

// computeSalaryComponents derives gross, deduction, net from a structure.
func computeSalaryComponents(s *SalaryStructure) (gross, deduction, net decimal.Decimal) {
	gross = s.BasicSalary.Add(s.HRA).Add(s.DA).Add(s.OtherAllowance)
	deduction = s.PF.Add(s.TDS).Add(s.OtherDeduction)
	net = gross.Sub(deduction)
	return
}

// prorateSalary reduces net salary based on attendance.
func prorateSalary(net decimal.Decimal, presentDays, workingDays int) decimal.Decimal {
	if workingDays == 0 {
		return net
	}
	return net.
		Mul(decimal.NewFromInt(int64(presentDays))).
		Div(decimal.NewFromInt(int64(workingDays))).
		Round(2)
}

// computeStatus derives SalaryStatus from paid vs net amounts.
func computeStatus(net, paid decimal.Decimal) database.SalaryStatus {
	switch {
	case paid.GreaterThanOrEqual(net):
		return database.SalaryStatusPaid
	case paid.IsZero():
		return database.SalaryStatusPending
	default:
		return database.SalaryStatusPartial
	}
}
