package seed

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type SubjectSeeder struct{}

func (s *SubjectSeeder) Name() string { return "subject" }

func (s *SubjectSeeder) Run(ctx context.Context, db *gorm.DB) error {
	subjects := []database.Subject{
		{Code: "MATH", Name: "Mathematics"},
		{Code: "ENG", Name: "English"},
	}
	for _, sub := range subjects {
		db.FirstOrCreate(&sub, database.Subject{Code: sub.Code})
	}
	return nil
}
