package attendance

import (
	"context"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Repository interface ─────────────────────────────────────────────────────

type Repository interface {
	// Student Attendance
	CreateAttendance(ctx context.Context, a *Attendance) error
	BulkCreateAttendance(ctx context.Context, records []*Attendance) error
	GetAttendanceByID(ctx context.Context, id string) (*Attendance, error)
	FindStudentAttendance(ctx context.Context, q query_params.Query[StudentFilterParams]) ([]*Attendance, int64, error)
	FindByClassSectionAndDate(ctx context.Context, classSectionID string, date time.Time) ([]*Attendance, error)
	FindByEnrollmentAndDate(ctx context.Context, enrollmentID string, date time.Time) (*Attendance, error)
	UpdateAttendance(ctx context.Context, a *Attendance) error
	DeleteAttendance(ctx context.Context, id string) error

	// Student Attendance helpers
	GetEnrollmentIDsByClassSection(ctx context.Context, classSectionID string) ([]string, error)
	CountByClassSectionAndDate(ctx context.Context, classSectionID string, date time.Time) (present, absent, halfDay, late, leave int64, err error)

	// Teacher Attendance
	CreateTeacherAttendance(ctx context.Context, a *TeacherAttendance) error
	BulkCreateTeacherAttendance(ctx context.Context, records []*TeacherAttendance) error
	GetTeacherAttendanceByID(ctx context.Context, id string) (*TeacherAttendance, error)
	FindTeacherAttendance(ctx context.Context, q query_params.Query[TeacherFilterParams]) ([]*TeacherAttendance, int64, error)
	FindTeacherAttendanceByDate(ctx context.Context, teacherID string, date time.Time) (*TeacherAttendance, error)
	UpdateTeacherAttendance(ctx context.Context, a *TeacherAttendance) error
	DeleteTeacherAttendance(ctx context.Context, id string) error
}

// ─── repositoryImpl ───────────────────────────────────────────────────────────

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

// ─── Student Attendance ───────────────────────────────────────────────────────

func (r *repositoryImpl) CreateAttendance(ctx context.Context, a *Attendance) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repositoryImpl) BulkCreateAttendance(ctx context.Context, records []*Attendance) error {
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

func (r *repositoryImpl) GetAttendanceByID(ctx context.Context, id string) (*Attendance, error) {
	var a Attendance
	if err := r.db.WithContext(ctx).
		Model(&database.Attendance{}).
		First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) FindStudentAttendance(ctx context.Context, q query_params.Query[StudentFilterParams]) ([]*Attendance, int64, error) {
	var records []*Attendance
	query := r.db.WithContext(ctx).Model(&database.Attendance{})

	f := q.Filter
	if f.StudentEnrollmentID != nil {
		query = query.Where("student_enrollment_id = ?", *f.StudentEnrollmentID)
	}
	if f.ClassSectionID != nil {
		// Join through student_enrollment to filter by class section
		query = query.Joins("JOIN student_enrollment ON student_enrollment.id = attendance.student_enrollment_id").
			Where("student_enrollment.class_section_id = ?", *f.ClassSectionID)
	}
	if f.DateFrom != nil {
		query = query.Where("attendance.date >= ?", f.DateFrom.Format("2006-01-02"))
	}
	if f.DateTo != nil {
		query = query.Where("attendance.date <= ?", f.DateTo.Format("2006-01-02"))
	}
	if f.Status != nil {
		query = query.Where("attendance.status = ?", *f.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Find(&records).Error
	return records, total, err
}

func (r *repositoryImpl) FindByClassSectionAndDate(ctx context.Context, classSectionID string, date time.Time) ([]*Attendance, error) {
	var records []*Attendance
	err := r.db.WithContext(ctx).
		Model(&database.Attendance{}).
		Joins("JOIN student_enrollment ON student_enrollment.id = attendance.student_enrollment_id").
		Where("student_enrollment.class_section_id = ? AND attendance.date = ?", classSectionID, date.Format("2006-01-02")).
		Find(&records).Error
	return records, err
}

func (r *repositoryImpl) FindByEnrollmentAndDate(ctx context.Context, enrollmentID string, date time.Time) (*Attendance, error) {
	var a Attendance
	err := r.db.WithContext(ctx).
		Model(&database.Attendance{}).
		Where("student_enrollment_id = ? AND date = ?", enrollmentID, date.Format("2006-01-02")).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) UpdateAttendance(ctx context.Context, a *Attendance) error {
	return r.db.WithContext(ctx).Where("id = ?", a.ID).Save(a).Error
}

func (r *repositoryImpl) DeleteAttendance(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.Attendance{}).Error
}

func (r *repositoryImpl) GetEnrollmentIDsByClassSection(ctx context.Context, classSectionID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Model(&database.StudentEnrollment{}).
		Select("id").
		Where("class_section_id = ? AND status = 'ENROLLED'", classSectionID).
		Find(&ids).Error
	return ids, err
}

func (r *repositoryImpl) CountByClassSectionAndDate(ctx context.Context, classSectionID string, date time.Time) (present, absent, halfDay, late, leave int64, err error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	err = r.db.WithContext(ctx).
		Model(&database.Attendance{}).
		Select("attendance.status AS status, COUNT(*) AS count").
		Joins("JOIN student_enrollment ON student_enrollment.id = attendance.student_enrollment_id").
		Where("student_enrollment.class_section_id = ? AND attendance.date = ?", classSectionID, date.Format("2006-01-02")).
		Group("attendance.status").
		Scan(&rows).Error
	if err != nil {
		return
	}
	for _, rw := range rows {
		switch rw.Status {
		case "PRESENT":
			present = rw.Count
		case "ABSENT":
			absent = rw.Count
		case "HALF_DAY":
			halfDay = rw.Count
		case "LATE":
			late = rw.Count
		case "LEAVE":
			leave = rw.Count
		}
	}
	return
}

// ─── Teacher Attendance ───────────────────────────────────────────────────────

func (r *repositoryImpl) CreateTeacherAttendance(ctx context.Context, a *TeacherAttendance) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repositoryImpl) BulkCreateTeacherAttendance(ctx context.Context, records []*TeacherAttendance) error {
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

func (r *repositoryImpl) GetTeacherAttendanceByID(ctx context.Context, id string) (*TeacherAttendance, error) {
	var a TeacherAttendance
	if err := r.db.WithContext(ctx).
		Model(&database.TeacherAttendance{}).
		First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) FindTeacherAttendance(ctx context.Context, q query_params.Query[TeacherFilterParams]) ([]*TeacherAttendance, int64, error) {
	var records []*TeacherAttendance
	query := r.db.WithContext(ctx).Model(&database.TeacherAttendance{})

	f := q.Filter
	if f.TeacherID != nil {
		query = query.Where("teacher_id = ?", *f.TeacherID)
	}
	if f.DateFrom != nil {
		query = query.Where("date >= ?", f.DateFrom.Format("2006-01-02"))
	}
	if f.DateTo != nil {
		query = query.Where("date <= ?", f.DateTo.Format("2006-01-02"))
	}
	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Find(&records).Error
	return records, total, err
}

func (r *repositoryImpl) FindTeacherAttendanceByDate(ctx context.Context, teacherID string, date time.Time) (*TeacherAttendance, error) {
	var a TeacherAttendance
	err := r.db.WithContext(ctx).
		Model(&database.TeacherAttendance{}).
		Where("teacher_id = ? AND date = ?", teacherID, date.Format("2006-01-02")).
		First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) UpdateTeacherAttendance(ctx context.Context, a *TeacherAttendance) error {
	return r.db.WithContext(ctx).Where("id = ?", a.ID).Save(a).Error
}

func (r *repositoryImpl) DeleteTeacherAttendance(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.TeacherAttendance{}).Error
}
