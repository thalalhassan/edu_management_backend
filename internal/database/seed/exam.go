package seed

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type ExamSeeder struct{}

func (s *ExamSeeder) Name() string { return "exam" }

func (s *ExamSeeder) Run(ctx context.Context, db *gorm.DB) error {
	var year database.AcademicYear
	db.First(&year)

	ex := database.Exam{Name: "MidTerm", AcademicYearID: year.ID}
	db.FirstOrCreate(&ex, database.Exam{Name: ex.Name})
	return nil
}
