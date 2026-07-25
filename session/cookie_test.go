package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultCookieOptionsAreOpaqueSafeDefaults(t *testing.T) {
	opts := DefaultCookieOptions()
	if opts.Name != SessionToken {
		t.Fatalf("Name=%q want=%q", opts.Name, SessionToken)
	}
	if opts.Path != "/" || opts.Domain != "" {
		t.Fatalf("unexpected path/domain: %+v", opts)
	}
	if !opts.Secure || !opts.HttpOnly || opts.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected Secure+HttpOnly+Strict, got %+v", opts)
	}
	if opts.MaxAge != 0 {
		t.Fatalf("default MaxAge should be 0 (session cookie), got %d", opts.MaxAge)
	}
}

func TestWriteCookieAppliesConfigurableAttributes(t *testing.T) {
	recorder := httptest.NewRecorder()
	opts := CookieOptions{
		Name:     "app_session",
		Path:     "/portal",
		Domain:   "example.test",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	}

	WriteCookie(recorder, "opaque-session-id", opts)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "app_session" || c.Value != "opaque-session-id" {
		t.Fatalf("unexpected name/value: %+v", c)
	}
	if c.Path != "/portal" || c.Domain != "example.test" {
		t.Fatalf("unexpected path/domain: %+v", c)
	}
	if !c.Secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.MaxAge != 3600 {
		t.Fatalf("unexpected attributes: %+v", c)
	}
}

func TestClearCookieExpiresNamedCookie(t *testing.T) {
	recorder := httptest.NewRecorder()
	opts := CookieOptions{
		Name:     "app_session",
		Path:     "/portal",
		Domain:   "example.test",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	ClearCookie(recorder, opts)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "app_session" || c.Value != "" || c.MaxAge >= 0 {
		t.Fatalf("expected expired empty cookie, got %+v", c)
	}
	if c.Path != "/portal" || c.Domain != "example.test" {
		t.Fatalf("clear must preserve path/domain, got %+v", c)
	}
	if !c.Expires.Equal(time.Unix(0, 0)) && c.Expires.After(time.Unix(1, 0)) {
		t.Fatalf("expected expire at epoch, got %v", c.Expires)
	}
}

func TestReadCookieValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.test/", nil)
	req.AddCookie(&http.Cookie{Name: "app_session", Value: "opaque-id"})

	if got := ReadCookieValue(req, "app_session"); got != "opaque-id" {
		t.Fatalf("got %q want opaque-id", got)
	}
	if got := ReadCookieValue(req, "missing"); got != "" {
		t.Fatalf("missing cookie should be empty, got %q", got)
	}
	if got := ReadCookieValue(nil, "app_session"); got != "" {
		t.Fatalf("nil request should be empty, got %q", got)
	}
}

func TestWriteSessionTokenToCookieUsesDefaultOptions(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteSessionTokenToCookie(recorder, "opaque-default")

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != SessionToken || c.Value != "opaque-default" {
		t.Fatalf("unexpected cookie: %+v", c)
	}
	if c.Path != "/" || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("default attributes mismatch: %+v", c)
	}
}

func TestClearSessionTokenCookieExpiresCookie(t *testing.T) {
	recorder := httptest.NewRecorder()

	ClearSessionTokenCookie(recorder)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != SessionToken {
		t.Fatalf("cookie name=%q want=%q", cookie.Name, SessionToken)
	}
	if cookie.Value != "" || cookie.MaxAge >= 0 || cookie.Expires.After(time.Unix(1, 0)) {
		t.Fatalf("expected expired empty cookie, got %+v", cookie)
	}
	if cookie.Path != "/" || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected cookie attributes: %+v", cookie)
	}
}
