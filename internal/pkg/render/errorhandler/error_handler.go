package errorhandler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/buni/wallet/internal/api/app/entity"
	"github.com/buni/wallet/internal/pkg/render"
	"github.com/buni/wallet/internal/pkg/syncx"
)

var errorHandlers = &syncx.SyncMap[string, HandleFunc]{} //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	RegisterErrorHandler("validation_error_handler", ValidationErrorHandler)
	RegisterErrorHandler("validation_field_errors_handler", ValidationFieldErrorsHandler)
	RegisterErrorHandler("validation_field_error_handler", ValidationFieldErrorHandler)
	RegisterErrorHandler("not_found_error_handler", NotFoundErrorHandler)
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

func NotFoundErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	if errors.Is(err, entity.ErrEntityNotFound) {
		render.NewNotFoundErrorResponse(ctx, w, err)
		return true
	}
	return false
}

func ConflictErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	if strings.Contains(err.Error(), "unique constraint") {
		render.NewErrorResponse(ctx, w, http.StatusConflict, render.ConflictError, errors.New("entity already exists")) //nolint:goerr113
		return true
	}
	return false
}

func ValidationFieldErrorsHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	var fieldErrors *render.FieldErrors
	if errors.As(err, &fieldErrors) {
		validationError := render.NewValidationError(*fieldErrors...)
		render.NewValidationErrorResponse(ctx, w, validationError)
		return true
	}
	return false
}

func ValidationFieldErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	var fieldError *render.FieldError
	if errors.As(err, &fieldError) {
		validationError := render.NewValidationError(fieldError)
		render.NewValidationErrorResponse(ctx, w, validationError)
		return true
	}
	return false
}
