package timetable

import (
	"context"
	"errors"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/shared/query_params"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*TimeTableResponse, error)
	GetByID(ctx context.Context, id string) (*TimeTableResponse, error)
	List(ctx context.Context, q query_params.Query[FilterParams]) ([]*TimeTableResponse, int64, error)
	GetClassSchedule(ctx context.Context, classSectionID string) ([]DaySchedule, error)
	GetTeacherSchedule(ctx context.Context, teacherID string) ([]DaySchedule, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*TimeTableResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*TimeTableResponse, error) {
	if !req.EndTime.After(req.StartTime) {
		return nil, errors.New("timetable.Service.Create: end_time must be after start_time")
	}

	// Guard: class section slot conflict
	conflict, err := s.repo.HasConflict(ctx, req.ClassSectionID, req.DayOfWeek, req.StartTime, req.EndTime, "")
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Create.HasConflict: %w", err)
	}
	if conflict {
		return nil, fmt.Errorf("timetable.Service.Create: class section already has a period overlapping %s–%s on %s",
			req.StartTime.Format("15:04"), req.EndTime.Format("15:04"), DayNames[req.DayOfWeek])
	}

	// Guard: teacher time conflict across all their classes
	teacherConflict, err := s.repo.HasTeacherConflict(ctx, req.TeacherID, req.DayOfWeek, req.StartTime, req.EndTime, "")
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Create.HasTeacherConflict: %w", err)
	}
	if teacherConflict {
		return nil, fmt.Errorf("timetable.Service.Create: teacher is already assigned to another class at this time on %s", DayNames[req.DayOfWeek])
	}

	t := &TimeTable{
		ClassSectionID: req.ClassSectionID,
		SubjectID:      req.SubjectID,
		TeacherID:      req.TeacherID,
		DayOfWeek:      req.DayOfWeek,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		RoomNumber:     req.RoomNumber,
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("timetable.Service.Create: %w", err)
	}

	created, err := s.repo.GetByID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Create.GetByID: %w", err)
	}
	return ToTimeTableResponse(created), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*TimeTableResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.GetByID: %w", err)
	}
	return ToTimeTableResponse(t), nil
}

func (s *service) List(ctx context.Context, q query_params.Query[FilterParams]) ([]*TimeTableResponse, int64, error) {
	entries, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("timetable.Service.List: %w", err)
	}
	responses := make([]*TimeTableResponse, len(entries))
	for i, t := range entries {
		responses[i] = ToTimeTableResponse(t)
	}
	return responses, total, nil
}

// GetClassSchedule returns the weekly timetable for a class section
// grouped by day — the primary view for students and class teachers.
func (s *service) GetClassSchedule(ctx context.Context, classSectionID string) ([]DaySchedule, error) {
	entries, err := s.repo.FindByClassSection(ctx, classSectionID)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.GetClassSchedule: %w", err)
	}
	periods := make([]PeriodEntry, len(entries))
	for i, t := range entries {
		periods[i] = ToPeriodEntry(t)
	}
	return GroupByDay(periods), nil
}

// GetTeacherSchedule returns the weekly schedule for a teacher
// across all their class sections — grouped by day.
func (s *service) GetTeacherSchedule(ctx context.Context, teacherID string) ([]DaySchedule, error) {
	entries, err := s.repo.FindByTeacher(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.GetTeacherSchedule: %w", err)
	}
	periods := make([]PeriodEntry, len(entries))
	for i, t := range entries {
		periods[i] = ToPeriodEntry(t)
	}
	return GroupByDay(periods), nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateRequest) (*TimeTableResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Update.GetByID: %w", err)
	}

	// Apply field updates
	if req.SubjectID != nil {
		t.SubjectID = *req.SubjectID
	}
	if req.TeacherID != nil {
		t.TeacherID = *req.TeacherID
	}
	if req.DayOfWeek != nil {
		t.DayOfWeek = *req.DayOfWeek
	}
	if req.StartTime != nil {
		t.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		t.EndTime = *req.EndTime
	}
	if req.RoomNumber != nil {
		t.RoomNumber = req.RoomNumber
	}

	if !t.EndTime.After(t.StartTime) {
		return nil, errors.New("timetable.Service.Update: end_time must be after start_time")
	}

	// Re-check conflicts with this entry excluded
	conflict, err := s.repo.HasConflict(ctx, t.ClassSectionID, t.DayOfWeek, t.StartTime, t.EndTime, id)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Update.HasConflict: %w", err)
	}
	if conflict {
		return nil, fmt.Errorf("timetable.Service.Update: class section already has an overlapping period on %s", DayNames[t.DayOfWeek])
	}

	teacherConflict, err := s.repo.HasTeacherConflict(ctx, t.TeacherID, t.DayOfWeek, t.StartTime, t.EndTime, id)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Update.HasTeacherConflict: %w", err)
	}
	if teacherConflict {
		return nil, fmt.Errorf("timetable.Service.Update: teacher already has a class at this time on %s", DayNames[t.DayOfWeek])
	}

	if err := s.repo.Update(ctx, id, t); err != nil {
		return nil, fmt.Errorf("timetable.Service.Update.Save: %w", err)
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("timetable.Service.Update.Reload: %w", err)
	}
	return ToTimeTableResponse(updated), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("timetable.Service.Delete.GetByID: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("timetable.Service.Delete: %w", err)
	}
	return nil
}
