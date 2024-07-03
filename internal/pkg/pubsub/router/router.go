package router

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/buni/wallet/internal/pkg/sloglog"
)

// Router ...
type Router struct {
	handlers    []HandleSubscribe
	middlewares []Middleware
	concurrency int
	logger      *slog.Logger
	wg          *sync.WaitGroup
}

// Option router option type.
type Option func(*Router) error

// WithConcurrency increases the number of concurrent listeners an individual handler can have.
// Defaults to 1.
// This is useful for handlers that need to do a lot of work and can benefit from parallelism.
// To make the most out of this option, you should also increase the prefetch/polling size of your Subscriber.
func WithConcurrency(concurrency int) Option {
	return func(r *Router) error {
		r.concurrency = concurrency
		return nil
	}
}

// WithLogger sets the logger for the router.
// Defaults to zap.NewProduction().
func WithLogger(logger *slog.Logger) Option {
	return func(r *Router) error {
		r.logger = logger
		return nil
	}
}

// WithMiddleware adds middleware to each handler registered with the router.
// Middleware order is left to right (first to last).
func WithMiddleware(middlewares ...Middleware) Option {
	return func(r *Router) error {
		r.middlewares = append(r.middlewares, middlewares...)
		return nil
	}
}

// NewRouter creates a new router.
func NewRouter(opts ...Option) (*Router, error) {
	r := &Router{
		concurrency: 1,
		logger: slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
		})),
		wg: &sync.WaitGroup{},
	}

	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// Register registers a handler(s) with the router,
// as well as applying any middleware provided to the router.
func (r *Router) Register(handlers ...HandleSubscribe) {
	for k := range handlers {
		handlers[k] = applyMiddleware(handlers[k], r.middlewares...)
	}

	r.handlers = append(r.handlers, handlers...)
}

// Start starts the router.
// This will start all handlers registered with the router.
// To cancel/stop the router, cancel the context passed to this function.
func (r *Router) Start(ctx context.Context) error {
	for _, h := range r.handlers {
		msgChan, err := h.Subscribe(ctx)
		if err != nil {
			return fmt.Errorf("failed to create subscriber for handler: %w", err)
		}
		for range r.concurrency {
			r.wg.Add(1)
			go r.processor(ctx, h, msgChan) //nolint:errcheck
		}
	}
	return nil
}

func (r *Router) processor(ctx context.Context, h HandleFunc, msgChan <-chan pubsub.SubscriberMessage) {
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			for len(msgChan) > 0 { // drain msgs
				select { // use a select in case the messages are drained by another goroutine
				case msg := <-msgChan:
					err := h.Handle(ctx, msg)
					if err != nil {
						r.logger.Error("failed to drain msgs", sloglog.Error(err))
					}
				default: // if there are no more messages, exit the function
					return
				}
			}
			return
		case msg := <-msgChan:
			err := h.Handle(ctx, msg)
			if err != nil {
				r.logger.Error("failed to execute handler", sloglog.Error(err))
			}
		}
	}
}

// Wait waits for all handlers to finish.
// This will block until all handlers have finished, after the context provided to Start(ctx) has been cancelled.
func (r *Router) Wait() {
	r.wg.Wait()
}
