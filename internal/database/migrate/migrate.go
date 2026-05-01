// Package migrate provides a versioned SQL migration runner for PostgreSQL.
//
// Migrations are embedded SQL files read from the migrations/ directory.
// State is tracked in a schema_migrations table. Each migration run acquires
// a PostgreSQL advisory lock, executes each pending file inside a transaction,
// and records a SHA-256 checksum. On subsequent runs the stored checksums are
// verified to detect tampering or drift.
package migrate

import (
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:embed sql/*.sql
var migrationsFS embed.FS

const (
	// advisoryLockKey is a fixed int64 used as the PostgreSQL session-level
	// advisory lock key. Change this if running multiple independent migrators
	// on the same PostgreSQL instance.
	advisoryLockKey = int64(7391823741239847)

	// migrationsTable is the name of the state-tracking table.
	migrationsTable = "schema_migrations"
)

// migrationRecord represents a row in schema_migrations.
type migrationRecord struct {
	Version   int       `gorm:"column:version;primaryKey"`
	Name      string    `gorm:"column:name;not null"`
	Checksum  string    `gorm:"column:checksum;not null"`
	AppliedAt time.Time `gorm:"column:applied_at;not null;autoCreateTime"`
}

func (migrationRecord) TableName() string { return migrationsTable }

// migration holds the parsed content of a .up/.down SQL file pair.
type migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
	// Checksum is the SHA-256 hex digest of UpSQL, computed at load time.
	Checksum string
}

// Migrator manages versioned SQL migrations against a PostgreSQL database.
// Create one via New; it carries no mutable state itself.
type Migrator struct {
	db *gorm.DB
}

// New returns a Migrator backed by db.
func New(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Up applies every pending migration in ascending version order.
//
// Before running, it acquires a PostgreSQL session-level advisory lock to
// prevent concurrent migration runs. Each migration file is executed inside
// its own transaction; a failure rolls back that migration and returns an
// error immediately, leaving all previously applied migrations intact.
//
// Checksums of already-applied migrations are verified against the embedded
// files before any new migration is attempted.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("migrate up: create tracking table: %w", err)
	}
	if err := m.acquireLock(ctx); err != nil {
		return err
	}
	defer m.releaseLock(ctx) //nolint:errcheck

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("migrate up: load SQL files: %w", err)
	}

	applied, err := m.fetchApplied(ctx)
	if err != nil {
		return fmt.Errorf("migrate up: fetch applied: %w", err)
	}
	if err := verifyChecksums(migrations, applied); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	appliedSet := make(map[int]struct{}, len(applied))
	for _, r := range applied {
		appliedSet[r.Version] = struct{}{}
	}

	pending := make([]migration, 0, len(migrations))
	for _, mg := range migrations {
		if _, done := appliedSet[mg.Version]; !done {
			pending = append(pending, mg)
		}
	}

	if len(pending) == 0 {
		slog.InfoContext(ctx, "migrate: no pending migrations")
		return nil
	}

	for _, mg := range pending {
		if err := m.applyUp(ctx, mg); err != nil {
			return err
		}
	}
	return nil
}

// Down rolls back the last `steps` applied migrations in descending version
// order. Pass steps ≤ 0 to roll back all applied migrations.
func (m *Migrator) Down(ctx context.Context, steps int) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("migrate down: create tracking table: %w", err)
	}
	if err := m.acquireLock(ctx); err != nil {
		return err
	}
	defer m.releaseLock(ctx) //nolint:errcheck

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("migrate down: load SQL files: %w", err)
	}

	byVersion := make(map[int]migration, len(migrations))
	for _, mg := range migrations {
		byVersion[mg.Version] = mg
	}

	applied, err := m.fetchApplied(ctx)
	if err != nil {
		return fmt.Errorf("migrate down: fetch applied: %w", err)
	}

	// Newest first.
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version > applied[j].Version
	})

	if steps > 0 && steps < len(applied) {
		applied = applied[:steps]
	}

	for _, rec := range applied {
		mg, ok := byVersion[rec.Version]
		if !ok {
			return fmt.Errorf("migrate down: applied version %d has no corresponding SQL file", rec.Version)
		}
		if err := m.applyDown(ctx, mg); err != nil {
			return err
		}
	}
	return nil
}

// Status logs the applied/pending state of every known migration via slog at
// INFO level. Safe to call concurrently; acquires no lock.
func (m *Migrator) Status(ctx context.Context) error {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("migrate status: create tracking table: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("migrate status: load SQL files: %w", err)
	}

	applied, err := m.fetchApplied(ctx)
	if err != nil {
		return fmt.Errorf("migrate status: fetch applied: %w", err)
	}

	appliedMap := make(map[int]migrationRecord, len(applied))
	for _, r := range applied {
		appliedMap[r.Version] = r
	}

	for _, mg := range migrations {
		if rec, ok := appliedMap[mg.Version]; ok {
			slog.InfoContext(ctx, "migration",
				"version", mg.Version,
				"name", mg.Name,
				"state", "applied",
				"applied_at", rec.AppliedAt.UTC().Format(time.RFC3339),
			)
		} else {
			slog.InfoContext(ctx, "migration",
				"version", mg.Version,
				"name", mg.Name,
				"state", "pending",
			)
		}
	}
	return nil
}

// Version returns the highest applied migration version, or 0 if no migrations
// have been applied yet.
func (m *Migrator) Version(ctx context.Context) (int, error) {
	if err := m.ensureMigrationsTable(ctx); err != nil {
		return 0, fmt.Errorf("migrate version: create tracking table: %w", err)
	}

	applied, err := m.fetchApplied(ctx)
	if err != nil {
		return 0, fmt.Errorf("migrate version: fetch applied: %w", err)
	}

	max := 0
	for _, r := range applied {
		if r.Version > max {
			max = r.Version
		}
	}
	return max, nil
}

// RunMigrationsOnStartup is a convenience wrapper intended for use in main
// before the HTTP server starts. It calls Up and — if the RUN_SEED environment
// variable is "true" — also calls Seed.
func RunMigrationsOnStartup(ctx context.Context, db *gorm.DB) error {
	m := New(db)
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("startup: run migrations: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal: lock + tracking table
// ─────────────────────────────────────────────────────────────────────────────

func (m *Migrator) ensureMigrationsTable(ctx context.Context) error {
	ddl := `CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (
    version    INTEGER     PRIMARY KEY,
    name       TEXT        NOT NULL,
    checksum   TEXT        NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`
	return m.db.WithContext(ctx).Exec(ddl).Error
}

func (m *Migrator) acquireLock(ctx context.Context) error {
	var acquired bool
	row := m.db.WithContext(ctx).
		Raw("SELECT pg_try_advisory_lock($1)", advisoryLockKey).
		Row()
	if err := row.Scan(&acquired); err != nil {
		return fmt.Errorf("migrate: advisory lock query failed: %w", err)
	}
	if !acquired {
		return fmt.Errorf("migrate: another process is running migrations (advisory lock key %d is held)", advisoryLockKey)
	}
	return nil
}

func (m *Migrator) releaseLock(ctx context.Context) error {
	return m.db.WithContext(ctx).
		Exec("SELECT pg_advisory_unlock($1)", advisoryLockKey).
		Error
}

func (m *Migrator) fetchApplied(ctx context.Context) ([]migrationRecord, error) {
	var records []migrationRecord
	if err := m.db.WithContext(ctx).
		Order("version ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal: apply / rollback
// ─────────────────────────────────────────────────────────────────────────────

func (m *Migrator) applyUp(ctx context.Context, mg migration) error {
	t0 := time.Now()
	slog.InfoContext(ctx, "migrate: applying", "version", mg.Version, "name", mg.Name)

	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(mg.UpSQL).Error; err != nil {
			return fmt.Errorf("execute SQL: %w", err)
		}
		return tx.Create(&migrationRecord{
			Version:  mg.Version,
			Name:     mg.Name,
			Checksum: mg.Checksum,
		}).Error
	})
	if err != nil {
		return fmt.Errorf("migrate up v%06d (%s): %w", mg.Version, mg.Name, err)
	}

	slog.InfoContext(ctx, "migrate: applied",
		"version", mg.Version,
		"name", mg.Name,
		"elapsed", time.Since(t0).String(),
	)
	return nil
}

func (m *Migrator) applyDown(ctx context.Context, mg migration) error {
	t0 := time.Now()
	slog.InfoContext(ctx, "migrate: rolling back", "version", mg.Version, "name", mg.Name)

	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(mg.DownSQL).Error; err != nil {
			return fmt.Errorf("execute SQL: %w", err)
		}
		return tx.Where("version = ?", mg.Version).Delete(&migrationRecord{}).Error
	})
	if err != nil {
		return fmt.Errorf("migrate down v%06d (%s): %w", mg.Version, mg.Name, err)
	}

	slog.InfoContext(ctx, "migrate: rolled back",
		"version", mg.Version,
		"name", mg.Name,
		"elapsed", time.Since(t0).String(),
	)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal: file loading
// ─────────────────────────────────────────────────────────────────────────────

// loadMigrations reads all embedded *.up.sql / *.down.sql files, pairs them by
// version, checksums the up content, and returns them sorted ascending.
func loadMigrations() ([]migration, error) {

	entries, err := fs.ReadDir(migrationsFS, "sql")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations dir: %w", err)
	}

	type pair struct {
		name string
		up   string
		down string
	}
	byVersion := make(map[int]*pair)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		ver, name, dir, err := parseFilename(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("invalid migration filename %q: %w", entry.Name(), err)
		}

		raw, err := fs.ReadFile(migrationsFS, filepath.Join("sql", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", entry.Name(), err)
		}

		if _, ok := byVersion[ver]; !ok {
			byVersion[ver] = &pair{name: name}
		}
		switch dir {
		case "up":
			byVersion[ver].up = string(raw)
		case "down":
			byVersion[ver].down = string(raw)
		}
	}

	result := make([]migration, 0, len(byVersion))
	for ver, p := range byVersion {
		if p.up == "" {
			return nil, fmt.Errorf("version %d: missing .up.sql file", ver)
		}
		if p.down == "" {
			return nil, fmt.Errorf("version %d: missing .down.sql file", ver)
		}
		h := sha256.Sum256([]byte(p.up))
		result = append(result, migration{
			Version:  ver,
			Name:     p.name,
			UpSQL:    p.up,
			DownSQL:  p.down,
			Checksum: fmt.Sprintf("%x", h),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	// Guard: no duplicate versions.
	for i := 1; i < len(result); i++ {
		if result[i].Version == result[i-1].Version {
			return nil, fmt.Errorf("duplicate migration version %d", result[i].Version)
		}
	}

	return result, nil
}

// parseFilename parses a filename such as "000002_create_auth_tables.up.sql"
// into (version=2, name="create_auth_tables", direction="up", nil).
func parseFilename(fname string) (version int, name, direction string, err error) {
	base := strings.TrimSuffix(fname, ".sql")

	dot := strings.LastIndex(base, ".")
	if dot < 0 {
		return 0, "", "", fmt.Errorf("missing .up/.down suffix")
	}
	direction = base[dot+1:]
	if direction != "up" && direction != "down" {
		return 0, "", "", fmt.Errorf("unknown direction %q (expected up or down)", direction)
	}
	base = base[:dot]

	under := strings.IndexByte(base, '_')
	if under < 0 {
		return 0, "", "", fmt.Errorf("missing underscore between version number and name")
	}
	v64, parseErr := strconv.ParseInt(base[:under], 10, 64)
	if parseErr != nil {
		return 0, "", "", fmt.Errorf("parse version number: %w", parseErr)
	}
	return int(v64), base[under+1:], direction, nil
}

// verifyChecksums checks that every applied migration's recorded checksum
// matches the current embedded file. Returns an error on the first mismatch.
// Applied versions with no corresponding file emit a warning but are not fatal.
func verifyChecksums(migrations []migration, applied []migrationRecord) error {
	byVersion := make(map[int]migration, len(migrations))
	for _, mg := range migrations {
		byVersion[mg.Version] = mg
	}

	for _, rec := range applied {
		mg, ok := byVersion[rec.Version]
		if !ok {
			slog.Warn("migrate: applied version has no embedded SQL file — skipping checksum check",
				"version", rec.Version,
				"name", rec.Name,
			)
			continue
		}
		if mg.Checksum != rec.Checksum {
			return fmt.Errorf(
				"checksum mismatch for version %d (%s): "+
					"file=%s recorded=%s — do not modify applied migration files",
				rec.Version, rec.Name, mg.Checksum, rec.Checksum,
			)
		}
	}
	return nil
}
