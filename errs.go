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

type KVPair[K any, V any] struct {
	Key K
	Val V
}

type Error struct {
	// Internal is the underlying cause.
	// By being an 'error' type, it allows for %w wrapping and stack traces.
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
}

// Error implements the error interface.
// Returns Internal error message
func (e Error) Error() string {
	return e.Internal.Error()
}

// Unwrap returns the underlying wrapped error to support errors.As and errors.Is.
func (e Error) Unwrap() error {
	return e.Internal
}

// Newf creates a new *Error with formatted internal message and optional wrapped error.
// Usage examples:
//
//	Newf("something failed: %w", err) // wraps err
//	Newf("simple error without wrapping")
func Newf(internalMsgFmt string, args ...any) *Error {
	return &Error{
		Internal:   fmt.Errorf(internalMsgFmt, args...),
		LogDetails: make([]any, 0),
	}
}

// New creates a new *Error
func New(internalMsg string) *Error {
	return &Error{
		Internal:   errors.New(internalMsg),
		LogDetails: make([]any, 0),
	}
}
