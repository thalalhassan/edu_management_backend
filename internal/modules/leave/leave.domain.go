package leave

import (
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type EmployeeLeave = database.EmployeeLeave
type LeaveStatus = database.LeaveStatus

// ─── Requests ────────────────────────────────────────────────────────────────

// ApplyRequest is submitted by an employee (or on their behalf) to create a leave request.
type ApplyRequest struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required,uuid"`
	FromDate   time.Time `json:"from_date"   binding:"required"`
	ToDate     time.Time `json:"to_date"     binding:"required"`
	Reason     string    `json:"reason"      binding:"required"`
}

// UpdateRequest allows a teacher to edit their own PENDING leave request.
type UpdateRequest struct {
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Reason   *string    `json:"reason,omitempty"`
}

// ReviewRequest is submitted by an admin / principal to approve or reject.
type ReviewRequest struct {
	// Status must be APPROVED or REJECTED — PENDING is not a valid review outcome.
	Status     LeaveStatus `json:"status"      binding:"required"`
	ReviewNote *string     `json:"review_note,omitempty"`
}

// ─── Response ────────────────────────────────────────────────────────────────

type LeaveResponse struct {
	ID         uuid.UUID   `json:"id"`
	EmployeeID uuid.UUID   `json:"employee_id"`
	FromDate   time.Time   `json:"from_date"`
	ToDate     time.Time   `json:"to_date"`
	Reason     string      `json:"reason"`
	Status     LeaveStatus `json:"status"`
	ReviewedBy *uuid.UUID  `json:"reviewed_by,omitempty"`
	ReviewNote *string     `json:"review_note,omitempty"`
	ReviewedAt *time.Time  `json:"reviewed_at,omitempty"`
	// Computed convenience field — number of calendar days
	DurationDays int       `json:"duration_days"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ─── Filter params ────────────────────────────────────────────────────────────

type FilterParams struct {
	EmployeeID *uuid.UUID   `form:"employee_id"`
	Status     *LeaveStatus `form:"status"`
	DateFrom   *time.Time   `form:"date_from"`
	DateTo     *time.Time   `form:"date_to"`
}

var allowedSortFields = map[string]bool{
	"from_date":  true,
	"to_date":    true,
	"status":     true,
	"created_at": true,
}

// ─── Mapper ──────────────────────────────────────────────────────────────────

func ToLeaveResponse(l *EmployeeLeave) *LeaveResponse {
	duration := int(l.ToDate.Sub(l.FromDate).Hours()/24) + 1
	if duration < 1 {
		duration = 1
	}
	return &LeaveResponse{
		ID:           l.ID,
		EmployeeID:   l.EmployeeID,
		FromDate:     l.FromDate,
		ToDate:       l.ToDate,
		Reason:       l.Reason,
		Status:       l.Status,
		ReviewedBy:   l.ReviewedBy,
		ReviewNote:   l.ReviewNote,
		ReviewedAt:   l.ReviewedAt,
		DurationDays: duration,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}
