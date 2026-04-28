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
	emp := database.Employee{
		EmployeeCode: "T001",
		FirstName:    "Test",
		LastName:     "Teacher",
		Category:     database.EmployeeCategoryTeacher,
		Designation:  "Teacher",
		JoiningDate:  time.Now(),
	}
	db.FirstOrCreate(&emp, database.Employee{EmployeeCode: emp.EmployeeCode})

	s1 := database.Student{
		AdmissionNo:   "S001",
		FirstName:     "Test",
		LastName:      "Student",
		AdmissionDate: time.Now(),
	}
	db.FirstOrCreate(&s1, database.Student{AdmissionNo: s1.AdmissionNo})

	s2 := database.Student{
		AdmissionNo:   "S002",
		FirstName:     "Test2",
		LastName:      "Student2",
		AdmissionDate: time.Now(),
	}
	db.FirstOrCreate(&s2, database.Student{AdmissionNo: s2.AdmissionNo})

	return nil
}
