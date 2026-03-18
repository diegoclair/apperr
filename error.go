package apperr

import "maps"

// Error is a domain error with context (cause, meta).
// Created when extra information needs to be added to a Definition,
// via WithMeta, Wrap, or WithMessage.
// Implements AppError and Go's error interface.
type Error struct {
	kind    Kind
	code    Code
	message string
	cause   error
	meta    map[string]any
}

// Compile-time check: *Error implements AppError.
var _ AppError = (*Error)(nil)

func (e *Error) Error() string {
	if e.cause != nil {
		return e.message + ": " + e.cause.Error()
	}
	return e.message
}

func (e *Error) Kind() Kind           { return e.kind }
func (e *Error) Code() Code           { return e.code }
func (e *Error) Message() string      { return e.message }
func (e *Error) Unwrap() error        { return e.cause }
func (e *Error) Meta() map[string]any { return e.meta }

// WithMeta returns a new Error with the metadata added (immutable).
func (e *Error) WithMeta(key string, value any) *Error {
	meta := make(map[string]any, len(e.meta)+1)
	maps.Copy(meta, e.meta)
	meta[key] = value
	return &Error{kind: e.kind, code: e.code, message: e.message, cause: e.cause, meta: meta}
}

// WithCause returns a new Error with the cause added (immutable).
func (e *Error) WithCause(cause error) *Error {
	return &Error{kind: e.kind, code: e.code, message: e.message, cause: cause, meta: e.meta}
}

// Is allows comparison with errors.Is() — compares by Code.
// Works with both *Error and *Definition as target.
func (e *Error) Is(target error) bool {
	if t, ok := target.(AppError); ok {
		return e.code == t.Code()
	}
	return false
}
