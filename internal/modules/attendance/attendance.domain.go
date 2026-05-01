package attendance

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

// Domain aliases
// Attendance and EmployeeAttendance are persisted models.
type Attendance = database.Attendance

type EmployeeAttendance = database.EmployeeAttendance

type AttendanceStatus = database.AttendanceStatus

// NormalizeDate trims time-of-day to UTC midnight so date comparisons stay stable.
func NormalizeDate(date time.Time) time.Time {
	if date.IsZero() {
		return date
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

// Attendance response shapes are defined in dto layer.
