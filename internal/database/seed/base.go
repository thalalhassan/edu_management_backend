package seed

import (
	"context"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type BaseSeeder struct{}

func (s *BaseSeeder) Name() string { return "base" }

func (s *BaseSeeder) Run(ctx context.Context, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		year := database.AcademicYear{
			Name:      "2025-2026",
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
			IsActive:  true,
		}
		tx.FirstOrCreate(&year, database.AcademicYear{Name: year.Name})
		return nil
	})
}
