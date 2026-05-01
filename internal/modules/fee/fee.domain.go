package fee

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
)

type FeeStructure = database.FeeStructure
type FeeRecord = database.FeeRecord
type FeeStatus = database.FeeStatus

// ──────────────────────────────────────────────────────────────
// SORT / FILTER
// ──────────────────────────────────────────────────────────────

var AllowedStructureSortFields = map[string]bool{
	"created_at":    true,
	"fee_component": true,
	"amount":        true,
	"due_date":      true,
}

var AllowedRecordSortFields = map[string]bool{
	"created_at":  true,
	"due_date":    true,
	"paid_date":   true,
	"amount_due":  true,
	"amount_paid": true,
}

type StructureFilterParams struct {
	AcademicYearID *uuid.UUID `form:"academic_year_id"`
	StandardID     *uuid.UUID `form:"standard_id"`
	FeeComponentID *uuid.UUID `form:"fee_component_id"`
}

type RecordFilterParams struct {
	StudentEnrollmentID *uuid.UUID `form:"student_enrollment_id"`
	FeeComponentID      *uuid.UUID `form:"fee_component_id"`
	Status              *FeeStatus `form:"status"`
	DueDateFrom         *time.Time `form:"due_date_from"`
	DueDateTo           *time.Time `form:"due_date_to"`
}

// ──────────────────────────────────────────────────────────────
// FEE STRUCTURE REQUESTS
// ──────────────────────────────────────────────────────────────

type CreateStructureRequest struct {
	AcademicYearID uuid.UUID       `json:"academic_year_id" binding:"required,uuid"`
	StandardID     uuid.UUID       `json:"standard_id"      binding:"required,uuid"`
	FeeComponentID uuid.UUID       `json:"fee_component_id" binding:"required,uuid"`
	Amount         decimal.Decimal `json:"amount"           binding:"required"`
	DueDate        *time.Time      `json:"due_date,omitempty"`
}

type UpdateStructureRequest struct {
	FeeComponentID *uuid.UUID       `json:"fee_component_id,omitempty"`
	Amount         *decimal.Decimal `json:"amount,omitempty"`
	DueDate        *time.Time       `json:"due_date,omitempty"`
}

// BulkCreateStructureRequest creates all fee components for a standard
// in one call — the typical setup flow at the start of an academic year.
type BulkCreateStructureRequest struct {
	AcademicYearID uuid.UUID           `json:"academic_year_id" binding:"required,uuid"`
	StandardID     uuid.UUID           `json:"standard_id"      binding:"required,uuid"`
	Components     []FeeComponentInput `json:"components"       binding:"required,min=1"`
}

type FeeComponentInput struct {
	FeeComponentID uuid.UUID       `json:"fee_component_id" binding:"required,uuid"`
	Amount         decimal.Decimal `json:"amount"           binding:"required"`
	DueDate        *time.Time      `json:"due_date,omitempty"`
}

// ──────────────────────────────────────────────────────────────
// FEE RECORD REQUESTS
// ──────────────────────────────────────────────────────────────

type CreateRecordRequest struct {
	StudentEnrollmentID uuid.UUID       `json:"student_enrollment_id" binding:"required,uuid"`
	FeeComponentID      uuid.UUID       `json:"fee_component_id"      binding:"required,uuid"`
	AmountDue           decimal.Decimal `json:"amount_due"            binding:"required"`
	DueDate             time.Time       `json:"due_date"              binding:"required"`
	Remarks             *string         `json:"remarks,omitempty"`
}

// RecordPaymentRequest is the primary mutation — records a payment
// against an existing fee record.
type RecordPaymentRequest struct {
	AmountPaid     decimal.Decimal `json:"amount_paid"     binding:"required"`
	PaidDate       time.Time       `json:"paid_date"       binding:"required"`
	TransactionRef *string         `json:"transaction_ref,omitempty"`
	Remarks        *string         `json:"remarks,omitempty"`
}

// WaiveRequest marks a fee record as waived with an optional reason.
type WaiveRequest struct {
	Remarks *string `json:"remarks,omitempty"`
}

// BulkGenerateRequest generates fee records for all enrolled students
// in a class section based on the fee structure for their standard.
type BulkGenerateRequest struct {
	ClassSectionID uuid.UUID `json:"class_section_id" binding:"required,uuid"`
}

// ──────────────────────────────────────────────────────────────
// RESPONSES
// ──────────────────────────────────────────────────────────────

type FeeStructureResponse struct {
	FeeStructure
}

type FeeRecordResponse struct {
	FeeRecord
	Balance decimal.Decimal `json:"balance"` // AmountDue - AmountPaid
}

// StudentFeeSummary gives a rolled-up view of all fee records
// for a student enrollment — used in the student fee dashboard.
type StudentFeeSummary struct {
	StudentEnrollmentID uuid.UUID           `json:"student_enrollment_id"`
	StudentName         string              `json:"student_name"`
	AdmissionNo         string              `json:"admission_no"`
	ClassSection        string              `json:"class_section"`
	TotalDue            decimal.Decimal     `json:"total_due"`
	TotalPaid           decimal.Decimal     `json:"total_paid"`
	TotalBalance        decimal.Decimal     `json:"total_balance"`
	Records             []FeeRecordResponse `json:"records"`
}

// ──────────────────────────────────────────────────────────────
// MAPPERS
// ──────────────────────────────────────────────────────────────

func ToFeeStructureResponse(f *FeeStructure) *FeeStructureResponse {
	return &FeeStructureResponse{FeeStructure: *f}
}

func ToFeeRecordResponse(r *FeeRecord) *FeeRecordResponse {
	balance := r.AmountDue.Sub(r.AmountPaid)
	return &FeeRecordResponse{
		FeeRecord: *r,
		Balance:   balance,
	}
}

// computeStatus derives the correct FeeStatus from amounts and due date.
func computeStatus(amountDue, amountPaid decimal.Decimal, dueDate time.Time) database.FeeStatus {
	switch {
	case amountPaid.GreaterThanOrEqual(amountDue):
		return database.FeeStatusPaid
	case amountPaid.IsZero() && time.Now().After(dueDate):
		return database.FeeStatusOverdue
	case !amountPaid.IsZero():
		return database.FeeStatusPartial
	default:
		return database.FeeStatusPending
	}
}
