package exam

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Exam = database.Exam
type ExamSchedule = database.ExamSchedule
type ExamResult = database.ExamResult

// ─── Repository interface ────────────────────────────────────────────────────

type Repository interface {
	// Exam CRUD
	CreateExam(ctx context.Context, e *Exam) error
	GetExamByID(ctx context.Context, id uuid.UUID) (*Exam, error)
	FindAllExams(ctx context.Context, q query_params.Query[FilterParams]) ([]*Exam, int64, error)
	UpdateExam(ctx context.Context, e *Exam) error
	DeleteExam(ctx context.Context, id uuid.UUID) error

	// Exam helpers
	IsDuplicateExamName(ctx context.Context, academicYearID uuid.UUID, name string) (bool, error)
	HasSchedules(ctx context.Context, examID uuid.UUID) (bool, error)
	WithTx(ctx context.Context, fn func(tx Repository) error) error
	GetExamByIDForUpdate(ctx context.Context, id uuid.UUID) (*Exam, error)

	// ExamSchedule CRUD
	CreateSchedule(ctx context.Context, s *ExamSchedule) error
	GetScheduleByID(ctx context.Context, id uuid.UUID) (*ExamSchedule, error)
	FindSchedulesByExam(ctx context.Context, examID uuid.UUID) ([]*ExamSchedule, error)
	FindSchedulesByClassSection(ctx context.Context, classSectionID uuid.UUID) ([]*ExamSchedule, error)
	UpdateSchedule(ctx context.Context, s *ExamSchedule) error
	DeleteSchedule(ctx context.Context, id uuid.UUID) error

	// ExamSchedule helpers
	IsDuplicateSchedule(ctx context.Context, examID, classSectionID, subjectID uuid.UUID) (bool, error)
	HasResults(ctx context.Context, scheduleID uuid.UUID) (bool, error)
	GetScheduleByIDForUpdate(ctx context.Context, id uuid.UUID) (*ExamSchedule, error)

	// ExamResult CRUD
	CreateResult(ctx context.Context, r *ExamResult) error
	BulkCreateResults(ctx context.Context, results []*ExamResult) error
	GetResultByID(ctx context.Context, id uuid.UUID) (*ExamResult, error)
	FindResultsBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]*ExamResult, error)
	FindResultsByStudent(ctx context.Context, studentEnrollmentID uuid.UUID) ([]*ExamResult, error)
	UpdateResult(ctx context.Context, r *ExamResult) error
	DeleteResult(ctx context.Context, id uuid.UUID) error

	// ExamResult helpers
	IsDuplicateResult(ctx context.Context, scheduleID, enrollmentID uuid.UUID) (bool, error)
	FindDuplicateResultEnrollmentIDs(ctx context.Context, scheduleID uuid.UUID, enrollmentIDs []uuid.UUID) ([]uuid.UUID, error)
	GetScheduleMarks(ctx context.Context, scheduleID uuid.UUID) (maxMarks decimal.Decimal, passingMarks decimal.Decimal, err error)
}

// ─── repositoryImpl ─────────────────────────────────────────────────────────

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) WithTx(ctx context.Context, fn func(tx Repository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repositoryImpl{db: tx})
	})
}

func (r *repositoryImpl) GetExamByIDForUpdate(ctx context.Context, id uuid.UUID) (*Exam, error) {
	var e Exam
	if err := r.db.WithContext(ctx).
		Model(&database.Exam{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repositoryImpl) GetScheduleByIDForUpdate(ctx context.Context, id uuid.UUID) (*ExamSchedule, error) {
	var s ExamSchedule
	if err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repositoryImpl) FindDuplicateResultEnrollmentIDs(ctx context.Context, scheduleID uuid.UUID, enrollmentIDs []uuid.UUID) ([]uuid.UUID, error) {
	var ids []string
	if len(enrollmentIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Select("student_enrollment_id").
		Where("exam_schedule_id = ? AND student_enrollment_id IN ?", scheduleID, enrollmentIDs).
		Pluck("student_enrollment_id", &ids).Error

	uuids := make([]uuid.UUID, 0, len(ids))
	for _, idStr := range ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, id)
	}
	return uuids, err
}

// ─── Exam ────────────────────────────────────────────────────────────────────

func (r *repositoryImpl) CreateExam(ctx context.Context, e *Exam) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *repositoryImpl) GetExamByID(ctx context.Context, id uuid.UUID) (*Exam, error) {
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

func (r *repositoryImpl) DeleteExam(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.Exam{}).Error
}

func (r *repositoryImpl) IsDuplicateExamName(ctx context.Context, academicYearID uuid.UUID, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.Exam{}).
		Where("academic_year_id = ? AND name = ?", academicYearID, name).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasSchedules(ctx context.Context, examID uuid.UUID) (bool, error) {
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

func (r *repositoryImpl) GetScheduleByID(ctx context.Context, id uuid.UUID) (*ExamSchedule, error) {
	var s ExamSchedule
	if err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repositoryImpl) FindSchedulesByExam(ctx context.Context, examID uuid.UUID) ([]*ExamSchedule, error) {
	var schedules []*ExamSchedule
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Where("exam_id = ?", examID).
		Order("exam_date ASC, start_time ASC").
		Find(&schedules).Error
	return schedules, err
}

func (r *repositoryImpl) FindSchedulesByClassSection(ctx context.Context, classSectionID uuid.UUID) ([]*ExamSchedule, error) {
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

func (r *repositoryImpl) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.ExamSchedule{}).Error
}

func (r *repositoryImpl) IsDuplicateSchedule(ctx context.Context, examID, classSectionID, subjectID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Where("exam_id = ? AND class_section_id = ? AND subject_id = ?", examID, classSectionID, subjectID).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) HasResults(ctx context.Context, scheduleID uuid.UUID) (bool, error) {
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

func (r *repositoryImpl) GetResultByID(ctx context.Context, id uuid.UUID) (*ExamResult, error) {
	var res ExamResult
	if err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		First(&res, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *repositoryImpl) FindResultsBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]*ExamResult, error) {
	var results []*ExamResult
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Where("exam_schedule_id = ?", scheduleID).
		Find(&results).Error
	return results, err
}

func (r *repositoryImpl) FindResultsByStudent(ctx context.Context, studentEnrollmentID uuid.UUID) ([]*ExamResult, error) {
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

func (r *repositoryImpl) DeleteResult(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.ExamResult{}).Error
}

func (r *repositoryImpl) IsDuplicateResult(ctx context.Context, scheduleID, enrollmentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&database.ExamResult{}).
		Where("exam_schedule_id = ? AND student_enrollment_id = ?", scheduleID, enrollmentID).
		Count(&count).Error
	return count > 0, err
}

func (r *repositoryImpl) GetScheduleMarks(ctx context.Context, scheduleID uuid.UUID) (decimal.Decimal, decimal.Decimal, error) {
	var s ExamSchedule
	err := r.db.WithContext(ctx).
		Model(&database.ExamSchedule{}).
		Select("max_marks, passing_marks").
		First(&s, "id = ?", scheduleID).Error
	return s.MaxMarks, s.PassingMarks, err
}
