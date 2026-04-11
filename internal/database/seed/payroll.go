package seed

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type PayrollSeeder struct{}

func (s *PayrollSeeder) Name() string { return "payroll" }

func (s *PayrollSeeder) Run(ctx context.Context, db *gorm.DB) error {
	var t database.Teacher
	db.First(&t)

	ps := database.SalaryStructure{
		TeacherID:   t.ID,
		BasicSalary: decimal.NewFromInt(50000),
	}
	db.FirstOrCreate(&ps, database.SalaryStructure{TeacherID: t.ID})
	return nil
}
