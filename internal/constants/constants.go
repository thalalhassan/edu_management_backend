package constants

const ApiBase = "/api"
const ApiVersion1 = "/v1"

const ApiV1 = ApiBase + ApiVersion1

const DateLayout = "2006-01-02"

// const DateLayout = "02-01-2006" // DD-MM-YYYY

const (
	UserIDContextKey         = "userID"
	EmailContextKey          = "email"
	RoleContextKey           = "role"
	AcademicYearIDContextKey = "academicYearID"
)

const (
	ApiUserPath = "/users"
	ApiAuthPath = "/auth"

	ApiStudentPath = "/students"
	ApiTeacherPath = "/teachers"
	ApiParentPath  = "/parents"
	ApiStaffPath   = "/staff" // or /employees

	ApiDepartmentPath   = "/departments"
	ApiClassSectionPath = "/class-sections"
	ApiStandardPath     = "/standards"

	ApiEnrollmentPath        = "/enrollments"
	ApiTeacherAssignmentPath = "/teacher-assignments"
	ApiAcademicYearPath      = "/academic-years"

	ApiFeePath   = "/fees"   // or split into /payments, /fee-structures
	ApiLeavePath = "/leaves" // or /leave-requests

	ApiNoticePath       = "/notices"
	ApiAnnouncementPath = "/announcements"
	ApiReportPath       = "/reports"

	ApiAttendancePath         = "/attendance"
	ApiEmployeeAttendancePath = "/employee-attendance"

	ApiExamPath         = "/exams"
	ApiScheduleExamPath = "/exam-schedule"
	ApiResultExamPath   = "/exam-result"

	ApiTimetablePath = "/timetables"

	ApiSubjectPath = "/subjects"

	ApiSalariesPath = "/salaries"
)
