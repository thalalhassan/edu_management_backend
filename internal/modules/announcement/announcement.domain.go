package announcement

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type Announcement = database.Announcement
type AnnouncementAudience = database.AnnouncementAudience
type AnnouncementRead = database.AnnouncementRead

// ─── Requests ────────────────────────────────────────────────────────────────

// CreateRequest creates an announcement in draft state (is_published = false).
// Use PATCH /:id/publish to make it visible to the target audience.
type CreateRequest struct {
	Title         string               `json:"title"          binding:"required"`
	Body          string               `json:"body"           binding:"required"`
	Audience      AnnouncementAudience `json:"audience"       binding:"required"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	AttachmentURL *string              `json:"attachment_url,omitempty"`
}

// UpdateRequest allows editing a draft announcement.
// A published announcement cannot be edited — it must be unpublished first.
type UpdateRequest struct {
	Title         *string               `json:"title,omitempty"`
	Body          *string               `json:"body,omitempty"`
	Audience      *AnnouncementAudience `json:"audience,omitempty"`
	ExpiresAt     *time.Time            `json:"expires_at,omitempty"`
	AttachmentURL *string               `json:"attachment_url,omitempty"`
}

// PublishRequest toggles the published state.
// Setting is_published = true stamps published_at with the current time.
// Setting is_published = false clears published_at (reverts to draft).
type PublishRequest struct {
	IsPublished bool `json:"is_published"`
}

// ─── Response ────────────────────────────────────────────────────────────────

type AnnouncementResponse struct {
	ID            string               `json:"id"`
	Title         string               `json:"title"`
	Body          string               `json:"body"`
	Audience      AnnouncementAudience `json:"audience"`
	IsPublished   bool                 `json:"is_published"`
	PublishedAt   *time.Time           `json:"published_at,omitempty"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	AttachmentURL *string              `json:"attachment_url,omitempty"`
	AuthorID      string               `json:"author_id"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

type FilterParams struct {
	Audience    *AnnouncementAudience `form:"audience"`
	IsPublished *bool                 `form:"is_published"`
	// Search matches against title
	Search *string `form:"search"`
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

// validAudiences is the set of accepted AnnouncementAudience values.
var validAudiences = map[AnnouncementAudience]bool{
	database.AnnouncementAudienceAll:      true,
	database.AnnouncementAudienceTeachers: true,
	database.AnnouncementAudienceStudents: true,
	database.AnnouncementAudienceParents:  true,
	database.AnnouncementAudienceStaff:    true,
}
