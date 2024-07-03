package sloglog

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey struct{}

// ToContext adds the logger to the context.
// If a logger is already present, it is replaced.
func ToContext(ctx context.Context, id *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext extracts the logger from the context.
// If no logger is found, a new one is created, using zap.NewProduction() as the base.
func FromContext(ctx context.Context) *slog.Logger {
	total, ok := ctx.Value(ctxKey{}).(*slog.Logger)
	if !ok {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
		}))
	}
	return total
}
