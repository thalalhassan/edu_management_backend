// Command migrate is the standalone CLI for running database migrations and
// seeding reference data.
//
// Usage:
//
//	migrate up                 – apply all pending migrations
//	migrate down [--steps N]   – rollback N migrations (default: 1; 0 = all)
//	migrate status             – print applied/pending state
//	migrate version            – print current schema version
//
// Environment variables:
//
//	DATABASE_URL        – PostgreSQL DSN (required)
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/database/migrate"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize the application and database connection
	ctx := context.Background()

	appInstance, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	db := appInstance.DB

	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	m := migrate.New(db.Gorm)

	switch os.Args[1] {
	case "up":
		if err := runUp(ctx, m); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}

	case "down":
		steps, err := parseDownSteps(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if err := m.Down(ctx, steps); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}

	case "status":
		if err := m.Status(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}

	case "version":
		v, err := m.Version(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		fmt.Printf("current version: %d\n", v)

	default:
		fmt.Fprintf(os.Stderr, "error: unknown subcommand %q\n", os.Args[1])
		printUsage()
		return 1
	}

	return 0
}

// runUp applies pending migrations, then conditionally seeds.
func runUp(ctx context.Context, m *migrate.Migrator) error {
	if err := m.Up(ctx); err != nil {
		return err
	}
	return nil
}

// parseDownSteps parses the optional --steps flag from args.
// Returns 1 if the flag is absent (roll back exactly one migration).
func parseDownSteps(args []string) (int, error) {
	fs := flag.NewFlagSet("down", flag.ContinueOnError)
	steps := fs.Int("steps", 1, "number of migrations to roll back (0 = all)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0, nil
		}
		return 0, err
	}
	if len(fs.Args()) == 1 {
		// positional fallback: `migrate down 3`
		n, err := strconv.Atoi(fs.Args()[0])
		if err != nil {
			return 0, fmt.Errorf("invalid steps value %q: %w", fs.Args()[0], err)
		}
		return n, nil
	}
	return *steps, nil
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: migrate <subcommand> [flags]

Subcommands:
  up                Apply all pending migrations (also runs seeder if RUN_SEED=true)
  down [--steps N]  Roll back N migrations (default 1; 0 = all)
  status            Print applied/pending migration state
  version           Print current schema version number
  seed              Run the reference-data seeder

Environment variables:
  DATABASE_URL         PostgreSQL DSN (required)
  RUN_SEED             Set to "true" to auto-seed after "up"
  SEED_ADMIN_EMAIL     Super-admin email address
  SEED_ADMIN_PASSWORD  Super-admin password (stored bcrypt-hashed)`)
}
