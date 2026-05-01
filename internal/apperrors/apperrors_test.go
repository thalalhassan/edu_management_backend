package apperrors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/thalalhassan/edu_management/internal/apperrors"
)

// ---------------------------------------------------------------------------
// Sentinel identity
// ---------------------------------------------------------------------------

func TestSentinelIdentity(t *testing.T) {
	// errors.Is must match through fmt.Errorf %w wrapping
	wrapped := fmt.Errorf("repo.GetStudent: %w", apperrors.ErrNotFound)
	if !errors.Is(wrapped, apperrors.ErrNotFound) {
		t.Fatal("expected errors.Is to match ErrNotFound through wrapping")
	}
}

// ---------------------------------------------------------------------------
// AppError
// ---------------------------------------------------------------------------

func TestAppError_ErrorString(t *testing.T) {
	e := apperrors.New(apperrors.ErrNotFound, "student with id 42 not found")
	want := "resource not found: student with id 42 not found"
	if e.Error() != want {
		t.Fatalf("got %q, want %q", e.Error(), want)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	e := apperrors.New(apperrors.ErrUnauthorized, "role instructor required")
	if !errors.Is(e, apperrors.ErrUnauthorized) {
		t.Fatal("errors.Is should match sentinel through AppError.Unwrap")
	}
}

func TestAppError_Newf(t *testing.T) {
	e := apperrors.Newf(apperrors.ErrNotFound, "course %d not found", 99)
	if !errors.Is(e, apperrors.ErrNotFound) {
		t.Fatal("Newf should unwrap to sentinel")
	}
}

func TestAppError_NewWithMeta(t *testing.T) {
	meta := map[string]any{"student_id": "abc123", "course_id": 7}
	e := apperrors.NewWithMeta(apperrors.ErrAlreadyEnrolled, "already enrolled", meta)
	if !errors.Is(e, apperrors.ErrAlreadyEnrolled) {
		t.Fatal("errors.Is should match ErrAlreadyEnrolled")
	}
	var ae *apperrors.AppError
	if !errors.As(e, &ae) {
		t.Fatal("errors.As should extract *AppError")
	}
	if ae.Meta["student_id"] != "abc123" {
		t.Fatalf("unexpected meta: %v", ae.Meta)
	}
}

// ---------------------------------------------------------------------------
// ValidationError
// ---------------------------------------------------------------------------

func TestValidationError_OrNil_NoErrors(t *testing.T) {
	ve := apperrors.NewValidationError()
	if ve.OrNil() != nil {
		t.Fatal("OrNil should return nil when no fields added")
	}
}

func TestValidationError_OrNil_WithErrors(t *testing.T) {
	ve := apperrors.NewValidationError()
	ve.Add("email", "invalid format")
	ve.Add("dob", "must be in the past")

	err := ve.OrNil()
	if err == nil {
		t.Fatal("OrNil should return error when fields present")
	}
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatal("ValidationError should unwrap to ErrValidation")
	}
	var ve2 *apperrors.ValidationError
	if !errors.As(err, &ve2) {
		t.Fatal("errors.As should extract *ValidationError")
	}
	if len(ve2.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(ve2.Fields))
	}
}

// ---------------------------------------------------------------------------
// HTTPStatus mapping
// ---------------------------------------------------------------------------

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{nil, http.StatusOK},
		{apperrors.ErrNotFound, http.StatusNotFound},
		{apperrors.ErrAlreadyExists, http.StatusConflict},
		{apperrors.ErrAlreadyEnrolled, http.StatusConflict},
		{apperrors.ErrCourseFull, http.StatusConflict},
		{apperrors.ErrInvalidCredentials, http.StatusUnauthorized},
		{apperrors.ErrTokenExpired, http.StatusUnauthorized},
		{apperrors.ErrUnauthorized, http.StatusForbidden},
		{apperrors.ErrAccountDisabled, http.StatusForbidden},
		{apperrors.ErrValidation, http.StatusUnprocessableEntity},
		{apperrors.ErrMissingField, http.StatusBadRequest},
		{apperrors.ErrEnrollmentClosed, http.StatusForbidden},
		{apperrors.ErrRateLimited, http.StatusTooManyRequests},
		{apperrors.ErrFileTooLarge, http.StatusRequestEntityTooLarge},
		{apperrors.ErrTimeout, http.StatusGatewayTimeout},
		{apperrors.ErrDatabase, http.StatusServiceUnavailable},
		{apperrors.ErrInternal, http.StatusInternalServerError},
		// Unknown error defaults to 500
		{errors.New("some unknown error"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		got := apperrors.HTTPStatus(tc.err)
		if got != tc.want {
			t.Errorf("HTTPStatus(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}

func TestHTTPStatus_ThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("service.Enroll: %w", apperrors.ErrCourseFull)
	if apperrors.HTTPStatus(wrapped) != http.StatusConflict {
		t.Fatal("HTTPStatus should resolve through %w wrapping")
	}
}

func TestIsClientError(t *testing.T) {
	if !apperrors.IsClientError(apperrors.ErrNotFound) {
		t.Fatal("ErrNotFound is a client error")
	}
	if apperrors.IsClientError(apperrors.ErrInternal) {
		t.Fatal("ErrInternal is not a client error")
	}
}

func TestIsServerError(t *testing.T) {
	if !apperrors.IsServerError(apperrors.ErrDatabase) {
		t.Fatal("ErrDatabase is a server error")
	}
	if apperrors.IsServerError(apperrors.ErrUnauthorized) {
		t.Fatal("ErrUnauthorized is not a server error")
	}
}
