package errs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func GetHTTPCode(err error) int {
	switch {
	case errors.Is(err, ErrNotImplemented):
		return http.StatusNotImplemented
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, ErrRemoteServiceErr):
		return http.StatusBadGateway
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests
	case IsAny(err,
		ErrInvalidArgument,
		ErrMissingArgument,
		ErrOutOfRange,
	):
		return http.StatusBadRequest
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case IsAny(err, ErrExists, ErrOutdated):
		return http.StatusConflict
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func HTTPGetLogLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}

type ErrorHTTPResponse struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func HandleHTTP(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	err error,
	opts ...LogErrOption,
) (handled bool) {
	if err == nil {
		return false
	}
	status := GetHTTPCode(err)
	config := LogErrOptions{
		Logger:      slog.Default(),
		LogLevel:    HTTPGetLogLevel(status),
		LoggerAttrs: []any{},
	}

	for _, opt := range opts {
		opt(&config)
	}
	httpAttrs := []any{
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"remote_addr", r.RemoteAddr,
	}

	LogErr(
		ctx,
		err,
		append(opts, LogErrUseLoggerAttrs(
			httpAttrs...,
		))...,
	)

	message := http.StatusText(status)

	var e *Error
	if !errors.As(err, &e) {
		config.Logger.Log(ctx, config.LogLevel, err.Error(), config.LoggerAttrs...)
		http.Error(w, message, status)
		return
	}

	LogErr(ctx, err, append(opts, LogErrUseLoggerAttrs(httpAttrs...))...)

	if e.SafeMessage != "" {
		message = e.SafeMessage
	} else if e.ExposeInternal {
		message = e.Internal.Error()
	}

	resp, marshalErr := json.Marshal(ErrorHTTPResponse{
		Error:   message,
		Details: e.UserDetails,
	})
	if marshalErr != nil {
		config.Logger.ErrorContext(ctx,
			fmt.Sprintf("failed to marshal error response: %v", marshalErr),
			httpAttrs...,
		)
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(resp); err != nil {
		config.Logger.WarnContext(ctx,
			fmt.Sprintf("failed to write error response: %v", err),
			httpAttrs...,
		)
	}
	return true
}
