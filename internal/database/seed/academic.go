package seed

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type AcademicSeeder struct{}

func (s *AcademicSeeder) Name() string { return "academic" }

func (s *AcademicSeeder) Run(ctx context.Context, db *gorm.DB) error {
	var year database.AcademicYear
	db.First(&year, "is_active = ?", true)

	std := database.Standard{Name: "Grade 1"}
	db.FirstOrCreate(&std, database.Standard{Name: std.Name})

	sec := database.ClassSection{
		AcademicYearID: year.ID,
		StandardID:     std.ID,
		SectionName:    "A",
	}
	db.FirstOrCreate(&sec, database.ClassSection{
		AcademicYearID: year.ID,
		StandardID:     std.ID,
		SectionName:    "A",
	})

	return nil
}
