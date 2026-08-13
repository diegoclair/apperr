package httpmap

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/diegoclair/apperr"
)

func TestToHTTP_Definition(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantMeta   bool
	}{
		{
			name:       "NotFound Definition",
			err:        apperr.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "Validation Definition",
			err:        apperr.ErrValidation,
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "Authentication Definition",
			err:        apperr.ErrUnauthenticated,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHENTICATED",
		},
		{
			name:       "Authorization Definition",
			err:        apperr.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantCode:   "FORBIDDEN",
		},
		{
			name:       "Conflict Definition",
			err:        apperr.ErrConflict,
			wantStatus: http.StatusConflict,
			wantCode:   "CONFLICT",
		},
		{
			name:       "Internal Definition",
			err:        apperr.ErrInternal,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
		{
			name:       "RateLimited Definition",
			err:        apperr.ErrRateLimited,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   "RATE_LIMITED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, resp := ToHTTP(tt.err)

			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if resp.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", resp.Code, tt.wantCode)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("resp.StatusCode = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if resp.Meta != nil {
				t.Errorf("meta should be nil for Definition, got %v", resp.Meta)
			}
		})
	}
}

func TestToHTTP_ErrorWithMeta(t *testing.T) {
	def := apperr.Define(apperr.KindAuthentication, "AUTH_FAILED", "auth failed")
	err := def.WithMeta("remaining_attempts", 2)

	status, resp := ToHTTP(err)

	if status != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", status, http.StatusUnauthorized)
	}
	if resp.Code != "AUTH_FAILED" {
		t.Errorf("code = %q, want %q", resp.Code, "AUTH_FAILED")
	}
	if resp.Message != "auth failed" {
		t.Errorf("message = %q, want %q", resp.Message, "auth failed")
	}
	if resp.Meta["remaining_attempts"] != 2 {
		t.Errorf("meta = %v, want remaining_attempts=2", resp.Meta)
	}
}

func TestToHTTP_ErrorWithCause(t *testing.T) {
	def := apperr.Define(apperr.KindInternal, "DB_ERROR", "database error")
	err := def.Wrap(fmt.Errorf("connection refused"))

	status, resp := ToHTTP(err)

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", status, http.StatusInternalServerError)
	}
	// The cause should NOT be exposed in the response message
	if resp.Message != "database error" {
		t.Errorf("message = %q, want %q (cause should not be in message)", resp.Message, "database error")
	}
}

func TestToHTTP_PlainError(t *testing.T) {
	err := fmt.Errorf("random error")

	status, resp := ToHTTP(err)

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", status, http.StatusInternalServerError)
	}
	if resp.Code != "INTERNAL_ERROR" {
		t.Errorf("code = %q, want %q", resp.Code, "INTERNAL_ERROR")
	}
	if resp.Message != "internal server error" {
		t.Errorf("message = %q, want generic message", resp.Message)
	}
}

func TestStatusFromKind(t *testing.T) {
	tests := []struct {
		kind apperr.Kind
		want int
	}{
		{apperr.KindValidation, http.StatusBadRequest},
		{apperr.KindAuthentication, http.StatusUnauthorized},
		{apperr.KindAuthorization, http.StatusForbidden},
		{apperr.KindNotFound, http.StatusNotFound},
		{apperr.KindConflict, http.StatusConflict},
		{apperr.KindRateLimited, http.StatusTooManyRequests},
		{apperr.KindInternal, http.StatusInternalServerError},
		{apperr.KindUnavailable, http.StatusServiceUnavailable},
		{apperr.Kind(99), http.StatusInternalServerError}, // unmapped
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			if got := StatusFromKind(tt.kind); got != tt.want {
				t.Errorf("StatusFromKind(%v) = %d, want %d", tt.kind, got, tt.want)
			}
		})
	}
}
