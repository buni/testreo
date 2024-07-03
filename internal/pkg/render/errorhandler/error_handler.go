package errorhandler

import (
	"context"
	"errors"
	"fmt"
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

func NegativeAmountErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	if errors.Is(err, entity.ErrNegativeAmount) {
		render.NewValidationErrorResponse(ctx, w, render.NewValidationError(&render.FieldError{
			Field:   "amount",
			Message: "amount must not be negative",
		}))
		return true
	}
	return false
}

func InsufficientBalanceErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	if errors.Is(err, entity.ErrInsufficientBalance) {
		render.NewValidationErrorResponse(ctx, w, render.NewValidationError(&render.FieldError{
			Field:   "balance",
			Message: "wallet balance is insufficient for this operation",
		}))
		return true
	}
	return false
}

func ConflictErrorHandler(ctx context.Context, w http.ResponseWriter, err error) bool {
	errStr := err.Error()
	if strings.Contains(errStr, "unique constraint") {
		field := "reference_id"
		if strings.Contains(errStr, "transfer_id") {
			field = "transfer_id"
		}

		render.NewValidationErrorResponse(ctx, w, render.NewValidationError(&render.FieldError{
			Field:   field,
			Message: fmt.Sprintf("value already used for different %s", strings.Trim(field, "_id")),
		}))
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
