package render

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/buni/wallet/internal/pkg/sloglog"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jinzhu/copier"
)

// ErrorResponse ...
type ErrorResponse struct {
	Error *Error `json:"error"`
}

// NewSuccessHandlerResponse ...
func NewSuccessResponse[Response any](ctx context.Context, w http.ResponseWriter, status int, response Response) {
	write(ctx, w, status, response)
}

// NewErrorResponse ...
func NewErrorResponse(ctx context.Context, w http.ResponseWriter, code int, status string, err error) {
	responseError := NewError(status, err.Error())

	errors.As(err, &responseError)

	write(
		ctx,
		w,
		code,
		ErrorResponse{Error: responseError},
	)
}

// NewSuccessOKResponse ...
func NewSuccessOKResponse[Response any](ctx context.Context, w http.ResponseWriter, response Response) {
	NewSuccessResponse(ctx, w, http.StatusOK, response)
}

// NewSuccessCreatedResponse ...
func NewSuccessCreatedResponse[Response any](ctx context.Context, w http.ResponseWriter, response Response) {
	NewSuccessResponse(ctx, w, http.StatusCreated, response)
}

// NewInternalServerErrorResponse ...
func NewInternalServerErrorResponse(ctx context.Context, w http.ResponseWriter, _ error) {
	NewErrorResponse(ctx, w, http.StatusInternalServerError, InternalServerError, errors.New("internal server error")) //nolint:goerr113
}

// NewUnauthorizedErrorResponse ...
func NewUnauthorizedErrorResponse(ctx context.Context, w http.ResponseWriter, _ error) {
	NewErrorResponse(ctx, w, http.StatusUnauthorized, UnauthorizedError, errors.New("unauthorized")) //nolint:goerr113
}

// NewNotFoundErrorResponse ...
func NewNotFoundErrorResponse(ctx context.Context, w http.ResponseWriter, _ error) {
	NewErrorResponse(ctx, w, http.StatusNotFound, NotFoundError, errors.New("not found")) //nolint:goerr113
}

// NewNotFoundErrorResponse ...
func NewBadRequestErrorResponse(ctx context.Context, w http.ResponseWriter, _ error) {
	NewErrorResponse(ctx, w, http.StatusBadRequest, BadRequestError, errors.New("bad request")) //nolint:goerr113
}

func NewConflictErrorResponse(ctx context.Context, w http.ResponseWriter, _ error) {
	NewErrorResponse(ctx, w, http.StatusConflict, ConflictError, errors.New("conflict with existing resource")) //nolint:goerr113
}

func NewValidationErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	NewErrorResponse(ctx, w, http.StatusBadRequest, RequestValidationError, err) //nolint:goerr113
}

func NewMethodNotAllowedErrorResponse(ctx context.Context, w http.ResponseWriter, _ error) {
	NewErrorResponse(ctx, w, http.StatusMethodNotAllowed, MethodNotAllowedError, errors.New("method not allowed")) //nolint:goerr113
}

func write[Response any](ctx context.Context, w http.ResponseWriter, status int, v Response) {
	log := sloglog.FromContext(ctx)

	ww, ok := w.(middleware.WrapResponseWriter)
	if !ok {
		ww = middleware.NewWrapResponseWriter(w, 1)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if ww.BytesWritten() != 0 {
		log.DebugContext(ctx, "response already written, in renderer")
		return
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	err := enc.Encode(v)
	if err != nil {
		log.ErrorContext(ctx, "error encoding response, in renderer", slog.Any("error", err))
		http.Error(ww, InternalServerError, http.StatusInternalServerError)
		return
	}

	if ww.Status() == 0 {
		ww.WriteHeader(status)
	} else {
		log.DebugContext(ctx, "status resp already written, in renderer")
	}

	_, err = ww.Write(buf.Bytes())
	if err != nil {
		log.ErrorContext(ctx, "error writing response, in renderer", slog.Any("error", err))
		return
	}
}

func NewResponse[Response any](entity any) (*Response, error) {
	resp := new(Response)

	err := copier.Copy(resp, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to copy entity to response1: %w", err)
	}
	return resp, nil
}

func NewResponses[Entity any, Response any](entity []Entity) (*[]Response, error) {
	resp := make([]Response, 0, len(entity))
	for _, e := range entity {
		e := e
		res, err := NewResponse[Response](&e)
		if err != nil {
			return nil, fmt.Errorf("failed to copy entity to response2: %w", err)
		}

		resp = append(resp, *res)
	}

	if len(resp) == 0 {
		return &[]Response{}, nil
	}

	return &resp, nil
}
