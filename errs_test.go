package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/4nd3r5on/errs"
)

type customErr struct{ msg string }

func (c customErr) Error() string { return c.msg }

func TestWrap(t *testing.T) {
	t.Run("returns nil when err is nil", func(t *testing.T) {
		if got := errs.Wrap(nil, "context"); got != nil {
			t.Errorf("Wrap(nil, msg) = %v, want nil", got)
		}
	})

	t.Run("prepends context to error message", func(t *testing.T) {
		base := errors.New("base error")
		wrapped := errs.Wrap(base, "context")

		want := "context: base error"
		if wrapped.Error() != want {
			t.Errorf("wrapped.Error() = %q, want %q", wrapped.Error(), want)
		}
	})

	t.Run("preserves original error for errors.Is", func(t *testing.T) {
		base := errors.New("base error")
		wrapped := errs.Wrap(base, "context")

		if !errors.Is(wrapped, base) {
			t.Error("errors.Is(wrapped, base) = false, want true")
		}
	})

	t.Run("preserves original error for errors.As", func(t *testing.T) {
		base := customErr{"custom"}
		wrapped := errs.Wrap(base, "context")

		var target customErr
		if !errors.As(wrapped, &target) {
			t.Error("errors.As(wrapped, &target) = false, want true")
		}
		if target.msg != "custom" {
			t.Errorf("target.msg = %q, want %q", target.msg, "custom")
		}
	})

	t.Run("chains multiple wraps", func(t *testing.T) {
		base := errors.New("base")
		w1 := errs.Wrap(base, "layer1")
		w2 := errs.Wrap(w1, "layer2")

		want := "layer2: layer1: base"
		if w2.Error() != want {
			t.Errorf("w2.Error() = %q, want %q", w2.Error(), want)
		}

		if !errors.Is(w2, base) {
			t.Error("errors.Is(w2, base) = false, want true")
		}
	})

	t.Run("preserves markers from wrapped Error", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		base := errors.New("base")

		marked := errs.Mark(base, sentinel)
		wrapped := errs.Wrap(marked, "context")

		if !errors.Is(wrapped, sentinel) {
			t.Error("errors.Is(wrapped, sentinel) = false, want true")
		}
		if !errors.Is(wrapped, base) {
			t.Error("errors.Is(wrapped, base) = false, want true")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		base := errors.New("base")
		wrapped := errs.Wrap(base, "context", func(e *errs.Error) {
			e.Domain = "test-domain"
		})

		asErr, ok := wrapped.(*errs.Error)
		if !ok {
			t.Fatal("wrapped is not *errs.Error")
		}
		if asErr.Domain != "test-domain" {
			t.Errorf("Domain = %q, want %q", asErr.Domain, "test-domain")
		}
	})

	t.Run("preserves fields from wrapped Error", func(t *testing.T) {
		base := errs.New("base", func(e *errs.Error) {
			e.ExposeInternal = true
			e.SafeMessage = "safe msg"
			e.Domain = "original-domain"
			e.LogDetails = []any{"key", "value"}
		})

		wrapped := errs.Wrap(base, "context")

		asErr, ok := wrapped.(*errs.Error)
		if !ok {
			t.Fatal("wrapped is not *errs.Error")
		}

		if !asErr.ExposeInternal {
			t.Error("ExposeInternal not preserved")
		}
		if asErr.SafeMessage != "safe msg" {
			t.Errorf("SafeMessage = %q, want %q", asErr.SafeMessage, "safe msg")
		}
		if asErr.Domain != "original-domain" {
			t.Errorf("Domain = %q, want %q", asErr.Domain, "original-domain")
		}
		if len(asErr.LogDetails) != 2 {
			t.Errorf("LogDetails length = %d, want 2", len(asErr.LogDetails))
		}
	})
}

func TestMark(t *testing.T) {
	t.Run("returns nil when err is nil", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		if got := errs.Mark(nil, sentinel); got != nil {
			t.Errorf("Mark(nil, sentinel) = %v, want nil", got)
		}
	})

	t.Run("preserves original error message", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		base := errors.New("base error")
		marked := errs.Mark(base, sentinel)

		if marked.Error() != "base error" {
			t.Errorf("marked.Error() = %q, want %q", marked.Error(), "base error")
		}
	})

	t.Run("enables errors.Is matching with marker", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		base := errors.New("base error")
		marked := errs.Mark(base, sentinel)

		if !errors.Is(marked, sentinel) {
			t.Error("errors.Is(marked, sentinel) = false, want true")
		}
	})

	t.Run("preserves errors.Is matching with original error", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		base := errors.New("base error")
		marked := errs.Mark(base, sentinel)

		if !errors.Is(marked, base) {
			t.Error("errors.Is(marked, base) = false, want true")
		}
	})

	t.Run("supports multiple markers", func(t *testing.T) {
		marker1 := errors.New("marker1")
		marker2 := errors.New("marker2")
		base := errors.New("base")

		marked := errs.Mark(base, marker1)
		marked = errs.Mark(marked, marker2)

		if !errors.Is(marked, marker1) {
			t.Error("errors.Is(marked, marker1) = false, want true")
		}
		if !errors.Is(marked, marker2) {
			t.Error("errors.Is(marked, marker2) = false, want true")
		}
		if !errors.Is(marked, base) {
			t.Error("errors.Is(marked, base) = false, want true")
		}
	})

	t.Run("wraps foreign errors", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		foreign := fmt.Errorf("foreign error")
		marked := errs.Mark(foreign, sentinel)

		if !errors.Is(marked, sentinel) {
			t.Error("errors.Is(marked, sentinel) = false, want true")
		}
		if !errors.Is(marked, foreign) {
			t.Error("errors.Is(marked, foreign) = false, want true")
		}
		if marked.Error() != "foreign error" {
			t.Errorf("marked.Error() = %q, want %q", marked.Error(), "foreign error")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		sentinel := errors.New("sentinel")
		base := errors.New("base")
		marked := errs.Mark(base, sentinel, func(e *errs.Error) {
			e.Domain = "marked-domain"
		})

		asErr, ok := marked.(*errs.Error)
		if !ok {
			t.Fatal("marked is not *errs.Error")
		}
		if asErr.Domain != "marked-domain" {
			t.Errorf("Domain = %q, want %q", asErr.Domain, "marked-domain")
		}
	})

	t.Run("does not mutate original Error", func(t *testing.T) {
		marker1 := errors.New("marker1")
		marker2 := errors.New("marker2")
		base := errs.New("base")

		marked1 := errs.Mark(base, marker1)
		marked2 := errs.Mark(base, marker2)

		// base should not have any markers
		asBase, _ := base.(*errs.Error)
		if len(asBase.Markers) != 0 {
			t.Errorf("original error mutated: markers = %v", asBase.Markers)
		}

		// marked1 should only have marker1
		if errors.Is(marked1, marker2) {
			t.Error("marked1 should not match marker2")
		}

		// marked2 should only have marker2
		if errors.Is(marked2, marker1) {
			t.Error("marked2 should not match marker1")
		}
	})
}

func TestIntegration(t *testing.T) {
	t.Run("realistic usage pattern", func(t *testing.T) {
		var (
			ErrPermission = errors.New("permission denied")
			ErrNotFound   = errors.New("not found")
		)

		// Simulate layered error handling
		dbErr := errors.New("row not found")

		// Repository layer marks as NotFound
		repoErr := errs.Mark(dbErr, ErrNotFound)
		repoErr = errs.Wrap(repoErr, "user lookup failed")

		// Service layer marks as Permission error
		svcErr := errs.Mark(repoErr, ErrPermission)
		svcErr = errs.Wrap(svcErr, "authorization check")

		// All markers should be present
		if !errors.Is(svcErr, ErrPermission) {
			t.Error("missing ErrPermission marker")
		}
		if !errors.Is(svcErr, ErrNotFound) {
			t.Error("missing ErrNotFound marker")
		}
		if !errors.Is(svcErr, dbErr) {
			t.Error("lost original dbErr")
		}

		// Error message should show wrapping chain
		want := "authorization check: user lookup failed: row not found"
		if svcErr.Error() != want {
			t.Errorf("Error() = %q, want %q", svcErr.Error(), want)
		}
	})

	t.Run("example from requirements", func(t *testing.T) {
		ErrPermission := errors.New("permission denied")
		SomeError := errors.New("Some random error")

		err := SomeError
		err = errs.Mark(err, ErrPermission)

		if !errors.Is(err, ErrPermission) {
			t.Error("errors.Is(err, ErrPermission) = false, want true")
		}
		if err.Error() != "Some random error" {
			t.Errorf("err.Error() = %q, want %q", err.Error(), "Some random error")
		}
		if !errors.Is(err, SomeError) {
			t.Error("errors.Is(err, SomeError) = false, want true")
		}
	})
}

func TestExistingFunctionality(t *testing.T) {
	t.Run("New creates error", func(t *testing.T) {
		err := errs.New("test error")
		if err.Error() != "test error" {
			t.Errorf("Error() = %q, want %q", err.Error(), "test error")
		}
	})

	t.Run("New applies options", func(t *testing.T) {
		err := errs.New("test", func(e *errs.Error) {
			e.ExposeInternal = true
			e.SafeMessage = "safe"
			e.Domain = "domain"
		})

		asErr, ok := err.(*errs.Error)
		if !ok {
			t.Fatal("not *errs.Error")
		}
		if !asErr.ExposeInternal {
			t.Error("ExposeInternal not set")
		}
		if asErr.SafeMessage != "safe" {
			t.Error("SafeMessage not set")
		}
		if asErr.Domain != "domain" {
			t.Error("Domain not set")
		}
	})

	t.Run("Newf formats message", func(t *testing.T) {
		err := errs.Newf("error: %s %d", "test", 42)
		want := "error: test 42"
		if err.Error() != want {
			t.Errorf("Error() = %q, want %q", err.Error(), want)
		}
	})

	t.Run("Newf wraps error", func(t *testing.T) {
		base := errors.New("base")
		err := errs.Newf("wrapped: %w", base)

		if !errors.Is(err, base) {
			t.Error("errors.Is(err, base) = false, want true")
		}
	})

	t.Run("Newf applies options", func(t *testing.T) {
		err := errs.Newf("test %s", "msg", func(e *errs.Error) {
			e.Domain = "domain"
		})

		asErr, ok := err.(*errs.Error)
		if !ok {
			t.Fatal("not *errs.Error")
		}
		if asErr.Domain != "domain" {
			t.Error("Domain not set")
		}
	})

	t.Run("Newf applies options", func(t *testing.T) {
		err := errs.Newf("test %s", "msg", func(e *errs.Error) {
			e.Domain = "domain"
		})

		t.Logf("err type: %T", err)
		t.Logf("err value: %#v", err)

		asErr, ok := err.(*errs.Error)
		if !ok {
			t.Fatal("not *errs.Error")
		}

		t.Logf("Domain value: %q", asErr.Domain)
		t.Logf("Full Error struct: %#v", asErr)

		if asErr.Domain != "domain" {
			t.Error("Domain not set")
		}
	})

	t.Run("Unwrap returns Internal", func(t *testing.T) {
		base := errors.New("base")
		err := errs.New("wrapper", func(e *errs.Error) {
			e.Internal = base
		})

		if errors.Unwrap(err) != base {
			t.Error("Unwrap() did not return Internal")
		}
	})
}
