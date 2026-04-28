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
	var emp database.Employee
	db.First(&emp)

	ps := database.SalaryStructure{
		EmployeeID:  emp.ID,
		BasicSalary: decimal.NewFromInt(50000),
	}
	db.FirstOrCreate(&ps, database.SalaryStructure{EmployeeID: emp.ID})
	return nil
}
