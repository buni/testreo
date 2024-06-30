package sloglog

import "log/slog"

// Error returns a slog.Attr with the given error message.
// It takes an error as input and returns a slog.Attr with the "error" key and the error message as the value.
func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}
