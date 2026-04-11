package seed

import (
	"context"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type PeopleSeeder struct{}

func (s *PeopleSeeder) Name() string { return "people" }

func (s *PeopleSeeder) Run(ctx context.Context, db *gorm.DB) error {
	t := database.Teacher{
		EmployeeID:  "T001",
		FirstName:   "Test",
		LastName:    "Teacher",
		JoiningDate: time.Now(),
	}
	db.FirstOrCreate(&t, database.Teacher{EmployeeID: t.EmployeeID})

	s1 := database.Student{
		AdmissionNo:   "S001",
		FirstName:     "Test",
		LastName:      "Student",
		AdmissionDate: time.Now(),
	}
	db.FirstOrCreate(&s1, database.Student{AdmissionNo: s1.AdmissionNo})

	return nil
}
