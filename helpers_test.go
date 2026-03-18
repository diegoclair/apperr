package apperr

import (
	"fmt"
	"testing"
)

func TestIsKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		kind Kind
		want bool
	}{
		{"Definition match", Define(KindNotFound, "NF", "nf"), KindNotFound, true},
		{"Definition no match", Define(KindNotFound, "NF", "nf"), KindValidation, false},
		{"Error match", &Error{kind: KindConflict, code: "C", message: "c"}, KindConflict, true},
		{"Error no match", &Error{kind: KindConflict, code: "C", message: "c"}, KindInternal, false},
		{"wrapped Definition", fmt.Errorf("repo: %w", Define(KindNotFound, "NF", "nf")), KindNotFound, true},
		{"wrapped Error", fmt.Errorf("service: %w", &Error{kind: KindConflict, code: "C", message: "c"}), KindConflict, true},
		{"wrapped no match", fmt.Errorf("repo: %w", Define(KindNotFound, "NF", "nf")), KindInternal, false},
		{"double wrapped", fmt.Errorf("handler: %w", fmt.Errorf("service: %w", ErrNotFound)), KindNotFound, true},
		{"plain error", fmt.Errorf("plain"), KindInternal, false},
		{"nil error", nil, KindInternal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsKind(tt.err, tt.kind); got != tt.want {
				t.Errorf("IsKind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		check  func(error) bool
		want   bool
	}{
		{"IsNotFound with Definition", ErrNotFound, IsNotFound, true},
		{"IsNotFound with Error", &Error{kind: KindNotFound}, IsNotFound, true},
		{"IsNotFound wrong kind", ErrConflict, IsNotFound, false},

		{"IsConflict", ErrConflict, IsConflict, true},
		{"IsAuthentication", ErrUnauthenticated, IsAuthentication, true},
		{"IsAuthorization", ErrForbidden, IsAuthorization, true},
		{"IsValidation", ErrValidation, IsValidation, true},
		{"IsInternal", ErrInternal, IsInternal, true},
		{"IsRateLimited", ErrRateLimited, IsRateLimited, true},

		// wrapped errors
		{"IsNotFound wrapped", fmt.Errorf("repo: %w", ErrNotFound), IsNotFound, true},
		{"IsConflict wrapped", fmt.Errorf("svc: %w", ErrConflict), IsConflict, true},
		{"IsInternal wrapped", fmt.Errorf("handler: %w", ErrInternal), IsInternal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check(tt.err); got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestHasCode(t *testing.T) {
	def := Define(KindNotFound, "USER_NOT_FOUND", "user not found")
	err := def.WithMeta("id", "123")

	if !HasCode(def, "USER_NOT_FOUND") {
		t.Error("HasCode should match Definition code")
	}
	if !HasCode(err, "USER_NOT_FOUND") {
		t.Error("HasCode should match Error code")
	}
	if HasCode(def, "OTHER_CODE") {
		t.Error("HasCode should not match different code")
	}
	if HasCode(fmt.Errorf("plain"), "ANY") {
		t.Error("HasCode should return false for plain errors")
	}

	// wrapped
	wrapped := fmt.Errorf("repo: %w", def)
	if !HasCode(wrapped, "USER_NOT_FOUND") {
		t.Error("HasCode should match wrapped Definition code")
	}
}

func TestGetCode(t *testing.T) {
	def := Define(KindNotFound, "USER_NOT_FOUND", "user not found")

	if got := GetCode(def); got != "USER_NOT_FOUND" {
		t.Errorf("GetCode() = %v, want USER_NOT_FOUND", got)
	}
	if got := GetCode(fmt.Errorf("plain")); got != "" {
		t.Errorf("GetCode() = %v, want empty", got)
	}

	// wrapped
	wrapped := fmt.Errorf("repo: %w", def)
	if got := GetCode(wrapped); got != "USER_NOT_FOUND" {
		t.Errorf("GetCode(wrapped) = %v, want USER_NOT_FOUND", got)
	}
}

func TestAsAppError(t *testing.T) {
	t.Run("with Definition", func(t *testing.T) {
		def := Define(KindNotFound, "NF", "not found")
		appErr, ok := AsAppError(def)
		if !ok || appErr == nil {
			t.Error("AsAppError should succeed for Definition")
		}
		if appErr.Code() != "NF" {
			t.Errorf("Code() = %v, want NF", appErr.Code())
		}
	})

	t.Run("with Error", func(t *testing.T) {
		err := &Error{kind: KindInternal, code: "INT", message: "internal"}
		appErr, ok := AsAppError(err)
		if !ok || appErr == nil {
			t.Error("AsAppError should succeed for Error")
		}
	})

	t.Run("with plain error", func(t *testing.T) {
		_, ok := AsAppError(fmt.Errorf("plain"))
		if ok {
			t.Error("AsAppError should fail for plain error")
		}
	})

	t.Run("with wrapped Definition", func(t *testing.T) {
		wrapped := fmt.Errorf("repo: %w", ErrNotFound)
		appErr, ok := AsAppError(wrapped)
		if !ok || appErr == nil {
			t.Error("AsAppError should succeed for wrapped Definition")
		}
		if appErr.Code() != "NOT_FOUND" {
			t.Errorf("Code() = %v, want NOT_FOUND", appErr.Code())
		}
	})

	t.Run("with wrapped Error", func(t *testing.T) {
		err := ErrNotFound.WithMeta("id", "123")
		wrapped := fmt.Errorf("repo: %w", err)
		appErr, ok := AsAppError(wrapped)
		if !ok || appErr == nil {
			t.Error("AsAppError should succeed for wrapped Error")
		}
	})
}

func TestGetMeta(t *testing.T) {
	t.Run("from Error with meta", func(t *testing.T) {
		err := &Error{kind: KindValidation, code: "V", message: "v", meta: map[string]any{"field": "email"}}
		meta := GetMeta(err)
		if meta["field"] != "email" {
			t.Errorf("GetMeta() = %v, want field=email", meta)
		}
	})

	t.Run("from Definition (no meta)", func(t *testing.T) {
		def := Define(KindNotFound, "NF", "not found")
		meta := GetMeta(def)
		if meta != nil {
			t.Errorf("GetMeta() = %v, want nil for Definition", meta)
		}
	})

	t.Run("from plain error", func(t *testing.T) {
		meta := GetMeta(fmt.Errorf("plain"))
		if meta != nil {
			t.Errorf("GetMeta() = %v, want nil for plain error", meta)
		}
	})

	t.Run("from wrapped Error with meta", func(t *testing.T) {
		err := ErrNotFound.WithMeta("id", "123")
		wrapped := fmt.Errorf("repo: %w", err)
		meta := GetMeta(wrapped)
		if meta["id"] != "123" {
			t.Errorf("GetMeta(wrapped) = %v, want id=123", meta)
		}
	})
}
