package apperr

// Generic sentinel errors — common errors reusable in any project.
// For business-specific errors, define in your project with Define().
var (
	// Validation
	ErrValidation    = Define(KindValidation, "VALIDATION_ERROR", "validation error")
	ErrInvalidInput  = Define(KindValidation, "INVALID_INPUT", "invalid input")
	ErrRequiredField = Define(KindValidation, "REQUIRED_FIELD", "required field is missing")
	ErrInvalidFormat = Define(KindValidation, "INVALID_FORMAT", "invalid format")

	// Not Found
	ErrNotFound       = Define(KindNotFound, "NOT_FOUND", "resource not found")
	ErrRecordNotFound = Define(KindNotFound, "RECORD_NOT_FOUND", "record not found")

	// Conflict
	ErrConflict       = Define(KindConflict, "CONFLICT", "resource already exists")
	ErrDuplicateEntry = Define(KindConflict, "DUPLICATE_ENTRY", "duplicate entry")

	// Authentication
	ErrUnauthenticated = Define(KindAuthentication, "UNAUTHENTICATED", "authentication required")
	ErrTokenInvalid    = Define(KindAuthentication, "TOKEN_INVALID", "token is invalid")
	ErrTokenExpired    = Define(KindAuthentication, "TOKEN_EXPIRED", "token has expired")
	ErrTokenRequired   = Define(KindAuthentication, "TOKEN_REQUIRED", "token is required")

	// Authorization
	ErrForbidden = Define(KindAuthorization, "FORBIDDEN", "access denied")

	// Internal
	ErrInternal = Define(KindInternal, "INTERNAL_ERROR", "internal server error")

	// Rate Limit
	ErrRateLimited = Define(KindRateLimited, "RATE_LIMITED", "rate limit exceeded")
)
