package exam

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Repository interface ────────────────────────────────────────────────────

type Repository interface {
	// Exam CRUD
	CreateExam(ctx context.Context, e *Exam) error
	GetExamByID(ctx context.Context, id string) (*Exam, error)
	FindAllExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*Exam, int64, error)
	UpdateExam(ctx context.Context, e *Exam) error
	DeleteExam(ctx context.Context, id string) error

	// Exam helpers
	IsDuplicateExamName(ctx context.Context, academicYearID, name string) (bool, error)
	HasSchedules(ctx context.Context, examID string) (bool, error)

	// ExamSchedule CRUD
	CreateSchedule(ctx context.Context, s *ExamSchedule) error
	GetScheduleByID(ctx context.Context, id string) (*ExamSchedule, error)
	FindSchedulesByExam(ctx context.Context, examID string) ([]*ExamSchedule, error)
	FindSchedulesByClassSection(ctx context.Context, classSectionID string) ([]*ExamSchedule, error)
	UpdateSchedule(ctx context.Context, s *ExamSchedule) error
	DeleteSchedule(ctx context.Context, id string) error

	// ExamSchedule helpers
	IsDuplicateSchedule(ctx context.Context, examID, classSectionID, subjectID string) (bool, error)
	HasResults(ctx context.Context, scheduleID string) (bool, error)

	// ExamResult CRUD
	CreateResult(ctx context.Context, r *ExamResult) error
	BulkCreateResults(ctx context.Context, results []*ExamResult) error
	GetResultByID(ctx context.Context, id string) (*ExamResult, error)
	FindResultsBySchedule(ctx context.Context, scheduleID string) ([]*ExamResult, error)
	FindResultsByStudent(ctx context.Context, studentEnrollmentID string) ([]*ExamResult, error)
	UpdateResult(ctx context.Context, r *ExamResult) error
	DeleteResult(ctx context.Context, id string) error

	// ExamResult helpers
	IsDuplicateResult(ctx context.Context, scheduleID, enrollmentID string) (bool, error)
	GetScheduleMarks(ctx context.Context, scheduleID string) (maxMarks decimal.Decimal, passingMarks decimal.Decimal, err error)
}

// ─── repositoryImpl ─────────────────────────────────────────────────────────

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// ─── Exam ────────────────────────────────────────────────────────────────────

func (r *repositoryImpl) CreateExam(ctx context.Context, e *Exam) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *repositoryImpl) GetExamByID(ctx context.Context, id string) (*Exam, error) {
	var e Exam
	if err := r.db.WithContext(ctx).
		Model(&database.Exam{}).
		First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repositoryImpl) FindAllExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*Exam, int64, error) {
	var exams []*Exam
	query := r.db.WithContext(ctx).Model(&database.Exam{})

	f := q.Filter
	if f.AcademicYearID != nil {
		query = query.Where("academic_year_id = ?", *f.AcademicYearID)
	}
	if f.ExamType != nil {
		query = query.Where("exam_type = ?", *f.ExamType)
	}
	if f.IsPublished != nil {
		query = query.Where("is_published = ?", *f.IsPublished)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Find(&exams).Error
	return exams, total, err
}

func (r *repositoryImpl) UpdateExam(ctx context.Context, e *Exam) error {
	return r.db.WithContext(ctx).Where("id = ?", e.ID).Save(e).Error
}

func (r *repositoryImpl) DeleteExam(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.Exam{}).Error
}

func (r *repositoryImpl) IsDuplicateExamName(ctx context.Context, academicYearID, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.Exam{}).
		Where("academic_year_id = ? AND name = ?", academicYearID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasSchedules(ctx context.Context, examID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Where("exam_id = ?", examID).
		Count(&count).Error
	return count > 0, err
}

// ─── ExamSchedule ────────────────────────────────────────────────────────────

func (r *repositoryImpl) CreateSchedule(ctx context.Context, s *ExamSchedule) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repositoryImpl) GetScheduleByID(ctx context.Context, id string) (*ExamSchedule, error) {
	var s ExamSchedule
	if err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repositoryImpl) FindSchedulesByExam(ctx context.Context, examID string) ([]*ExamSchedule, error) {
	var schedules []*ExamSchedule
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Where("exam_id = ?", examID).
		Order("exam_date ASC, start_time ASC").
		Find(&schedules).Error
	return schedules, err
}

func (r *repositoryImpl) FindSchedulesByClassSection(ctx context.Context, classSectionID string) ([]*ExamSchedule, error) {
	var schedules []*ExamSchedule
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Where("class_section_id = ?", classSectionID).
		Order("exam_date ASC").
		Find(&schedules).Error
	return schedules, err
}

func (r *repositoryImpl) UpdateSchedule(ctx context.Context, s *ExamSchedule) error {
	return r.db.WithContext(ctx).Where("id = ?", s.ID).Save(s).Error
}

func (r *repositoryImpl) DeleteSchedule(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.ExamSchedule{}).Error
}

func (r *repositoryImpl) IsDuplicateSchedule(ctx context.Context, examID, classSectionID, subjectID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Where("exam_id = ? AND class_section_id = ? AND subject_id = ?", examID, classSectionID, subjectID).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasResults(ctx context.Context, scheduleID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Where("exam_schedule_id = ?", scheduleID).
		Count(&count).Error
	return count > 0, err
}

// ─── ExamResult ──────────────────────────────────────────────────────────────

func (r *repositoryImpl) CreateResult(ctx context.Context, res *ExamResult) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *repositoryImpl) BulkCreateResults(ctx context.Context, results []*ExamResult) error {
	return r.db.WithContext(ctx).CreateInBatches(results, 100).Error
}

func (r *repositoryImpl) GetResultByID(ctx context.Context, id string) (*ExamResult, error) {
	var res ExamResult
	if err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		First(&res, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *repositoryImpl) FindResultsBySchedule(ctx context.Context, scheduleID string) ([]*ExamResult, error) {
	var results []*ExamResult
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Where("exam_schedule_id = ?", scheduleID).
		Find(&results).Error
	return results, err
}

func (r *repositoryImpl) FindResultsByStudent(ctx context.Context, studentEnrollmentID string) ([]*ExamResult, error) {
	var results []*ExamResult
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Where("student_enrollment_id = ?", studentEnrollmentID).
		Find(&results).Error
	return results, err
}

func (r *repositoryImpl) UpdateResult(ctx context.Context, res *ExamResult) error {
	return r.db.WithContext(ctx).Where("id = ?", res.ID).Save(res).Error
}

func (r *repositoryImpl) DeleteResult(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.ExamResult{}).Error
}

func (r *repositoryImpl) IsDuplicateResult(ctx context.Context, scheduleID, enrollmentID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Where("exam_schedule_id = ? AND student_enrollment_id = ?", scheduleID, enrollmentID).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) GetScheduleMarks(ctx context.Context, scheduleID string) (decimal.Decimal, decimal.Decimal, error) {
	var s ExamSchedule
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Select("max_marks, passing_marks").
		First(&s, "id = ?", scheduleID).Error
	return s.MaxMarks, s.PassingMarks, err
}
