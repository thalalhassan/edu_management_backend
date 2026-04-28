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

	// Create FeeComponent first
	feeComp := database.FeeComponent{
		Name:        "Tuition",
		Description: stringPtr("Monthly tuition fee"),
		IsActive:    true,
	}
	db.FirstOrCreate(&feeComp, database.FeeComponent{Name: "Tuition"})

	f := database.FeeStructure{
		StandardID:     std.ID,
		AcademicYearID: year.ID,
		FeeComponentID: feeComp.ID,
		Amount:         decimal.NewFromInt(10000),
	}

	db.FirstOrCreate(&f, database.FeeStructure{
		StandardID:     std.ID,
		AcademicYearID: year.ID,
		FeeComponentID: feeComp.ID,
	})

	return nil
}
