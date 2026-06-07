package module

import "testing"

func TestRegisterEReturnsValidationError(t *testing.T) {
	if err := RegisterE(nil); err == nil {
		t.Fatalf("expected RegisterE to return validation error")
	}
}

func TestMustRegisterPanicsOnValidationError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected MustRegister to panic")
		}
	}()

	MustRegister(nil)
}
