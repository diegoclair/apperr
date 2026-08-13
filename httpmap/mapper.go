// Package httpmap provides Kind → HTTP status code mapping for REST APIs.
// Import this package only in your transport/HTTP layer — the core apperr
// package has zero dependency on net/http.
package httpmap

import (
	"net/http"

	"github.com/diegoclair/apperr"
)

// DefaultStatusMap is the default Kind → HTTP status mapping.
var DefaultStatusMap = map[apperr.Kind]int{
	apperr.KindValidation:     http.StatusBadRequest,          // 400
	apperr.KindAuthentication: http.StatusUnauthorized,        // 401
	apperr.KindAuthorization:  http.StatusForbidden,           // 403
	apperr.KindNotFound:       http.StatusNotFound,            // 404
	apperr.KindConflict:       http.StatusConflict,            // 409
	apperr.KindRateLimited:    http.StatusTooManyRequests,     // 429
	apperr.KindInternal:       http.StatusInternalServerError, // 500
	apperr.KindUnavailable:    http.StatusServiceUnavailable,  // 503
}

// ErrorResponse is the JSON structure returned by REST APIs.
// Maintains compatibility with common error formats (message, status_code, error)
// plus new fields (code, meta) for i18n and programmatic handling.
type ErrorResponse struct {
	Message    string         `json:"message"`
	StatusCode int            `json:"status_code"`
	Error      string         `json:"error"`
	Code       string         `json:"code,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
}

// ToHTTP converts an error to (statusCode, ErrorResponse).
// Accepts both *Definition (direct return) and *Error (with meta/cause).
// If the error is not an AppError, returns 500 with a generic message.
func ToHTTP(err error) (int, ErrorResponse) {
	appErr, ok := apperr.AsAppError(err)
	if !ok {
		return http.StatusInternalServerError, ErrorResponse{
			Message:    "internal server error",
			StatusCode: http.StatusInternalServerError,
			Error:      http.StatusText(http.StatusInternalServerError),
			Code:       string(apperr.ErrInternal.Code()),
		}
	}

	status := StatusFromKind(appErr.Kind())

	return status, ErrorResponse{
		Message:    appErr.Message(),
		StatusCode: status,
		Error:      http.StatusText(status),
		Code:       string(appErr.Code()),
		Meta:       apperr.GetMeta(err),
	}
}

// StatusFromKind returns the HTTP status code for a Kind.
// Returns 500 if the Kind is not mapped.
func StatusFromKind(kind apperr.Kind) int {
	if status, ok := DefaultStatusMap[kind]; ok {
		return status
	}
	return http.StatusInternalServerError
}
