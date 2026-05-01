package salary

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/middleware"
	"github.com/thalalhassan/edu_management/internal/shared/helper"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type Handler struct {
	structureSvc StructureService
	recordSvc    RecordService
	config       *config.Config
	app          *app.App
}

func NewHandler(structureSvc StructureService, recordSvc RecordService, a *app.App) *Handler {
	return &Handler{structureSvc: structureSvc, recordSvc: recordSvc, config: a.Config, app: a}
}

func RegisterRouter(r *gin.RouterGroup, a *app.App) {
	structRepo := NewStructureRepository(a.DB.Gorm)
	recordRepo := NewRecordRepository(a.DB.Gorm)
	h := NewHandler(
		NewStructureService(structRepo),
		NewRecordService(recordRepo, structRepo),
		a,
	)
	h.Routes(r)
}

func (h *Handler) Routes(r *gin.RouterGroup) {
	salary := r.Group(constants.ApiSalariesPath)
	salary.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		// ── Salary Structure ───────────────────────────────
		structure := salary.Group("/structure")
		{
			structure.POST("/", h.createStructure)
			structure.GET("/", h.listStructures)
			structure.GET("/:id", h.getStructureByID)
			// All structures for a teacher — revision history
			structure.GET("/teacher/:teacher_id", h.listByTeacher)
			// Active (current) structure for a teacher
			structure.GET("/teacher/:teacher_id/active", h.getActiveForTeacher)
			structure.PUT("/:id", h.updateStructure)
			structure.DELETE("/:id", h.deleteStructure)
		}

		// ── Salary Records ─────────────────────────────────
		record := salary.Group("/record")
		{
			record.POST("/bulk-generate", h.bulkGenerate)
			record.GET("/", h.listRecords)
			record.GET("/:id", h.getRecordByID)
			// All records for a teacher — payslip history
			record.GET("/teacher/:teacher_id", h.listByTeacherRecords)
			// Monthly rollup — summary card for dashboard
			record.GET("/summary", h.getMonthlySummary)
			record.PATCH("/:id/pay", h.recordPayment)
			record.DELETE("/:id", h.deleteRecord)
		}
	}
}

// ── STRUCTURE HANDLERS ────────────────────────────────────────

func (h *Handler) createStructure(c *gin.Context) {
	var req CreateStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.structureSvc.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Salary structure created successfully")
}

func (h *Handler) getStructureByID(c *gin.Context) {

	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}

	resp, err := h.structureSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Salary structure retrieved successfully")
}

func (h *Handler) getActiveForTeacher(c *gin.Context) {
	teacherID, valid := helper.ParseParamUUIDWithAbort(c, "teacher_id")
	if !valid {
		return
	}
	resp, err := h.structureSvc.GetActiveForTeacher(c.Request.Context(), teacherID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Active salary structure retrieved successfully")
}

func (h *Handler) listStructures(c *gin.Context) {
	var f StructureFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	q := query_params.Query[StructureFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, AllowedStructureSortFields, "effective_from"),
		Filter:     f,
	}
	resp, total, err := h.structureSvc.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, q.Pagination), "Salary structures listed successfully")
}

func (h *Handler) listByTeacher(c *gin.Context) {
	teacherID, valid := helper.ParseParamUUIDWithAbort(c, "teacher_id")
	if !valid {
		return
	}
	resp, err := h.structureSvc.ListByTeacher(c.Request.Context(), teacherID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Salary structures retrieved successfully")
}

func (h *Handler) updateStructure(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	var req UpdateStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.structureSvc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Salary structure updated successfully")
}

func (h *Handler) deleteStructure(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	if err := h.structureSvc.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Salary structure deleted successfully")
}

// ── RECORD HANDLERS ───────────────────────────────────────────

func (h *Handler) bulkGenerate(c *gin.Context) {
	var req BulkGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.recordSvc.BulkGenerate(c.Request.Context(), req, h.app.DB.Gorm)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Salary records generated successfully")
}

func (h *Handler) getRecordByID(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	resp, err := h.recordSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Salary record retrieved successfully")
}

func (h *Handler) listRecords(c *gin.Context) {
	var f RecordFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	q := query_params.Query[RecordFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, AllowedRecordSortFields, "year"),
		Filter:     f,
	}
	resp, total, err := h.recordSvc.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, q.Pagination), "Salary records listed successfully")
}

func (h *Handler) listByTeacherRecords(c *gin.Context) {
	teacherID, valid := helper.ParseParamUUIDWithAbort(c, "teacher_id")
	if !valid {
		return
	}
	resp, err := h.recordSvc.ListByTeacher(c.Request.Context(), teacherID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Salary records retrieved successfully")
}

// getMonthlySummary godoc
// GET /salary/record/summary?academic_year_id=xxx&month=7&year=2024
func (h *Handler) getMonthlySummary(c *gin.Context) {
	academicYearID, err := middleware.GetAcademicYearIDFromContext(c)
	if err != nil {
		response.BadRequest(c, "invalid academic year ID")
		return
	}

	month, err := strconv.Atoi(c.Query("month"))
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "month must be an integer between 1 and 12")
		return
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil || year < 2000 {
		response.BadRequest(c, "year must be a valid 4-digit year")
		return
	}

	resp, err := h.recordSvc.GetMonthlySummary(c.Request.Context(), academicYearID, month, year)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Monthly salary summary retrieved successfully")
}

func (h *Handler) recordPayment(c *gin.Context) {
	var req RecordPaymentRequest

	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.recordSvc.RecordPayment(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Payment recorded successfully")
}

func (h *Handler) deleteRecord(c *gin.Context) {
	id, valid := helper.ParseParamUUIDWithAbort(c, "id")
	if !valid {
		return
	}
	if err := h.recordSvc.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Salary record deleted successfully")
}
