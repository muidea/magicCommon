package session

import "testing"

func TestBaseClientCloneSnapshotsContextAndAuthSecret(t *testing.T) {
	base := NewBaseClient("http://example.com")
	ctx := NewDefaultHeaderContext()
	ctx.Set("X-Mp-Namespace", "alpha")
	base.AttachContext(ctx)

	secret := &AuthSecret{Endpoint: "service", AuthToken: "token-a"}
	base.BindAuthSecret(secret)

	clone := base.Clone()

	ctx.Set("X-Mp-Namespace", "beta")
	secret.Endpoint = "changed"

	if got := clone.GetContextValues()["X-Mp-Namespace"]; len(got) == 0 || got[0] != "alpha" {
		t.Fatalf("clone should snapshot header context, got %#v", got)
	}
	if clone.sessionAuthSecret == nil || clone.sessionAuthSecret.Endpoint != "service" {
		t.Fatalf("clone should copy auth secret, got %#v", clone.sessionAuthSecret)
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
