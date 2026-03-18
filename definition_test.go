package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestDefinition_Error(t *testing.T) {
	def := Define(KindValidation, "TEST_CODE", "test message")
	if got := def.Error(); got != "test message" {
		t.Errorf("Error() = %q, want %q", got, "test message")
	}
}

func TestDefinition_Accessors(t *testing.T) {
	def := Define(KindAuthentication, "AUTH_FAILED", "auth failed")

	if def.Kind() != KindAuthentication {
		t.Errorf("Kind() = %v, want %v", def.Kind(), KindAuthentication)
	}
	if def.Code() != "AUTH_FAILED" {
		t.Errorf("Code() = %v, want %v", def.Code(), "AUTH_FAILED")
	}
	if def.Message() != "auth failed" {
		t.Errorf("Message() = %v, want %v", def.Message(), "auth failed")
	}
}

func TestDefinition_DirectReturn(t *testing.T) {
	// Definition can be used as error directly (compile-time guarantee via AppError interface)
	var err error = Define(KindNotFound, "NOT_FOUND", "not found")
	if err.Error() != "not found" {
		t.Errorf("Error() = %q, want %q", err.Error(), "not found")
	}
}

func TestDefinition_Wrap(t *testing.T) {
	def := Define(KindInternal, "INTERNAL", "internal error")
	cause := fmt.Errorf("connection refused")
	wrapped := def.Wrap(cause)

	if wrapped.Kind() != KindInternal {
		t.Errorf("Kind() = %v, want %v", wrapped.Kind(), KindInternal)
	}
	if wrapped.Code() != "INTERNAL" {
		t.Errorf("Code() = %v, want %v", wrapped.Code(), "INTERNAL")
	}
	if !errors.Is(wrapped, cause) {
		t.Error("wrapped error should contain the cause")
	}
	want := "internal error: connection refused"
	if got := wrapped.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestDefinition_WithMessage(t *testing.T) {
	def := Define(KindValidation, "VALIDATION", "validation error")
	custom := def.WithMessage("email is invalid")

	if custom.Message() != "email is invalid" {
		t.Errorf("Message() = %q, want %q", custom.Message(), "email is invalid")
	}
	// Code should be preserved
	if custom.Code() != "VALIDATION" {
		t.Errorf("Code() = %v, want %v", custom.Code(), "VALIDATION")
	}
	// Original should not be affected
	if def.Message() != "validation error" {
		t.Error("original Definition should not be affected")
	}
}

func TestDefinition_WithMeta(t *testing.T) {
	def := Define(KindAuthentication, "AUTH_FAILED", "auth failed")
	withMeta := def.WithMeta("remaining_attempts", 2)

	if withMeta.Meta()["remaining_attempts"] != 2 {
		t.Errorf("meta = %v, want remaining_attempts=2", withMeta.Meta())
	}
	if withMeta.Kind() != KindAuthentication {
		t.Errorf("Kind() = %v, want %v", withMeta.Kind(), KindAuthentication)
	}
	if withMeta.Code() != "AUTH_FAILED" {
		t.Errorf("Code() = %v, want %v", withMeta.Code(), "AUTH_FAILED")
	}
}

func TestDefinition_Is(t *testing.T) {
	defA := Define(KindNotFound, "USER_NOT_FOUND", "user not found")
	defB := Define(KindNotFound, "USER_NOT_FOUND", "different message")
	defC := Define(KindNotFound, "COMPANY_NOT_FOUND", "company not found")

	t.Run("same code matches", func(t *testing.T) {
		if !errors.Is(defA, defB) {
			t.Error("Definitions with same code should match")
		}
	})

	t.Run("different code does not match", func(t *testing.T) {
		if errors.Is(defA, defC) {
			t.Error("Definitions with different code should not match")
		}
	})

	t.Run("matches Error with same code", func(t *testing.T) {
		err := &Error{kind: KindNotFound, code: "USER_NOT_FOUND", message: "user not found"}
		if !errors.Is(defA, err) {
			t.Error("Definition should match Error with same code")
		}
	})

	t.Run("does not match non-AppError", func(t *testing.T) {
		plainErr := fmt.Errorf("plain error")
		if errors.Is(defA, plainErr) {
			t.Error("should not match non-AppError")
		}
	})
}

func TestDefinition_IsWithWrappedError(t *testing.T) {
	def := Define(KindNotFound, "USER_NOT_FOUND", "user not found")
	cause := fmt.Errorf("sql: no rows")
	wrapped := def.Wrap(cause)

	// errors.Is should match the wrapped *Error against the Definition
	if !errors.Is(wrapped, def) {
		t.Error("wrapped error should match its Definition")
	}

	// errors.Is should also find the original cause
	if !errors.Is(wrapped, cause) {
		t.Error("wrapped error should also match the original cause via Unwrap")
	}
}

func TestDefinition_WithMetaChained(t *testing.T) {
	def := Define(KindValidation, "VALIDATION", "validation error")
	withOne := def.WithMeta("field", "email")
	withTwo := withOne.WithMeta("rule", "required")

	// First should only have "field"
	if _, exists := withOne.Meta()["rule"]; exists {
		t.Error("first WithMeta result should not have 'rule'")
	}

	// Second should have both
	if withTwo.Meta()["field"] != "email" || withTwo.Meta()["rule"] != "required" {
		t.Error("chained WithMeta should accumulate both keys")
	}
}
