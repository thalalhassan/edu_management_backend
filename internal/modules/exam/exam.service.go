package exam

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

type Code string

const (
	CodeNotFound     Code = "not_found"
	CodeDuplicate    Code = "duplicate"
	CodeValidation   Code = "validation_error"
	CodeBusinessRule Code = "business_rule"
	CodeInternal     Code = "internal"
)

type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Detail == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Detail)
}

func NewAppError(code Code, message string, detail ...string) *AppError {
	err := &AppError{Code: code, Message: message}
	if len(detail) > 0 {
		err.Detail = detail[0]
	}
	return err
}

func (e *AppError) Unwrap() error {
	return nil
}

func mapRepoError(err error, entity string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewAppError(CodeNotFound, fmt.Sprintf("%s not found", entity), err.Error())
	}
	return err
}

// Service defines the exam use case contract.
type Service interface {
	CreateExam(ctx context.Context, req CreateExamRequest) (*ExamResponse, error)
	GetExamByID(ctx context.Context, id uuid.UUID) (*ExamResponse, error)
	ListExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*ExamResponse, int64, error)
	UpdateExam(ctx context.Context, id uuid.UUID, req UpdateExamRequest) (*ExamResponse, error)
	PublishExam(ctx context.Context, id uuid.UUID, req PublishExamRequest) (*ExamResponse, error)
	DeleteExam(ctx context.Context, id uuid.UUID) error

	CreateSchedule(ctx context.Context, examID uuid.UUID, req CreateScheduleRequest) (*ExamScheduleResponse, error)
	GetScheduleByID(ctx context.Context, id uuid.UUID) (*ExamScheduleResponse, error)
	ListSchedulesByExam(ctx context.Context, examID uuid.UUID) ([]*ExamScheduleResponse, error)
	ListSchedulesByClassSection(ctx context.Context, classSectionID uuid.UUID) ([]*ExamScheduleResponse, error)
	UpdateSchedule(ctx context.Context, id uuid.UUID, req UpdateScheduleRequest) (*ExamScheduleResponse, error)
	DeleteSchedule(ctx context.Context, id uuid.UUID) error

	CreateResult(ctx context.Context, scheduleID uuid.UUID, req CreateResultRequest) (*ExamResultResponse, error)
	BulkCreateResults(ctx context.Context, scheduleID uuid.UUID, req BulkCreateResultRequest) ([]*ExamResultResponse, error)
	GetResultByID(ctx context.Context, id uuid.UUID) (*ExamResultResponse, error)
	ListResultsBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]*ExamResultResponse, error)
	ListResultsByStudent(ctx context.Context, studentEnrollmentID uuid.UUID) ([]*ExamResultResponse, error)
	UpdateResult(ctx context.Context, id uuid.UUID, req UpdateResultRequest) (*ExamResultResponse, error)
	DeleteResult(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateExam(ctx context.Context, req CreateExamRequest) (*ExamResponse, error) {
	if err := ValidateExamWindow(req.StartDate, req.EndDate); err != nil {
		return nil, NewAppError(CodeValidation, err.Error())
	}
	trimmedName := strings.TrimSpace(req.Name)
	isDup, err := s.repo.IsDuplicateExamName(ctx, req.AcademicYearID, trimmedName)
	if err != nil {
		return nil, err
	}
	if isDup {
		return nil, NewAppError(CodeDuplicate, "exam already exists for this academic year")
	}

	e := &Exam{
		AcademicYearID: req.AcademicYearID,
		Name:           trimmedName,
		Description:    normalizeStringPtr(req.Description),
		ExamType:       req.ExamType,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		IsPublished:    false,
	}
	if err := s.repo.CreateExam(ctx, e); err != nil {
		return nil, err
	}
	return ToExamResponse(e), nil
}

func (s *service) GetExamByID(ctx context.Context, id uuid.UUID) (*ExamResponse, error) {
	e, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "exam")
	}
	return ToExamResponse(e), nil
}

func (s *service) ListExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*ExamResponse, int64, error) {
	exams, total, err := s.repo.FindAllExams(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*ExamResponse, len(exams))
	for i, exam := range exams {
		responses[i] = ToExamResponse(exam)
	}
	return responses, total, nil
}

func (s *service) UpdateExam(ctx context.Context, id uuid.UUID, req UpdateExamRequest) (*ExamResponse, error) {
	exam, err := s.repo.GetExamByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if exam.IsPublished {
		if req.Name != nil && strings.TrimSpace(*req.Name) != exam.Name {
			return nil, NewAppError(CodeBusinessRule, "cannot rename a published exam")
		}
		if req.ExamType != nil && *req.ExamType != exam.ExamType {
			return nil, NewAppError(CodeBusinessRule, "cannot change exam_type for a published exam")
		}
		if req.StartDate != nil && !req.StartDate.Equal(exam.StartDate) {
			return nil, NewAppError(CodeBusinessRule, "cannot change start_date for a published exam")
		}
		if req.EndDate != nil && !req.EndDate.Equal(exam.EndDate) {
			return nil, NewAppError(CodeBusinessRule, "cannot change end_date for a published exam")
		}
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != exam.Name {
		isDup, err := s.repo.IsDuplicateExamName(ctx, exam.AcademicYearID, strings.TrimSpace(*req.Name))
		if err != nil {
			return nil, err
		}
		if isDup {
			return nil, NewAppError(CodeDuplicate, "exam name already exists for the academic year")
		}
	}

	if req.Name != nil {
		exam.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		exam.Description = normalizeStringPtr(req.Description)
	}
	if req.ExamType != nil {
		exam.ExamType = *req.ExamType
	}
	if req.StartDate != nil {
		exam.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		exam.EndDate = *req.EndDate
	}

	if err := ValidateExamWindow(exam.StartDate, exam.EndDate); err != nil {
		return nil, NewAppError(CodeValidation, err.Error())
	}

	if err := s.repo.UpdateExam(ctx, exam); err != nil {
		return nil, err
	}
	return ToExamResponse(exam), nil
}

func (s *service) PublishExam(ctx context.Context, id uuid.UUID, req PublishExamRequest) (*ExamResponse, error) {
	if req.IsPublished == nil {
		return nil, NewAppError(CodeValidation, "is_published is required")
	}

	var exam *Exam
	if err := s.repo.WithTx(ctx, func(tx Repository) error {
		var err error
		exam, err = tx.GetExamByIDForUpdate(ctx, id)
		if err != nil {
			return mapRepoError(err, "exam")
		}
		if *req.IsPublished {
			hasSchedules, err := tx.HasSchedules(ctx, id)
			if err != nil {
				return err
			}
			if !hasSchedules {
				return NewAppError(CodeValidation, "cannot publish exam with no schedules")
			}
		}
		exam.IsPublished = *req.IsPublished
		return tx.UpdateExam(ctx, exam)
	}); err != nil {
		return nil, err
	}

	return ToExamResponse(exam), nil
}

func (s *service) DeleteExam(ctx context.Context, id uuid.UUID) error {
	return s.repo.WithTx(ctx, func(tx Repository) error {
		exam, err := tx.GetExamByIDForUpdate(ctx, id)
		if err != nil {
			return mapRepoError(err, "exam")
		}
		if exam.IsPublished {
			return NewAppError(CodeBusinessRule, "cannot delete a published exam — unpublish it first")
		}
		return tx.DeleteExam(ctx, id)
	})
}

func (s *service) CreateSchedule(ctx context.Context, examID uuid.UUID, req CreateScheduleRequest) (*ExamScheduleResponse, error) {
	exam, err := s.repo.GetExamByID(ctx, examID)
	if err != nil {
		return nil, err
	}

	if err := ValidateScheduleWindow(req.ExamDate, exam.StartDate, exam.EndDate); err != nil {
		return nil, NewAppError(CodeValidation, err.Error())
	}
	if err := ValidateScheduleTimeBounds(req.StartTime, req.EndTime); err != nil {
		return nil, NewAppError(CodeValidation, err.Error())
	}
	if err := ValidateScheduleMarks(req.MaxMarks, req.PassingMarks); err != nil {
		return nil, NewAppError(CodeValidation, err.Error())
	}

	schedule := &ExamSchedule{
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
	if isDup, err := s.repo.IsDuplicateSchedule(ctx, examID, req.ClassSectionID, req.SubjectID); err != nil {
		return nil, err
	} else if isDup {
		return nil, NewAppError(CodeDuplicate, "a schedule for this class section and subject already exists in this exam")
	}
	if err := s.repo.CreateSchedule(ctx, schedule); err != nil {
		return nil, err
	}
	return ToScheduleResponse(schedule), nil
}

func (s *service) GetScheduleByID(ctx context.Context, id uuid.UUID) (*ExamScheduleResponse, error) {
	schedule, err := s.repo.GetScheduleByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "exam schedule")
	}
	return ToScheduleResponse(schedule), nil
}

func (s *service) ListSchedulesByExam(ctx context.Context, examID uuid.UUID) ([]*ExamScheduleResponse, error) {
	schedules, err := s.repo.FindSchedulesByExam(ctx, examID)
	if err != nil {
		return nil, err
	}
	responses := make([]*ExamScheduleResponse, len(schedules))
	for i, sch := range schedules {
		responses[i] = ToScheduleResponse(sch)
	}
	return responses, nil
}

func (s *service) ListSchedulesByClassSection(ctx context.Context, classSectionID uuid.UUID) ([]*ExamScheduleResponse, error) {
	schedules, err := s.repo.FindSchedulesByClassSection(ctx, classSectionID)
	if err != nil {
		return nil, err
	}
	responses := make([]*ExamScheduleResponse, len(schedules))
	for i, sch := range schedules {
		responses[i] = ToScheduleResponse(sch)
	}
	return responses, nil
}

func (s *service) UpdateSchedule(ctx context.Context, id uuid.UUID, req UpdateScheduleRequest) (*ExamScheduleResponse, error) {
	var schedule *ExamSchedule
	if err := s.repo.WithTx(ctx, func(tx Repository) error {
		var err error
		schedule, err = tx.GetScheduleByIDForUpdate(ctx, id)
		if err != nil {
			return mapRepoError(err, "exam schedule")
		}
		hasResults, err := tx.HasResults(ctx, id)
		if err != nil {
			return err
		}
		if hasResults {
			return NewAppError(CodeBusinessRule, "cannot update schedule that already has results — delete results first")
		}
		if req.ExamDate != nil {
			schedule.ExamDate = *req.ExamDate
		}
		if req.StartTime != nil {
			schedule.StartTime = req.StartTime
		}
		if req.EndTime != nil {
			schedule.EndTime = req.EndTime
		}
		if req.MaxMarks != nil {
			schedule.MaxMarks = *req.MaxMarks
		}
		if req.PassingMarks != nil {
			schedule.PassingMarks = *req.PassingMarks
		}
		if req.RoomID != nil {
			schedule.RoomID = req.RoomID
		}

		if err := ValidateScheduleTimeBounds(schedule.StartTime, schedule.EndTime); err != nil {
			return NewAppError(CodeValidation, err.Error())
		}
		if err := ValidateScheduleMarks(schedule.MaxMarks, schedule.PassingMarks); err != nil {
			return NewAppError(CodeValidation, err.Error())
		}
		return tx.UpdateSchedule(ctx, schedule)
	}); err != nil {
		return nil, err
	}
	return ToScheduleResponse(schedule), nil
}

func (s *service) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	return s.repo.WithTx(ctx, func(tx Repository) error {
		schedule, err := tx.GetScheduleByIDForUpdate(ctx, id)
		if err != nil {
			return mapRepoError(err, "exam schedule")
		}
		exam, err := tx.GetExamByID(ctx, schedule.ExamID)
		if err != nil {
			return err
		}
		if exam.IsPublished {
			return NewAppError(CodeBusinessRule, "cannot delete a schedule from a published exam")
		}
		hasResults, err := tx.HasResults(ctx, id)
		if err != nil {
			return err
		}
		if hasResults {
			return NewAppError(CodeBusinessRule, "cannot delete schedule with existing results")
		}
		return tx.DeleteSchedule(ctx, id)
	})
}

func (s *service) CreateResult(ctx context.Context, scheduleID uuid.UUID, req CreateResultRequest) (*ExamResultResponse, error) {
	maxMarks, passingMarks, err := s.repo.GetScheduleMarks(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	if err := ValidateResultMarks(req.MarksObtained, maxMarks); err != nil {
		return nil, NewAppError(CodeValidation, err.Error())
	}
	isDup, err := s.repo.IsDuplicateResult(ctx, scheduleID, req.StudentEnrollmentID)
	if err != nil {
		return nil, err
	}
	if isDup {
		return nil, NewAppError(CodeDuplicate, "result already exists for this student in this schedule")
	}

	res := &ExamResult{
		ExamScheduleID:      scheduleID,
		StudentEnrollmentID: req.StudentEnrollmentID,
		MarksObtained:       req.MarksObtained,
		Grade:               normalizeStringPtr(req.Grade),
		Status:              DeriveExamResultStatus(req.MarksObtained, passingMarks),
		Remarks:             normalizeStringPtr(req.Remarks),
	}
	if err := s.repo.CreateResult(ctx, res); err != nil {
		return nil, err
	}
	return ToResultResponse(res), nil
}

func (s *service) BulkCreateResults(ctx context.Context, scheduleID uuid.UUID, req BulkCreateResultRequest) ([]*ExamResultResponse, error) {
	if len(req.Results) == 0 {
		return nil, NewAppError(CodeValidation, "results must not be empty")
	}
	if len(req.Results) > BulkResultMaxItems {
		return nil, NewAppError(CodeValidation, fmt.Sprintf("results cannot contain more than %d entries", BulkResultMaxItems))
	}

	seen := make(map[uuid.UUID]int)
	var duplicateIndexes []int
	studentIDs := make([]uuid.UUID, 0, len(req.Results))
	for idx, item := range req.Results {
		if _, ok := seen[item.StudentEnrollmentID]; ok {
			duplicateIndexes = append(duplicateIndexes, idx)
		}
		seen[item.StudentEnrollmentID] = idx
		studentIDs = append(studentIDs, item.StudentEnrollmentID)
	}
	if len(duplicateIndexes) > 0 {
		return nil, NewAppError(CodeValidation, "duplicate student_enrollment_id entries in batch", fmt.Sprintf("duplicate indices: %v", duplicateIndexes))
	}

	maxMarks, passingMarks, err := s.repo.GetScheduleMarks(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindDuplicateResultEnrollmentIDs(ctx, scheduleID, studentIDs)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return nil, NewAppError(CodeDuplicate, "results already exist for one or more students", fmt.Sprintf("duplicate student_enrollment_ids: %v", existing))
	}

	results := make([]*ExamResult, 0, len(req.Results))
	for _, item := range req.Results {
		if err := ValidateResultMarks(item.MarksObtained, maxMarks); err != nil {
			return nil, NewAppError(CodeValidation, err.Error())
		}
		results = append(results, &ExamResult{
			ExamScheduleID:      scheduleID,
			StudentEnrollmentID: item.StudentEnrollmentID,
			MarksObtained:       item.MarksObtained,
			Grade:               normalizeStringPtr(item.Grade),
			Status:              DeriveExamResultStatus(item.MarksObtained, passingMarks),
			Remarks:             normalizeStringPtr(item.Remarks),
		})
	}

	if err := s.repo.BulkCreateResults(ctx, results); err != nil {
		return nil, err
	}

	responses := make([]*ExamResultResponse, len(results))
	for i, result := range results {
		responses[i] = ToResultResponse(result)
	}
	return responses, nil
}

func (s *service) GetResultByID(ctx context.Context, id uuid.UUID) (*ExamResultResponse, error) {
	result, err := s.repo.GetResultByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "exam result")
	}
	return ToResultResponse(result), nil
}

func (s *service) ListResultsBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]*ExamResultResponse, error) {
	results, err := s.repo.FindResultsBySchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	responses := make([]*ExamResultResponse, len(results))
	for i, res := range results {
		responses[i] = ToResultResponse(res)
	}
	return responses, nil
}

func (s *service) ListResultsByStudent(ctx context.Context, studentEnrollmentID uuid.UUID) ([]*ExamResultResponse, error) {
	results, err := s.repo.FindResultsByStudent(ctx, studentEnrollmentID)
	if err != nil {
		return nil, err
	}
	responses := make([]*ExamResultResponse, len(results))
	for i, res := range results {
		responses[i] = ToResultResponse(res)
	}
	return responses, nil
}

func (s *service) UpdateResult(ctx context.Context, id uuid.UUID, req UpdateResultRequest) (*ExamResultResponse, error) {
	result, err := s.repo.GetResultByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "exam result")
	}

	if req.MarksObtained != nil {
		maxMarks, passingMarks, err := s.repo.GetScheduleMarks(ctx, result.ExamScheduleID)
		if err != nil {
			return nil, err
		}
		if err := ValidateResultMarks(req.MarksObtained, maxMarks); err != nil {
			return nil, NewAppError(CodeValidation, err.Error())
		}
		result.MarksObtained = req.MarksObtained
		result.Status = DeriveExamResultStatus(req.MarksObtained, passingMarks)
	}
	if req.Grade != nil {
		result.Grade = normalizeStringPtr(req.Grade)
	}
	if req.Remarks != nil {
		result.Remarks = normalizeStringPtr(req.Remarks)
	}

	if err := s.repo.UpdateResult(ctx, result); err != nil {
		return nil, err
	}
	return ToResultResponse(result), nil
}

func (s *service) DeleteResult(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetResultByID(ctx, id); err != nil {
		return mapRepoError(err, "exam result")
	}
	return s.repo.DeleteResult(ctx, id)
}
