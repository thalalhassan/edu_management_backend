package seed

import (
	"context"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, command string) error {
	r := NewRunner(db)

	switch command {
	case "reset":
		r.Register(&ResetAndBootstrapSeeder{})
	default:
		r.Register(
			&BaseSeeder{},
			&SubjectSeeder{},
			&AcademicSeeder{},
			&PeopleSeeder{},
			&EnrollmentSeeder{},
			&ExamSeeder{},
			&FeeSeeder{},
			&PayrollSeeder{},
			// &ScenarioSmallSchool{},
		)
	}

	return r.Run(context.Background())
}
