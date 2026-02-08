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
err := errs.Newf("user %s not found", userID)
err.SafeMessage = "User not found"
err.UserDetails = map[string]any{"user_id": userID}
err.LogDetails = []any{
    "attempted_id", userID,
    "query_duration_ms", 42,
}
err.Domain = "users"
return err
```

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
  "message": "User not found",
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
