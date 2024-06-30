package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/buni/wallet/internal/pkg/render"
	"github.com/buni/wallet/internal/pkg/render/errorhandler"
	"github.com/buni/wallet/internal/pkg/requestdecoder"
	"github.com/buni/wallet/internal/pkg/requestvalidator"
	"github.com/buni/wallet/internal/pkg/sloglog"
	"github.com/go-chi/chi/v5/middleware"
)

var ErrValidationMiddlewareNilRequestBody = errors.New("validation middleware expects non nil request body")

// ValidationMiddleware validates the request body body after it has been decoded.
// Uses requestvalidator.PlaygroundValidator to validate the request body.
func ValidationMiddleware[Request, Response any](validator *requestvalidator.PlaygroundValidator) MiddlewareFunc[Request, Response] {
	return func(handler HandlerFunc[Request, Response]) HandlerFunc[Request, Response] {
		return func(w http.ResponseWriter, r *http.Request, reqBody *Request) (*Response, error) {
			if reqBody == nil {
				return nil, ErrValidationMiddlewareNilRequestBody
			}

			err := validator.Validate(r.Context(), reqBody)
			if err != nil {
				return nil, fmt.Errorf("request validation error: %w", err)
			}
			return handler(w, r, reqBody)
		}
	}
}

// RequestDecoderMiddleware decodes the request body into the request body type.
func RequestDecoderMiddleware[Request, Response any]() MiddlewareFunc[Request, Response] {
	return func(handler HandlerFunc[Request, Response]) HandlerFunc[Request, Response] {
		return func(w http.ResponseWriter, r *http.Request, reqBody *Request) (*Response, error) {
			reqBody, err := requestdecoder.Decode[Request](r, reqBody)
			if err != nil {
				return nil, fmt.Errorf("failed to decode request: %w", err)
			}
			return handler(w, r, reqBody)
		}
	}
}

// ResponseHandlerMiddleware handles the response written by the handler.
// It handles errors and empty responses.
// It allows for the handler to set the status code or write the response body.
func ResponseHandlerMiddleware[Request, Response any]() MiddlewareFunc[Request, Response] {
	return func(handler HandlerFunc[Request, Response]) HandlerFunc[Request, Response] {
		return func(w http.ResponseWriter, r *http.Request, reqBody *Request) (*Response, error) {
			ctx := r.Context()
			log := sloglog.FromContext(ctx)

			ww, ok := w.(middleware.WrapResponseWriter)
			if !ok {
				ww = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			}

			resp, err := handler(ww, r, reqBody)
			if err != nil {
				log.ErrorContext(ctx, "handler execution error", slog.Any("error", err))
				errorhandler.NewDefaultErrorResponse(ctx, ww, err)
				return nil, fmt.Errorf("failed to execute handler: %w", err)
			}

			if resp == nil { // check for empty response
				if ww.Status() == 0 { // no status was set, set it to 204
					log.DebugContext(ctx, "empty response, setting status to 204")
					ww.WriteHeader(http.StatusNoContent)
					return resp, nil
				}
				log.DebugContext(ctx, "empty response, status was already set by the handler")
				return resp, nil
			}

			render.NewSuccessOKResponse(ctx, ww, resp)
			return resp, nil
		}
	}
}
