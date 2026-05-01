package announcement

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

// ─── Service interface ────────────────────────────────────────────────────────

type Service interface {
	Create(ctx context.Context, req CreateRequest, authorID uuid.UUID) (*AnnouncementResponse, error)
	GetByID(ctx context.Context, id uuid.UUID, userAudiences []AnnouncementAudience) (*AnnouncementResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams], userAudiences []AnnouncementAudience) ([]*AnnouncementResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*AnnouncementResponse, error)
	Publish(ctx context.Context, id uuid.UUID, req PublishRequest) (*AnnouncementResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ─── service struct ───────────────────────────────────────────────────────────

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func (s *service) Create(ctx context.Context, req CreateRequest, authorID uuid.UUID) (*AnnouncementResponse, error) {
	if !IsValidAudience(req.Audience) {
		return nil, ValidationError{Field: "audience", Message: "invalid audience"}
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return nil, ValidationError{Field: "expires_at", Message: "must be a future date"}
	}

	a := &Announcement{
		Title:         req.Title,
		Body:          req.Body,
		Audience:      req.Audience,
		ExpiresAt:     req.ExpiresAt,
		AttachmentURL: req.AttachmentURL,
		AuthorID:      authorID,
		IsPublished:   false, // always starts as draft
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create announcement: %w", err)
	}
	return ToAnnouncementResponse(a), nil
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func (s *service) GetByID(ctx context.Context, id uuid.UUID, userAudiences []AnnouncementAudience) (*AnnouncementResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.canAccessAnnouncement(a, userAudiences) {
		return nil, ErrUnauthorized
	}
	return ToAnnouncementResponse(a), nil
}

// ─── List ────────────────────────────────────────────────────────────────────

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams], userAudiences []AnnouncementAudience) ([]*AnnouncementResponse, int64, error) {
	// For simplicity, fetch all and filter in memory. In production, modify query to filter by audiences.
	announcements, _, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("list announcements: %w", err)
	}
	// Filter in memory for user audiences
	var filtered []*Announcement
	for _, a := range announcements {
		if s.canAccessAnnouncement(a, userAudiences) {
			filtered = append(filtered, a)
		}
	}
	responses := make([]*AnnouncementResponse, len(filtered))
	for i, a := range filtered {
		responses[i] = ToAnnouncementResponse(a)
	}
	// Note: total is approximate since we filter after count
	return responses, int64(len(responses)), nil
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*AnnouncementResponse, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// A published announcement is read-only — unpublish it first via PATCH /publish.
	if a.IsPublished {
		return nil, BusinessError{Message: "cannot edit a published announcement — unpublish it first"}
	}

	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Body != nil {
		a.Body = *req.Body
	}
	if req.Audience != nil {
		if !IsValidAudience(*req.Audience) {
			return nil, ValidationError{Field: "audience", Message: "invalid audience"}
		}
		a.Audience = *req.Audience
	}
	if req.ExpiresAt != nil {
		if !req.ExpiresAt.After(time.Now()) {
			return nil, ValidationError{Field: "expires_at", Message: "must be a future date"}
		}
		a.ExpiresAt = req.ExpiresAt
	}
	if req.AttachmentURL != nil {
		a.AttachmentURL = req.AttachmentURL
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update announcement: %w", err)
	}
	return ToAnnouncementResponse(a), nil
}

// ─── Publish ─────────────────────────────────────────────────────────────────

func (s *service) Publish(ctx context.Context, id uuid.UUID, req PublishRequest) (*AnnouncementResponse, error) {
	tx := s.repo.WithTx(s.repo.(*repositoryImpl).db.Begin())
	defer func() {
		if r := recover(); r != nil {
			tx.(*repositoryImpl).db.Rollback()
			panic(r)
		}
	}()

	a, err := tx.GetByID(ctx, id)
	if err != nil {
		tx.(*repositoryImpl).db.Rollback()
		return nil, err
	}

	now := time.Now()
	if req.IsPublished {
		if a.IsPublished {
			tx.(*repositoryImpl).db.Rollback()
			return nil, BusinessError{Message: "already published"}
		}
		a.IsPublished = true
		a.PublishedAt = &now
	} else {
		if !a.IsPublished {
			tx.(*repositoryImpl).db.Rollback()
			return nil, BusinessError{Message: "already unpublished"}
		}
		a.IsPublished = false
		a.PublishedAt = nil
	}

	if err := tx.Update(ctx, a); err != nil {
		tx.(*repositoryImpl).db.Rollback()
		return nil, fmt.Errorf("publish announcement: %w", err)
	}

	if err := tx.(*repositoryImpl).db.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit publish: %w", err)
	}

	return ToAnnouncementResponse(a), nil
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if a.IsPublished {
		return BusinessError{Message: "cannot delete a published announcement"}
	}
	return s.repo.Delete(ctx, id)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *service) canAccessAnnouncement(a *Announcement, userAudiences []AnnouncementAudience) bool {
	for _, ua := range userAudiences {
		if ua == a.Audience || ua == database.AnnouncementAudienceAll {
			return true
		}
	}
	return false
}
