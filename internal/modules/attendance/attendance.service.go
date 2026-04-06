package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Service interface ────────────────────────────────────────────────────────

type Service interface {
	// Student attendance
	MarkAttendance(ctx context.Context, req MarkStudentAttendanceRequest) (*AttendanceResponse, error)
	BulkMarkAttendance(ctx context.Context, req BulkMarkRequest) ([]*AttendanceResponse, error)
	GetAttendanceByID(ctx context.Context, id string) (*AttendanceResponse, error)
	ListStudentAttendance(ctx context.Context, q query_params.Query[StudentFilterParams]) ([]*AttendanceResponse, int64, error)
	GetClassAttendanceSummary(ctx context.Context, classSectionID, date string) (*ClassAttendanceSummary, error)
	UpdateAttendance(ctx context.Context, id string, req UpdateStudentAttendanceRequest) (*AttendanceResponse, error)
	DeleteAttendance(ctx context.Context, id string) error

	// Teacher attendance
	MarkTeacherAttendance(ctx context.Context, req MarkTeacherAttendanceRequest) (*TeacherAttendanceResponse, error)
	BulkMarkTeacherAttendance(ctx context.Context, req BulkMarkTeacherRequest) ([]*TeacherAttendanceResponse, error)
	GetTeacherAttendanceByID(ctx context.Context, id string) (*TeacherAttendanceResponse, error)
	ListTeacherAttendance(ctx context.Context, q query_params.Query[TeacherFilterParams]) ([]*TeacherAttendanceResponse, int64, error)
	UpdateTeacherAttendance(ctx context.Context, id string, req UpdateTeacherAttendanceRequest) (*TeacherAttendanceResponse, error)
	DeleteTeacherAttendance(ctx context.Context, id string) error
}

// ─── service struct ───────────────────────────────────────────────────────────

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ─── Student Attendance ───────────────────────────────────────────────────────

func (s *service) MarkAttendance(ctx context.Context, req MarkStudentAttendanceRequest) (*AttendanceResponse, error) {
	existing, err := s.repo.FindByEnrollmentAndDate(ctx, req.StudentEnrollmentID, req.Date)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("attendance.Service.MarkAttendance.FindExisting: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("attendance.Service.MarkAttendance: attendance already marked for enrollment %s on %s — use PUT to update", req.StudentEnrollmentID, req.Date.Format("2006-01-02"))
	}
	if !isValidStatus(req.Status) {
		return nil, fmt.Errorf("attendance.Service.MarkAttendance: invalid status %q", req.Status)
	}

	a := &Attendance{
		StudentEnrollmentID: req.StudentEnrollmentID,
		Date:                req.Date,
		Status:              req.Status,
		Remark:              req.Remark,
		RecordedByTeacherID: req.RecordedByTeacherID,
	}
	if err := s.repo.CreateAttendance(ctx, a); err != nil {
		return nil, fmt.Errorf("attendance.Service.MarkAttendance: %w", err)
	}
	return ToAttendanceResponse(a), nil
}

func (s *service) BulkMarkAttendance(ctx context.Context, req BulkMarkRequest) ([]*AttendanceResponse, error) {
	records := make([]*Attendance, 0, len(req.Records))
	for i, row := range req.Records {
		if !isValidStatus(row.Status) {
			return nil, fmt.Errorf("attendance.Service.BulkMarkAttendance: invalid status %q at index %d", row.Status, i)
		}
		existing, err := s.repo.FindByEnrollmentAndDate(ctx, row.StudentEnrollmentID, req.Date)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("attendance.Service.BulkMarkAttendance.FindExisting[%d]: %w", i, err)
		}
		if existing != nil {
			return nil, fmt.Errorf("attendance.Service.BulkMarkAttendance: attendance already marked for enrollment %s", row.StudentEnrollmentID)
		}
		records = append(records, &Attendance{
			StudentEnrollmentID: row.StudentEnrollmentID,
			Date:                req.Date,
			Status:              row.Status,
			Remark:              row.Remark,
			RecordedByTeacherID: req.RecordedByTeacherID,
		})
	}

	if err := s.repo.BulkCreateAttendance(ctx, records); err != nil {
		return nil, fmt.Errorf("attendance.Service.BulkMarkAttendance: %w", err)
	}

	responses := make([]*AttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToAttendanceResponse(a)
	}
	return responses, nil
}

func (s *service) GetAttendanceByID(ctx context.Context, id string) (*AttendanceResponse, error) {
	a, err := s.repo.GetAttendanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("attendance.Service.GetAttendanceByID: %w", err)
	}
	return ToAttendanceResponse(a), nil
}

func (s *service) ListStudentAttendance(ctx context.Context, q query_params.Query[StudentFilterParams]) ([]*AttendanceResponse, int64, error) {
	records, total, err := s.repo.FindStudentAttendance(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("attendance.Service.ListStudentAttendance: %w", err)
	}
	responses := make([]*AttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToAttendanceResponse(a)
	}
	return responses, total, nil
}

func (s *service) GetClassAttendanceSummary(ctx context.Context, classSectionID, dateStr string) (*ClassAttendanceSummary, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("attendance.Service.GetClassAttendanceSummary: invalid date format, use YYYY-MM-DD")
	}

	records, err := s.repo.FindByClassSectionAndDate(ctx, classSectionID, date)
	if err != nil {
		return nil, fmt.Errorf("attendance.Service.GetClassAttendanceSummary.FindRecords: %w", err)
	}

	present, absent, halfDay, late, leave, err := s.repo.CountByClassSectionAndDate(ctx, classSectionID, date)
	if err != nil {
		return nil, fmt.Errorf("attendance.Service.GetClassAttendanceSummary.Count: %w", err)
	}

	responses := make([]*AttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToAttendanceResponse(a)
	}

	return &ClassAttendanceSummary{
		Date:           date,
		ClassSectionID: classSectionID,
		TotalStudents:  len(records),
		Present:        int(present),
		Absent:         int(absent),
		HalfDay:        int(halfDay),
		Late:           int(late),
		OnLeave:        int(leave),
		Records:        responses,
	}, nil
}

func (s *service) UpdateAttendance(ctx context.Context, id string, req UpdateStudentAttendanceRequest) (*AttendanceResponse, error) {
	a, err := s.repo.GetAttendanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("attendance.Service.UpdateAttendance.GetByID: %w", err)
	}
	if !isValidStatus(req.Status) {
		return nil, fmt.Errorf("attendance.Service.UpdateAttendance: invalid status %q", req.Status)
	}
	a.Status = req.Status
	a.Remark = req.Remark
	if err := s.repo.UpdateAttendance(ctx, a); err != nil {
		return nil, fmt.Errorf("attendance.Service.UpdateAttendance: %w", err)
	}
	return ToAttendanceResponse(a), nil
}

func (s *service) DeleteAttendance(ctx context.Context, id string) error {
	if _, err := s.repo.GetAttendanceByID(ctx, id); err != nil {
		return fmt.Errorf("attendance.Service.DeleteAttendance.GetByID: %w", err)
	}
	if err := s.repo.DeleteAttendance(ctx, id); err != nil {
		return fmt.Errorf("attendance.Service.DeleteAttendance: %w", err)
	}
	return nil
}

// ─── Teacher Attendance ───────────────────────────────────────────────────────

func (s *service) MarkTeacherAttendance(ctx context.Context, req MarkTeacherAttendanceRequest) (*TeacherAttendanceResponse, error) {
	existing, err := s.repo.FindTeacherAttendanceByDate(ctx, req.TeacherID, req.Date)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("attendance.Service.MarkTeacherAttendance.FindExisting: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("attendance.Service.MarkTeacherAttendance: attendance already marked for teacher %s on %s — use PUT to update", req.TeacherID, req.Date.Format("2006-01-02"))
	}
	if !isValidStatus(req.Status) {
		return nil, fmt.Errorf("attendance.Service.MarkTeacherAttendance: invalid status %q", req.Status)
	}

	a := &TeacherAttendance{
		TeacherID: req.TeacherID,
		Date:      req.Date,
		Status:    req.Status,
		Remark:    req.Remark,
	}
	if err := s.repo.CreateTeacherAttendance(ctx, a); err != nil {
		return nil, fmt.Errorf("attendance.Service.MarkTeacherAttendance: %w", err)
	}
	return ToTeacherAttendanceResponse(a), nil
}

func (s *service) BulkMarkTeacherAttendance(ctx context.Context, req BulkMarkTeacherRequest) ([]*TeacherAttendanceResponse, error) {
	records := make([]*TeacherAttendance, 0, len(req.Records))
	for i, row := range req.Records {
		if !isValidStatus(row.Status) {
			return nil, fmt.Errorf("attendance.Service.BulkMarkTeacherAttendance: invalid status %q at index %d", row.Status, i)
		}
		existing, err := s.repo.FindTeacherAttendanceByDate(ctx, row.TeacherID, req.Date)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("attendance.Service.BulkMarkTeacherAttendance.FindExisting[%d]: %w", i, err)
		}
		if existing != nil {
			return nil, fmt.Errorf("attendance.Service.BulkMarkTeacherAttendance: attendance already marked for teacher %s", row.TeacherID)
		}
		records = append(records, &TeacherAttendance{
			TeacherID: row.TeacherID,
			Date:      req.Date,
			Status:    row.Status,
			Remark:    row.Remark,
		})
	}

	if err := s.repo.BulkCreateTeacherAttendance(ctx, records); err != nil {
		return nil, fmt.Errorf("attendance.Service.BulkMarkTeacherAttendance: %w", err)
	}

	responses := make([]*TeacherAttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToTeacherAttendanceResponse(a)
	}
	return responses, nil
}

func (s *service) GetTeacherAttendanceByID(ctx context.Context, id string) (*TeacherAttendanceResponse, error) {
	a, err := s.repo.GetTeacherAttendanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("attendance.Service.GetTeacherAttendanceByID: %w", err)
	}
	return ToTeacherAttendanceResponse(a), nil
}

func (s *service) ListTeacherAttendance(ctx context.Context, q query_params.Query[TeacherFilterParams]) ([]*TeacherAttendanceResponse, int64, error) {
	records, total, err := s.repo.FindTeacherAttendance(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("attendance.Service.ListTeacherAttendance: %w", err)
	}
	responses := make([]*TeacherAttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToTeacherAttendanceResponse(a)
	}
	return responses, total, nil
}

func (s *service) UpdateTeacherAttendance(ctx context.Context, id string, req UpdateTeacherAttendanceRequest) (*TeacherAttendanceResponse, error) {
	a, err := s.repo.GetTeacherAttendanceByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("attendance.Service.UpdateTeacherAttendance.GetByID: %w", err)
	}
	if !isValidStatus(req.Status) {
		return nil, fmt.Errorf("attendance.Service.UpdateTeacherAttendance: invalid status %q", req.Status)
	}
	a.Status = req.Status
	a.Remark = req.Remark
	if err := s.repo.UpdateTeacherAttendance(ctx, a); err != nil {
		return nil, fmt.Errorf("attendance.Service.UpdateTeacherAttendance: %w", err)
	}
	return ToTeacherAttendanceResponse(a), nil
}

func (s *service) DeleteTeacherAttendance(ctx context.Context, id string) error {
	if _, err := s.repo.GetTeacherAttendanceByID(ctx, id); err != nil {
		return fmt.Errorf("attendance.Service.DeleteTeacherAttendance.GetByID: %w", err)
	}
	if err := s.repo.DeleteTeacherAttendance(ctx, id); err != nil {
		return fmt.Errorf("attendance.Service.DeleteTeacherAttendance: %w", err)
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func isValidStatus(s database.AttendanceStatus) bool {
	switch s {
	case database.AttendanceStatusPresent,
		database.AttendanceStatusAbsent,
		database.AttendanceStatusHalfDay,
		database.AttendanceStatusLate,
		database.AttendanceStatusLeave:
		return true
	}
	return false
}
