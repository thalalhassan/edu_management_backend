package notice

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type Notice = database.Notice
type NoticeAudience = database.NoticeAudience

// ─── Requests ────────────────────────────────────────────────────────────────

// CreateRequest creates a notice in draft state (is_published = false).
// Use PATCH /:id/publish to make it visible to the target audience.
type CreateRequest struct {
	Title          string         `json:"title"            binding:"required"`
	Content        string         `json:"content"          binding:"required"`
	Audience       NoticeAudience `json:"audience"         binding:"required"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	// ClassSectionID scopes the notice to a specific section.
	// Only meaningful when Audience is STUDENTS or ALL.
	ClassSectionID *string        `json:"class_section_id,omitempty"`
}

// UpdateRequest allows editing a draft notice.
// A published notice cannot be edited — it must be unpublished first.
type UpdateRequest struct {
	Title          *string         `json:"title,omitempty"`
	Content        *string         `json:"content,omitempty"`
	Audience       *NoticeAudience `json:"audience,omitempty"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	ClassSectionID *string         `json:"class_section_id,omitempty"`
}

// PublishRequest toggles the published state.
// Setting is_published = true stamps published_at with the current time.
// Setting is_published = false clears published_at (reverts to draft).
type PublishRequest struct {
	IsPublished bool `json:"is_published"`
}

// ─── Response ────────────────────────────────────────────────────────────────

type NoticeResponse struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Content        string         `json:"content"`
	Audience       NoticeAudience `json:"audience"`
	IsPublished    bool           `json:"is_published"`
	PublishedAt    *time.Time     `json:"published_at,omitempty"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	ClassSectionID *string        `json:"class_section_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ─── Filter params ────────────────────────────────────────────────────────────

type FilterParams struct {
	Audience       *NoticeAudience `form:"audience"`
	IsPublished    *bool           `form:"is_published"`
	ClassSectionID *string         `form:"class_section_id"`
	// Search matches against title
	Search         *string         `form:"search"`
}

var allowedSortFields = map[string]bool{
	"title":        true,
	"audience":     true,
	"published_at": true,
	"expires_at":   true,
	"created_at":   true,
}

// validAudiences is the set of accepted NoticeAudience values.
var validAudiences = map[NoticeAudience]bool{
	database.NoticeAudienceAll:      true,
	database.NoticeAudienceTeachers: true,
	database.NoticeAudienceStudents: true,
	database.NoticeAudienceParents:  true,
	database.NoticeAudienceStaff:    true,
}

// ─── Mapper ──────────────────────────────────────────────────────────────────

func ToNoticeResponse(n *Notice) *NoticeResponse {
	return &NoticeResponse{
		ID:             n.ID,
		Title:          n.Title,
		Content:        n.Content,
		Audience:       n.Audience,
		IsPublished:    n.IsPublished,
		PublishedAt:    n.PublishedAt,
		ExpiresAt:      n.ExpiresAt,
		ClassSectionID: n.ClassSectionID,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
	}
}
