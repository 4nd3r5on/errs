package errs

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cockroachdb/errors"
)

func GetHTTPCode(err error) int {
	switch {
	case errors.Is(err, ErrNotImplemented):
		return http.StatusNotImplemented
	case errors.Is(err, ErrInternal):
		return http.StatusInternalServerError
	case errors.Is(err, ErrDeadlineExceeded):
		return http.StatusGatewayTimeout
	case errors.Is(err, ErrRemoteServiceErr):
		return http.StatusBadGateway
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.IsAny(err,
		ErrInvalidArgument,
		ErrMissingArgument,
		ErrOutOfRange,
	):
		return http.StatusBadRequest
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.IsAny(err, ErrExists, ErrOutdated):
		return http.StatusConflict
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	default:
		// ErrInternal
		// ErrCanceled
		// ErrOOM
		return http.StatusInternalServerError
	}
}

// HTTPErrResponse is the standard JSON error response body
type HTTPErrResponse struct {
	Error string      `json:"error"`                 // User-facing message
	Code  string      `json:"code,omitempty"`        // Machine-readable error code
	Hints []string    `json:"hints,omitempty"`       // User-facing suggestions
	Links []IssueLink `json:"issue_links,omitempty"` // Bug tracker references
}

type HandleHTTPErrOpts struct {
	Logger   *slog.Logger
	LogLevel slog.Level

	// Response body control
	IncludeDetails    bool // Developer-facing details (PII risk)
	IncludeHints      bool // User-facing hints
	IncludeIssueLinks bool // Bug tracker links
	IncludeErrorCode  bool // Telemetry key as error code

	// Error handling behavior
	CreateBarrier   bool // Use Handled() to hide internal errors from clients
	SanitizeMessage bool // Only show generic message for 500s
}

var DefaultHandleHTTPErrOpts = HandleHTTPErrOpts{
	Logger:   slog.Default(),
	LogLevel: slog.LevelError,

	IncludeHints:      true,
	IncludeIssueLinks: true,
	IncludeErrorCode:  true,
	SanitizeMessage:   true,
}

func HandleHTTPErr(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	err error,
	opts *HandleHTTPErrOpts,
) (handled bool) {
	if err == nil {
		return false
	}

	if opts == nil {
		opts = &DefaultHandleHTTPErrOpts
	}

	status := GetHTTPCode(err)

	resp := HTTPErrResponse{
		Error: err.Error(),
	}

	if opts.SanitizeMessage && status >= 500 {
		resp.Error = http.StatusText(status)
	}
	if opts.IncludeHints {
		resp.Hints = errors.GetAllHints(err)
	}
	if opts.IncludeIssueLinks {
		resp.Links = getIssueLinks(err)
	}

	respBytes, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		logMarshalErr := errors.Wrap(marshalErr, "failed to marshal HTTP error response")
		LogErr(ctx, logMarshalErr,
			LogErrUseLogger(opts.Logger),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	LogErr(ctx, err,
		LogErrUseLogger(opts.Logger),
		LogErrUseLogLevel(opts.LogLevel),
		LogErrUseLoggerArgs(
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"remote_addr", r.RemoteAddr,
		),
		LogErrUseLogDetails(true),
		LogErrUseLogHints(false),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, writeErr := w.Write(respBytes); writeErr != nil {
		writeErr = errors.Wrap(writeErr, "failed to write error response body")
		LogErr(ctx, writeErr,
			LogErrUseLogger(opts.Logger),
		)
	}
	return true
}
