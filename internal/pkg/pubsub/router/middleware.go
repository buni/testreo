package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/buni/wallet/internal/pkg/database"
	"github.com/buni/wallet/internal/pkg/pubsub"
	"github.com/buni/wallet/internal/pkg/sloglog"
)

type ComplexMiddleware func(next HandleSubscribe) HandleSubscribe

type ComplexMiddlewareWrapper struct {
	next           HandleSubscribe
	nextMiddleware Middleware
}

func (m *ComplexMiddlewareWrapper) Handle(ctx context.Context, msg pubsub.SubscriberMessage) error {
	err := m.nextMiddleware(m.next).Handle(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to handle message: %w", err)
	}
	return nil
}

func (m *ComplexMiddlewareWrapper) Subscribe(ctx context.Context) (<-chan pubsub.SubscriberMessage, error) {
	return m.next.Subscribe(ctx)
}

func (m *ComplexMiddlewareWrapper) HandlerName() string {
	return m.next.HandlerName()
}

func (m *ComplexMiddlewareWrapper) Topic() string {
	return m.next.Topic()
}

type Middleware func(next HandleFunc) HandleFunc

func applyMiddleware(hs HandleSubscribe, middlewares ...Middleware) HandleSubscribe {
	for _, middleware := range middlewares {
		hs = applyComplexMiddleware(hs, func(next HandleSubscribe) HandleSubscribe {
			return &ComplexMiddlewareWrapper{next: next, nextMiddleware: middleware}
		})
	}

	return hs
}

func applyComplexMiddleware(hs HandleSubscribe, middlewares ...ComplexMiddleware) HandleSubscribe {
	for _, middleware := range middlewares {
		hs = middleware(hs)
	}

	return hs
}

type autoAckNackMiddleware struct {
	next HandleFunc
}

// AutoAckNackMiddleware is a middleware that automatically acks or nacks the message based on the error returned by the handler
// if the handler returns an error, the message is nacked, otherwise it's acked.
func AutoAckNackMiddleware(next HandleFunc) HandleFunc {
	return &autoAckNackMiddleware{next: next}
}

// Handle implements the HandleFunc interface.
func (m *autoAckNackMiddleware) Handle(ctx context.Context, msg pubsub.SubscriberMessage) (err error) {
	err = m.next.Handle(ctx, msg)
	if err != nil {
		ok, processedErr := msg.IsProcessed(ctx)
		if processedErr != nil {
			return errors.Join(err, processedErr)
		}

		if !ok {
			nackErr := msg.Nack(ctx)
			if nackErr != nil {
				return errors.Join(err, nackErr)
			}
		}

		return err //nolint:wrapcheck
	}

	ok, err := msg.IsProcessed(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if message is processed: %w", err)
	}

	if !ok {
		err = msg.Ack(ctx)
		if err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
	}

	return nil
}

type atomicTransactionMiddleware struct {
	next HandleFunc
	txm  database.TransactionManager
}

// AtomicTransactionMiddleware is a middleware that wraps the handler in a transaction.
// If the handler returns an error, the transaction is rolled back, otherwise it's committed.
func AtomicTransactionMiddleware(txm database.TransactionManager) func(next HandleFunc) HandleFunc {
	return func(next HandleFunc) HandleFunc {
		return &atomicTransactionMiddleware{next: next, txm: txm}
	}
}

func (m *atomicTransactionMiddleware) Handle(ctx context.Context, msg pubsub.SubscriberMessage) (err error) {
	err = m.txm.Run(ctx, func(ctx context.Context) error {
		err = m.next.Handle(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to handle message: %w", err)
		}

		return nil
	})
	if err != nil {
		nackErr := msg.Nack(ctx)
		if nackErr != nil {
			return fmt.Errorf("failed to nack message: %w", errors.Join(err, nackErr))
		}
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return nil
}

type loggerMiddleware struct {
	logger *slog.Logger
	next   HandleFunc
}

func LoggerMiddleware() func(next HandleFunc) HandleFunc {
	return func(next HandleFunc) HandleFunc {
		return &loggerMiddleware{next: next}
	}
}

func LoggerMiddlewareWithLogger(logger *slog.Logger) func(next HandleFunc) HandleFunc {
	return func(next HandleFunc) HandleFunc {
		return &loggerMiddleware{next: next, logger: logger}
	}
}

func (m *loggerMiddleware) Handle(ctx context.Context, msg pubsub.SubscriberMessage) (err error) {
	logger := sloglog.FromContext(ctx)

	if m.logger != nil {
		logger = m.logger
	}

	handler, ok := m.next.(HandleSubscribe)
	if !ok {
		logger.Error("failed to cast handler to HandleSubscribe")
		return m.next.Handle(ctx, msg) //nolint:wrapcheck
	}

	logger = logger.With(slog.String("consumer", "pubsub"), slog.String("handler", handler.HandlerName()), slog.String("topic", handler.Topic()))

	msgBody, err := msg.Message(ctx)
	if err != nil {
		logger.Error("failed to get message", sloglog.Error(err))
		return err //nolint:wrapcheck
	}

	if msgBody.ID != nil {
		logger = logger.With(slog.String("message_id", *msgBody.ID))
	}

	if msgBody.Key != "" {
		logger = logger.With(slog.String("message_key", msgBody.Key))
	}

	ctx = sloglog.ToContext(ctx, logger)

	err = m.next.Handle(ctx, msg)
	if err != nil {
		logger.Error("failed to handle message", sloglog.Error(err))
		return err //nolint:wrapcheck
	}

	return nil
}

type panicRecoveryMiddleware struct {
	next HandleFunc
}

func PanicRecoveryMiddleware(next HandleFunc) HandleFunc {
	return &panicRecoveryMiddleware{next: next}
}

func (m *panicRecoveryMiddleware) Handle(ctx context.Context, msg pubsub.SubscriberMessage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			sloglog.FromContext(ctx).Error("panic recovered", sloglog.Error(err), slog.Any("stack", string(debug.Stack())))
		}
	}()

	err = m.next.Handle(ctx, msg)
	if err != nil {
		return
	}
	return
}
