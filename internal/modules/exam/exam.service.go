package exam

import (
	"context"
	"errors"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

// ─── Service interface ────────────────────────────────────────────────────────

type Service interface {
	// Exam
	CreateExam(ctx context.Context, req CreateExamRequest) (*ExamResponse, error)
	GetExamByID(ctx context.Context, id string) (*ExamResponse, error)
	ListExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*ExamResponse, int64, error)
	UpdateExam(ctx context.Context, id string, req UpdateExamRequest) (*ExamResponse, error)
	PublishExam(ctx context.Context, id string, req PublishExamRequest) (*ExamResponse, error)
	DeleteExam(ctx context.Context, id string) error

	// ExamSchedule
	CreateSchedule(ctx context.Context, examID string, req CreateScheduleRequest) (*ExamScheduleResponse, error)
	GetScheduleByID(ctx context.Context, id string) (*ExamScheduleResponse, error)
	ListSchedulesByExam(ctx context.Context, examID string) ([]*ExamScheduleResponse, error)
	ListSchedulesByClassSection(ctx context.Context, classSectionID string) ([]*ExamScheduleResponse, error)
	UpdateSchedule(ctx context.Context, id string, req UpdateScheduleRequest) (*ExamScheduleResponse, error)
	DeleteSchedule(ctx context.Context, id string) error

	// ExamResult
	CreateResult(ctx context.Context, scheduleID string, req CreateResultRequest) (*ExamResultResponse, error)
	BulkCreateResults(ctx context.Context, scheduleID string, req BulkCreateResultRequest) ([]*ExamResultResponse, error)
	GetResultByID(ctx context.Context, id string) (*ExamResultResponse, error)
	ListResultsBySchedule(ctx context.Context, scheduleID string) ([]*ExamResultResponse, error)
	ListResultsByStudent(ctx context.Context, studentEnrollmentID string) ([]*ExamResultResponse, error)
	UpdateResult(ctx context.Context, id string, req UpdateResultRequest) (*ExamResultResponse, error)
	DeleteResult(ctx context.Context, id string) error
}

// ─── service struct ───────────────────────────────────────────────────────────

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// deriveStatus computes PASS / FAIL / ABSENT from marks and passing threshold.
func deriveStatus(marks, passingMarks interface {
	Cmp(interface{ Cmp(interface{}) int }) int
}) database.ExamResultStatus {
	// Use shopspring/decimal comparison via marks.Cmp
	return database.ExamResultStatusPass // handled inline below using decimal
}

// ─── Exam ─────────────────────────────────────────────────────────────────────

func (s *service) CreateExam(ctx context.Context, req CreateExamRequest) (*ExamResponse, error) {
	if !req.EndDate.After(req.StartDate) {
		return nil, errors.New("exam.Service.CreateExam: end_date must be after start_date")
	}

	isDup, err := s.repo.IsDuplicateExamName(ctx, req.AcademicYearID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.CreateExam.IsDuplicateName: %w", err)
	}
	if isDup {
		return nil, fmt.Errorf("exam.Service.CreateExam: exam %q already exists for this academic year", req.Name)
	}

	e := &Exam{
		AcademicYearID: req.AcademicYearID,
		Name:           req.Name,
		Description:    req.Description,
		ExamType:       req.ExamType,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		IsPublished:    false,
	}
	if err := s.repo.CreateExam(ctx, e); err != nil {
		return nil, fmt.Errorf("exam.Service.CreateExam: %w", err)
	}
	return ToExamResponse(e), nil
}

func (s *service) GetExamByID(ctx context.Context, id string) (*ExamResponse, error) {
	e, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.GetExamByID: %w", err)
	}
	return ToExamResponse(e), nil
}

func (s *service) ListExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*ExamResponse, int64, error) {
	exams, total, err := s.repo.FindAllExams(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("exam.Service.ListExams: %w", err)
	}
	responses := make([]*ExamResponse, len(exams))
	for i, e := range exams {
		responses[i] = ToExamResponse(e)
	}
	return responses, total, nil
}

func (s *service) UpdateExam(ctx context.Context, id string, req UpdateExamRequest) (*ExamResponse, error) {
	e, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateExam.GetByID: %w", err)
	}

	if req.Name != nil && *req.Name != e.Name {
		isDup, err := s.repo.IsDuplicateExamName(ctx, e.AcademicYearID, *req.Name)
		if err != nil {
			return nil, fmt.Errorf("exam.Service.UpdateExam.IsDuplicateName: %w", err)
		}
		if isDup {
			return nil, fmt.Errorf("exam.Service.UpdateExam: exam name %q already in use for this academic year", *req.Name)
		}
		e.Name = *req.Name
	}
	if req.Description != nil {
		e.Description = req.Description
	}
	if req.ExamType != nil {
		e.ExamType = *req.ExamType
	}
	if req.StartDate != nil {
		e.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		e.EndDate = *req.EndDate
	}

	if !e.EndDate.After(e.StartDate) {
		return nil, errors.New("exam.Service.UpdateExam: end_date must be after start_date")
	}

	if err := s.repo.UpdateExam(ctx, e); err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateExam: %w", err)
	}
	return ToExamResponse(e), nil
}

func (s *service) PublishExam(ctx context.Context, id string, req PublishExamRequest) (*ExamResponse, error) {
	e, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.PublishExam.GetByID: %w", err)
	}

	// Cannot publish an exam with no schedules
	if req.IsPublished {
		hasSchedules, err := s.repo.HasSchedules(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("exam.Service.PublishExam.HasSchedules: %w", err)
		}
		if !hasSchedules {
			return nil, errors.New("exam.Service.PublishExam: cannot publish exam with no schedules — add at least one schedule first")
		}
	}

	e.IsPublished = req.IsPublished
	if err := s.repo.UpdateExam(ctx, e); err != nil {
		return nil, fmt.Errorf("exam.Service.PublishExam: %w", err)
	}
	return ToExamResponse(e), nil
}

func (s *service) DeleteExam(ctx context.Context, id string) error {
	e, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		return fmt.Errorf("exam.Service.DeleteExam.GetByID: %w", err)
	}
	if e.IsPublished {
		return errors.New("exam.Service.DeleteExam: cannot delete a published exam — unpublish it first")
	}
	if err := s.repo.DeleteExam(ctx, id); err != nil {
		return fmt.Errorf("exam.Service.DeleteExam: %w", err)
	}
	return nil
}

// ─── ExamSchedule ─────────────────────────────────────────────────────────────

func (s *service) CreateSchedule(ctx context.Context, examID string, req CreateScheduleRequest) (*ExamScheduleResponse, error) {
	// Confirm parent exam exists
	e, err := s.repo.GetExamByID(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.CreateSchedule.GetExam: %w", err)
	}

	// Guard: exam date must fall within exam window
	if req.ExamDate.Before(e.StartDate) || req.ExamDate.After(e.EndDate) {
		return nil, fmt.Errorf("exam.Service.CreateSchedule: exam_date must be between exam start_date and end_date")
	}

	// Guard: passing marks cannot exceed max marks
	if req.PassingMarks.GreaterThan(req.MaxMarks) {
		return nil, errors.New("exam.Service.CreateSchedule: passing_marks cannot exceed max_marks")
	}

	// Guard: no duplicate subject for the same class section within this exam
	isDup, err := s.repo.IsDuplicateSchedule(ctx, examID, req.ClassSectionID, req.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.CreateSchedule.IsDuplicateSchedule: %w", err)
	}
	if isDup {
		return nil, errors.New("exam.Service.CreateSchedule: a schedule for this class section and subject already exists in this exam")
	}

	sch := &ExamSchedule{
		ExamID:         examID,
		ClassSectionID: req.ClassSectionID,
		SubjectID:      req.SubjectID,
		ExamDate:       req.ExamDate,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		MaxMarks:       req.MaxMarks,
		PassingMarks:   req.PassingMarks,
		RoomID:         req.RoomID,
	}
	if err := s.repo.CreateSchedule(ctx, sch); err != nil {
		return nil, fmt.Errorf("exam.Service.CreateSchedule: %w", err)
	}
	return ToScheduleResponse(sch), nil
}

func (s *service) GetScheduleByID(ctx context.Context, id string) (*ExamScheduleResponse, error) {
	sch, err := s.repo.GetScheduleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.GetScheduleByID: %w", err)
	}
	return ToScheduleResponse(sch), nil
}

func (s *service) ListSchedulesByExam(ctx context.Context, examID string) ([]*ExamScheduleResponse, error) {
	if _, err := s.repo.GetExamByID(ctx, examID); err != nil {
		return nil, fmt.Errorf("exam.Service.ListSchedulesByExam.GetExam: %w", err)
	}
	schedules, err := s.repo.FindSchedulesByExam(ctx, examID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.ListSchedulesByExam: %w", err)
	}
	responses := make([]*ExamScheduleResponse, len(schedules))
	for i, sch := range schedules {
		responses[i] = ToScheduleResponse(sch)
	}
	return responses, nil
}

func (s *service) ListSchedulesByClassSection(ctx context.Context, classSectionID string) ([]*ExamScheduleResponse, error) {
	schedules, err := s.repo.FindSchedulesByClassSection(ctx, classSectionID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.ListSchedulesByClassSection: %w", err)
	}
	responses := make([]*ExamScheduleResponse, len(schedules))
	for i, sch := range schedules {
		responses[i] = ToScheduleResponse(sch)
	}
	return responses, nil
}

func (s *service) UpdateSchedule(ctx context.Context, id string, req UpdateScheduleRequest) (*ExamScheduleResponse, error) {
	sch, err := s.repo.GetScheduleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateSchedule.GetByID: %w", err)
	}

	// Cannot update a schedule that already has results entered
	hasResults, err := s.repo.HasResults(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateSchedule.HasResults: %w", err)
	}
	if hasResults {
		return nil, errors.New("exam.Service.UpdateSchedule: cannot update schedule that already has results — delete results first")
	}

	if req.ExamDate != nil {
		sch.ExamDate = *req.ExamDate
	}
	if req.StartTime != nil {
		sch.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		sch.EndTime = req.EndTime
	}
	if req.MaxMarks != nil {
		sch.MaxMarks = *req.MaxMarks
	}
	if req.PassingMarks != nil {
		sch.PassingMarks = *req.PassingMarks
	}
	if req.RoomID != nil {
		sch.RoomID = req.RoomID
	}

	if sch.PassingMarks.GreaterThan(sch.MaxMarks) {
		return nil, errors.New("exam.Service.UpdateSchedule: passing_marks cannot exceed max_marks")
	}

	if err := s.repo.UpdateSchedule(ctx, sch); err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateSchedule: %w", err)
	}
	return ToScheduleResponse(sch), nil
}

func (s *service) DeleteSchedule(ctx context.Context, id string) error {
	hasResults, err := s.repo.HasResults(ctx, id)
	if err != nil {
		return fmt.Errorf("exam.Service.DeleteSchedule.HasResults: %w", err)
	}
	if hasResults {
		return errors.New("exam.Service.DeleteSchedule: cannot delete schedule with existing results")
	}
	if err := s.repo.DeleteSchedule(ctx, id); err != nil {
		return fmt.Errorf("exam.Service.DeleteSchedule: %w", err)
	}
	return nil
}

// ─── ExamResult ───────────────────────────────────────────────────────────────

func (s *service) CreateResult(ctx context.Context, scheduleID string, req CreateResultRequest) (*ExamResultResponse, error) {
	maxMarks, passingMarks, err := s.repo.GetScheduleMarks(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.CreateResult.GetScheduleMarks: %w", err)
	}

	if req.MarksObtained.GreaterThan(maxMarks) {
		return nil, fmt.Errorf("exam.Service.CreateResult: marks_obtained (%s) exceeds max_marks (%s)", req.MarksObtained, maxMarks)
	}

	isDup, err := s.repo.IsDuplicateResult(ctx, scheduleID, req.StudentEnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.CreateResult.IsDuplicate: %w", err)
	}
	if isDup {
		return nil, errors.New("exam.Service.CreateResult: result already exists for this student in this schedule")
	}

	status := database.ExamResultStatusPass
	if req.MarksObtained.LessThan(passingMarks) {
		status = database.ExamResultStatusFail
	}

	res := &ExamResult{
		ExamScheduleID:      scheduleID,
		StudentEnrollmentID: req.StudentEnrollmentID,
		MarksObtained:       req.MarksObtained,
		Grade:               req.Grade,
		Status:              status,
		Remarks:             req.Remarks,
	}
	if err := s.repo.CreateResult(ctx, res); err != nil {
		return nil, fmt.Errorf("exam.Service.CreateResult: %w", err)
	}
	return ToResultResponse(res), nil
}

func (s *service) BulkCreateResults(ctx context.Context, scheduleID string, req BulkCreateResultRequest) ([]*ExamResultResponse, error) {
	maxMarks, passingMarks, err := s.repo.GetScheduleMarks(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.BulkCreateResults.GetScheduleMarks: %w", err)
	}

	results := make([]*ExamResult, 0, len(req.Results))
	for i, r := range req.Results {
		if (*r.MarksObtained).GreaterThan(maxMarks) {
			return nil, fmt.Errorf("exam.Service.BulkCreateResults: marks_obtained at index %d exceeds max_marks", i)
		}

		isDup, err := s.repo.IsDuplicateResult(ctx, scheduleID, r.StudentEnrollmentID)
		if err != nil {
			return nil, fmt.Errorf("exam.Service.BulkCreateResults.IsDuplicate[%d]: %w", i, err)
		}
		if isDup {
			return nil, fmt.Errorf("exam.Service.BulkCreateResults: duplicate result at index %d for enrollment %s", i, r.StudentEnrollmentID)
		}

		status := database.ExamResultStatusPass
		if (*r.MarksObtained).LessThan(passingMarks) {
			status = database.ExamResultStatusFail
		}
		results = append(results, &ExamResult{
			ExamScheduleID:      scheduleID,
			StudentEnrollmentID: r.StudentEnrollmentID,
			MarksObtained:       r.MarksObtained,
			Grade:               r.Grade,
			Status:              status,
			Remarks:             r.Remarks,
		})
	}

	if err := s.repo.BulkCreateResults(ctx, results); err != nil {
		return nil, fmt.Errorf("exam.Service.BulkCreateResults: %w", err)
	}

	responses := make([]*ExamResultResponse, len(results))
	for i, res := range results {
		responses[i] = ToResultResponse(res)
	}
	return responses, nil
}

func (s *service) GetResultByID(ctx context.Context, id string) (*ExamResultResponse, error) {
	res, err := s.repo.GetResultByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.GetResultByID: %w", err)
	}
	return ToResultResponse(res), nil
}

func (s *service) ListResultsBySchedule(ctx context.Context, scheduleID string) ([]*ExamResultResponse, error) {
	results, err := s.repo.FindResultsBySchedule(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.ListResultsBySchedule: %w", err)
	}
	responses := make([]*ExamResultResponse, len(results))
	for i, r := range results {
		responses[i] = ToResultResponse(r)
	}
	return responses, nil
}

func (s *service) ListResultsByStudent(ctx context.Context, studentEnrollmentID string) ([]*ExamResultResponse, error) {
	results, err := s.repo.FindResultsByStudent(ctx, studentEnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.ListResultsByStudent: %w", err)
	}
	responses := make([]*ExamResultResponse, len(results))
	for i, r := range results {
		responses[i] = ToResultResponse(r)
	}
	return responses, nil
}

func (s *service) UpdateResult(ctx context.Context, id string, req UpdateResultRequest) (*ExamResultResponse, error) {
	res, err := s.repo.GetResultByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateResult.GetByID: %w", err)
	}

	if req.MarksObtained != nil {
		maxMarks, passingMarks, err := s.repo.GetScheduleMarks(ctx, res.ExamScheduleID)
		if err != nil {
			return nil, fmt.Errorf("exam.Service.UpdateResult.GetScheduleMarks: %w", err)
		}
		if (*req.MarksObtained).GreaterThan(maxMarks) {
			return nil, fmt.Errorf("exam.Service.UpdateResult: marks_obtained exceeds max_marks (%s)", maxMarks)
		}
		res.MarksObtained = req.MarksObtained
		// Recompute status
		if (*res.MarksObtained).LessThan(passingMarks) {
			res.Status = database.ExamResultStatusFail
		} else {
			res.Status = database.ExamResultStatusPass
		}
	}
	if req.Grade != nil {
		res.Grade = req.Grade
	}
	if req.Remarks != nil {
		res.Remarks = req.Remarks
	}

	if err := s.repo.UpdateResult(ctx, res); err != nil {
		return nil, fmt.Errorf("exam.Service.UpdateResult: %w", err)
	}
	return ToResultResponse(res), nil
}

func (s *service) DeleteResult(ctx context.Context, id string) error {
	if _, err := s.repo.GetResultByID(ctx, id); err != nil {
		return fmt.Errorf("exam.Service.DeleteResult.GetByID: %w", err)
	}
	if err := s.repo.DeleteResult(ctx, id); err != nil {
		return fmt.Errorf("exam.Service.DeleteResult: %w", err)
	}
	return nil
}
