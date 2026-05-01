package seed

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/crypto/bcrypt"
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

// seedResource describes a domain resource exposed through the permission table.
var seedResources = []string{
	"user", "student", "teacher", "staff", "parent",
	"department", "standard", "subject", "class_section",
	"academic_year", "exam", "attendance", "fee", "notice", "salary",
}

// seedActions lists every action that can be granted on a resource.
var seedActions = []string{"CREATE", "READ", "UPDATE", "DELETE", "MANAGE"}

// permissionRow is a minimal struct used solely for the seed upsert.
type permissionRow struct {
	ID       string `gorm:"column:id;primaryKey"`
	Resource string `gorm:"column:resource"`
	Action   string `gorm:"column:action"`
}

func (permissionRow) TableName() string { return "permission" }

// adminUserRow is used only for the seed INSERT.
type adminUserRow struct {
	ID           string `gorm:"column:id;primaryKey"`
	Email        string `gorm:"column:email"`
	PasswordHash string `gorm:"column:password_hash"`
	Role         string `gorm:"column:role"`
	IsActive     bool   `gorm:"column:is_active"`
}

func (adminUserRow) TableName() string { return "user" }

// Seed populates baseline reference data required for the application to
// function. It is safe to call multiple times; all inserts use
// ON CONFLICT DO NOTHING, so existing rows are never overwritten.
//
// Required environment variables:
//
//	SEED_ADMIN_EMAIL    — email address for the super-admin account
//	SEED_ADMIN_PASSWORD — plain-text password (bcrypt-hashed before storage)
func SeedV2(ctx context.Context, db *gorm.DB) error {
	slog.InfoContext(ctx, "seed: starting")

	if err := seedPermissions(ctx, db); err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}

	if err := seedSuperAdmin(ctx, db); err != nil {
		return fmt.Errorf("seed super admin: %w", err)
	}

	slog.InfoContext(ctx, "seed: complete")
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Permissions
// ─────────────────────────────────────────────────────────────────────────────

// seedPermissions inserts one row per (resource, action) pair derived from
// seedResources × seedActions. The composite unique index on the permission
// table (resource, action) WHERE deleted_at IS NULL guarantees idempotency;
// the ON CONFLICT clause makes that explicit.
func seedPermissions(ctx context.Context, db *gorm.DB) error {
	total := len(seedResources) * len(seedActions)
	slog.InfoContext(ctx, "seed: inserting permissions", "count", total)

	// Build the insert in a single batch for efficiency.
	sql := `
INSERT INTO permission (id, resource, action, description)
SELECT
    gen_random_uuid(),
    r.resource,
    a.action,
    r.resource || ':' || a.action
FROM (VALUES %s) AS r(resource)
CROSS JOIN (VALUES %s) AS a(action)
ON CONFLICT (resource, action)
WHERE deleted_at IS NULL
DO NOTHING`

	resourcePlaceholders := buildValuesList(seedResources)
	actionPlaceholders := buildValuesList(seedActions)

	query := fmt.Sprintf(sql, resourcePlaceholders, actionPlaceholders)
	if err := db.WithContext(ctx).Exec(query).Error; err != nil {
		return err
	}

	return nil
}

// buildValuesList converts []string{"a","b"} to "('a'),('b')" for use in
// a SQL VALUES clause.
func buildValuesList(items []string) string {
	out := make([]byte, 0, len(items)*8)
	for i, s := range items {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, '(', '\'')
		out = append(out, []byte(s)...)
		out = append(out, '\'', ')')
	}
	return string(out)
}

// ─────────────────────────────────────────────────────────────────────────────
// Super-admin user
// ─────────────────────────────────────────────────────────────────────────────

// seedSuperAdmin creates the initial SUPER_ADMIN user if one does not already
// exist with the given email. The password is bcrypt-hashed at cost 12.
func seedSuperAdmin(ctx context.Context, db *gorm.DB) error {
	email := os.Getenv("SEED_ADMIN_EMAIL")
	if email == "" {
		slog.WarnContext(ctx, "seed: SEED_ADMIN_EMAIL not set — skipping super-admin seed")
		return nil
	}

	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if password == "" {
		return fmt.Errorf("SEED_ADMIN_EMAIL is set but SEED_ADMIN_PASSWORD is empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("bcrypt password: %w", err)
	}

	sql := `
INSERT INTO "user" (id, email, password_hash, role, is_active)
VALUES (gen_random_uuid(), ?, ?, 'SUPER_ADMIN', true)
ON CONFLICT (email)
WHERE deleted_at IS NULL
DO NOTHING`

	if err := db.WithContext(ctx).Exec(sql, email, string(hashed)).Error; err != nil {
		return err
	}

	slog.InfoContext(ctx, "seed: super-admin user upserted", "email", email)
	return nil
}
