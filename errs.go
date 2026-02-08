// Package errs provides opinionated error primitives and error handling
// built on top of cockroachdb/errors.
//
// It defines a small set of canonical error values, maps them to HTTP semantics,
// and exposes helpers for rendering safe, structured error responses while
// preserving rich diagnostic context for logging and tracing.
package errs

import (
	"github.com/cockroachdb/errors"
)

// IssueLink has the same structure as errors.IssueLink
// but also has additional JSON tags
type IssueLink struct {
	// URL to the issue on a tracker.
	IssueURL string `json:"issue_url"`
	// Annotation that characterizes a sub-issue.
	Detail string `json:"detail,omitempty"`
}

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInternal       = errors.New("internal error")

	ErrCanceled         = errors.New("canceled")
	ErrOOM              = errors.New("out of memory")
	ErrDeadlineExceeded = errors.New("deadline exceeded")
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

func getIssueLinks(err error) []IssueLink {
	links := errors.GetAllIssueLinks(err)
	outLinks := make([]IssueLink, len(links))
	for i, link := range links {
		outLinks[i] = IssueLink{
			IssueURL: link.IssueURL,
			Detail:   link.Detail,
		}
	}
	return outLinks
}
