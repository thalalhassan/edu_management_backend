package dashboard

import "github.com/shopspring/decimal"

// DashboardResponse is the complete payload returned by GET /dashboard.
// Structured to match the three UI sections exactly.
type DashboardResponse struct {
	AcademicYear   string             `json:"academic_year"`
	Institution    InstitutionOverview `json:"institution"`
	Fees           FeesSummary         `json:"fees"`
	Payroll        PayrollSummary      `json:"payroll"`
}

// ──────────────────────────────────────────────────────────────
// INSTITUTION OVERVIEW
// ──────────────────────────────────────────────────────────────

type InstitutionOverview struct {
	ActiveStudents CountStat `json:"active_students"` // active vs total
	ActiveTeachers CountStat `json:"active_teachers"` // active vs total
	TotalSubjects  int       `json:"total_subjects"`
	TotalClasses   int       `json:"total_classes"`  // class sections in active AY
}

// CountStat holds an active/current count alongside the total.
// Renders as "3 out of 10 total" in the UI.
type CountStat struct {
	Active int `json:"active"`
	Total  int `json:"total"`
}

// ──────────────────────────────────────────────────────────────
// FEES & COLLECTIONS
// ──────────────────────────────────────────────────────────────

type FeesSummary struct {
	StudentsPaid      PaidStat        `json:"students_paid"`       // paid count vs pending count
	TotalCollected    decimal.Decimal `json:"total_collected"`
	TotalPending      decimal.Decimal `json:"total_pending"`
	CollectedThisMonth decimal.Decimal `json:"collected_this_month"`
	CollectedThisYear  decimal.Decimal `json:"collected_this_year"`
}

// PaidStat holds a paid count alongside the pending count.
// Renders as "3 vs 5 pending" in the UI.
type PaidStat struct {
	Paid    int `json:"paid"`
	Pending int `json:"pending"`
}

// ──────────────────────────────────────────────────────────────
// PAYROLL & SALARIES
// ──────────────────────────────────────────────────────────────

// Note: Payroll is not modelled in the DB yet. The values below
// are stubs that will return zero until a payroll module is built.
// The domain shape is defined now so the UI contract is stable.
type PayrollSummary struct {
	TeachersPaid       PaidStat        `json:"teachers_paid"`
	TotalSalariesPaid  decimal.Decimal `json:"total_salaries_paid"`
	TotalSalariesPending decimal.Decimal `json:"total_salaries_pending"`
	ToPayThisMonth     decimal.Decimal `json:"to_pay_this_month"`
	InstitutionBalance decimal.Decimal `json:"institution_balance"`
	BalanceStatus      string          `json:"balance_status"` // "Positive" | "Negative" | "Neutral"
}
