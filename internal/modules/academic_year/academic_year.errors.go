package academic_year

import "github.com/google/uuid"

type ValidationError struct {
	Message string
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{Message: msg}
}

func (e *ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
	ID       uuid.UUID
}

func NewNotFoundError(resource string, id uuid.UUID) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

func (e *NotFoundError) Error() string {
	if e.ID == uuid.Nil {
		return e.Resource + " not found"
	}
	return e.Resource + " not found: " + e.ID.String()
}

type BusinessError struct {
	Message string
}

func NewBusinessError(msg string) *BusinessError {
	return &BusinessError{Message: msg}
}

func (e *BusinessError) Error() string {
	return e.Message
}
