package session

import "testing"

func TestBaseClientCloneSnapshotsContext(t *testing.T) {
	base := NewBaseClient("http://example.com")
	ctx := NewDefaultHeaderContext()
	ctx.Set("X-Mp-Namespace", "alpha")
	base.AttachContext(ctx)

	clone := base.Clone()

	ctx.Set("X-Mp-Namespace", "beta")
	if got := clone.GetContextValues()["X-Mp-Namespace"]; len(got) == 0 || got[0] != "alpha" {
		t.Fatalf("clone should snapshot header context, got %#v", got)
	}
}

func TestBaseClientWithContextDoesNotMutateOriginal(t *testing.T) {
	base := NewBaseClient("http://example.com")
	ctx := NewDefaultHeaderContext()
	ctx.Set("X-Mp-Application", "app-001")

	derived := base.WithContext(ctx)

	if got := base.GetContextValues().Get("X-Mp-Application"); got != "" {
		t.Fatalf("base client should remain unchanged, got %q", got)
	}
	if got := derived.GetContextValues().Get("X-Mp-Application"); got != "app-001" {
		t.Fatalf("derived client should contain context, got %q", got)
	}
}
