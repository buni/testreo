package errorhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/buni/wallet/internal/pkg/render"
	"github.com/buni/wallet/internal/pkg/syncx"
)

var errorHandlers = &syncx.SyncMap[string, HandleFunc]{} //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	RegisterErrorHandler("validation_error_handler", ValidationErrorHandler)
	RegisterErrorHandler("validation_field_errors_handler", ValidationFieldErrorsHandler)
	RegisterErrorHandler("validation_field_error_handler", ValidationFieldErrorHandler)
}

// RegisterErrorHandler registers an error handler, the error handlers are called in the order they are registered.
func RegisterErrorHandler(name string, errorHandler HandleFunc) {
	errorHandlers.Store(name, errorHandler)
}

// HandleFunc is a function that handles an error, if it returns true, the error is considered handled.
type HandleFunc func(ctx context.Context, w http.ResponseWriter, err error) bool

// ErrorHandler ...
type ErrorHandler struct {
	errorHandlers []HandleFunc
}

// NewErrorHandler ...
func NewErrorHandler(errorHandlers ...HandleFunc) *ErrorHandler {
	return &ErrorHandler{
		errorHandlers: errorHandlers,
	}
}

// NewErrorResponse write an error response, if an error handler is found for the error, it will be handled by it,
// otherwise a generic internal server error will be returned, this is done to prevent leaking internal errors.
func (e *ErrorHandler) NewErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	for _, errorHandler := range e.errorHandlers {
		if errorHandler(ctx, w, err) {
			return
		}
	}

	render.NewInternalServerErrorResponse(ctx, w, err)
}

// NewDefaultErrorResponse writes an error response, and uses the globally registered error handlers.
func NewDefaultErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	eh := NewErrorHandler(
		errorHandlers.Values()...,
	)

	eh.NewErrorResponse(ctx, w, err)
}

// ValidationErrorHandler handles validation errors, if err is of type render.Error it will be handled as a validation error.
func ValidationErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	var validationError *render.Error
	if errors.As(err, &validationError) {
		render.NewValidationErrorResponse(ctx, w, validationError)
		return true
	}
	return false
}

// ValidationFieldErrorsHandler handles FieldErrors errors, if err is of type render.FieldErrors it will be handled as a validation error.
func ValidationFieldErrorsHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	var fieldErrors *render.FieldErrors
	if errors.As(err, &fieldErrors) {
		validationError := render.NewValidationError(*fieldErrors...)
		render.NewValidationErrorResponse(ctx, w, validationError)
		return true
	}
	return false
}

// ValidationFieldErrorsHandler handles FieldError errors, if err is of type render.FieldError it will be handled as a validation error.
func ValidationFieldErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	var fieldError *render.FieldError
	if errors.As(err, &fieldError) {
		validationError := render.NewValidationError(fieldError)
		render.NewValidationErrorResponse(ctx, w, validationError)
		return true
	}
	return false
}
