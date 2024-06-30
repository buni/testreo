package sloglog

import "log/slog"

func ApplyMiddleware(logger slog.Handler, middlewares ...func(next slog.Handler) slog.Handler) slog.Handler {
	for _, h := range middlewares {
		logger = h(logger)
	}
	return logger
}
