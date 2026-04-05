package enrollment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Service interface {
	Enroll(ctx context.Context, req EnrollRequest) (*EnrollmentResponse, error)
	GetByID(ctx context.Context, id string) (*EnrollmentResponse, error)
	GetByStudentID(ctx context.Context, studentID string, p pagination.Params) ([]*EnrollmentResponse, int64, error)
	GetRoster(ctx context.Context, classSectionID string) ([]*RosterEntry, error)
	UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*EnrollmentResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo        Repository
	classSvcDB  *gorm.DB // raw db for class section lookups — avoids circular module import
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{repo: repo, classSvcDB: db}
}

func (s *service) Enroll(ctx context.Context, req EnrollRequest) (*EnrollmentResponse, error) {
	// 1. Fetch the class section to validate it exists and check capacity
	var classSection database.ClassSection
	if err := s.classSvcDB.WithContext(ctx).
		Preload("AcademicYear").
		First(&classSection, "id = ?", req.ClassSectionID).Error; err != nil {
		return nil, fmt.Errorf("enrollment.Service.Enroll: class section not found: %w", err)
	}

	// 2. Reject if academic year is not active
	if !classSection.AcademicYear.IsActive {
		return nil, errors.New("enrollment.Service.Enroll: cannot enroll into an inactive academic year")
	}

	// 3. Reject if student is already enrolled in this academic year
	alreadyEnrolled, err := s.repo.IsStudentEnrolledInYear(ctx, req.StudentID, classSection.AcademicYearID)
	if err != nil {
		return nil, fmt.Errorf("enrollment.Service.Enroll.IsStudentEnrolledInYear: %w", err)
	}
	if alreadyEnrolled {
		return nil, errors.New("enrollment.Service.Enroll: student is already enrolled in this academic year")
	}

	// 4. Enforce class section capacity
	enrolled, err := s.repo.CountEnrolledInClassSection(ctx, req.ClassSectionID)
	if err != nil {
		return nil, fmt.Errorf("enrollment.Service.Enroll.CountEnrolled: %w", err)
	}
	if int(enrolled) >= classSection.MaxStrength {
		return nil, fmt.Errorf("enrollment.Service.Enroll: class section is full (%d/%d)", enrolled, classSection.MaxStrength)
	}

	e := &Enrollment{
		StudentID:      req.StudentID,
		ClassSectionID: req.ClassSectionID,
		RollNumber:     req.RollNumber,
		Status:         database.EnrollmentStatusEnrolled,
		EnrollmentDate: req.EnrollmentDate,
	}
	if e.EnrollmentDate.IsZero() {
		e.EnrollmentDate = time.Now()
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("enrollment.Service.Enroll.Create: %w", err)
	}

	// Reload with full relations for the response
	created, err := s.repo.GetByID(ctx, e.ID)
	if err != nil {
		return nil, fmt.Errorf("enrollment.Service.Enroll.GetByID: %w", err)
	}
	return ToEnrollmentResponse(created), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*EnrollmentResponse, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("enrollment.Service.GetByID: %w", err)
	}
	return ToEnrollmentResponse(e), nil
}

func (s *service) GetByStudentID(ctx context.Context, studentID string, p pagination.Params) ([]*EnrollmentResponse, int64, error) {
	enrollments, total, err := s.repo.GetByStudentID(ctx, studentID, p)
	if err != nil {
		return nil, 0, fmt.Errorf("enrollment.Service.GetByStudentID: %w", err)
	}
	responses := make([]*EnrollmentResponse, len(enrollments))
	for i, e := range enrollments {
		responses[i] = ToEnrollmentResponse(e)
	}
	return responses, total, nil
}

func (s *service) GetRoster(ctx context.Context, classSectionID string) ([]*RosterEntry, error) {
	enrollments, err := s.repo.GetRosterByClassSection(ctx, classSectionID)
	if err != nil {
		return nil, fmt.Errorf("enrollment.Service.GetRoster: %w", err)
	}
	entries := make([]*RosterEntry, len(enrollments))
	for i, e := range enrollments {
		entries[i] = ToRosterEntry(e)
	}
	return entries, nil
}

// UpdateStatus handles promote, detain, and withdraw transitions.
// Business rules:
//   - PROMOTED / DETAINED: terminal states for the academic year — no further changes allowed
//   - WITHDRAWN: requires a LeftDate
//   - ENROLLED → any transition is valid
func (s *service) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*EnrollmentResponse, error) {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("enrollment.Service.UpdateStatus.GetByID: %w", err)
	}

	// Guard: cannot transition out of a terminal state
	if e.Status == database.EnrollmentStatusPromoted || e.Status == database.EnrollmentStatusDetained {
		return nil, fmt.Errorf("enrollment.Service.UpdateStatus: cannot change status from %s", e.Status)
	}

	// Guard: withdrawal requires a left date
	if req.Status == database.EnrollmentStatusWithdrawn && req.LeftDate == nil {
		return nil, errors.New("enrollment.Service.UpdateStatus: left_date is required when withdrawing")
	}

	e.Status = req.Status
	if req.LeftDate != nil {
		e.LeftDate = req.LeftDate
	}

	if err := s.repo.Update(ctx, id, e); err != nil {
		return nil, fmt.Errorf("enrollment.Service.UpdateStatus.Update: %w", err)
	}
	return ToEnrollmentResponse(e), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	e, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("enrollment.Service.Delete.GetByID: %w", err)
	}
	// Prevent deleting an enrollment that has attendance or exam results
	if len(e.Attendances) > 0 || len(e.ExamResults) > 0 {
		return errors.New("enrollment.Service.Delete: cannot delete enrollment with existing attendance or exam records — withdraw instead")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("enrollment.Service.Delete: %w", err)
	}
	return nil
}
