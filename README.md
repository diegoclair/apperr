# apperr

<p align="center">
  <b>Transport-agnostic application error handling for Go</b><br>
  Define errors once, return them directly, and let the transport layer handle the mapping.
  <br><br>
  <a href="https://github.com/diegoclair/apperr/actions/workflows/ci.yml">
    <img src="https://github.com/diegoclair/apperr/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://github.com/diegoclair/apperr/tags">
    <img src="https://img.shields.io/github/tag/diegoclair/apperr.svg" alt="GitHub tag" />
  </a>
  <a href="https://pkg.go.dev/github.com/diegoclair/apperr">
    <img src="https://pkg.go.dev/badge/github.com/diegoclair/apperr.svg" alt="Go Reference" />
  </a>
  <a href="https://goreportcard.com/report/github.com/diegoclair/apperr">
    <img src="https://goreportcard.com/badge/github.com/diegoclair/apperr" alt="Go Report Card" />
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License" />
  </a>
</p>

## Introduction

### Why

A common pattern in Go backends — even those following Clean Architecture — is leaking transport concerns into business logic:

```go
func (s *userService) Login(ctx context.Context, email, password string) (User, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        // ❌ HTTP status in business logic (imports net/http)
        return User{}, echo.NewHTTPError(http.StatusNotFound, "user not found")
    }

    if user.IsBlocked {
        // ❌ No error code — frontend can't programmatically handle this
        // ❌ No metadata — frontend can't show "try again in 5 minutes"
        // ❌ Hardcoded message — can't do i18n
        return User{}, echo.NewHTTPError(http.StatusForbidden, "account blocked")
    }

    return user, nil
}
```

This couples your service layer to HTTP, makes errors impossible to translate (i18n), and gives the frontend no structured way to handle specific error cases.

### How

With `apperr`, errors are defined once as `Definition`s — transport-agnostic, with a unique `Code` and `Kind`. Return them directly from any layer:

```go
func (s *userService) Login(ctx context.Context, email, password string) (User, error) {
    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return User{}, errcodes.ErrUserNotFound       // ✅ transport-agnostic, no net/http
    }

    if user.IsBlocked {
        return User{}, errcodes.ErrAccountBlocked.     // ✅ error code: "AUTH_ACCOUNT_BLOCKED"
            WithMeta("retry_after_minutes", 5)          // ✅ structured metadata for the frontend
    }

    return user, nil
}
```

The transport layer maps `Kind` to the appropriate status code automatically. Your business logic never imports `net/http`, and the frontend gets structured `code` + `meta` for i18n and programmatic handling.

## Install

```bash
go get github.com/diegoclair/apperr
```

## Getting Started

### 1. Define your errors (once)

Create a package in your project with all error definitions. Each `Definition` has a `Kind` (category), a `Code` (unique identifier for the frontend), and a default message:

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

The library also provides [built-in sentinel errors](#built-in-sentinel-errors) for common cases like `ErrNotFound`, `ErrInternal`, `ErrTokenExpired`, etc.

### 2. Return errors from your services

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

All helpers use `errors.As` under the hood, so they work with wrapped errors (e.g., `fmt.Errorf("repo: %w", err)`):

```go
// By Kind (category)
if apperr.IsNotFound(err) { ... }
if apperr.IsValidation(err) { ... }
if apperr.IsAuthentication(err) { ... }

// By specific error (compares Code via errors.Is)
if errors.Is(err, errcodes.ErrAccountBlocked) { ... }

// Extract code
code := apperr.GetCode(err) // "AUTH_ACCOUNT_BLOCKED"
```

### 4. Map to HTTP (in your transport layer)

The `httpmap` sub-package converts any `AppError` into an HTTP response. Import it only in your transport layer — the core `apperr` package has zero dependency on `net/http`:

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

The frontend uses `code` for i18n translations and programmatic handling, while `meta` carries dynamic data. The `message` field serves as a fallback for logging and debugging.

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

Generic errors reusable in any project. For business-specific errors, define your own with `Define()`.

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

`Definition` implements `error` directly. You just `return errcodes.ErrSomething`. Only when you need to add context (meta, cause, custom message) do the methods create an `*Error` instance. Less verbosity, same safety.

## Roadmap

- [ ] `grpcmap/` — Kind → gRPC status code mapping
- [ ] `gqlmap/` — Kind → GraphQL error extensions mapping

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## License

[MIT](./LICENSE)
