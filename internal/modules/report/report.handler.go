package report

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type Handler struct {
	service Service
	config  *config.Config
}

func NewHandler(s Service, config *config.Config) *Handler {
	return &Handler{service: s, config: config}
}

func RegisterRouter(r *gin.RouterGroup, a *app.App) {
	svc := NewService(NewRepository(a.DB.Gorm))
	h := NewHandler(svc, a.Config)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	rpt := r.Group("/report")
	rpt.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		// Student report card
		rpt.GET("/report-card", h.reportCard)

		// Attendance reports
		rpt.GET("/attendance/student", h.studentAttendance)
		rpt.GET("/attendance/class", h.classAttendance)
		rpt.GET("/attendance/teacher", h.teacherAttendance)

		// Academic performance
		rpt.GET("/performance/class", h.classPerformance)

		// Fee collection
		rpt.GET("/fee-collection", h.feeCollection)
	}
}

// reportCard godoc
// GET /report/report-card?student_enrollment_id=xxx&exam_id=xxx
// exam_id is optional — omit to get results across all exams in the AY
func (h *Handler) reportCard(c *gin.Context) {
	var req ReportCardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.GetReportCard(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Report card retrieved successfully")
}

// studentAttendance godoc
// GET /report/attendance/student?student_enrollment_id=xxx&from_date=xxx&to_date=xxx
func (h *Handler) studentAttendance(c *gin.Context) {
	var req StudentAttendanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.GetStudentAttendanceSummary(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Student attendance summary retrieved successfully")
}

// classAttendance godoc
// GET /report/attendance/class?class_section_id=xxx&month=7&year=2024
// OR ?class_section_id=xxx&date=2024-07-15
func (h *Handler) classAttendance(c *gin.Context) {
	var req ClassAttendanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.GetClassAttendanceSummary(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Class attendance summary retrieved successfully")
}

// classPerformance godoc
// GET /report/performance/class?class_section_id=xxx&exam_id=xxx
func (h *Handler) classPerformance(c *gin.Context) {
	var req ClassPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.GetClassPerformanceReport(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Class performance report retrieved successfully")
}

// feeCollection godoc
// GET /report/fee-collection?academic_year_id=xxx&standard_id=xxx&class_section_id=xxx
// standard_id and class_section_id are optional filters
func (h *Handler) feeCollection(c *gin.Context) {
	var req FeeCollectionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.GetFeeCollectionReport(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Fee collection report retrieved successfully")
}

// teacherAttendance godoc
// GET /report/attendance/teacher?teacher_id=xxx&month=7&year=2024
// teacher_id is optional — omit to get summary for all teachers
func (h *Handler) teacherAttendance(c *gin.Context) {
	var req TeacherAttendanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.GetTeacherAttendanceSummary(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Teacher attendance summary retrieved successfully")
}
