# apperr

Transport-agnostic application error handling for Go.

Define errors once as `Definition`s, return them directly from any layer, and let the transport layer (HTTP, gRPC, GraphQL) map them to the appropriate response format.

## Why?

Common problem in Go backends using Clean Architecture:

```go
// service layer — should NOT know about HTTP
func (s *userService) GetByID(ctx context.Context, id string) (User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return User{}, resterrors.NewNotFoundError("user not found") // ❌ HTTP in business logic
    }
    return user, nil
}
```

With `apperr`:

```go
// service layer — transport-agnostic
func (s *userService) GetByID(ctx context.Context, id string) (User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return User{}, errcodes.ErrUserNotFound // ✅ domain error, returned directly
    }
    return user, nil
}
```

The HTTP layer handles the mapping automatically — your business logic never imports `net/http`.

## Install

```bash
go get github.com/diegoclair/apperr
```

## Quick Start

### 1. Define your errors (once)

```go
package errcodes

import "github.com/diegoclair/apperr"

var (
    ErrUserNotFound   = apperr.Define(apperr.KindNotFound, "USER_NOT_FOUND", "user not found")
    ErrEmailExists    = apperr.Define(apperr.KindConflict, "USER_EMAIL_EXISTS", "email already exists")
    ErrLoginFailed    = apperr.Define(apperr.KindAuthentication, "AUTH_LOGIN_FAILED", "invalid credentials")
    ErrAccountBlocked = apperr.Define(apperr.KindAuthentication, "AUTH_ACCOUNT_BLOCKED", "account blocked")
)
```

### 2. Return them from your services

```go
// Simple — return the Definition directly (it implements error)
return errcodes.ErrUserNotFound

// With metadata — creates a new *Error instance (immutable)
return errcodes.ErrLoginFailed.WithMeta("remaining_attempts", 2)

// With cause — preserves the original error for logging
return errcodes.ErrUserNotFound.Wrap(err)

// With custom message — overrides the default
return errcodes.ErrEmailExists.WithMessage("the email john@example.com is already in use")
```

### 3. Check errors

```go
// By Kind (category)
if apperr.IsNotFound(err) { ... }
if apperr.IsValidation(err) { ... }
if apperr.IsAuthentication(err) { ... }

// By specific error (compares Code)
if errors.Is(err, errcodes.ErrAccountBlocked) { ... }

// Extract code
code := apperr.GetCode(err) // "AUTH_ACCOUNT_BLOCKED"
```

### 4. Map to HTTP (in your transport layer)

```go
import "github.com/diegoclair/apperr/httpmap"

func handleError(w http.ResponseWriter, err error) {
    status, response := httpmap.ToHTTP(err)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}
```

JSON response:

```json
{
    "message": "invalid credentials",
    "status_code": 401,
    "error": "Unauthorized",
    "code": "AUTH_LOGIN_FAILED",
    "meta": {
        "remaining_attempts": 2
    }
}
```

## Architecture

```
apperr (core — zero net/http dependency)
├── AppError    interface (common between Definition and Error)
├── Definition  pre-defined error, returned directly as error
├── Error       error with context (cause, meta)
├── Kind        error category (Validation, NotFound, Conflict, ...)
├── Code        unique error identifier (string)
├── Helpers     IsNotFound(), IsValidation(), HasCode(), GetMeta(), ...
└── Sentinels   ErrNotFound, ErrInternal, ErrConflict, ... (generic, reusable)

httpmap/ (sub-package — maps Kind → HTTP status)
├── ToHTTP()        converts error → (statusCode, ErrorResponse)
├── StatusFromKind()
└── ErrorResponse   JSON-serializable response struct
```

### Kind → HTTP Status Mapping

| Kind | HTTP Status |
|------|-------------|
| `KindValidation` | 400 Bad Request |
| `KindAuthentication` | 401 Unauthorized |
| `KindAuthorization` | 403 Forbidden |
| `KindNotFound` | 404 Not Found |
| `KindConflict` | 409 Conflict |
| `KindRateLimited` | 429 Too Many Requests |
| `KindInternal` | 500 Internal Server Error |

### Built-in Sentinel Errors

These are generic errors usable in any project. For business-specific errors, define your own with `Define()`.

| Sentinel | Kind | Code |
|----------|------|------|
| `ErrValidation` | Validation | `VALIDATION_ERROR` |
| `ErrInvalidInput` | Validation | `INVALID_INPUT` |
| `ErrRequiredField` | Validation | `REQUIRED_FIELD` |
| `ErrInvalidFormat` | Validation | `INVALID_FORMAT` |
| `ErrNotFound` | NotFound | `NOT_FOUND` |
| `ErrRecordNotFound` | NotFound | `RECORD_NOT_FOUND` |
| `ErrConflict` | Conflict | `CONFLICT` |
| `ErrDuplicateEntry` | Conflict | `DUPLICATE_ENTRY` |
| `ErrUnauthenticated` | Authentication | `UNAUTHENTICATED` |
| `ErrTokenInvalid` | Authentication | `TOKEN_INVALID` |
| `ErrTokenExpired` | Authentication | `TOKEN_EXPIRED` |
| `ErrTokenRequired` | Authentication | `TOKEN_REQUIRED` |
| `ErrForbidden` | Authorization | `FORBIDDEN` |
| `ErrInternal` | Internal | `INTERNAL_ERROR` |
| `ErrRateLimited` | RateLimited | `RATE_LIMITED` |

## Design Decisions

### Why Definition + Error (two types)?

- **Definition** is a static, pre-defined error — cheap to return, no allocations
- **Error** adds context (cause, meta) — only created when needed via `WithMeta()`, `Wrap()`, or `WithMessage()`
- Both implement `AppError` interface, so helpers and mappers work with either transparently

### Why no global registry?

Definitions are package-level variables (`var ErrXxx = apperr.Define(...)`). This gives you:
- IDE autocomplete (`errcodes.Err` → see all errors)
- Compile-time safety (typos are caught)
- Zero global state, zero `init()` functions
- `errors.Is()` works via Code comparison

### Why no `New()` method?

Definition implements `error` directly. You just `return errcodes.ErrSomething`. Only when you need to add context (meta, cause, custom message) do the methods create an `*Error` instance. Less verbosity, same safety.

## Roadmap

- [ ] `grpcmap/` — Kind → gRPC status code mapping
- [ ] `gqlmap/` — Kind → GraphQL error extensions mapping

## License

MIT
