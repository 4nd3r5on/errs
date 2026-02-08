package errs

import "log/slog"

type LogErrOptions struct {
	Logger      *slog.Logger
	LogLevel    slog.Level
	LoggerAttrs []any
}

type LogErrOption func(*LogErrOptions)
