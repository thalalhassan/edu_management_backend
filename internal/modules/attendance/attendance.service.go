package attendance

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/apperrors"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Service interface ───────────────────────────────────────────────────────

type Service interface {
	// Student attendance
	MarkAttendance(ctx context.Context, req MarkStudentAttendanceRequest) (*AttendanceResponse, error)
	BulkMarkAttendance(ctx context.Context, req BulkMarkRequest) ([]*AttendanceResponse, error)
	GetAttendanceByID(ctx context.Context, id uuid.UUID) (*AttendanceResponse, error)
	ListStudentAttendance(ctx context.Context, q query_params.Query[StudentFilterParams]) ([]*AttendanceResponse, int64, error)
	GetClassAttendanceSummary(ctx context.Context, classSectionID uuid.UUID, date string) (*ClassAttendanceSummary, error)
	UpdateAttendance(ctx context.Context, id uuid.UUID, req UpdateStudentAttendanceRequest) (*AttendanceResponse, error)
	DeleteAttendance(ctx context.Context, id uuid.UUID) error

	// Employee attendance
	MarkEmployeeAttendance(ctx context.Context, req MarkEmployeeAttendanceRequest) (*EmployeeAttendanceResponse, error)
	BulkMarkEmployeeAttendance(ctx context.Context, req BulkMarkEmployeeRequest) ([]*EmployeeAttendanceResponse, error)
	GetEmployeeAttendanceByID(ctx context.Context, id uuid.UUID) (*EmployeeAttendanceResponse, error)
	ListEmployeeAttendance(ctx context.Context, q query_params.Query[EmployeeFilterParams]) ([]*EmployeeAttendanceResponse, int64, error)
	UpdateEmployeeAttendance(ctx context.Context, id uuid.UUID, req UpdateEmployeeAttendanceRequest) (*EmployeeAttendanceResponse, error)
	DeleteEmployeeAttendance(ctx context.Context, id uuid.UUID) error
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
	if err := req.Validate(); err != nil {
		return nil, err
	}

	req.Date = NormalizeDate(req.Date)
	existing, err := s.repo.FindByEnrollmentAndDate(ctx, req.StudentEnrollmentID, req.Date)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to check existing attendance")
	}
	if existing != nil {
		return nil, apperrors.Newf(apperrors.ErrAlreadyExists, "attendance already exists for enrollment %s on %s", req.StudentEnrollmentID, req.Date.Format("2006-01-02"))
	}

	a := &Attendance{
		StudentEnrollmentID: req.StudentEnrollmentID,
		Date:                req.Date,
		Status:              req.Status,
		Remark:              req.Remark,
		RecordedByID:        req.RecordedByID,
	}
	if err := s.repo.CreateAttendance(ctx, a); err != nil {
		return nil, mapCreateError(err, "failed to create attendance")
	}
	return ToAttendanceResponse(a), nil
}

func (s *service) BulkMarkAttendance(ctx context.Context, req BulkMarkRequest) ([]*AttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	date := NormalizeDate(req.Date)
	enrollmentIDs := make([]uuid.UUID, 0, len(req.Records))
	for _, row := range req.Records {
		enrollmentIDs = append(enrollmentIDs, row.StudentEnrollmentID)
	}

	existing, err := s.repo.FindExistingAttendanceForEnrollmentsAndDate(ctx, enrollmentIDs, date)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to check existing attendance")
	}
	if len(existing) > 0 {
		return nil, apperrors.Newf(apperrors.ErrAlreadyExists, "attendance already exists for %d student(s) on %s", len(existing), date.Format("2006-01-02"))
	}

	records := make([]*Attendance, 0, len(req.Records))
	for _, row := range req.Records {
		records = append(records, &Attendance{
			StudentEnrollmentID: row.StudentEnrollmentID,
			Date:                date,
			Status:              row.Status,
			Remark:              row.Remark,
			RecordedByID:        req.RecordedByID,
		})
	}

	if err := s.repo.BulkCreateAttendance(ctx, records); err != nil {
		return nil, mapCreateError(err, "failed to create attendance records")
	}

	responses := make([]*AttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToAttendanceResponse(a)
	}
	return responses, nil
}

func (s *service) GetAttendanceByID(ctx context.Context, id uuid.UUID) (*AttendanceResponse, error) {
	a, err := s.repo.GetAttendanceByID(ctx, id)
	if err != nil {
		return nil, mapGetError(err, "failed to retrieve attendance")
	}
	return ToAttendanceResponse(a), nil
}

func (s *service) ListStudentAttendance(ctx context.Context, q query_params.Query[StudentFilterParams]) ([]*AttendanceResponse, int64, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, 0, err
	}

	records, total, err := s.repo.FindStudentAttendance(ctx, q)
	if err != nil {
		return nil, 0, apperrors.New(apperrors.ErrDatabase, "failed to list attendance")
	}
	responses := make([]*AttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToAttendanceResponse(a)
	}
	return responses, total, nil
}

func (s *service) GetClassAttendanceSummary(ctx context.Context, classSectionID uuid.UUID, dateStr string) (*ClassAttendanceSummary, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrInvalidFormat, "date must be YYYY-MM-DD")
	}

	records, err := s.repo.FindByClassSectionAndDate(ctx, classSectionID, date)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to load class attendance")
	}

	present, absent, halfDay, late, leave, err := s.repo.CountByClassSectionAndDate(ctx, classSectionID, date)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to count class attendance")
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

func (s *service) UpdateAttendance(ctx context.Context, id uuid.UUID, req UpdateStudentAttendanceRequest) (*AttendanceResponse, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	a, err := s.repo.GetAttendanceByID(ctx, id)
	if err != nil {
		return nil, mapGetError(err, "failed to retrieve attendance")
	}
	a.Status = req.Status
	a.Remark = req.Remark
	if err := s.repo.UpdateAttendance(ctx, a); err != nil {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to update attendance")
	}
	return ToAttendanceResponse(a), nil
}

func (s *service) DeleteAttendance(ctx context.Context, id uuid.UUID) error {

	if err := s.repo.DeleteAttendance(ctx, id); err != nil {
		return mapGetError(err, "failed to delete attendance")
	}
	return nil
}

// ─── Employee Attendance ───────────────────────────────────────────────────────

func (s *service) MarkEmployeeAttendance(ctx context.Context, req MarkEmployeeAttendanceRequest) (*EmployeeAttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	req.Date = NormalizeDate(req.Date)
	existing, err := s.repo.FindEmployeeAttendanceByDate(ctx, req.EmployeeID, req.Date)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to check existing employee attendance")
	}
	if existing != nil {
		return nil, apperrors.Newf(apperrors.ErrAlreadyExists, "employee attendance already exists for %s on %s", req.EmployeeID, req.Date.Format("2006-01-02"))
	}

	a := &EmployeeAttendance{
		EmployeeID: req.EmployeeID,
		Date:       req.Date,
		Status:     req.Status,
		Remark:     req.Remark,
	}
	if err := s.repo.CreateEmployeeAttendance(ctx, a); err != nil {
		return nil, mapCreateError(err, "failed to create employee attendance")
	}
	return ToEmployeeAttendanceResponse(a), nil
}

func (s *service) BulkMarkEmployeeAttendance(ctx context.Context, req BulkMarkEmployeeRequest) ([]*EmployeeAttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	date := NormalizeDate(req.Date)
	employeeIDs := make([]uuid.UUID, 0, len(req.Records))
	for _, row := range req.Records {
		employeeIDs = append(employeeIDs, row.EmployeeID)
	}

	existing, err := s.repo.FindExistingEmployeeAttendanceForEmployeeIDsAndDate(ctx, employeeIDs, date)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to check existing employee attendance")
	}
	if len(existing) > 0 {
		return nil, apperrors.Newf(apperrors.ErrAlreadyExists, "employee attendance already exists for %d record(s) on %s", len(existing), date.Format("2006-01-02"))
	}

	records := make([]*EmployeeAttendance, 0, len(req.Records))
	for _, row := range req.Records {
		records = append(records, &EmployeeAttendance{
			EmployeeID: row.EmployeeID,
			Date:       date,
			Status:     row.Status,
			Remark:     row.Remark,
		})
	}

	if err := s.repo.BulkCreateEmployeeAttendance(ctx, records); err != nil {
		return nil, mapCreateError(err, "failed to create employee attendance records")
	}

	responses := make([]*EmployeeAttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToEmployeeAttendanceResponse(a)
	}
	return responses, nil
}

func (s *service) GetEmployeeAttendanceByID(ctx context.Context, id uuid.UUID) (*EmployeeAttendanceResponse, error) {
	a, err := s.repo.GetEmployeeAttendanceByID(ctx, id)
	if err != nil {
		return nil, mapGetError(err, "failed to retrieve employee attendance")
	}
	return ToEmployeeAttendanceResponse(a), nil
}

func (s *service) ListEmployeeAttendance(ctx context.Context, q query_params.Query[EmployeeFilterParams]) ([]*EmployeeAttendanceResponse, int64, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, 0, err
	}

	records, total, err := s.repo.FindEmployeeAttendance(ctx, q)
	if err != nil {
		return nil, 0, apperrors.New(apperrors.ErrDatabase, "failed to list employee attendance")
	}
	responses := make([]*EmployeeAttendanceResponse, len(records))
	for i, a := range records {
		responses[i] = ToEmployeeAttendanceResponse(a)
	}
	return responses, total, nil
}

func (s *service) UpdateEmployeeAttendance(ctx context.Context, id uuid.UUID, req UpdateEmployeeAttendanceRequest) (*EmployeeAttendanceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	a, err := s.repo.GetEmployeeAttendanceByID(ctx, id)
	if err != nil {
		return nil, mapGetError(err, "failed to retrieve employee attendance")
	}
	a.Status = req.Status
	a.Remark = req.Remark
	if err := s.repo.UpdateEmployeeAttendance(ctx, a); err != nil {
		return nil, apperrors.New(apperrors.ErrDatabase, "failed to update employee attendance")
	}
	return ToEmployeeAttendanceResponse(a), nil
}

func (s *service) DeleteEmployeeAttendance(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteEmployeeAttendance(ctx, id); err != nil {
		return mapGetError(err, "failed to delete employee attendance")
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func isValidStatus(s AttendanceStatus) bool {
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

func ToAttendanceResponse(a *Attendance) *AttendanceResponse {
	if a == nil {
		return nil
	}
	return &AttendanceResponse{
		ID:                  a.ID,
		StudentEnrollmentID: a.StudentEnrollmentID,
		Date:                a.Date,
		Status:              a.Status,
		Remark:              a.Remark,
		RecordedByID:        a.RecordedByID,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}

func ToEmployeeAttendanceResponse(a *EmployeeAttendance) *EmployeeAttendanceResponse {
	if a == nil {
		return nil
	}
	return &EmployeeAttendanceResponse{
		ID:         a.ID,
		EmployeeID: a.EmployeeID,
		Date:       a.Date,
		Status:     a.Status,
		Remark:     a.Remark,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
