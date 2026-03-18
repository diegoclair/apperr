package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := &Error{kind: KindValidation, code: "TEST", message: "test error"}
		if got := err.Error(); got != "test error" {
			t.Errorf("Error() = %q, want %q", got, "test error")
		}
	})

	t.Run("with cause", func(t *testing.T) {
		cause := fmt.Errorf("database timeout")
		err := &Error{kind: KindInternal, code: "TEST", message: "internal error", cause: cause}
		want := "internal error: database timeout"
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestError_Accessors(t *testing.T) {
	err := &Error{
		kind:    KindNotFound,
		code:    "USER_NOT_FOUND",
		message: "user not found",
		cause:   fmt.Errorf("sql: no rows"),
		meta:    map[string]any{"id": "123"},
	}

	if err.Kind() != KindNotFound {
		t.Errorf("Kind() = %v, want %v", err.Kind(), KindNotFound)
	}
	if err.Code() != "USER_NOT_FOUND" {
		t.Errorf("Code() = %v, want %v", err.Code(), "USER_NOT_FOUND")
	}
	if err.Message() != "user not found" {
		t.Errorf("Message() = %v, want %v", err.Message(), "user not found")
	}
	if err.Unwrap() == nil {
		t.Error("Unwrap() should not be nil")
	}
	if err.Meta()["id"] != "123" {
		t.Errorf("Meta()[\"id\"] = %v, want %v", err.Meta()["id"], "123")
	}
}

func TestError_WithMeta(t *testing.T) {
	original := &Error{kind: KindValidation, code: "TEST", message: "test"}
	withMeta := original.WithMeta("key", "value")

	// Should create a new instance
	if original == withMeta {
		t.Error("WithMeta should return a new instance")
	}

	// Original should not have meta
	if original.Meta() != nil {
		t.Error("original should not have meta")
	}

	// New instance should have meta
	if withMeta.Meta()["key"] != "value" {
		t.Errorf("meta[\"key\"] = %v, want %v", withMeta.Meta()["key"], "value")
	}

	// Chained WithMeta should accumulate
	withTwo := withMeta.WithMeta("key2", "value2")
	if withTwo.Meta()["key"] != "value" || withTwo.Meta()["key2"] != "value2" {
		t.Error("chained WithMeta should accumulate metadata")
	}

	// Previous should not be affected by chain
	if _, exists := withMeta.Meta()["key2"]; exists {
		t.Error("previous instance should not be affected by chained WithMeta")
	}
}

func TestError_WithCause(t *testing.T) {
	original := &Error{kind: KindInternal, code: "TEST", message: "test"}
	cause := fmt.Errorf("db error")
	withCause := original.WithCause(cause)

	if original == withCause {
		t.Error("WithCause should return a new instance")
	}
	if original.Unwrap() != nil {
		t.Error("original should not have cause")
	}
	if !errors.Is(withCause.Unwrap(), cause) {
		t.Error("new instance should have the cause")
	}
}

func TestError_Is(t *testing.T) {
	errA := &Error{kind: KindNotFound, code: "USER_NOT_FOUND", message: "user not found"}
	errB := &Error{kind: KindNotFound, code: "USER_NOT_FOUND", message: "different message"}
	errC := &Error{kind: KindNotFound, code: "COMPANY_NOT_FOUND", message: "company not found"}

	t.Run("same code matches", func(t *testing.T) {
		if !errors.Is(errA, errB) {
			t.Error("errors with same code should match")
		}
	})

	t.Run("different code does not match", func(t *testing.T) {
		if errors.Is(errA, errC) {
			t.Error("errors with different code should not match")
		}
	})

	t.Run("matches Definition with same code", func(t *testing.T) {
		def := Define(KindNotFound, "USER_NOT_FOUND", "user not found")
		if !errors.Is(errA, def) {
			t.Error("Error should match Definition with same code")
		}
	})

	t.Run("does not match non-AppError", func(t *testing.T) {
		plainErr := fmt.Errorf("plain error")
		if errors.Is(errA, plainErr) {
			t.Error("should not match non-AppError")
		}
	})
}

func TestError_ImplementsError(t *testing.T) {
	// Compile-time guarantee via AppError interface check (var _ AppError = (*Error)(nil))
	var err error = &Error{kind: KindValidation, code: "TEST", message: "test"}
	if err.Error() != "test" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test")
	}
}
