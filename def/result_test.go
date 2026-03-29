package def

import (
	"errors"
	"testing"
)

func TestToStdErrorNilHandling(t *testing.T) {
	if got := ToStdError(nil); got != nil {
		t.Fatalf("ToStdError(nil) = %v, want nil", got)
	}

	var typedNilErr *Error
	if got := ToStdError(typedNilErr); got != nil {
		t.Fatalf("ToStdError(typed nil) = %v, want nil", got)
	}

	cdErr := NewError(Unexpected, "boom")
	got := ToStdError(cdErr)
	if got == nil {
		t.Fatal("ToStdError(non-nil) returned nil")
	}

	var target *Error
	if !errors.As(got, &target) {
		t.Fatalf("ToStdError(non-nil) did not preserve *Error type: %T", got)
	}
}

func TestErrorMethodsAreNilSafe(t *testing.T) {
	var err *Error

	if got := err.Error(); got != "<nil>" {
		t.Fatalf("nil Error() = %q, want %q", got, "<nil>")
	}

	if got := err.Unwrap(); got != nil {
		t.Fatalf("nil Unwrap() = %v, want nil", got)
	}

	if err.HasStackTrace() {
		t.Fatal("nil HasStackTrace() = true, want false")
	}

	if got := err.GetFullStackTrace(); got != "" {
		t.Fatalf("nil GetFullStackTrace() = %q, want empty string", got)
	}
}
