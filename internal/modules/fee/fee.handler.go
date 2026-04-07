package fee

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

type Handler struct {
	structureSvc StructureService
	recordSvc    RecordService
	config       *config.Config
	db           interface{ DB() interface{} } // raw gorm.DB for BulkGenerate
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
	fee := r.Group(constants.ApiFeePath)
	fee.Use(middleware.AuthCheckMiddleware(&h.config.JWT))
	{
		// ── Fee Structure ──────────────────────────────────
		structure := fee.Group("/structure")
		{
			structure.POST("/", h.createStructure)
			structure.POST("/bulk", h.bulkCreateStructure)
			structure.GET("/", h.listStructures)
			structure.GET("/:id", h.getStructureByID)
			// Convenience: all components for a standard in an AY
			structure.GET("/standard/:standard_id", h.listByStandardAndYear)
			structure.PUT("/:id", h.updateStructure)
			structure.DELETE("/:id", h.deleteStructure)
		}

		// ── Fee Records ────────────────────────────────────
		record := fee.Group("/record")
		{
			record.POST("/", h.createRecord)
			record.POST("/bulk-generate", h.bulkGenerate)
			record.GET("/", h.listRecords)
			record.GET("/:id", h.getRecordByID)
			// Student fee summary dashboard
			record.GET("/student/:enrollment_id/summary", h.getStudentSummary)
			// Payment actions
			record.PATCH("/:id/pay", h.recordPayment)
			record.PATCH("/:id/waive", h.waive)
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
	response.Created(c, resp, "Fee structure created successfully")
}

func (h *Handler) bulkCreateStructure(c *gin.Context) {
	var req BulkCreateStructureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.structureSvc.BulkCreate(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Fee structures created successfully")
}

func (h *Handler) getStructureByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.structureSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Fee structure retrieved successfully")
}

func (h *Handler) listStructures(c *gin.Context) {
	var f StructureFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	q := query_params.Query[StructureFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, AllowedStructureSortFields, "fee_component"),
		Filter:     f,
	}
	resp, total, err := h.structureSvc.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, q.Pagination), "Fee structures listed successfully")
}

func (h *Handler) listByStandardAndYear(c *gin.Context) {
	standardID := c.Param("standard_id")
	academicYearID := c.GetHeader("X-Academic-Year-ID")
	if academicYearID == "" {
		academicYearID = c.Query("academic_year_id")
	}
	if academicYearID == "" {
		response.BadRequest(c, "academic_year_id is required")
		return
	}
	resp, err := h.structureSvc.ListByStandardAndYear(c.Request.Context(), standardID, academicYearID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Fee structures retrieved successfully")
}

func (h *Handler) updateStructure(c *gin.Context) {
	id := c.Param("id")
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
	response.Success(c, resp, "Fee structure updated successfully")
}

func (h *Handler) deleteStructure(c *gin.Context) {
	id := c.Param("id")
	if err := h.structureSvc.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Fee structure deleted successfully")
}

// ── RECORD HANDLERS ───────────────────────────────────────────

func (h *Handler) createRecord(c *gin.Context) {
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.recordSvc.Create(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, resp, "Fee record created successfully")
}

// bulkGenerate creates fee records for all enrolled students in a class section.
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
	response.Created(c, resp, "Fee records generated successfully")
}

func (h *Handler) getRecordByID(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.recordSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp, "Fee record retrieved successfully")
}

func (h *Handler) listRecords(c *gin.Context) {
	var f RecordFilterParams
	if err := c.ShouldBindQuery(&f); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	q := query_params.Query[RecordFilterParams]{
		Pagination: pagination.NewFromRequest(c),
		Sort:       query_params.NewSortFromRequest(c, AllowedRecordSortFields, "due_date"),
		Filter:     f,
	}
	resp, total, err := h.recordSvc.List(c.Request.Context(), q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessPaginated(c, resp, pagination.NewMeta(total, q.Pagination), "Fee records listed successfully")
}

func (h *Handler) getStudentSummary(c *gin.Context) {
	enrollmentID := c.Param("enrollment_id")
	resp, err := h.recordSvc.GetStudentSummary(c.Request.Context(), enrollmentID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, resp, "Student fee summary retrieved successfully")
}

func (h *Handler) recordPayment(c *gin.Context) {
	id := c.Param("id")
	var req RecordPaymentRequest
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

func (h *Handler) waive(c *gin.Context) {
	id := c.Param("id")
	var req WaiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := h.recordSvc.Waive(c.Request.Context(), id, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, resp, "Fee waived successfully")
}

func (h *Handler) deleteRecord(c *gin.Context) {
	id := c.Param("id")
	if err := h.recordSvc.Delete(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success[any](c, nil, "Fee record deleted successfully")
}
