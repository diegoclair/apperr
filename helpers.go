package apperr

// IsKind checks if an error implements AppError with the specified Kind.
// Works with both *Definition and *Error.
func IsKind(err error, kind Kind) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Kind() == kind
	}
	return false
}

func IsNotFound(err error) bool       { return IsKind(err, KindNotFound) }
func IsConflict(err error) bool       { return IsKind(err, KindConflict) }
func IsAuthentication(err error) bool  { return IsKind(err, KindAuthentication) }
func IsAuthorization(err error) bool   { return IsKind(err, KindAuthorization) }
func IsValidation(err error) bool      { return IsKind(err, KindValidation) }
func IsInternal(err error) bool        { return IsKind(err, KindInternal) }
func IsRateLimited(err error) bool     { return IsKind(err, KindRateLimited) }

// HasCode checks if the error has a specific Code.
func HasCode(err error, code Code) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Code() == code
	}
	return false
}

// GetCode extracts the Code from an error. Returns "" if not an AppError.
func GetCode(err error) Code {
	if appErr, ok := err.(AppError); ok {
		return appErr.Code()
	}
	return ""
}

// AsAppError attempts to convert an error to AppError (works with *Definition and *Error).
func AsAppError(err error) (AppError, bool) {
	appErr, ok := err.(AppError)
	return appErr, ok
}

// GetMeta extracts the meta from an error. Returns nil if it's a *Definition or non-AppError.
// Only *Error carries meta.
func GetMeta(err error) map[string]any {
	if e, ok := err.(*Error); ok {
		return e.Meta()
	}
	return nil
}
