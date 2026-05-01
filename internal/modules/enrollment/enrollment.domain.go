package enrollment

import (
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
)

type Enrollment = database.StudentEnrollment

// EnrollRequest creates a new enrollment for a student into a class section.
type EnrollRequest struct {
	StudentID      uuid.UUID `json:"student_id"       binding:"required,uuid"`
	ClassSectionID uuid.UUID `json:"class_section_id" binding:"required,uuid"`
	RollNumber     int       `json:"roll_number"      binding:"required,min=1"`
	EnrollmentDate time.Time `json:"enrollment_date"  binding:"required"`
}

// UpdateStatusRequest is used for promote / detain / withdraw actions.
// A dedicated request keeps the intent explicit — callers can't accidentally
// set an arbitrary status string via a generic update.
type UpdateStatusRequest struct {
	Status   database.EnrollmentStatus `json:"status"    binding:"required,oneof=ENROLLED PROMOTED DETAINED WITHDRAWN"`
	LeftDate *time.Time                `json:"left_date,omitempty"` // required when status = WITHDRAWN
}

// ──────────────────────────────────────────────────────────────
// RESPONSES
// ──────────────────────────────────────────────────────────────

type EnrollmentResponse struct {
	Enrollment
}

// RosterEntry is a lightweight row used in the class roster list —
// avoids sending full nested objects for every student.
type RosterEntry struct {
	EnrollmentID   uuid.UUID                 `json:"enrollment_id"`
	RollNumber     int                       `json:"roll_number"`
	Status         database.EnrollmentStatus `json:"status"`
	EnrollmentDate time.Time                 `json:"enrollment_date"`
	StudentID      uuid.UUID                 `json:"student_id"`
	FirstName      string                    `json:"first_name"`
	LastName       string                    `json:"last_name"`
	AdmissionNo    string                    `json:"admission_no"`
	Gender         database.Gender           `json:"gender"`
}

// ──────────────────────────────────────────────────────────────
// MAPPERS
// ──────────────────────────────────────────────────────────────

func ToEnrollmentResponse(e *Enrollment) *EnrollmentResponse {
	return &EnrollmentResponse{Enrollment: *e}
}

func ToRosterEntry(e *Enrollment) *RosterEntry {
	entry := &RosterEntry{
		EnrollmentID:   e.ID,
		RollNumber:     e.RollNumber,
		Status:         e.Status,
		EnrollmentDate: e.EnrollmentDate,
		StudentID:      e.StudentID,
	}
	// Student is preloaded by the roster query
	entry.FirstName = e.Student.FirstName
	entry.LastName = e.Student.LastName
	entry.AdmissionNo = e.Student.AdmissionNo
	entry.Gender = e.Student.Gender
	return entry
}
