package salary

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────────────────────────
// STRUCTURE SERVICE
// ──────────────────────────────────────────────────────────────

type StructureService interface {
	Create(ctx context.Context, req CreateStructureRequest) (*SalaryStructureResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SalaryStructureResponse, error)
	GetActiveForTeacher(ctx context.Context, teacherID uuid.UUID) (*SalaryStructureResponse, error)
	List(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*SalaryStructureResponse, int64, error)
	ListByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*SalaryStructureResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateStructureRequest) (*SalaryStructureResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type structureService struct {
	repo StructureRepository
}

func NewStructureService(repo StructureRepository) StructureService {
	return &structureService{repo: repo}
}

func (s *structureService) Create(ctx context.Context, req CreateStructureRequest) (*SalaryStructureResponse, error) {
	str := &SalaryStructure{
		EmployeeID:     req.EmployeeID,
		BasicSalary:    req.BasicSalary,
		HRA:            req.HRA,
		DA:             req.DA,
		OtherAllowance: req.OtherAllowance,
		PF:             req.PF,
		TDS:            req.TDS,
		OtherDeduction: req.OtherDeduction,
		EffectiveFrom:  req.EffectiveFrom,
		Remarks:        req.Remarks,
	}
	if err := s.repo.Create(ctx, str); err != nil {
		return nil, fmt.Errorf("salary.StructureService.Create: %w", err)
	}
	created, err := s.repo.GetByID(ctx, str.ID)
	if err != nil {
		return nil, fmt.Errorf("salary.StructureService.Create.GetByID: %w", err)
	}
	return ToStructureResponse(created), nil
}

func (s *structureService) GetByID(ctx context.Context, id uuid.UUID) (*SalaryStructureResponse, error) {
	str, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("salary.StructureService.GetByID: %w", err)
	}
	return ToStructureResponse(str), nil
}

func (s *structureService) GetActiveForTeacher(ctx context.Context, teacherID uuid.UUID) (*SalaryStructureResponse, error) {
	str, err := s.repo.GetActiveForTeacher(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("salary.StructureService.GetActiveForTeacher: no active structure found for this teacher")
	}
	return ToStructureResponse(str), nil
}

func (s *structureService) List(ctx context.Context, q query_params.Query[StructureFilterParams]) ([]*SalaryStructureResponse, int64, error) {
	structs, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("salary.StructureService.List: %w", err)
	}
	responses := make([]*SalaryStructureResponse, len(structs))
	for i, str := range structs {
		responses[i] = ToStructureResponse(str)
	}
	return responses, total, nil
}

func (s *structureService) ListByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*SalaryStructureResponse, error) {
	structs, err := s.repo.FindByTeacher(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("salary.StructureService.ListByTeacher: %w", err)
	}
	responses := make([]*SalaryStructureResponse, len(structs))
	for i, str := range structs {
		responses[i] = ToStructureResponse(str)
	}
	return responses, nil
}

func (s *structureService) Update(ctx context.Context, id uuid.UUID, req UpdateStructureRequest) (*SalaryStructureResponse, error) {
	str, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("salary.StructureService.Update.GetByID: %w", err)
	}
	if req.BasicSalary != nil {
		str.BasicSalary = *req.BasicSalary
	}
	if req.HRA != nil {
		str.HRA = *req.HRA
	}
	if req.DA != nil {
		str.DA = *req.DA
	}
	if req.OtherAllowance != nil {
		str.OtherAllowance = *req.OtherAllowance
	}
	if req.PF != nil {
		str.PF = *req.PF
	}
	if req.TDS != nil {
		str.TDS = *req.TDS
	}
	if req.OtherDeduction != nil {
		str.OtherDeduction = *req.OtherDeduction
	}
	if req.EffectiveFrom != nil {
		str.EffectiveFrom = *req.EffectiveFrom
	}
	if req.Remarks != nil {
		str.Remarks = req.Remarks
	}
	if err := s.repo.Update(ctx, id, str); err != nil {
		return nil, fmt.Errorf("salary.StructureService.Update.Save: %w", err)
	}
	return ToStructureResponse(str), nil
}

func (s *structureService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("salary.StructureService.Delete.GetByID: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("salary.StructureService.Delete: %w", err)
	}
	return nil
}

// ──────────────────────────────────────────────────────────────
// RECORD SERVICE
// ──────────────────────────────────────────────────────────────

type RecordService interface {
	BulkGenerate(ctx context.Context, req BulkGenerateRequest, db *gorm.DB) ([]*SalaryRecordResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SalaryRecordResponse, error)
	List(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*SalaryRecordResponse, int64, error)
	ListByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*SalaryRecordResponse, error)
	GetMonthlySummary(ctx context.Context, academicYearID uuid.UUID, month, year int) (*MonthlySummary, error)
	RecordPayment(ctx context.Context, id uuid.UUID, req RecordPaymentRequest) (*SalaryRecordResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type recordService struct {
	repo       RecordRepository
	structRepo StructureRepository
}

func NewRecordService(repo RecordRepository, structRepo StructureRepository) RecordService {
	return &recordService{repo: repo, structRepo: structRepo}
}

// BulkGenerate creates salary records for all active teachers for the given month.
// Reads each teacher's active salary structure, computes gross/deductions/net,
// optionally prorates based on attendance, then inserts all records atomically.
func (s *recordService) BulkGenerate(ctx context.Context, req BulkGenerateRequest, db *gorm.DB) ([]*SalaryRecordResponse, error) {
	// Fetch all active teachers
	var teachers []database.Employee
	if err := db.WithContext(ctx).Where("is_active = true").Find(&teachers).Error; err != nil {
		return nil, fmt.Errorf("salary.RecordService.BulkGenerate.FetchTeachers: %w", err)
	}
	if len(teachers) == 0 {
		return nil, errors.New("salary.RecordService.BulkGenerate: no active teachers found")
	}

	var records []*SalaryRecord
	for _, t := range teachers {
		// Skip if record already exists for this teacher+month+year
		isDup, err := s.repo.IsDuplicate(ctx, t.ID, req.Month, req.Year)
		if err != nil {
			return nil, fmt.Errorf("salary.RecordService.BulkGenerate.IsDuplicate(%s): %w", t.ID, err)
		}
		if isDup {
			continue
		}

		// Fetch active salary structure
		structure, err := s.structRepo.GetActiveForTeacher(ctx, t.ID)
		if err != nil {
			// Skip teachers with no structure rather than failing the whole batch
			continue
		}

		gross, deduction, net := computeSalaryComponents(structure)

		// Attendance-based proration
		presentDays, workingDays := 0, 0
		if req.Prorate {
			var att struct {
				WorkingDays int
				PresentDays int
			}
			db.WithContext(ctx).
				Model(&database.EmployeeAttendance{}).
				Select(`COUNT(*) as working_days,
					SUM(CASE WHEN status = 'PRESENT' OR status = 'LATE' THEN 1
					         WHEN status = 'HALF_DAY' THEN 0 ELSE 0 END) as present_days`).
				Where("employee_id = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?",
					t.ID, req.Month, req.Year).
				Scan(&att)
			workingDays = att.WorkingDays
			presentDays = att.PresentDays
			if workingDays > 0 {
				net = prorateSalary(net, presentDays, workingDays)
			}
		}

		records = append(records, &SalaryRecord{
			EmployeeID:     t.ID,
			AcademicYearID: req.AcademicYearID,
			Month:          req.Month,
			Year:           req.Year,
			WorkingDays:    workingDays,
			PresentDays:    presentDays,
			GrossSalary:    gross,
			TotalDeduction: deduction,
			NetSalary:      net,
			Status:         database.SalaryStatusPending,
		})
	}

	if len(records) == 0 {
		return nil, errors.New("salary.RecordService.BulkGenerate: all records already exist for this month or no teachers have a salary structure")
	}

	if err := s.repo.BulkCreate(ctx, records); err != nil {
		return nil, fmt.Errorf("salary.RecordService.BulkGenerate.BulkCreate: %w", err)
	}

	responses := make([]*SalaryRecordResponse, len(records))
	for i, r := range records {
		responses[i] = ToRecordResponse(r)
	}
	return responses, nil
}

func (s *recordService) GetByID(ctx context.Context, id uuid.UUID) (*SalaryRecordResponse, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("salary.RecordService.GetByID: %w", err)
	}
	return ToRecordResponse(rec), nil
}

func (s *recordService) List(ctx context.Context, q query_params.Query[RecordFilterParams]) ([]*SalaryRecordResponse, int64, error) {
	records, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("salary.RecordService.List: %w", err)
	}
	responses := make([]*SalaryRecordResponse, len(records))
	for i, r := range records {
		responses[i] = ToRecordResponse(r)
	}
	return responses, total, nil
}

func (s *recordService) ListByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*SalaryRecordResponse, error) {
	records, err := s.repo.FindByTeacher(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("salary.RecordService.ListByTeacher: %w", err)
	}
	responses := make([]*SalaryRecordResponse, len(records))
	for i, r := range records {
		responses[i] = ToRecordResponse(r)
	}
	return responses, nil
}

func (s *recordService) GetMonthlySummary(ctx context.Context, academicYearID uuid.UUID, month, year int) (*MonthlySummary, error) {
	summary, err := s.repo.GetMonthlySummary(ctx, academicYearID, month, year)
	if err != nil {
		return nil, fmt.Errorf("salary.RecordService.GetMonthlySummary: %w", err)
	}
	return summary, nil
}

// RecordPayment marks a salary record as paid (or partial).
func (s *recordService) RecordPayment(ctx context.Context, id uuid.UUID, req RecordPaymentRequest) (*SalaryRecordResponse, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("salary.RecordService.RecordPayment.GetByID: %w", err)
	}
	if rec.Status == database.SalaryStatusPaid {
		return nil, errors.New("salary.RecordService.RecordPayment: salary is already fully paid")
	}
	if rec.Status == database.SalaryStatusOnHold {
		return nil, errors.New("salary.RecordService.RecordPayment: salary is on hold — remove hold before recording payment")
	}

	balance := rec.NetSalary.Sub(rec.PaidAmount)
	if req.PaidAmount.GreaterThan(balance) {
		return nil, fmt.Errorf("salary.RecordService.RecordPayment: payment %.2f exceeds outstanding balance %.2f",
			req.PaidAmount.InexactFloat64(), balance.InexactFloat64())
	}

	rec.PaidAmount = rec.PaidAmount.Add(req.PaidAmount)
	rec.PaidDate = &req.PaidDate
	rec.TransactionRef = req.TransactionRef
	rec.Remarks = req.Remarks
	rec.Status = computeStatus(rec.NetSalary, rec.PaidAmount)

	if err := s.repo.Update(ctx, id, rec); err != nil {
		return nil, fmt.Errorf("salary.RecordService.RecordPayment.Update: %w", err)
	}
	return ToRecordResponse(rec), nil
}

func (s *recordService) Delete(ctx context.Context, id uuid.UUID) error {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("salary.RecordService.Delete.GetByID: %w", err)
	}
	if rec.Status == database.SalaryStatusPaid {
		return errors.New("salary.RecordService.Delete: cannot delete a paid salary record")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("salary.RecordService.Delete: %w", err)
	}
	return nil
}
