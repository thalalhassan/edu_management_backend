package announcement

import (
	"errors"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

// Type aliases
type Announcement = database.Announcement
type AnnouncementAudience = database.AnnouncementAudience
type AnnouncementRead = database.AnnouncementRead

// ─── Custom Errors ───────────────────────────────────────────────────────────

var (
	ErrValidation   = errors.New("validation error")
	ErrBusinessRule = errors.New("business rule violation")
	ErrNotFound     = errors.New("not found")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized = errors.New("unauthorized")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

type BusinessError struct {
	Message string
}

func (e BusinessError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
	ID       uuid.UUID
}

func (e NotFoundError) Error() string {
	if e.ID == uuid.Nil {
		return e.Resource + " not found"
	}
	return e.Resource + " not found: " + e.ID.String()
}

// ─── Domain Invariants ────────────────────────────────────────────────────────

// validAudiences is the set of accepted AnnouncementAudience values.
var validAudiences = map[AnnouncementAudience]bool{
	database.AnnouncementAudienceAll:      true,
	database.AnnouncementAudienceTeachers: true,
	database.AnnouncementAudienceStudents: true,
	database.AnnouncementAudienceParents:  true,
	database.AnnouncementAudienceStaff:    true,
}

// IsValidAudience checks if the audience is valid.
func IsValidAudience(audience AnnouncementAudience) bool {
	return validAudiences[audience]
}

// GetUserAudience maps user role to announcement audience.
// Assumption: roles are "teacher", "student", "parent", "staff", "admin", "principal".
// Admin and principal can see all.
func GetUserAudience(role string) []AnnouncementAudience {
	switch role {
	case "teacher":
		return []AnnouncementAudience{database.AnnouncementAudienceAll, database.AnnouncementAudienceTeachers}
	case "student":
		return []AnnouncementAudience{database.AnnouncementAudienceAll, database.AnnouncementAudienceStudents}
	case "parent":
		return []AnnouncementAudience{database.AnnouncementAudienceAll, database.AnnouncementAudienceParents}
	case "staff":
		return []AnnouncementAudience{database.AnnouncementAudienceAll, database.AnnouncementAudienceStaff}
	default: // admin, principal, etc.
		return []AnnouncementAudience{
			database.AnnouncementAudienceAll,
			database.AnnouncementAudienceTeachers,
			database.AnnouncementAudienceStudents,
			database.AnnouncementAudienceParents,
			database.AnnouncementAudienceStaff,
		}
	}
}
