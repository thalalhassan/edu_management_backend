package exam

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
)

const (
	ExamNameMaxLength        = 255
	ExamDescriptionMaxLength = 1000
	GradeMaxLength           = 10
	RemarksMaxLength         = 500
	BulkResultMaxItems       = 500
)

func DeriveExamResultStatus(marksObtained *decimal.Decimal, passingMarks decimal.Decimal) database.ExamResultStatus {
	if marksObtained == nil {
		return database.ExamResultStatusAbsent
	}
	if marksObtained.Cmp(passingMarks) >= 0 {
		return database.ExamResultStatusPass
	}
	return database.ExamResultStatusFail
}

func ValidateExamWindow(startDate, endDate time.Time) error {
	if !endDate.After(startDate) {
		return fmt.Errorf("end_date must be after start_date")
	}
	return nil
}

func ValidateScheduleWindow(examDate, startDate, endDate time.Time) error {
	if examDate.Before(startDate) || examDate.After(endDate) {
		return fmt.Errorf("exam_date must be between exam start_date and end_date")
	}
	return nil
}

func ValidateScheduleTimeBounds(startTime, endTime *time.Time) error {
	if startTime != nil && endTime != nil && !endTime.After(*startTime) {
		return fmt.Errorf("end_time must be after start_time")
	}
	return nil
}

func ValidateScheduleMarks(maxMarks, passingMarks decimal.Decimal) error {
	if maxMarks.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("max_marks must be greater than zero")
	}
	if passingMarks.LessThanOrEqual(decimal.Zero) || passingMarks.GreaterThan(maxMarks) {
		return fmt.Errorf("passing_marks must be greater than zero and less than or equal to max_marks")
	}
	return nil
}

func ValidateResultMarks(marksObtained *decimal.Decimal, maxMarks decimal.Decimal) error {
	if marksObtained == nil {
		return nil
	}
	if marksObtained.LessThan(decimal.Zero) {
		return fmt.Errorf("marks_obtained must be greater than or equal to zero")
	}
	if marksObtained.GreaterThan(maxMarks) {
		return fmt.Errorf("marks_obtained cannot exceed max_marks")
	}
	return nil
}

func normalizeString(value string) string {
	return strings.TrimSpace(value)
}

func normalizeStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}
