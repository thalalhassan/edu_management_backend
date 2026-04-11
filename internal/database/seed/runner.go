
package seed

import (
	"context"
	"gorm.io/gorm"
)

type Seeder interface {
	Name() string
	Run(ctx context.Context, db *gorm.DB) error
}

type Runner struct {
	db *gorm.DB
	seeders []Seeder
}

func NewRunner(db *gorm.DB) *Runner {
	return &Runner{db: db}
}

func (r *Runner) Register(s ...Seeder) {
	r.seeders = append(r.seeders, s...)
}

func (r *Runner) Run(ctx context.Context) error {
	for _, s := range r.seeders {
		if err := s.Run(ctx, r.db); err != nil {
			return err
		}
	}
	return nil
}
