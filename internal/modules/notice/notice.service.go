package notice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

// ─── Service interface ────────────────────────────────────────────────────────

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*NoticeResponse, error)
	GetByID(ctx context.Context, id string) (*NoticeResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*NoticeResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*NoticeResponse, error)
	Publish(ctx context.Context, id string, req PublishRequest) (*NoticeResponse, error)
	Delete(ctx context.Context, id string) error
}

// ─── service struct ───────────────────────────────────────────────────────────

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func (s *service) Create(ctx context.Context, req CreateRequest) (*NoticeResponse, error) {
	if !validAudiences[req.Audience] {
		return nil, fmt.Errorf("notice.Service.Create: invalid audience %q — must be one of ALL, TEACHERS, STUDENTS, PARENTS, STAFF", req.Audience)
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return nil, errors.New("notice.Service.Create: expires_at must be a future date")
	}

	n := &Notice{
		Title:          req.Title,
		Content:        req.Content,
		Audience:       req.Audience,
		ExpiresAt:      req.ExpiresAt,
		ClassSectionID: req.ClassSectionID,
		IsPublished:    false, // always starts as draft
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("notice.Service.Create: %w", err)
	}
	return ToNoticeResponse(n), nil
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func (s *service) GetByID(ctx context.Context, id string) (*NoticeResponse, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("notice.Service.GetByID: %w", err)
	}
	return ToNoticeResponse(n), nil
}

// ─── List ────────────────────────────────────────────────────────────────────

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*NoticeResponse, int64, error) {
	notices, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("notice.Service.List: %w", err)
	}
	responses := make([]*NoticeResponse, len(notices))
	for i, n := range notices {
		responses[i] = ToNoticeResponse(n)
	}
	return responses, total, nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*NoticeResponse, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("notice.Service.Update.GetByID: %w", err)
	}

	// A published notice is read-only — unpublish it first via PATCH /publish.
	if n.IsPublished {
		return nil, errors.New("notice.Service.Update: cannot edit a published notice — unpublish it first")
	}

	if req.Title != nil {
		n.Title = *req.Title
	}
	if req.Content != nil {
		n.Content = *req.Content
	}
	if req.Audience != nil {
		if !validAudiences[*req.Audience] {
			return nil, fmt.Errorf("notice.Service.Update: invalid audience %q", *req.Audience)
		}
		n.Audience = *req.Audience
	}
	if req.ExpiresAt != nil {
		if !req.ExpiresAt.After(time.Now()) {
			return nil, errors.New("notice.Service.Update: expires_at must be a future date")
		}
		n.ExpiresAt = req.ExpiresAt
	}
	// Explicit null clears the class section scope (makes it school-wide).
	// Omitting the field entirely leaves it unchanged.
	if req.ClassSectionID != nil {
		n.ClassSectionID = req.ClassSectionID
	}

	if err := s.repo.Update(ctx, n); err != nil {
		return nil, fmt.Errorf("notice.Service.Update: %w", err)
	}
	return ToNoticeResponse(n), nil
}

// ─── Publish ─────────────────────────────────────────────────────────────────

func (s *service) Publish(ctx context.Context, id string, req PublishRequest) (*NoticeResponse, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("notice.Service.Publish.GetByID: %w", err)
	}

	// Guard: publishing an already-expired notice makes no sense.
	if req.IsPublished && n.ExpiresAt != nil && !n.ExpiresAt.After(time.Now()) {
		return nil, errors.New("notice.Service.Publish: cannot publish a notice whose expires_at is already in the past — update expires_at first")
	}

	now := time.Now()
	n.IsPublished = req.IsPublished
	if req.IsPublished {
		n.PublishedAt = &now
	} else {
		// Revert to draft — clear published_at so re-publishing stamps a fresh time.
		n.PublishedAt = nil
	}

	if err := s.repo.Update(ctx, n); err != nil {
		return nil, fmt.Errorf("notice.Service.Publish: %w", err)
	}
	return ToNoticeResponse(n), nil
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func (s *service) Delete(ctx context.Context, id string) error {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("notice.Service.Delete.GetByID: %w", err)
	}
	// Prevent accidental deletion of live notices — unpublish first.
	if n.IsPublished {
		return errors.New("notice.Service.Delete: cannot delete a published notice — unpublish it first")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("notice.Service.Delete: %w", err)
	}
	return nil
}
