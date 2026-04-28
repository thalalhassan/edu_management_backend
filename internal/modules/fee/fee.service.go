package fee

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────────────────────────
// STRUCTURE SERVICE
// ──────────────────────────────────────────────────────────────

type StructureService interface {
	Create(ctx context.Context, req CreateStructureRequest) (*FeeStructureResponse, error)
	BulkCreate(ctx context.Context, req BulkCreateStructureRequest) ([]*FeeStructureResponse, error)
	GetByID(ctx context.Context, id string) (*FeeStructureResponse, error)
	List(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*FeeStructureResponse, int64, error)
	ListByStandardAndYear(ctx context.Context, standardID, academicYearID string) ([]*FeeStructureResponse, error)
	Update(ctx context.Context, id string, req UpdateStructureRequest) (*FeeStructureResponse, error)
	Delete(ctx context.Context, id string) error
}

type structureService struct {
	repo StructureRepository
}

func NewStructureService(repo StructureRepository) StructureService {
	return &structureService{repo: repo}
}

func (s *structureService) Create(ctx context.Context, req CreateStructureRequest) (*FeeStructureResponse, error) {
	isDup, err := s.repo.IsDuplicate(ctx, req.AcademicYearID, req.StandardID, req.FeeComponentID)
	if err != nil {
		return nil, fmt.Errorf("fee.StructureService.Create.IsDuplicate: %w", err)
	}
	if isDup {
		return nil, fmt.Errorf("fee.StructureService.Create: fee component already exists for this standard and academic year")
	}

	f := &FeeStructure{
		AcademicYearID: req.AcademicYearID,
		StandardID:     req.StandardID,
		FeeComponentID: req.FeeComponentID,
		Amount:         req.Amount,
		DueDate:        req.DueDate,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("fee.StructureService.Create: %w", err)
	}
	created, err := s.repo.GetByID(ctx, f.ID)
	if err != nil {
		return nil, fmt.Errorf("fee.StructureService.Create.GetByID: %w", err)
	}
	return ToFeeStructureResponse(created), nil
}

// BulkCreate creates all fee components for a standard in a single call.
// Duplicate components within the batch are rejected entirely — no partial inserts.
func (s *structureService) BulkCreate(ctx context.Context, req BulkCreateStructureRequest) ([]*FeeStructureResponse, error) {
	// Check for duplicates within the batch itself
	seen := make(map[string]bool)
	for _, c := range req.Components {
		if seen[c.FeeComponentID] {
			return nil, fmt.Errorf("fee.StructureService.BulkCreate: duplicate component in request")
		}
		seen[c.FeeComponentID] = true

		isDup, err := s.repo.IsDuplicate(ctx, req.AcademicYearID, req.StandardID, c.FeeComponentID)
		if err != nil {
			return nil, fmt.Errorf("fee.StructureService.BulkCreate.IsDuplicate: %w", err)
		}
		if isDup {
			return nil, fmt.Errorf("fee.StructureService.BulkCreate: component already exists for this standard and year")
		}
	}

	structs := make([]*FeeStructure, len(req.Components))
	for i, c := range req.Components {
		structs[i] = &FeeStructure{
			AcademicYearID: req.AcademicYearID,
			StandardID:     req.StandardID,
			FeeComponentID: c.FeeComponentID,
			Amount:         c.Amount,
			DueDate:        c.DueDate,
		}
	}
	if err := s.repo.BulkCreate(ctx, structs); err != nil {
		return nil, fmt.Errorf("fee.StructureService.BulkCreate: %w", err)
	}

	responses := make([]*FeeStructureResponse, len(structs))
	for i, f := range structs {
		responses[i] = ToFeeStructureResponse(f)
	}
	return responses, nil
}

func (s *structureService) GetByID(ctx context.Context, id string) (*FeeStructureResponse, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fee.StructureService.GetByID: %w", err)
	}
	return ToFeeStructureResponse(f), nil
}

func (s *structureService) List(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*FeeStructureResponse, int64, error) {
	structs, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("fee.StructureService.List: %w", err)
	}
	responses := make([]*FeeStructureResponse, len(structs))
	for i, f := range structs {
		responses[i] = ToFeeStructureResponse(f)
	}
	return responses, total, nil
}

func (s *structureService) ListByStandardAndYear(ctx context.Context, standardID, academicYearID string) ([]*FeeStructureResponse, error) {
	structs, err := s.repo.FindByStandardAndYear(ctx, standardID, academicYearID)
	if err != nil {
		return nil, fmt.Errorf("fee.StructureService.ListByStandardAndYear: %w", err)
	}
	responses := make([]*FeeStructureResponse, len(structs))
	for i, f := range structs {
		responses[i] = ToFeeStructureResponse(f)
	}
	return responses, nil
}

func (s *structureService) Update(ctx context.Context, id string, req UpdateStructureRequest) (*FeeStructureResponse, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fee.StructureService.Update.GetByID: %w", err)
	}
	if req.FeeComponentID != nil {
		// Guard: new component ID must not clash
		isDup, err := s.repo.IsDuplicate(ctx, f.AcademicYearID, f.StandardID, *req.FeeComponentID)
		if err != nil {
			return nil, fmt.Errorf("fee.StructureService.Update.IsDuplicate: %w", err)
		}
		if isDup && *req.FeeComponentID != f.FeeComponentID {
			return nil, fmt.Errorf("fee.StructureService.Update: component already exists")
		}
		f.FeeComponentID = *req.FeeComponentID
	}
	if req.Amount != nil {
		f.Amount = *req.Amount
	}
	if req.DueDate != nil {
		f.DueDate = req.DueDate
	}
	if err := s.repo.Update(ctx, id, f); err != nil {
		return nil, fmt.Errorf("fee.StructureService.Update.Save: %w", err)
	}
	return ToFeeStructureResponse(f), nil
}

func (s *structureService) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("fee.StructureService.Delete.GetByID: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("fee.StructureService.Delete: %w", err)
	}
	return nil
}

// ──────────────────────────────────────────────────────────────
// RECORD SERVICE
// ──────────────────────────────────────────────────────────────

type RecordService interface {
	Create(ctx context.Context, req CreateRecordRequest) (*FeeRecordResponse, error)
	BulkGenerate(ctx context.Context, req BulkGenerateRequest, db *gorm.DB) ([]*FeeRecordResponse, error)
	GetByID(ctx context.Context, id string) (*FeeRecordResponse, error)
	List(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*FeeRecordResponse, int64, error)
	GetStudentSummary(ctx context.Context, studentEnrollmentID string) (*StudentFeeSummary, error)
	RecordPayment(ctx context.Context, id string, req RecordPaymentRequest) (*FeeRecordResponse, error)
	Waive(ctx context.Context, id string, req WaiveRequest) (*FeeRecordResponse, error)
	Delete(ctx context.Context, id string) error
}

type recordService struct {
	repo       RecordRepository
	structRepo StructureRepository
}

func NewRecordService(repo RecordRepository, structRepo StructureRepository) RecordService {
	return &recordService{repo: repo, structRepo: structRepo}
}

func (s *recordService) Create(ctx context.Context, req CreateRecordRequest) (*FeeRecordResponse, error) {
	rec := &FeeRecord{
		StudentEnrollmentID: req.StudentEnrollmentID,
		FeeComponentID:      req.FeeComponentID,
		AmountDue:           req.AmountDue,
		DueDate:             req.DueDate,
		Status:              database.FeeStatusPending,
		Remarks:             req.Remarks,
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, fmt.Errorf("fee.RecordService.Create: %w", err)
	}
	created, err := s.repo.GetByID(ctx, rec.ID)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.Create.GetByID: %w", err)
	}
	return ToFeeRecordResponse(created), nil
}

// BulkGenerate creates fee records for all enrolled students in a class section
// by reading the fee structure for their standard + academic year.
// Typical use: beginning of term, admin generates fees for the whole class.
func (s *recordService) BulkGenerate(ctx context.Context, req BulkGenerateRequest, db *gorm.DB) ([]*FeeRecordResponse, error) {
	// Fetch class section with standard + academic year
	var classSection database.ClassSection
	if err := db.WithContext(ctx).
		Preload("Standard").
		Preload("AcademicYear").
		First(&classSection, "id = ?", req.ClassSectionID).Error; err != nil {
		return nil, fmt.Errorf("fee.RecordService.BulkGenerate: class section not found: %w", err)
	}

	// Fetch fee structures for this standard + AY
	structures, err := s.structRepo.FindByStandardAndYear(ctx, classSection.StandardID, classSection.AcademicYearID)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.BulkGenerate.FindStructures: %w", err)
	}
	if len(structures) == 0 {
		return nil, errors.New("fee.RecordService.BulkGenerate: no fee structure defined for this standard and academic year")
	}

	// Fetch enrolled students
	var enrollments []database.StudentEnrollment
	if err := db.WithContext(ctx).
		Preload("Student").
		Where("class_section_id = ? AND status = ?", req.ClassSectionID, database.EnrollmentStatusEnrolled).
		Find(&enrollments).Error; err != nil {
		return nil, fmt.Errorf("fee.RecordService.BulkGenerate.FindEnrollments: %w", err)
	}
	if len(enrollments) == 0 {
		return nil, errors.New("fee.RecordService.BulkGenerate: no enrolled students found in this class section")
	}

	// Build records — one per student per fee component
	var records []*FeeRecord
	for _, enrollment := range enrollments {
		for _, structure := range structures {
			dueDate := time.Now()
			if structure.DueDate != nil {
				dueDate = *structure.DueDate
			}
			records = append(records, &FeeRecord{
				StudentEnrollmentID: enrollment.ID,
				FeeComponentID:      structure.FeeComponentID,
				AmountDue:           structure.Amount,
				DueDate:             dueDate,
				Status:              database.FeeStatusPending,
			})
		}
	}

	if err := s.repo.BulkCreate(ctx, records); err != nil {
		return nil, fmt.Errorf("fee.RecordService.BulkGenerate.BulkCreate: %w", err)
	}

	responses := make([]*FeeRecordResponse, len(records))
	for i, r := range records {
		responses[i] = ToFeeRecordResponse(r)
	}
	return responses, nil
}

func (s *recordService) GetByID(ctx context.Context, id string) (*FeeRecordResponse, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.GetByID: %w", err)
	}
	return ToFeeRecordResponse(rec), nil
}

func (s *recordService) List(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*FeeRecordResponse, int64, error) {
	records, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("fee.RecordService.List: %w", err)
	}
	responses := make([]*FeeRecordResponse, len(records))
	for i, r := range records {
		responses[i] = ToFeeRecordResponse(r)
	}
	return responses, total, nil
}

// GetStudentSummary returns a rolled-up view of all fee records for a student
// with totals — the primary view for the student fee dashboard.
func (s *recordService) GetStudentSummary(ctx context.Context, studentEnrollmentID string) (*StudentFeeSummary, error) {
	records, err := s.repo.FindByEnrollment(ctx, studentEnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.GetStudentSummary.FindByEnrollment: %w", err)
	}

	totalDue, totalPaid, err := s.repo.SumByEnrollment(ctx, studentEnrollmentID)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.GetStudentSummary.Sum: %w", err)
	}

	summary := &StudentFeeSummary{
		StudentEnrollmentID: studentEnrollmentID,
		TotalDue:            totalDue,
		TotalPaid:           totalPaid,
		TotalBalance:        totalDue.Sub(totalPaid),
		Records:             make([]FeeRecordResponse, len(records)),
	}

	// Populate student and class section info from first record if available
	if len(records) > 0 && records[0].StudentEnrollment.Student.ID != "" {
		s := records[0].StudentEnrollment.Student
		summary.StudentName = s.FirstName + " " + s.LastName
		summary.AdmissionNo = s.AdmissionNo
		cs := records[0].StudentEnrollment.ClassSection
		summary.ClassSection = cs.Standard.Name + " - " + cs.SectionName
	}

	for i, r := range records {
		summary.Records[i] = *ToFeeRecordResponse(r)
	}
	return summary, nil
}

// RecordPayment applies a payment against a fee record and recomputes status.
func (s *recordService) RecordPayment(ctx context.Context, id string, req RecordPaymentRequest) (*FeeRecordResponse, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.RecordPayment.GetByID: %w", err)
	}

	// Guard: cannot pay a waived record
	if rec.Status == database.FeeStatusWaived {
		return nil, errors.New("fee.RecordService.RecordPayment: cannot record payment on a waived fee")
	}

	// Guard: payment cannot exceed the outstanding balance
	balance := rec.AmountDue.Sub(rec.AmountPaid)
	if req.AmountPaid.GreaterThan(balance) {
		return nil, fmt.Errorf("fee.RecordService.RecordPayment: payment amount %.2f exceeds outstanding balance %.2f",
			req.AmountPaid.InexactFloat64(), balance.InexactFloat64())
	}

	rec.AmountPaid = rec.AmountPaid.Add(req.AmountPaid)
	rec.PaidDate = &req.PaidDate
	rec.TransactionRef = req.TransactionRef
	rec.Remarks = req.Remarks
	rec.Status = computeStatus(rec.AmountDue, rec.AmountPaid, rec.DueDate)

	if err := s.repo.Update(ctx, id, rec); err != nil {
		return nil, fmt.Errorf("fee.RecordService.RecordPayment.Update: %w", err)
	}
	return ToFeeRecordResponse(rec), nil
}

// Waive marks a fee record as fully waived — amount_paid stays as-is
// but status is set to WAIVED and no further payments are accepted.
func (s *recordService) Waive(ctx context.Context, id string, req WaiveRequest) (*FeeRecordResponse, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fee.RecordService.Waive.GetByID: %w", err)
	}
	if rec.Status == database.FeeStatusPaid {
		return nil, errors.New("fee.RecordService.Waive: cannot waive an already paid fee")
	}
	if rec.Status == database.FeeStatusWaived {
		return nil, errors.New("fee.RecordService.Waive: fee is already waived")
	}

	rec.Status = database.FeeStatusWaived
	rec.Remarks = req.Remarks

	if err := s.repo.Update(ctx, id, rec); err != nil {
		return nil, fmt.Errorf("fee.RecordService.Waive.Update: %w", err)
	}
	return ToFeeRecordResponse(rec), nil
}

func (s *recordService) Delete(ctx context.Context, id string) error {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("fee.RecordService.Delete.GetByID: %w", err)
	}
	if rec.Status == database.FeeStatusPaid {
		return errors.New("fee.RecordService.Delete: cannot delete a paid fee record")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("fee.RecordService.Delete: %w", err)
	}
	return nil
}
