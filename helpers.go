package apperr

import "errors"

// IsKind checks if an error (or any error in its chain) is an AppError with the specified Kind.
// Uses errors.As under the hood, so it works with wrapped errors (e.g., fmt.Errorf("...: %w", err)).
func IsKind(err error, kind Kind) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Kind() == kind
	}
	return false
}

func IsNotFound(err error) bool      { return IsKind(err, KindNotFound) }
func IsConflict(err error) bool      { return IsKind(err, KindConflict) }
func IsAuthentication(err error) bool { return IsKind(err, KindAuthentication) }
func IsAuthorization(err error) bool  { return IsKind(err, KindAuthorization) }
func IsValidation(err error) bool     { return IsKind(err, KindValidation) }
func IsInternal(err error) bool       { return IsKind(err, KindInternal) }
func IsRateLimited(err error) bool    { return IsKind(err, KindRateLimited) }

// HasCode checks if the error (or any error in its chain) has a specific Code.
func HasCode(err error, code Code) bool {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Code() == code
	}
	return false
}

// GetCode extracts the Code from an error (or any error in its chain).
// Returns "" if no AppError is found.
func GetCode(err error) Code {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Code()
	}
	return ""
}

// AsAppError extracts the AppError from an error chain.
// Works with both *Definition and *Error, and traverses wrapped errors.
func AsAppError(err error) (AppError, bool) {
	var appErr AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

// GetMeta extracts the meta from an error chain. Returns nil if no *Error is found.
// Only *Error carries meta — *Definition and plain errors return nil.
func GetMeta(err error) map[string]any {
	var e *Error
	if errors.As(err, &e) {
		return e.Meta()
	}
	return nil
}
