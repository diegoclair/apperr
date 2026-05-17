package apperr

import "errors"

// IsKind checks if an error (or any error in its chain) is an AppError with the specified Kind.
// Uses errors.AsType under the hood, so it works with wrapped errors (e.g., fmt.Errorf("...: %w", err)).
func IsKind(err error, kind Kind) bool {
	if appErr, ok := errors.AsType[AppError](err); ok {
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
	if appErr, ok := errors.AsType[AppError](err); ok {
		return appErr.Code() == code
	}
	return false
}

// GetCode extracts the Code from an error (or any error in its chain).
// Returns "" if no AppError is found.
func GetCode(err error) Code {
	if appErr, ok := errors.AsType[AppError](err); ok {
		return appErr.Code()
	}
	return ""
}

// AsAppError extracts the AppError from an error chain.
// Works with both *Definition and *Error, and traverses wrapped errors.
func AsAppError(err error) (AppError, bool) {
	return errors.AsType[AppError](err)
}

// GetMeta extracts the meta from an error chain. Returns nil if no *Error is found.
// Only *Error carries meta — *Definition and plain errors return nil.
func GetMeta(err error) map[string]any {
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Meta()
	}
	return nil
}
