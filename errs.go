// Package errs provides opinionated error primitives and error handling
//
// It defines a small set of canonical error values, maps them to HTTP semantics,
// and exposes helpers for rendering safe, structured error responses.
package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrRemoteServiceErr = errors.New("remote service error")
	ErrRateLimited      = errors.New("rate limited")

	ErrInvalidArgument = errors.New("invalid argument")
	ErrMissingArgument = errors.New("missing argument")
	ErrOutOfRange      = errors.New("out of range")

	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")

	ErrExists   = errors.New("already exists")
	ErrNotFound = errors.New("not found")
	ErrOutdated = errors.New("outdated")
)

type Error struct {
	// Internal is the underlying cause.
	// By being an 'error' type, it allows for %w wrapping.
	Internal error

	// Whether or not show user external message if Message field is empty
	ExposeInternal bool

	// SafeMessage is the "Safe" human-readable message intended for the end-user.
	SafeMessage string

	// LogDetails contains data for slog.
	LogDetails []any

	// UserDetails gets marshaled to the JSON response and sent to the user
	UserDetails any

	// TraceID or Domain can be added here for "Marking" where the error originated.
	Domain string

	// Markers holds sentinel errors for errors.Is matching
	Markers []error
}

// Error implements the error interface.
// Returns Internal error message
func (e *Error) Error() string {
	return e.Internal.Error()
}

// Unwrap returns the underlying wrapped error to support errors.As and errors.Is.
func (e *Error) Unwrap() error {
	return e.Internal
}

// Is implements errors.Is matching for marked sentinel errors
func (e *Error) Is(target error) bool {
	// Check if target matches any marker
	for _, m := range e.Markers {
		if errors.Is(m, target) {
			return true
		}
	}
	// Fall back to unwrapping Internal
	return errors.Is(e.Internal, target)
}

// Option can be provided in args to New and Newf
// to change error's parameters
type Option func(*Error)

// Newf creates a new *Error with formatted internal message and optional wrapped error.
// Usage examples:
//
//	Newf("something failed: %w", err) // wraps err
//	Newf("simple error without wrapping")
func Newf(internalMsgFmt string, args ...any) error {
	e := &Error{LogDetails: make([]any, 0)}
	cleanArgs := make([]any, 0, len(args))

	for _, arg := range args {
		if opt, ok := arg.(Option); ok {
			opt(e)
			continue
		}
		if opt, ok := arg.(func(*Error)); ok {
			opt(e)
			continue
		}
		cleanArgs = append(cleanArgs, arg)
	}

	e.Internal = fmt.Errorf(internalMsgFmt, cleanArgs...)
	if e.Internal == nil {
		return nil
	}
	return e
}

// New creates a new *Error
func New(internalMsg string, opts ...Option) error {
	err := &Error{
		Internal:   errors.New(internalMsg),
		LogDetails: make([]any, 0),
	}

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Wrap wraps an error with additional context string.
// Returns nil if err is nil.
// Preserves original error for errors.Is/As.
func Wrap(err error, msg string, opts ...Option) error {
	if err == nil {
		return nil
	}

	e := &Error{
		Internal:   fmt.Errorf("%s: %w", msg, err),
		LogDetails: make([]any, 0),
	}

	// Preserve markers if wrapping another *Error
	if prev, ok := err.(*Error); ok {
		e.Markers = prev.Markers
		e.ExposeInternal = prev.ExposeInternal
		e.SafeMessage = prev.SafeMessage
		e.UserDetails = prev.UserDetails
		e.Domain = prev.Domain
		if prev.LogDetails != nil {
			e.LogDetails = prev.LogDetails
		}
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Mark marks an error with a sentinel error for errors.Is matching.
// Returns nil if err is nil.
// The original error message is preserved; marker is only for Is() matching.
func Mark(err error, marker error, opts ...Option) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		// Wrap foreign error into *Error
		e = &Error{
			Internal:   err,
			LogDetails: make([]any, 0),
			Markers:    []error{marker},
		}
	} else {
		// Clone to avoid mutating original
		clone := *e
		clone.Markers = append([]error{}, e.Markers...)
		clone.Markers = append(clone.Markers, marker)
		e = &clone
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}
