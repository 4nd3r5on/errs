package errs

import (
	"context"
	"log/slog"

	"github.com/cockroachdb/errors"
)

type LogErrOptions struct {
	Logger     *slog.Logger
	LogLevel   slog.Level
	LoggerArgs []any
	LogDetails bool
	LogHints   bool
	LogLinks   bool
	LogSource  bool
}

type LogErrOption func(*LogErrOptions)

var DefaultLogErrOptions = LogErrOptions{
	Logger:     slog.Default(),
	LogLevel:   slog.LevelError,
	LoggerArgs: []any{},
	LogDetails: true,
	LogHints:   false,
	LogLinks:   true,
	LogSource:  true,
}

func LogErr(ctx context.Context, err error, opts ...LogErrOption) (errIsNotNil bool) {
	if err == nil {
		return false
	}

	config := DefaultLogErrOptions
	for _, opt := range opts {
		opt(&config)
	}
	loggerArgs := make([]any, 0)

	if config.LogSource {
		file, line, fn, ok := errors.GetOneLineSource(err)
		if ok {
			loggerArgs = append(loggerArgs, slog.Group("source",
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("function", fn),
			))
		} else {
			loggerArgs = append(loggerArgs, slog.String("source", "not found"))
		}
	}

	if config.LogDetails {
		details := errors.GetAllDetails(err)
		if len(details) > 0 {
			loggerArgs = append(loggerArgs,
				slog.Any("details", details),
			)
		}
	}

	if config.LogHints {
		hints := errors.GetAllHints(err)
		if len(hints) > 0 {
			loggerArgs = append(loggerArgs,
				slog.Any("hints", hints),
			)
		}
	}

	if config.LogLinks {
		links := getIssueLinks(err)
		linksLogVals := make([]any, 0)
		for _, link := range links {
			linksLogVals = append(linksLogVals, slog.GroupValue(
				slog.String("issue_url", link.IssueURL),
				slog.String("detail", link.Detail),
			))
		}
		loggerArgs = append(loggerArgs, slog.Any("links", linksLogVals))
	}

	loggerArgs = append(
		loggerArgs,
		config.LoggerArgs...,
	)
	config.Logger.Log(
		ctx,
		config.LogLevel,
		err.Error(),
		loggerArgs...,
	)

	return true
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

func LogErrUseLoggerArgs(args ...any) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LoggerArgs = args
	}
}

func LogErrUseLogDetails(log bool) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LogDetails = log
	}
}

func LogErrUseLogHints(log bool) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LogHints = log
	}
}

func LogErrUseLogLinks(log bool) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LogLinks = log
	}
}

func LogErrUseLogSource(log bool) LogErrOption {
	return func(opts *LogErrOptions) {
		opts.LogSource = log
	}
}
