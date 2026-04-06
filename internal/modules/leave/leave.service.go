package leave

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

// ─── Service interface ────────────────────────────────────────────────────────

type Service interface {
	Apply(ctx context.Context, req ApplyRequest) (*LeaveResponse, error)
	GetByID(ctx context.Context, id string) (*LeaveResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*LeaveResponse, int64, error)

	// Update allows editing a leave request that is still PENDING.
	// Only the teacher who owns the leave (or an admin) should call this.
	Update(ctx context.Context, id string, req UpdateRequest) (*LeaveResponse, error)

	// Review is the approval workflow endpoint — sets status to APPROVED or REJECTED.
	// reviewerID is the user_id extracted from the JWT by the handler.
	Review(ctx context.Context, id string, reviewerID string, req ReviewRequest) (*LeaveResponse, error)

	// Cancel allows a teacher to withdraw their own PENDING leave request.
	Cancel(ctx context.Context, id string) (*LeaveResponse, error)

	// Delete hard-deletes a PENDING leave request (admin safety valve).
	Delete(ctx context.Context, id string) error
}

// ─── service struct ───────────────────────────────────────────────────────────

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ─── Apply ───────────────────────────────────────────────────────────────────

func (s *service) Apply(ctx context.Context, req ApplyRequest) (*LeaveResponse, error) {
	if err := validateDateRange(req.FromDate, req.ToDate); err != nil {
		return nil, fmt.Errorf("leave.Service.Apply: %w", err)
	}

	hasOverlap, err := s.repo.HasOverlap(ctx, req.TeacherID, req.FromDate, req.ToDate, "")
	if err != nil {
		return nil, fmt.Errorf("leave.Service.Apply.HasOverlap: %w", err)
	}
	if hasOverlap {
		return nil, errors.New("leave.Service.Apply: an overlapping leave request already exists for this teacher in the requested date range")
	}

	l := &TeacherLeave{
		TeacherID: req.TeacherID,
		FromDate:  req.FromDate,
		ToDate:    req.ToDate,
		Reason:    req.Reason,
		Status:    database.LeaveStatusPending,
	}
	if err := s.repo.Create(ctx, l); err != nil {
		return nil, fmt.Errorf("leave.Service.Apply: %w", err)
	}
	return ToLeaveResponse(l), nil
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func (s *service) GetByID(ctx context.Context, id string) (*LeaveResponse, error) {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("leave.Service.GetByID: %w", err)
	}
	return ToLeaveResponse(l), nil
}

// ─── List ────────────────────────────────────────────────────────────────────

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*LeaveResponse, int64, error) {
	leaves, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("leave.Service.List: %w", err)
	}
	responses := make([]*LeaveResponse, len(leaves))
	for i, l := range leaves {
		responses[i] = ToLeaveResponse(l)
	}
	return responses, total, nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*LeaveResponse, error) {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("leave.Service.Update.GetByID: %w", err)
	}
	if l.Status != database.LeaveStatusPending {
		return nil, fmt.Errorf("leave.Service.Update: only PENDING leave requests can be edited — current status is %s", l.Status)
	}

	if req.FromDate != nil {
		l.FromDate = *req.FromDate
	}
	if req.ToDate != nil {
		l.ToDate = *req.ToDate
	}
	if req.Reason != nil {
		l.Reason = *req.Reason
	}

	if err := validateDateRange(l.FromDate, l.ToDate); err != nil {
		return nil, fmt.Errorf("leave.Service.Update: %w", err)
	}

	// Re-check overlap excluding the current record
	hasOverlap, err := s.repo.HasOverlap(ctx, l.TeacherID, l.FromDate, l.ToDate, id)
	if err != nil {
		return nil, fmt.Errorf("leave.Service.Update.HasOverlap: %w", err)
	}
	if hasOverlap {
		return nil, errors.New("leave.Service.Update: updated dates overlap with another existing leave request")
	}

	if err := s.repo.Update(ctx, l); err != nil {
		return nil, fmt.Errorf("leave.Service.Update: %w", err)
	}
	return ToLeaveResponse(l), nil
}

// ─── Review ──────────────────────────────────────────────────────────────────

func (s *service) Review(ctx context.Context, id string, reviewerID string, req ReviewRequest) (*LeaveResponse, error) {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("leave.Service.Review.GetByID: %w", err)
	}

	// Guard: only PENDING leaves can be reviewed
	if l.Status != database.LeaveStatusPending {
		return nil, fmt.Errorf("leave.Service.Review: leave request is already %s — only PENDING requests can be reviewed", l.Status)
	}

	// Guard: review status must be APPROVED or REJECTED, not PENDING
	if req.Status != database.LeaveStatusApproved && req.Status != database.LeaveStatusRejected {
		return nil, fmt.Errorf("leave.Service.Review: invalid review status %q — must be APPROVED or REJECTED", req.Status)
	}

	now := time.Now()
	l.Status = req.Status
	l.ReviewedBy = &reviewerID
	l.ReviewNote = req.ReviewNote
	l.ReviewedAt = &now

	if err := s.repo.Update(ctx, l); err != nil {
		return nil, fmt.Errorf("leave.Service.Review: %w", err)
	}
	return ToLeaveResponse(l), nil
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

// Cancel transitions a PENDING leave request to REJECTED (self-withdrawal).
// Using REJECTED rather than a new status keeps the model enum clean while
// still preventing the cancelled request from blocking future leave applications.
func (s *service) Cancel(ctx context.Context, id string) (*LeaveResponse, error) {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("leave.Service.Cancel.GetByID: %w", err)
	}
	if l.Status != database.LeaveStatusPending {
		return nil, fmt.Errorf("leave.Service.Cancel: only PENDING leave requests can be cancelled — current status is %s", l.Status)
	}

	l.Status = database.LeaveStatusRejected
	note := "Cancelled by teacher"
	l.ReviewNote = &note

	if err := s.repo.Update(ctx, l); err != nil {
		return nil, fmt.Errorf("leave.Service.Cancel: %w", err)
	}
	return ToLeaveResponse(l), nil
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func (s *service) Delete(ctx context.Context, id string) error {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leave.Service.Delete.GetByID: %w", err)
	}
	if l.Status == database.LeaveStatusApproved {
		return errors.New("leave.Service.Delete: cannot delete an APPROVED leave request — reject it first if it needs to be reversed")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("leave.Service.Delete: %w", err)
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func validateDateRange(from, to time.Time) error {
	if to.Before(from) {
		return errors.New("to_date must be on or after from_date")
	}
	return nil
}
