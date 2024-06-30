package handler

import (
	"context"
	"net/http"

	"github.com/buni/wallet/internal/pkg/requestvalidator"
)

// MiddlewareFunc middleware func signature for HandlerFunc.
type MiddlewareFunc[Request, Response any] func(HandlerFunc[Request, Response]) HandlerFunc[Request, Response]

// HandlerFunc extends the http.HandlerFunc type to include a request body and response.
// The request body is decoded decoded into Request and the Response is automatically marshaled.
// Responses are typically handled by the ResponseHandlerMiddleware.
type HandlerFunc[Request, Response any] func(w http.ResponseWriter, req *http.Request, reqBody *Request) (*Response, error) //nolint

type BasicHandlerFunc[Request, Response any] func(ctx context.Context, reqBody *Request) (*Response, error) //nolint

// Wrap wraps a HandlerFunc with middleware, returning a http.HandlerFunc.
// Middleware are applied front to back, so the first middleware in the list will be executed last.
func Wrap[Request, Response any](handler HandlerFunc[Request, Response], middleware ...MiddlewareFunc[Request, Response]) http.HandlerFunc {
	for _, m := range middleware {
		handler = m(handler)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := handler(w, r, nil)
		if err != nil {
			return
		}
	}
}

func ConvertBasicHandler[Request, Response any](handler BasicHandlerFunc[Request, Response]) HandlerFunc[Request, Response] {
	return func(_ http.ResponseWriter, r *http.Request, reqBody *Request) (*Response, error) {
		return handler(r.Context(), reqBody)
	}
}

// WrapDefault wraps a HandlerFunc with default middleware, returning a http.HandlerFunc.
// This function should be used instead of Wrap in most cases.
// The default middleware are handling validation ValidationMiddleware
// decoding the request body RequestDecoderMiddleware, and handling the response ResponseHandlerMiddleware.
// For execution order refer to the inline comments.
func WrapDefault[Request, Response any](handler HandlerFunc[Request, Response]) http.HandlerFunc {
	validator, err := requestvalidator.NewValidator() // this has a very small chance of returning an error, if it does we panic to prevent the server from starting
	if err != nil {
		panic(err)
	}

	return Wrap(
		handler, // runs last and executes second to last (the actual handler)
		ValidationMiddleware[Request, Response](validator), // runs third and executes second
		RequestDecoderMiddleware[Request, Response](),      // runs second and executes first
		ResponseHandlerMiddleware[Request, Response]())     // runs first and executes last (the response handling part)
}
