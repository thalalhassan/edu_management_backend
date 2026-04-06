package report

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	// Report card
	GetEnrollmentWithStudent(ctx context.Context, enrollmentID string) (*database.StudentEnrollment, error)
	GetExamResultsForEnrollment(ctx context.Context, enrollmentID, examID string) ([]database.ExamResult, error)

	// Student attendance
	GetStudentAttendanceCounts(ctx context.Context, enrollmentID string, from, to *time.Time) (map[database.AttendanceStatus]int, error)

	// Class attendance
	GetClassRoster(ctx context.Context, classSectionID string) ([]database.StudentEnrollment, error)
	GetClassAttendanceCounts(ctx context.Context, classSectionID string, from, to *time.Time) ([]attendanceCount, error)

	// Class performance
	GetExamResultsForClassSection(ctx context.Context, classSectionID, examID string) ([]database.ExamResult, error)
	GetExamWithSchedules(ctx context.Context, examID, classSectionID string) (*database.Exam, error)

	// Fee collection
	GetFeeCollectionAggregates(ctx context.Context, academicYearID string, standardID, classSectionID *string) ([]feeAggregate, error)

	// Teacher attendance
	GetTeacherAttendanceCounts(ctx context.Context, teacherID *string, from, to *time.Time) ([]teacherAttCount, error)
}

// Internal scan targets — not exported, used only within the repository.
type attendanceCount struct {
	EnrollmentID string
	Status       database.AttendanceStatus
	Count        int
}

type feeAggregate struct {
	ClassSection string
	FeeComponent string
	Status       database.FeeStatus
	Count        int
	TotalDue     decimal.Decimal
	TotalPaid    decimal.Decimal
}

type teacherAttCount struct {
	TeacherID  string
	EmployeeID string
	FirstName  string
	LastName   string
	Status     database.AttendanceStatus
	Count      int
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) GetEnrollmentWithStudent(ctx context.Context, enrollmentID string) (*database.StudentEnrollment, error) {
	var e database.StudentEnrollment
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("ClassSection.Standard.Department").
		Preload("ClassSection.AcademicYear").
		First(&e, "id = ?", enrollmentID).Error
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetExamResultsForEnrollment returns results for a student.
// If examID is empty, returns all results across every exam in the AY.
func (r *repositoryImpl) GetExamResultsForEnrollment(ctx context.Context, enrollmentID, examID string) ([]database.ExamResult, error) {
	var results []database.ExamResult
	query := r.db.WithContext(ctx).
		Preload("ExamSchedule.Exam").
		Preload("ExamSchedule.Subject").
		Where("student_enrollment_id = ?", enrollmentID)

	if examID != "" {
		query = query.Joins("JOIN exam_schedule ON exam_schedule.id = exam_results.exam_schedule_id").
			Where("exam_schedule.exam_id = ?", examID)
	}

	err := query.Find(&results).Error
	return results, err
}

func (r *repositoryImpl) GetStudentAttendanceCounts(ctx context.Context, enrollmentID string, from, to *time.Time) (map[database.AttendanceStatus]int, error) {
	type row struct {
		Status database.AttendanceStatus
		Count  int
	}
	var rows []row

	query := r.db.WithContext(ctx).
		Model(&database.Attendance{}).
		Select("status, COUNT(*) as count").
		Where("student_enrollment_id = ?", enrollmentID).
		Group("status")

	if from != nil {
		query = query.Where("date >= ?", from)
	}
	if to != nil {
		query = query.Where("date <= ?", to)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[database.AttendanceStatus]int)
	for _, row := range rows {
		counts[row.Status] = row.Count
	}
	return counts, nil
}

func (r *repositoryImpl) GetClassRoster(ctx context.Context, classSectionID string) ([]database.StudentEnrollment, error) {
	var enrollments []database.StudentEnrollment
	err := r.db.WithContext(ctx).
		Preload("Student").
		Where("class_section_id = ? AND status = ?", classSectionID, database.EnrollmentStatusEnrolled).
		Order("roll_number ASC").
		Find(&enrollments).Error
	return enrollments, err
}

func (r *repositoryImpl) GetClassAttendanceCounts(ctx context.Context, classSectionID string, from, to *time.Time) ([]attendanceCount, error) {
	var rows []attendanceCount

	query := r.db.WithContext(ctx).
		Model(&database.Attendance{}).
		Select("attendance.student_enrollment_id as enrollment_id, attendance.status, COUNT(*) as count").
		Joins("JOIN student_enrollment ON student_enrollment.id = attendance.student_enrollment_id").
		Where("student_enrollment.class_section_id = ?", classSectionID).
		Group("attendance.student_enrollment_id, attendance.status")

	if from != nil {
		query = query.Where("attendance.date >= ?", from)
	}
	if to != nil {
		query = query.Where("attendance.date <= ?", to)
	}

	err := query.Scan(&rows).Error
	return rows, err
}

func (r *repositoryImpl) GetExamResultsForClassSection(ctx context.Context, classSectionID, examID string) ([]database.ExamResult, error) {
	var results []database.ExamResult
	err := r.db.WithContext(ctx).
		Preload("ExamSchedule.Subject").
		Preload("ExamSchedule.Exam").
		Preload("StudentEnrollment.Student").
		Joins("JOIN exam_schedule ON exam_schedule.id = exam_results.exam_schedule_id").
		Joins("JOIN student_enrollment ON student_enrollment.id = exam_results.student_enrollment_id").
		Where("exam_schedule.exam_id = ? AND exam_schedule.class_section_id = ?", examID, classSectionID).
		Find(&results).Error
	return results, err
}

func (r *repositoryImpl) GetExamWithSchedules(ctx context.Context, examID, classSectionID string) (*database.Exam, error) {
	var exam database.Exam
	err := r.db.WithContext(ctx).
		Preload("Schedules", "class_section_id = ?", classSectionID).
		Preload("Schedules.Subject").
		First(&exam, "id = ?", examID).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *repositoryImpl) GetFeeCollectionAggregates(ctx context.Context, academicYearID string, standardID, classSectionID *string) ([]feeAggregate, error) {
	var rows []feeAggregate

	query := r.db.WithContext(ctx).
		Model(&database.FeeRecord{}).
		Select(`
			CONCAT(standard.name, ' - ', class_section.section_name) as class_section,
			fee_record.fee_component,
			fee_record.status,
			COUNT(*) as count,
			COALESCE(SUM(fee_record.amount_due), 0)  as total_due,
			COALESCE(SUM(fee_record.amount_paid), 0) as total_paid
		`).
		Joins("JOIN student_enrollment ON student_enrollment.id = fee_record.student_enrollment_id").
		Joins("JOIN class_section ON class_section.id = student_enrollment.class_section_id").
		Joins("JOIN standard ON standard.id = class_section.standard_id").
		Where("class_section.academic_year_id = ?", academicYearID).
		Group("class_section, fee_record.fee_component, fee_record.status").
		Order("class_section ASC, fee_record.fee_component ASC")

	if standardID != nil {
		query = query.Where("class_section.standard_id = ?", *standardID)
	}
	if classSectionID != nil {
		query = query.Where("student_enrollment.class_section_id = ?", *classSectionID)
	}

	err := query.Scan(&rows).Error
	return rows, err
}

func (r *repositoryImpl) GetTeacherAttendanceCounts(ctx context.Context, teacherID *string, from, to *time.Time) ([]teacherAttCount, error) {
	var rows []teacherAttCount

	query := r.db.WithContext(ctx).
		Model(&database.TeacherAttendance{}).
		Select(`
			teacher_attendance.teacher_id,
			teacher.employee_id,
			teacher.first_name,
			teacher.last_name,
			teacher_attendance.status,
			COUNT(*) as count
		`).
		Joins("JOIN teacher ON teacher.id = teacher_attendance.teacher_id").
		Group("teacher_attendance.teacher_id, teacher.employee_id, teacher.first_name, teacher.last_name, teacher_attendance.status")

	if teacherID != nil {
		query = query.Where("teacher_attendance.teacher_id = ?", *teacherID)
	}
	if from != nil {
		query = query.Where("teacher_attendance.date >= ?", from)
	}
	if to != nil {
		query = query.Where("teacher_attendance.date <= ?", to)
	}

	err := query.Scan(&rows).Error
	return rows, err
}
