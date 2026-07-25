package session

import "testing"

func TestFormatBearerAuthorization(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "spaces", input: "   ", want: ""},
		{name: "raw token", input: "access-token", want: "Bearer access-token"},
		{name: "already bearer", input: "Bearer access-token", want: "Bearer access-token"},
		{name: "already bearer lower", input: "bearer access-token", want: "bearer access-token"},
		{name: "trim", input: "  access-token  ", want: "Bearer access-token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBearerAuthorization(tt.input); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestAttachBearerAndWithBearerInjectAuthorization(t *testing.T) {
	base := NewBaseClient("http://example.com")
	base.AttachBearer("token-a")
	if got := base.GetContextValues().Get(Authorization); got != "Bearer token-a" {
		t.Fatalf("AttachBearer header=%q", got)
	}

	derived := base.WithBearer("token-b")
	if got := derived.GetContextValues().Get(Authorization); got != "Bearer token-b" {
		t.Fatalf("WithBearer header=%q", got)
	}
	if got := base.GetContextValues().Get(Authorization); got != "Bearer token-a" {
		t.Fatalf("WithBearer must not mutate base, got %q", got)
	}
}

func TestWithAuthorizationExplicitBearerPassthrough(t *testing.T) {
	base := NewBaseClient("http://example.com")
	derived := base.WithAuthorization("Bearer explicit-token")
	if got := derived.GetContextValues().Get(Authorization); got != "Bearer explicit-token" {
		t.Fatalf("explicit bearer passthrough failed, got %q", got)
	}
}

func TestParseAndDetectAuthorizationSchemes(t *testing.T) {
	scheme, creds := ParseAuthorizationScheme("Bearer abc.def")
	if scheme != "Bearer" || creds != "abc.def" {
		t.Fatalf("bearer parse got scheme=%q creds=%q", scheme, creds)
	}
	if !IsBearerAuthorization("Bearer x") {
		t.Fatal("expected bearer")
	}
	if IsBearerAuthorization("Sig x") {
		t.Fatal("sig must not be treated as bearer")
	}
	if IsBearerAuthorization("not-a-scheme") {
		t.Fatal("invalid authorization must not match schemes")
	}
}
