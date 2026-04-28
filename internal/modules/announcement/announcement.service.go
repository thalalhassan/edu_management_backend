package announcement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

// ─── Service interface ────────────────────────────────────────────────────────

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*AnnouncementResponse, error)
	GetByID(ctx context.Context, id string) (*AnnouncementResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*AnnouncementResponse, int64, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*AnnouncementResponse, error)
	Publish(ctx context.Context, id string, req PublishRequest) (*AnnouncementResponse, error)
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

func (s *service) Create(ctx context.Context, req CreateRequest) (*AnnouncementResponse, error) {
	if !validAudiences[req.Audience] {
		return nil, fmt.Errorf("announcement.Service.Create: invalid audience %q — must be one of ALL, TEACHERS, STUDENTS, PARENTS, STAFF", req.Audience)
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return nil, errors.New("announcement.Service.Create: expires_at must be a future date")
	}

	a := &Announcement{
		Title:         req.Title,
		Body:          req.Body,
		Audience:      req.Audience,
		ExpiresAt:     req.ExpiresAt,
		AttachmentURL: req.AttachmentURL,
		IsPublished:   false, // always starts as draft
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("announcement.Service.Create: %w", err)
	}
	return ToAnnouncementResponse(a), nil
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func (s *service) GetByID(ctx context.Context, id string) (*AnnouncementResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("announcement.Service.GetByID: %w", err)
	}
	return ToAnnouncementResponse(a), nil
}

// ─── List ────────────────────────────────────────────────────────────────────

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*AnnouncementResponse, int64, error) {
	announcements, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("announcement.Service.List: %w", err)
	}
	responses := make([]*AnnouncementResponse, len(announcements))
	for i, a := range announcements {
		responses[i] = ToAnnouncementResponse(a)
	}
	return responses, total, nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*AnnouncementResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("announcement.Service.Update.GetByID: %w", err)
	}

	// A published announcement is read-only — unpublish it first via PATCH /publish.
	if a.IsPublished {
		return nil, errors.New("announcement.Service.Update: cannot edit a published announcement — unpublish it first")
	}

	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Body != nil {
		a.Body = *req.Body
	}
	if req.Audience != nil {
		if !validAudiences[*req.Audience] {
			return nil, fmt.Errorf("announcement.Service.Update: invalid audience %q", *req.Audience)
		}
		a.Audience = *req.Audience
	}
	if req.ExpiresAt != nil {
		if !req.ExpiresAt.After(time.Now()) {
			return nil, errors.New("announcement.Service.Update: expires_at must be a future date")
		}
		a.ExpiresAt = req.ExpiresAt
	}
	if req.AttachmentURL != nil {
		a.AttachmentURL = req.AttachmentURL
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("announcement.Service.Update: %w", err)
	}
	return ToAnnouncementResponse(a), nil
}

// ─── Publish ─────────────────────────────────────────────────────────────────

func (s *service) Publish(ctx context.Context, id string, req PublishRequest) (*AnnouncementResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("announcement.Service.Publish.GetByID: %w", err)
	}

	now := time.Now()
	if req.IsPublished {
		if a.IsPublished {
			return nil, errors.New("announcement.Service.Publish: already published")
		}
		a.IsPublished = true
		a.PublishedAt = &now
	} else {
		if !a.IsPublished {
			return nil, errors.New("announcement.Service.Publish: already unpublished")
		}
		a.IsPublished = false
		a.PublishedAt = nil
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("announcement.Service.Publish: %w", err)
	}
	return ToAnnouncementResponse(a), nil
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────
