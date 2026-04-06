package report

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
)

type Service interface {
	GetReportCard(ctx context.Context, req ReportCardRequest) (*ReportCard, error)
	GetStudentAttendanceSummary(ctx context.Context, req StudentAttendanceRequest) (*StudentAttendanceSummary, error)
	GetClassAttendanceSummary(ctx context.Context, req ClassAttendanceRequest) (*ClassAttendanceSummary, error)
	GetClassPerformanceReport(ctx context.Context, req ClassPerformanceRequest) (*ClassPerformanceReport, error)
	GetFeeCollectionReport(ctx context.Context, req FeeCollectionRequest) (*FeeCollectionReport, error)
	GetTeacherAttendanceSummary(ctx context.Context, req TeacherAttendanceRequest) ([]TeacherAttendanceSummary, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ──────────────────────────────────────────────────────────────
// REPORT CARD
// ──────────────────────────────────────────────────────────────

func (s *service) GetReportCard(ctx context.Context, req ReportCardRequest) (*ReportCard, error) {
	enrollment, err := s.repo.GetEnrollmentWithStudent(ctx, req.StudentEnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetReportCard.GetEnrollment: %w", err)
	}

	results, err := s.repo.GetExamResultsForEnrollment(ctx, req.StudentEnrollmentID, req.ExamID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetReportCard.GetResults: %w", err)
	}

	cs := enrollment.ClassSection
	card := &ReportCard{
		StudentEnrollmentID: enrollment.ID,
		AdmissionNo:         enrollment.Student.AdmissionNo,
		StudentName:         enrollment.Student.FirstName + " " + enrollment.Student.LastName,
		ClassSection:        cs.Standard.Name + " - " + cs.SectionName,
		AcademicYear:        cs.AcademicYear.Name,
	}

	// Group results by exam
	examMap := make(map[string]*ExamReportSection)
	examOrder := []string{}

	for _, r := range results {
		eid := r.ExamSchedule.ExamID
		if _, ok := examMap[eid]; !ok {
			examMap[eid] = &ExamReportSection{
				ExamID:   eid,
				ExamName: r.ExamSchedule.Exam.Name,
				ExamType: r.ExamSchedule.Exam.ExamType,
			}
			examOrder = append(examOrder, eid)
		}

		pct := safeDiv(r.MarksObtained, r.ExamSchedule.MaxMarks)
		sub := SubjectResult{
			SubjectCode:   r.ExamSchedule.Subject.Code,
			SubjectName:   r.ExamSchedule.Subject.Name,
			MaxMarks:      r.ExamSchedule.MaxMarks,
			PassingMarks:  r.ExamSchedule.PassingMarks,
			MarksObtained: r.MarksObtained,
			Percentage:    pct,
			Grade:         gradeFromPercentage(pct),
			Status:        string(r.Status),
		}
		if r.Grade != nil {
			sub.Grade = *r.Grade
		}
		examMap[eid].Subjects = append(examMap[eid].Subjects, sub)
	}

	// Compute per-exam totals and overall totals
	var overallMax, overallObtained decimal.Decimal
	for _, eid := range examOrder {
		section := examMap[eid]
		var maxSum, obtainedSum decimal.Decimal
		for _, sub := range section.Subjects {
			maxSum = maxSum.Add(sub.MaxMarks)
			obtainedSum = obtainedSum.Add(sub.MarksObtained)
		}
		pct := safeDiv(obtainedSum, maxSum)
		section.Total = ExamTotals{
			MaxMarks:      maxSum,
			MarksObtained: obtainedSum,
			Percentage:    pct,
			Grade:         gradeFromPercentage(pct),
		}
		overallMax = overallMax.Add(maxSum)
		overallObtained = overallObtained.Add(obtainedSum)
		card.Exams = append(card.Exams, *section)
	}

	card.OverallPercentage = safeDiv(overallObtained, overallMax)
	card.OverallGrade = gradeFromPercentage(card.OverallPercentage)
	return card, nil
}

// ──────────────────────────────────────────────────────────────
// STUDENT ATTENDANCE SUMMARY
// ──────────────────────────────────────────────────────────────

func (s *service) GetStudentAttendanceSummary(ctx context.Context, req StudentAttendanceRequest) (*StudentAttendanceSummary, error) {
	enrollment, err := s.repo.GetEnrollmentWithStudent(ctx, req.StudentEnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetStudentAttendanceSummary.GetEnrollment: %w", err)
	}

	counts, err := s.repo.GetStudentAttendanceCounts(ctx, req.StudentEnrollmentID, req.FromDate, req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetStudentAttendanceSummary.GetCounts: %w", err)
	}

	present := counts[database.AttendanceStatusPresent]
	absent := counts[database.AttendanceStatusAbsent]
	halfDay := counts[database.AttendanceStatusHalfDay]
	late := counts[database.AttendanceStatusLate]
	leave := counts[database.AttendanceStatusLeave]
	total := present + absent + halfDay + late + leave

	// Count half days as 0.5 for attendance percent
	effectivePresent := decimal.NewFromInt(int64(present)).
		Add(decimal.NewFromFloat(float64(halfDay) * 0.5)).
		Add(decimal.NewFromInt(int64(late)))
	pct := safeDiv(effectivePresent, decimal.NewFromInt(int64(total)))

	cs := enrollment.ClassSection
	return &StudentAttendanceSummary{
		StudentEnrollmentID: enrollment.ID,
		StudentName:         enrollment.Student.FirstName + " " + enrollment.Student.LastName,
		AdmissionNo:         enrollment.Student.AdmissionNo,
		ClassSection:        cs.Standard.Name + " - " + cs.SectionName,
		FromDate:            req.FromDate,
		ToDate:              req.ToDate,
		TotalDays:           total,
		Present:             present,
		Absent:              absent,
		HalfDay:             halfDay,
		Late:                late,
		Leave:               leave,
		AttendancePercent:   pct,
	}, nil
}

// ──────────────────────────────────────────────────────────────
// CLASS ATTENDANCE SUMMARY
// ──────────────────────────────────────────────────────────────

func (s *service) GetClassAttendanceSummary(ctx context.Context, req ClassAttendanceRequest) (*ClassAttendanceSummary, error) {
	roster, err := s.repo.GetClassRoster(ctx, req.ClassSectionID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetClassAttendanceSummary.GetRoster: %w", err)
	}

	from, to := resolveDateRange(req.Month, req.Year, req.Date)
	rawCounts, err := s.repo.GetClassAttendanceCounts(ctx, req.ClassSectionID, from, to)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetClassAttendanceSummary.GetCounts: %w", err)
	}

	// Build per-enrollment count maps
	countMap := make(map[string]map[database.AttendanceStatus]int)
	for _, row := range rawCounts {
		if _, ok := countMap[row.EnrollmentID]; !ok {
			countMap[row.EnrollmentID] = make(map[database.AttendanceStatus]int)
		}
		countMap[row.EnrollmentID][row.Status] = row.Count
	}

	summary := &ClassAttendanceSummary{
		ClassSectionID: req.ClassSectionID,
		TotalStudents:  len(roster),
		Date:           req.Date,
		Month:          req.Month,
		Year:           req.Year,
	}
	if len(roster) > 0 {
		summary.ClassSection = roster[0].ClassSection.Standard.Name + " - " + roster[0].ClassSection.SectionName
	}

	for _, e := range roster {
		c := countMap[e.ID]
		present := c[database.AttendanceStatusPresent]
		absent := c[database.AttendanceStatusAbsent]
		halfDay := c[database.AttendanceStatusHalfDay]
		late := c[database.AttendanceStatusLate]
		total := present + absent + halfDay + late + c[database.AttendanceStatusLeave]

		effective := decimal.NewFromInt(int64(present)).
			Add(decimal.NewFromFloat(float64(halfDay) * 0.5)).
			Add(decimal.NewFromInt(int64(late)))
		pct := safeDiv(effective, decimal.NewFromInt(int64(total)))

		summary.Students = append(summary.Students, StudentAttendanceRow{
			EnrollmentID:      e.ID,
			RollNumber:        e.RollNumber,
			StudentName:       e.Student.FirstName + " " + e.Student.LastName,
			AdmissionNo:       e.Student.AdmissionNo,
			TotalDays:         total,
			Present:           present,
			Absent:            absent,
			AttendancePercent: pct,
		})
	}
	return summary, nil
}

// ──────────────────────────────────────────────────────────────
// CLASS PERFORMANCE
// ──────────────────────────────────────────────────────────────

func (s *service) GetClassPerformanceReport(ctx context.Context, req ClassPerformanceRequest) (*ClassPerformanceReport, error) {
	exam, err := s.repo.GetExamWithSchedules(ctx, req.ExamID, req.ClassSectionID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetClassPerformanceReport.GetExam: %w", err)
	}

	results, err := s.repo.GetExamResultsForClassSection(ctx, req.ClassSectionID, req.ExamID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetClassPerformanceReport.GetResults: %w", err)
	}

	// Build subject performance map
	type subjectAccumulator struct {
		schedule    database.ExamSchedule
		marks       []decimal.Decimal
		passCount   int
		failCount   int
		absentCount int
	}
	subjectMap := make(map[string]*subjectAccumulator)
	for _, sch := range exam.Schedules {
		subjectMap[sch.SubjectID] = &subjectAccumulator{schedule: sch}
	}

	// Per-student totals for ranking
	type studentTotal struct {
		name        string
		admissionNo string
		total       decimal.Decimal
		maxTotal    decimal.Decimal
	}
	studentTotals := make(map[string]*studentTotal)

	for _, r := range results {
		sid := r.ExamSchedule.SubjectID
		acc, ok := subjectMap[sid]
		if !ok {
			continue
		}

		switch r.Status {
		case database.ExamResultStatusAbsent:
			acc.absentCount++
		case database.ExamResultStatusPass, database.ExamResultStatusGrace:
			acc.passCount++
			acc.marks = append(acc.marks, r.MarksObtained)
		case database.ExamResultStatusFail:
			acc.failCount++
			acc.marks = append(acc.marks, r.MarksObtained)
		}

		// Accumulate student totals
		eid := r.StudentEnrollmentID
		if _, ok := studentTotals[eid]; !ok {
			st := r.StudentEnrollment.Student
			studentTotals[eid] = &studentTotal{
				name:        st.FirstName + " " + st.LastName,
				admissionNo: st.AdmissionNo,
			}
		}
		studentTotals[eid].total = studentTotals[eid].total.Add(r.MarksObtained)
		studentTotals[eid].maxTotal = studentTotals[eid].maxTotal.Add(r.ExamSchedule.MaxMarks)
	}

	report := &ClassPerformanceReport{
		ClassSectionID: req.ClassSectionID,
		ExamName:       exam.Name,
		TotalStudents:  len(studentTotals),
	}

	// Build subject performance rows
	for _, acc := range subjectMap {
		var sum decimal.Decimal
		for _, m := range acc.marks {
			sum = sum.Add(m)
		}
		n := len(acc.marks)
		var avg decimal.Decimal
		if n > 0 {
			avg = sum.Div(decimal.NewFromInt(int64(n))).Round(2)
		}

		var highest, lowest decimal.Decimal
		if n > 0 {
			highest = acc.marks[0]
			lowest = acc.marks[0]
			for _, m := range acc.marks[1:] {
				if m.GreaterThan(highest) {
					highest = m
				}
				if m.LessThan(lowest) {
					lowest = m
				}
			}
		}

		total := acc.passCount + acc.failCount + acc.absentCount
		passPct := safeDiv(decimal.NewFromInt(int64(acc.passCount)), decimal.NewFromInt(int64(total)))

		report.Subjects = append(report.Subjects, SubjectPerformance{
			SubjectCode:    acc.schedule.Subject.Code,
			SubjectName:    acc.schedule.Subject.Name,
			MaxMarks:       acc.schedule.MaxMarks,
			ClassAverage:   avg,
			HighestMarks:   highest,
			LowestMarks:    lowest,
			PassCount:      acc.passCount,
			FailCount:      acc.failCount,
			AbsentCount:    acc.absentCount,
			PassPercentage: passPct,
		})
	}

	// Build top 5 student ranking
	type rankedStudent struct {
		name        string
		admissionNo string
		pct         decimal.Decimal
		total       decimal.Decimal
		maxTotal    decimal.Decimal
	}
	var ranked []rankedStudent
	for _, st := range studentTotals {
		pct := safeDiv(st.total, st.maxTotal)
		ranked = append(ranked, rankedStudent{
			name:        st.name,
			admissionNo: st.admissionNo,
			pct:         pct,
			total:       st.total,
			maxTotal:    st.maxTotal,
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].pct.GreaterThan(ranked[j].pct)
	})
	limit := 5
	if len(ranked) < limit {
		limit = len(ranked)
	}
	for i, st := range ranked[:limit] {
		report.TopStudents = append(report.TopStudents, StudentRank{
			Rank:        i + 1,
			StudentName: st.name,
			AdmissionNo: st.admissionNo,
			TotalMarks:  st.total,
			MaxMarks:    st.maxTotal,
			Percentage:  st.pct,
			Grade:       gradeFromPercentage(st.pct),
		})
	}

	return report, nil
}

// ──────────────────────────────────────────────────────────────
// FEE COLLECTION REPORT
// ──────────────────────────────────────────────────────────────

func (s *service) GetFeeCollectionReport(ctx context.Context, req FeeCollectionRequest) (*FeeCollectionReport, error) {
	rows, err := s.repo.GetFeeCollectionAggregates(ctx, req.AcademicYearID, req.StandardID, req.ClassSectionID)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetFeeCollectionReport: %w", err)
	}

	// Group by class_section + fee_component, fan out by status
	type key struct{ classSection, component string }
	type accumulator struct {
		paid, pending, overdue, waived, partial int
		totalDue, totalPaid                     decimal.Decimal
	}
	accMap := make(map[key]*accumulator)
	keyOrder := []key{}

	for _, row := range rows {
		k := key{row.ClassSection, row.FeeComponent}
		if _, ok := accMap[k]; !ok {
			accMap[k] = &accumulator{}
			keyOrder = append(keyOrder, k)
		}
		a := accMap[k]
		a.totalDue = a.totalDue.Add(row.TotalDue)
		a.totalPaid = a.totalPaid.Add(row.TotalPaid)
		switch row.Status {
		case database.FeeStatusPaid:
			a.paid += row.Count
		case database.FeeStatusPending:
			a.pending += row.Count
		case database.FeeStatusOverdue:
			a.overdue += row.Count
		case database.FeeStatusWaived:
			a.waived += row.Count
		case database.FeeStatusPartial:
			a.partial += row.Count
		}
	}

	report := &FeeCollectionReport{}
	for _, k := range keyOrder {
		a := accMap[k]
		balance := a.totalDue.Sub(a.totalPaid)
		total := a.paid + a.pending + a.overdue + a.waived + a.partial

		report.Rows = append(report.Rows, FeeCollectionRow{
			ClassSection:   k.classSection,
			FeeComponent:   k.component,
			TotalStudents:  total,
			PaidCount:      a.paid,
			PendingCount:   a.pending,
			OverdueCount:   a.overdue,
			WaivedCount:    a.waived,
			TotalDue:       a.totalDue,
			TotalCollected: a.totalPaid,
			TotalBalance:   balance,
		})

		report.TotalDue = report.TotalDue.Add(a.totalDue)
		report.TotalCollected = report.TotalCollected.Add(a.totalPaid)
	}
	report.TotalBalance = report.TotalDue.Sub(report.TotalCollected)
	return report, nil
}

// ──────────────────────────────────────────────────────────────
// TEACHER ATTENDANCE
// ──────────────────────────────────────────────────────────────

func (s *service) GetTeacherAttendanceSummary(ctx context.Context, req TeacherAttendanceRequest) ([]TeacherAttendanceSummary, error) {
	from, to := resolveDateRange(req.Month, req.Year, nil)
	if req.FromDate != nil {
		from = req.FromDate
	}
	if req.ToDate != nil {
		to = req.ToDate
	}

	rows, err := s.repo.GetTeacherAttendanceCounts(ctx, req.TeacherID, from, to)
	if err != nil {
		return nil, fmt.Errorf("report.Service.GetTeacherAttendanceSummary: %w", err)
	}

	// Group by teacher
	type acc struct {
		employeeID string
		name       string
		counts     map[database.AttendanceStatus]int
	}
	accMap := make(map[string]*acc)
	order := []string{}

	for _, row := range rows {
		if _, ok := accMap[row.TeacherID]; !ok {
			accMap[row.TeacherID] = &acc{
				employeeID: row.EmployeeID,
				name:       row.FirstName + " " + row.LastName,
				counts:     make(map[database.AttendanceStatus]int),
			}
			order = append(order, row.TeacherID)
		}
		accMap[row.TeacherID].counts[row.Status] = row.Count
	}

	summaries := make([]TeacherAttendanceSummary, 0, len(order))
	for _, tid := range order {
		a := accMap[tid]
		present := a.counts[database.AttendanceStatusPresent]
		absent := a.counts[database.AttendanceStatusAbsent]
		halfDay := a.counts[database.AttendanceStatusHalfDay]
		late := a.counts[database.AttendanceStatusLate]
		leave := a.counts[database.AttendanceStatusLeave]
		total := present + absent + halfDay + late + leave

		effective := decimal.NewFromInt(int64(present)).
			Add(decimal.NewFromFloat(float64(halfDay) * 0.5)).
			Add(decimal.NewFromInt(int64(late)))
		pct := safeDiv(effective, decimal.NewFromInt(int64(total)))

		summaries = append(summaries, TeacherAttendanceSummary{
			TeacherID:         tid,
			EmployeeID:        a.employeeID,
			TeacherName:       a.name,
			TotalDays:         total,
			Present:           present,
			Absent:            absent,
			HalfDay:           halfDay,
			Late:              late,
			Leave:             leave,
			AttendancePercent: pct,
		})
	}
	return summaries, nil
}

// ──────────────────────────────────────────────────────────────
// HELPERS
// ──────────────────────────────────────────────────────────────

// resolveDateRange converts month+year or a single date into a from/to pair.
func resolveDateRange(month, year *int, date *time.Time) (from, to *time.Time) {
	if date != nil {
		d := *date
		return &d, &d
	}
	if month != nil && year != nil {
		start := time.Date(*year, time.Month(*month), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		return &start, &end
	}
	return nil, nil
}
