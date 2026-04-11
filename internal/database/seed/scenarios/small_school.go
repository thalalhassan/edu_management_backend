package seed

import (
	"context"

	"gorm.io/gorm"
)

type ScenarioSmallSchool struct{}

func (s *ScenarioSmallSchool) Name() string { return "scenario_small" }

func (s *ScenarioSmallSchool) Run(ctx context.Context, db *gorm.DB) error {
	// simple scenario extension point
	return nil
}
