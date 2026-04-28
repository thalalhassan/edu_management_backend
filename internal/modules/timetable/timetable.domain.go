package timetable

import (
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
)

type TimeTable = database.TimeTable

// DayOfWeek constants — 0=Sunday, 1=Monday … 6=Saturday
const (
	Sunday    = 0
	Monday    = 1
	Tuesday   = 2
	Wednesday = 3
	Thursday  = 4
	Friday    = 5
	Saturday  = 6
)

var DayNames = map[int]string{
	0: "Sunday",
	1: "Monday",
	2: "Tuesday",
	3: "Wednesday",
	4: "Thursday",
	5: "Friday",
	6: "Saturday",
}

// AllowedSortFields whitelists columns safe to ORDER BY.
var AllowedSortFields = map[string]bool{
	"day_of_week": true,
	"start_time":  true,
	"created_at":  true,
}

// FilterParams binds from query string.
type FilterParams struct {
	ClassSectionID *string `form:"class_section_id"`
	TeacherID      *string `form:"teacher_id"`
	SubjectID      *string `form:"subject_id"`
	DayOfWeek      *int    `form:"day_of_week"` // 0–6
}

type CreateRequest struct {
	ClassSectionID string    `json:"class_section_id" binding:"required,uuid"`
	SubjectID      string    `json:"subject_id"       binding:"required,uuid"`
	EmployeeID     string    `json:"employee_id"      binding:"required,uuid"`
	DayOfWeek      int       `json:"day_of_week"      binding:"required,min=0,max=6"`
	StartTime      time.Time `json:"start_time"       binding:"required"`
	EndTime        time.Time `json:"end_time"         binding:"required"`
	RoomID         *string   `json:"room_id,omitempty"`
}

type UpdateRequest struct {
	SubjectID  *string    `json:"subject_id,omitempty"  binding:"omitempty,uuid"`
	EmployeeID *string    `json:"employee_id,omitempty" binding:"omitempty,uuid"`
	DayOfWeek  *int       `json:"day_of_week,omitempty" binding:"omitempty,min=0,max=6"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	RoomID     *string    `json:"room_id,omitempty"`
}

// TimeTableResponse is the full response with nested relations.
type TimeTableResponse struct {
	TimeTable
	DayName string `json:"day_name"` // human-readable day
}

// PeriodEntry is a lightweight row used in the class/teacher schedule view.
type PeriodEntry struct {
	ID           string    `json:"id"`
	DayOfWeek    int       `json:"day_of_week"`
	DayName      string    `json:"day_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	SubjectCode  string    `json:"subject_code"`
	SubjectName  string    `json:"subject_name"`
	TeacherName  string    `json:"teacher_name"`
	ClassSection string    `json:"class_section"` // "Grade 6 - A"
	RoomNumber   *string   `json:"room_number,omitempty"`
}

// DaySchedule groups periods by day — used in the weekly schedule view.
type DaySchedule struct {
	DayOfWeek int           `json:"day_of_week"`
	DayName   string        `json:"day_name"`
	Periods   []PeriodEntry `json:"periods"`
}

func ToTimeTableResponse(t *TimeTable) *TimeTableResponse {
	return &TimeTableResponse{
		TimeTable: *t,
		DayName:   DayNames[t.DayOfWeek],
	}
}

func ToPeriodEntry(t *TimeTable) PeriodEntry {
	entry := PeriodEntry{
		ID:        t.ID,
		DayOfWeek: t.DayOfWeek,
		DayName:   DayNames[t.DayOfWeek],
		StartTime: t.StartTime,
		EndTime:   t.EndTime,
	}
	if t.Room != nil {
		entry.RoomNumber = &t.Room.RoomNumber
	}
	entry.SubjectCode = t.Subject.Code
	entry.SubjectName = t.Subject.Name
	entry.TeacherName = t.Employee.FirstName + " " + t.Employee.LastName
	entry.ClassSection = t.ClassSection.Standard.Name + " - " + t.ClassSection.SectionName
	return entry
}

// GroupByDay takes a flat list and organises it into a weekly schedule.
func GroupByDay(entries []PeriodEntry) []DaySchedule {
	dayMap := make(map[int][]PeriodEntry)
	for _, e := range entries {
		dayMap[e.DayOfWeek] = append(dayMap[e.DayOfWeek], e)
	}
	// Emit in Mon–Sat order, skip days with no periods
	var schedule []DaySchedule
	for day := Monday; day <= Saturday; day++ {
		periods, ok := dayMap[day]
		if !ok {
			continue
		}
		schedule = append(schedule, DaySchedule{
			DayOfWeek: day,
			DayName:   DayNames[day],
			Periods:   periods,
		})
	}
	return schedule
}
