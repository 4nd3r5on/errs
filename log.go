package errs

import (
	"context"
	"errors"
	"log/slog"
)

var DefaultLogErrOptions = LogErrOptions{
	Logger:      slog.Default(),
	LogLevel:    slog.LevelError,
	LoggerAttrs: []any{},
}

func LogErr(ctx context.Context, err error, opts ...LogErrOption) {
	if err == nil {
		return
	}
	config := DefaultLogErrOptions
	for _, opt := range opts {
		opt(&config)
	}

	var e *Error
	if errors.As(err, &e) {
		attrs := make([]any, 0)
		attrs = append(attrs, e.LogDetails...)
		if e.Domain != "" {
			attrs = append(attrs, "domain", e.Domain)
		}
		attrs = append(attrs, config.LoggerAttrs...)
		config.Logger.Log(ctx, config.LogLevel, err.Error(), attrs...)
	} else {
		config.Logger.Log(ctx, config.LogLevel, err.Error(), config.LoggerAttrs...)
	}
}

func LogErrUseLogger(logger *slog.Logger) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.Logger = logger
	}
}

func LogErrUseLogLevel(level slog.Level) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LogLevel = level
	}
}

func LogErrUseLoggerAttrs(args ...any) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LoggerAttrs = args
	}
}
