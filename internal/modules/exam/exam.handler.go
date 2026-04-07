package exam

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

// ─── Handler ─────────────────────────────────────────────────────────────────

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
	exam := r.Group(constants.ApiExamPath)
	exam.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		// ── Exam CRUD ──────────────────────────────────────────────────────
		exam.POST("/", h.createExam)
		exam.GET("/", h.listExams)
		exam.GET("/:id", h.getExam)
		exam.PUT("/:id", h.updateExam)
		exam.PATCH("/:id/publish", h.publishExam)
		exam.DELETE("/:id", h.deleteExam)

		// ── ExamSchedule — nested under /exam/:id/schedule ────────────
		exam.POST("/:id/schedule", h.createSchedule)
		exam.GET("/:id/schedule", h.listSchedulesByExam)

		// ── ExamSchedule — flat routes for easy cross-section queries ──────
		schedule := r.Group(constants.ApiScheduleExamPath)
		schedule.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
		{
			schedule.GET("/:id", h.getSchedule)
			schedule.PUT("/:id", h.updateSchedule)
			schedule.DELETE("/:id", h.deleteSchedule)
			schedule.GET("/class-section/:class_section_id", h.listSchedulesByClassSection)

			// ── ExamResult nested under schedule ───────────────────────────
			schedule.POST("/:id/result", h.createResult)
			schedule.POST("/:id/result/bulk", h.bulkCreateResults)
			schedule.GET("/:id/result", h.listResultsBySchedule)
		}

		// ── ExamResult — flat routes ───────────────────────────────────────
		result := r.Group(constants.ApiResultExamPath)
		result.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
		{
			result.GET("/:id", h.getResult)
			result.PUT("/:id", h.updateResult)
			result.DELETE("/:id", h.deleteResult)
			result.GET("/student/:student_enrollment_id", h.listResultsByStudent)
		}
	}
}

// ─── Exam handlers ────────────────────────────────────────────────────────────

func (h *Handler) createExam(c *gin.Context) {
	var req CreateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.CreateExam(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Exam created successfully")
}

func (h *Handler) getExam(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetExamByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam retrieved successfully")
}

func (h *Handler) listExams(c *gin.Context) {
	var f FilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	q := query_params.Query[FilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, allowedSortFields, "start_date"),
		Filter:     f,
	}

	exams, total, err := h.service.ListExams(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	meta := pagination.NewMeta(total, q.Pagination)
	response.SuccessPaginated(c, exams, meta, "Exams listed successfully")
}

func (h *Handler) updateExam(c *gin.Context) {
	id := c.Param("id")
	var req UpdateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateExam(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam updated successfully")
}

func (h *Handler) publishExam(c *gin.Context) {
	id := c.Param("id")
	var req PublishExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.PublishExam(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	msg := "Exam unpublished successfully"
	if req.IsPublished {
		msg = "Exam published successfully"
	}
	response.Success(c, resp, msg)
}

func (h *Handler) deleteExam(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteExam(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Exam deleted successfully")
}

// ─── ExamSchedule handlers ────────────────────────────────────────────────────

func (h *Handler) createSchedule(c *gin.Context) {
	examID := c.Param("id")
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.CreateSchedule(c.Request.Context(), examID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Exam schedule created successfully")
}

func (h *Handler) getSchedule(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetScheduleByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam schedule retrieved successfully")
}

func (h *Handler) listSchedulesByExam(c *gin.Context) {
	examID := c.Param("id")
	resp, err := h.service.ListSchedulesByExam(c.Request.Context(), examID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam schedules listed successfully")
}

func (h *Handler) listSchedulesByClassSection(c *gin.Context) {
	classSectionID := c.Param("class_section_id")
	resp, err := h.service.ListSchedulesByClassSection(c.Request.Context(), classSectionID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam schedules listed successfully")
}

func (h *Handler) updateSchedule(c *gin.Context) {
	id := c.Param("id")
	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateSchedule(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam schedule updated successfully")
}

func (h *Handler) deleteSchedule(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteSchedule(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Exam schedule deleted successfully")
}

// ─── ExamResult handlers ──────────────────────────────────────────────────────

func (h *Handler) createResult(c *gin.Context) {
	scheduleID := c.Param("id")
	var req CreateResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.CreateResult(c.Request.Context(), scheduleID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Exam result created successfully")
}

func (h *Handler) bulkCreateResults(c *gin.Context) {
	scheduleID := c.Param("id")
	var req BulkCreateResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.BulkCreateResults(c.Request.Context(), scheduleID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Exam results created successfully")
}

func (h *Handler) getResult(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.service.GetResultByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam result retrieved successfully")
}

func (h *Handler) listResultsBySchedule(c *gin.Context) {
	scheduleID := c.Param("id")
	resp, err := h.service.ListResultsBySchedule(c.Request.Context(), scheduleID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam results listed successfully")
}

func (h *Handler) listResultsByStudent(c *gin.Context) {
	enrollmentID := c.Param("student_enrollment_id")
	resp, err := h.service.ListResultsByStudent(c.Request.Context(), enrollmentID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Student exam results listed successfully")
}

func (h *Handler) updateResult(c *gin.Context) {
	id := c.Param("id")
	var req UpdateResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.service.UpdateResult(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Exam result updated successfully")
}

func (h *Handler) deleteResult(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteResult(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Exam result deleted successfully")
}
