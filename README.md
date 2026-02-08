# errs

Opinionated error primitives built on [`cockroachdb/errors`](https://github.com/cockroachdb/errors).

## Why

- Canonical error values → HTTP status codes
- Structured JSON responses with user/developer concerns separated
- Rich diagnostics (hints, issue links, source location) for logging
- Configurable sanitization: show stack traces in dev, hide in prod

## Usage

### Basic

```go
import "github.com/4nd3r5on/errs"

func getUser(id string) error {
    if id == "" {
        return errors.Wrap(errs.ErrMissingArgument, "user_id required")
    }
    // ...
}
```

### HTTP handler

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.svc.GetUser(r.Context(), userID)
    if errs.HandleHTTPErr(r.Context(), w, r, err, nil) {
        return // already responded
    }
    json.NewEncoder(w).Write(user)
}
```

**Default behavior** (`DefaultHandleHTTPErrOpts`):
- Logs with structured fields (method, path, status, source location)
- Returns hints + issue links in JSON
- Sanitizes 5xx messages to `http.StatusText(500)` (hides stack traces)
- Uses `slog.Default()` at `ERROR` level

**Dev override**:
```go
errs.HandleHTTPErr(ctx, w, r, err, &errs.HandleHTTPErrOpts{
    SanitizeMessage: false, // show full error text even for 500s
})
```

### Explicit logging

```go
err := doThing()
errs.LogErr(ctx, err,
    errs.LogErrUseLogger(customLogger),
    errs.LogErrUseLogLevel(slog.LevelWarn),
    errs.LogErrUseLoggerArgs("trace_id", traceID),
)
```

Extracts: source location, details, hints, issue links into structured log fields.

## Error → HTTP mapping

```
ErrInvalidArgument  → 400
ErrUnauthorized     → 401
ErrPermissionDenied → 403
ErrNotFound         → 404
ErrExists/Outdated  → 409
ErrRateLimited      → 429
ErrNotImplemented   → 501
ErrRemoteServiceErr → 502
ErrDeadlineExceeded → 504
// everything else   → 500
```

**Wins**:
- `HandleHTTPErr` call replaces manual status code mapping + JSON marshaling

**Not included**:
- No middleware (use `HandleHTTPErr` in handlers directly)
- No i18n (user-facing messages are English error strings)
