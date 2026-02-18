# errs

Opinionated error primitives for Go HTTP services. Maps canonical errors to HTTP semantics, structured logging, and safe JSON responses.

## Install
```bash
go get github.com/4nd3r5on/errs
```

## Usage

### Basic errors
```go
// Canonical errors
return errs.ErrNotFound
return errs.ErrInvalidArgument

// Custom errors
return errs.New("database connection failed")
return errs.Newf("invalid user_id: %d", id)
```

### Structured errors
```go
// arguments of type errs.Option can modify error's internals
// but they don't affect the formatting.
// can be used with both errs.New and errs.Newf
err := errs.Newf("user %s not found", userID, func(err *Error) {
    err.SafeMessage = "User not found"
    err.UserDetails = map[string]any{"user_id": userID}
    err.LogDetails = []any{
        "attempted_id", userID,
        "query_duration_ms", 42,
    }
    err.Domain = "users"
})
return err
```

### Factory API
```go
// Explicit control over visibility and structure
return errs.F().
    Message("user query failed: %v", dbErr).
    UserMessage("User not found").
    Logs([]any{"user_id", id, "duration_ms", 42}).
    Mark(dbErr).  // infers public/private from marked error's HTTP code
    Domain("users").
    Err()

// Force visibility regardless of marked errors
return errs.F().
    Message("rate limit exceeded").
    Mark(errs.ErrRateLimited).
    Private().  // lock as private (500) despite 429 marker
    Err()

// Minimal usage
return errs.F().Message("db timeout").Mark(context.DeadlineExceeded).Err()
```

**Visibility inference**: `Mark()` auto-sets public if any marked error maps to `<500`. Override with `Private()`/`Public()`.

**Immutability**: Each method returns a new factory instance. Safe to reuse base factories.

### HTTP handling
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    user, err := getUser(ctx, id)
    if errs.HandleHTTP(ctx, w, r, err) {
        return // Error logged and JSON response sent
    }
    json.NewEncoder(w).Encode(user)
}
```

Response on error:
```json
{
  "error": "User not found",
  "details": {"user_id": "123"}
}
```

### Direct logging
```go
errs.LogErr(ctx, err,
    errs.LogErrUseLogLevel(slog.LevelWarn),
    errs.LogErrUseLoggerAttrs("request_id", reqID),
)
```

## Error mapping

| Error | HTTP Status |
|-------|-------------|
| `ErrNotFound` | 404 |
| `ErrInvalidArgument`, `ErrMissingArgument`, `ErrOutOfRange` | 400 |
| `ErrUnauthorized` | 401 |
| `ErrPermissionDenied` | 403 |
| `ErrExists`, `ErrOutdated` | 409 |
| `ErrRateLimited` | 429 |
| `ErrNotImplemented` | 501 |
| `ErrRemoteServiceErr` | 502 |
| `context.DeadlineExceeded` | 504 |
| Others | 500 |

## Features

- **Standard library only**
- **Error wrapping**: Compatible with `errors.Is/As` and `%w`
- **Safe by default**: Internal errors hidden unless `ExposeInternal=true`
- **Structured logging**: Attach arbitrary data for logs and JSON responses separately
- **HTTP-aware**: Automatic status code mapping and JSON rendering
- **Fluent factories**: Declarative error construction with visibility inference
