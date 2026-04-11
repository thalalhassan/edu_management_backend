package seed

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type FeeSeeder struct{}

func (s *FeeSeeder) Name() string { return "fee" }

func (s *FeeSeeder) Run(ctx context.Context, db *gorm.DB) error {
	var std database.Standard
	var year database.AcademicYear

	db.First(&std)
	db.First(&year)

	f := database.FeeStructure{
		StandardID:     std.ID,
		AcademicYearID: year.ID,
		FeeComponent:   "Tuition",
		Amount:         decimal.NewFromInt(10000),
	}

	db.FirstOrCreate(&f, database.FeeStructure{
		StandardID:     std.ID,
		AcademicYearID: year.ID,
		FeeComponent:   "Tuition",
	})

	return nil
}
