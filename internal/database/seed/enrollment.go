package seed

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type EnrollmentSeeder struct{}

func (s *EnrollmentSeeder) Name() string { return "enrollment" }

func (s *EnrollmentSeeder) Run(ctx context.Context, db *gorm.DB) error {
	var student database.Student
	var section database.ClassSection

	db.First(&student)
	db.First(&section)

	en := database.StudentEnrollment{
		StudentID:      student.ID,
		ClassSectionID: section.ID,
		RollNumber:     1,
	}
	db.FirstOrCreate(&en, database.StudentEnrollment{
		StudentID:      student.ID,
		ClassSectionID: section.ID,
	})

	return nil
}
