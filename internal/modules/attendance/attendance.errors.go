package attendance

import (
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/thalalhassan/edu_management/internal/apperrors"
	"gorm.io/gorm"
)

func isUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key") || strings.Contains(strings.ToLower(err.Error()), "unique")
}

func mapCreateError(err error, context string) error {
	if err == nil {
		return nil
	}
	if isUniqueConstraintError(err) {
		return apperrors.New(apperrors.ErrAlreadyExists, "attendance already exists")
	}
	return apperrors.New(apperrors.ErrDatabase, context)
}

func mapGetError(err error, context string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound
	}
	return apperrors.New(apperrors.ErrDatabase, context)
}
