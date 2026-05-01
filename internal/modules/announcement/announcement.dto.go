package announcement

import (
	"time"

	"github.com/google/uuid"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// CreateRequest creates an announcement in draft state (is_published = false).
// Use PATCH /:id/publish to make it visible to the target audience.
type CreateRequest struct {
	Title         string               `json:"title"          binding:"required,min=1,max=255"`
	Body          string               `json:"body"           binding:"required,min=1,max=10000"`
	Audience      AnnouncementAudience `json:"audience"       binding:"required,oneof=ALL TEACHERS STUDENTS PARENTS STAFF"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	AttachmentURL *string              `json:"attachment_url,omitempty" binding:"omitempty,url"`
}

// UpdateRequest allows editing a draft announcement.
// A published announcement cannot be edited — it must be unpublished first.
type UpdateRequest struct {
	Title         *string               `json:"title,omitempty"          binding:"omitempty,min=1,max=255"`
	Body          *string               `json:"body,omitempty"           binding:"omitempty,min=1,max=10000"`
	Audience      *AnnouncementAudience `json:"audience,omitempty"       binding:"omitempty,oneof=ALL TEACHERS STUDENTS PARENTS STAFF"`
	ExpiresAt     *time.Time            `json:"expires_at,omitempty"`
	AttachmentURL *string               `json:"attachment_url,omitempty" binding:"omitempty,url"`
}

// PublishRequest toggles the published state.
// Setting is_published = true stamps published_at with the current time.
// Setting is_published = false clears published_at (reverts to draft).
type PublishRequest struct {
	IsPublished bool `json:"is_published" binding:"required"`
}

// ─── Response ─────────────────────────────────────────────────────────────────

type AnnouncementResponse struct {
	ID            uuid.UUID            `json:"id"`
	Title         string               `json:"title"`
	Body          string               `json:"body"`
	Audience      AnnouncementAudience `json:"audience"`
	IsPublished   bool                 `json:"is_published"`
	PublishedAt   *time.Time           `json:"published_at,omitempty"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	AttachmentURL *string              `json:"attachment_url,omitempty"`
	AuthorID      uuid.UUID            `json:"author_id"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type FilterParams struct {
	Audience    *AnnouncementAudience `form:"audience"    binding:"omitempty,oneof=ALL TEACHERS STUDENTS PARENTS STAFF"`
	IsPublished *bool                 `form:"is_published" binding:"omitempty"`
	// Search matches against title
	Search *string `form:"search" binding:"omitempty,max=255"`
}

var allowedSortFields = map[string]bool{
	"title":        true,
	"audience":     true,
	"published_at": true,
	"expires_at":   true,
	"created_at":   true,
}

func ToAnnouncementResponse(a *Announcement) *AnnouncementResponse {
	return &AnnouncementResponse{
		ID:            a.ID,
		Title:         a.Title,
		Body:          a.Body,
		Audience:      a.Audience,
		IsPublished:   a.IsPublished,
		PublishedAt:   a.PublishedAt,
		ExpiresAt:     a.ExpiresAt,
		AttachmentURL: a.AttachmentURL,
		AuthorID:      a.AuthorID,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}
