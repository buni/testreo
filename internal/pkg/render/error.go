package render

import (
	"fmt"
)

const (
	// InternalServerError ...
	InternalServerError string = "internal_server_error"
	// NotFoundError ...
	NotFoundError string = "not_found_error"
	// BadRequestError ...
	BadRequestError string = "bad_request_error"
	// UnauthorizedError ...
	UnauthorizedError string = "unauthorized_error"
	// RequestValidationError ...
	RequestValidationError string = "validation_error"
	// ConflictError ...
	ConflictError string = "conflict_error"
	// MethodNotAllowedError ...
	MethodNotAllowedError string = "method_not_allowed_error"
)

// Error ...
type Error struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Errors  *FieldErrors `json:"errors"`
}

type FieldErrors []*FieldError

func (e *FieldErrors) Error() string {
	return "validation error"
}

// Add appends a new error to the list of field errors.
func (e *FieldErrors) Add(field, message string, err error) {
	*e = append(*e, &FieldError{
		Field:   field,
		Message: message,
		err:     err,
	})
}

// ValidationError ...
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	err     error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *FieldError) Unwrap() error {
	return e.err
}

// NewError ...
func NewError(status, message string, errs ...*FieldError) *Error {
	return &Error{
		Status:  status,
		Message: message,
		Errors:  (*FieldErrors)(&errs),
	}
}

func NewValidationError(errs ...*FieldError) *Error {
	errsPtr := FieldErrors(errs)
	return &Error{
		Status:  RequestValidationError,
		Message: "validation error",
		Errors:  &errsPtr,
	}
}

func (e *Error) Error() string {
	return e.Message
}
