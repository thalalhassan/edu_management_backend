// Usage pattern:
//
//	// repository layer
//	return nil, fmt.Errorf("GetStudentByID: %w", apperrors.ErrNotFound)
//
//	// service layer - check and re-wrap if needed
//	if errors.Is(err, apperrors.ErrNotFound) { ... }
//
//	// handler layer
//	code := apperrors.HTTPStatus(err)
package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// ---------------------------------------------------------------------------
// Sentinel errors
// ---------------------------------------------------------------------------

// Generic / infrastructure
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned on duplicate create attempts.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrConflict is returned when a state transition is invalid
	// (e.g. enrolling a student in a course that is already full).
	ErrConflict = errors.New("operation conflicts with current state")

	// ErrInternal is returned for unexpected server-side failures.
	ErrInternal = errors.New("internal server error")

	// ErrDatabase is returned when a persistence operation fails.
	ErrDatabase = errors.New("database error")

	// ErrTimeout is returned when an operation exceeds its deadline.
	ErrTimeout = errors.New("operation timed out")
)

// Authentication & authorisation
var (
	// ErrInvalidCredentials is returned when login credentials are wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUnauthenticated is returned when no valid session/token is present.
	ErrUnauthenticated = errors.New("unauthenticated")

	// ErrUnauthorized is returned when the caller lacks permission.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTokenExpired is returned when a JWT/session token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenInvalid is returned when a token cannot be parsed or verified.
	ErrTokenInvalid = errors.New("token invalid")

	// ErrAccountDisabled is returned when the account is suspended/deactivated.
	ErrAccountDisabled = errors.New("account disabled")
)

// Input validation
var (
	// ErrValidation is returned when request data fails validation rules.
	ErrValidation = errors.New("validation error")

	// ErrMissingField is returned when a required field is absent.
	ErrMissingField = errors.New("required field missing")

	// ErrInvalidFormat is returned when a field value has an unexpected format.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrOutOfRange is returned when a numeric value is outside allowed bounds.
	ErrOutOfRange = errors.New("value out of range")
)

// Education domain
var (
	// ErrEnrollmentClosed is returned when a course is not accepting enrolments.
	ErrEnrollmentClosed = errors.New("enrollment is closed")

	// ErrCourseFull is returned when a course has reached maximum capacity.
	ErrCourseFull = errors.New("course is full")

	// ErrAlreadyEnrolled is returned when a student is already enrolled.
	ErrAlreadyEnrolled = errors.New("student already enrolled")

	// ErrNotEnrolled is returned when an operation requires active enrolment.
	ErrNotEnrolled = errors.New("student not enrolled")

	// ErrPrerequisiteNotMet is returned when prerequisite courses are incomplete.
	ErrPrerequisiteNotMet = errors.New("prerequisite not met")

	// ErrGradeAlreadySubmitted is returned on duplicate grade submission.
	ErrGradeAlreadySubmitted = errors.New("grade already submitted")

	// ErrExamNotActive is returned when an exam is not in an active window.
	ErrExamNotActive = errors.New("exam is not active")

	// ErrSubmissionDeadlinePassed is returned when a submission window has closed.
	ErrSubmissionDeadlinePassed = errors.New("submission deadline has passed")

	// ErrInvalidAcademicYear is returned when the academic year reference is invalid.
	ErrInvalidAcademicYear = errors.New("invalid academic year")

	// ErrFeeUnpaid is returned when an action is blocked by an outstanding fee.
	ErrFeeUnpaid = errors.New("outstanding fee must be paid")
)

// File / media
var (
	// ErrFileTooLarge is returned when an upload exceeds the size limit.
	ErrFileTooLarge = errors.New("file exceeds maximum allowed size")

	// ErrUnsupportedFileType is returned when the MIME type is not allowed.
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

// Rate limiting / quota
var (
	// ErrRateLimited is returned when the caller has exceeded request limits.
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrQuotaExceeded is returned when a business quota is breached
	// (e.g. max courses per instructor).
	ErrQuotaExceeded = errors.New("quota exceeded")
)

// ---------------------------------------------------------------------------
// Typed error – carries structured context alongside a sentinel
// ---------------------------------------------------------------------------

// AppError wraps a sentinel with human-readable detail and optional metadata.
// Use this when callers need both errors.Is matching and diagnostic context.
//
//	return apperrors.New(apperrors.ErrValidation, "email must be a valid address")
//	return apperrors.NewWithMeta(apperrors.ErrNotFound, "student not found", map[string]any{"id": id})
type AppError struct {
	// Sentinel is the canonical sentinel that errors.Is will match.
	Sentinel error
	// Detail is a human-readable explanation safe to surface in API responses.
	Detail string
	// Meta is optional structured data (IDs, field names, etc.) for logging.
	Meta map[string]any
}

func (e *AppError) Error() string {
	if e.Detail == "" {
		return e.Sentinel.Error()
	}
	return fmt.Sprintf("%s: %s", e.Sentinel.Error(), e.Detail)
}

// Unwrap makes errors.Is / errors.As traverse to the sentinel.
func (e *AppError) Unwrap() error { return e.Sentinel }

// New returns an AppError wrapping sentinel with a detail message.
func New(sentinel error, detail string) *AppError {
	return &AppError{Sentinel: sentinel, Detail: detail}
}

// NewWithMeta returns an AppError with structured metadata attached.
func NewWithMeta(sentinel error, detail string, meta map[string]any) *AppError {
	return &AppError{Sentinel: sentinel, Detail: detail, Meta: meta}
}

// Newf returns an AppError with a formatted detail message.
func Newf(sentinel error, format string, args ...any) *AppError {
	return &AppError{Sentinel: sentinel, Detail: fmt.Sprintf(format, args...)}
}

// ---------------------------------------------------------------------------
// Validation error – aggregates multiple field errors
// ---------------------------------------------------------------------------

// FieldError represents a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError aggregates one or more FieldErrors.
// It always unwraps to ErrValidation.
//
//	ve := apperrors.NewValidationError()
//	ve.Add("email", "must be a valid address")
//	ve.Add("dob", "must be in the past")
//	return ve.OrNil()
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %d field(s) failed", ErrValidation.Error(), len(e.Fields))
}

func (e *ValidationError) Unwrap() error { return ErrValidation }

// Add appends a field-level error.
func (e *ValidationError) Add(field, message string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Message: message})
}

// HasErrors reports whether any field errors were collected.
func (e *ValidationError) HasErrors() bool { return len(e.Fields) > 0 }

// OrNil returns the ValidationError if it has errors, otherwise nil.
// Useful at the end of a validation block to avoid returning an empty error.
func (e *ValidationError) OrNil() error {
	if e.HasErrors() {
		return e
	}
	return nil
}

// NewValidationError creates an empty ValidationError ready to collect fields.
func NewValidationError() *ValidationError {
	return &ValidationError{}
}

// ---------------------------------------------------------------------------
// HTTP status mapping
// ---------------------------------------------------------------------------

// httpStatusMap maps sentinel errors to HTTP status codes.
// Order matters for Is-chain matching: more specific sentinels first.
var httpStatusMap = []struct {
	sentinel error
	code     int
}{
	{ErrNotFound, http.StatusNotFound},
	{ErrAlreadyExists, http.StatusConflict},
	{ErrAlreadyEnrolled, http.StatusConflict},
	{ErrCourseFull, http.StatusConflict},
	{ErrConflict, http.StatusConflict},
	{ErrGradeAlreadySubmitted, http.StatusConflict},

	{ErrInvalidCredentials, http.StatusUnauthorized},
	{ErrUnauthenticated, http.StatusUnauthorized},
	{ErrTokenExpired, http.StatusUnauthorized},
	{ErrTokenInvalid, http.StatusUnauthorized},
	{ErrAccountDisabled, http.StatusForbidden},
	{ErrUnauthorized, http.StatusForbidden},

	{ErrValidation, http.StatusUnprocessableEntity},
	{ErrMissingField, http.StatusBadRequest},
	{ErrInvalidFormat, http.StatusBadRequest},
	{ErrOutOfRange, http.StatusBadRequest},

	{ErrEnrollmentClosed, http.StatusForbidden},
	{ErrPrerequisiteNotMet, http.StatusForbidden},
	{ErrSubmissionDeadlinePassed, http.StatusForbidden},
	{ErrFeeUnpaid, http.StatusForbidden},
	{ErrExamNotActive, http.StatusForbidden},

	{ErrRateLimited, http.StatusTooManyRequests},
	{ErrQuotaExceeded, http.StatusForbidden},

	{ErrFileTooLarge, http.StatusRequestEntityTooLarge},
	{ErrUnsupportedFileType, http.StatusUnsupportedMediaType},

	{ErrTimeout, http.StatusGatewayTimeout},
	{ErrDatabase, http.StatusServiceUnavailable},
	{ErrInternal, http.StatusInternalServerError},
}

// HTTPStatus returns the HTTP status code for err by walking its unwrap chain.
// Defaults to 500 if no mapping is found.
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	for _, m := range httpStatusMap {
		if errors.Is(err, m.sentinel) {
			return m.code
		}
	}
	return http.StatusInternalServerError
}

// IsClientError reports whether err maps to a 4xx status code.
func IsClientError(err error) bool {
	code := HTTPStatus(err)
	return code >= 400 && code < 500
}

// IsServerError reports whether err maps to a 5xx status code.
func IsServerError(err error) bool {
	return HTTPStatus(err) >= 500
}
