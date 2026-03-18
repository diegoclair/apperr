package apperr

// Definition is a pre-defined (static) error.
// Implements AppError and error — can be returned directly as error.
// Each project defines its Definitions as package-level variables.
//
// Simple usage (direct return):
//
//	return errcodes.ErrAccountBlocked
//
// With context (creates *Error):
//
//	return errcodes.ErrLoginFailed.WithMeta("remaining_attempts", 2)
//	return errcodes.ErrInternal.Wrap(err)
type Definition struct {
	kind    Kind
	code    Code
	message string
}

// Compile-time check: *Definition implements AppError.
var _ AppError = (*Definition)(nil)

// Define creates an error Definition. Called once at package initialization.
//
// Example:
//
//	var ErrAccountBlocked = apperr.Define(apperr.KindAuthentication, "AUTH_ACCOUNT_BLOCKED", "account blocked due to failed login attempts")
func Define(kind Kind, code Code, message string) *Definition {
	return &Definition{kind: kind, code: code, message: message}
}

// Error implements the error interface — allows direct return as error.
func (d *Definition) Error() string   { return d.message }
func (d *Definition) Kind() Kind      { return d.kind }
func (d *Definition) Code() Code      { return d.code }
func (d *Definition) Message() string { return d.message }

// Wrap creates an *Error from this Definition, wrapping an original error.
// Use when you want to preserve the original error for logging.
func (d *Definition) Wrap(cause error) *Error {
	return &Error{kind: d.kind, code: d.code, message: d.message, cause: cause}
}

// WithMessage creates an *Error with a custom message (overrides the default).
// Useful when the message needs dynamic context.
func (d *Definition) WithMessage(message string) *Error {
	return &Error{kind: d.kind, code: d.code, message: message}
}

// WithMeta creates an *Error with metadata. Used for dynamic data
// that the frontend needs (e.g., remaining_attempts, field_name).
func (d *Definition) WithMeta(key string, value any) *Error {
	return &Error{
		kind: d.kind, code: d.code, message: d.message,
		meta: map[string]any{key: value},
	}
}

// Is allows comparison with errors.Is() — compares by Code.
// Works with both *Error and *Definition as target.
func (d *Definition) Is(target error) bool {
	if t, ok := target.(AppError); ok {
		return d.code == t.Code()
	}
	return false
}
