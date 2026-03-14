package helper

import (
	"context"
	"testing"
)

type contextKey string

func TestGetValueFromContext(t *testing.T) {
	key := contextKey("key")
	ctx := context.WithValue(context.Background(), key, "value")

	val, ok := GetValueFromContext[string](ctx, key)
	if !ok || val != "value" {
		t.Fatalf("expected context value to be returned")
	}
}

func TestGetValueFromNilContext(t *testing.T) {
	key := contextKey("key")
	var ctx context.Context

	val, ok := GetValueFromContext[string](ctx, key)
	if ok {
		t.Fatalf("expected nil context to return not found")
	}
	if val != "" {
		t.Fatalf("expected zero value for nil context, got %q", val)
	}
}
