package session

import (
	"net/http"
	"time"
)

// CookieOptions configures opaque session-id cookie attributes.
// The cookie value must remain an opaque application session identifier only;
// it must not carry user, role, scope, JWT, or refresh token material.
type CookieOptions struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
	// MaxAge is cookie Max-Age in seconds.
	// Zero means session cookie (no Max-Age attribute).
	// Negative means delete the cookie.
	MaxAge int
}

// DefaultCookieOptions returns conservative defaults for opaque session cookies.
// Product session policy (BFF vs pure public client) must not be hard-coded here;
// callers override Name/Domain/MaxAge as needed.
func DefaultCookieOptions() CookieOptions {
	return CookieOptions{
		Name:     SessionToken,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   0,
	}
}

func (o CookieOptions) normalized() CookieOptions {
	if o.Name == "" {
		o.Name = SessionToken
	}
	if o.Path == "" {
		o.Path = "/"
	}
	if o.SameSite == 0 {
		o.SameSite = http.SameSiteStrictMode
	}
	return o
}

// ReadCookieValue reads the named cookie value from the request.
func ReadCookieValue(req *http.Request, name string) string {
	if req == nil || name == "" {
		return ""
	}
	cookie, err := req.Cookie(name)
	if err != nil || cookie == nil {
		return ""
	}
	return cookie.Value
}

// WriteCookie writes a cookie using the provided options.
// value should be an opaque session id when used for app sessions.
func WriteCookie(res http.ResponseWriter, value string, opts CookieOptions) {
	if res == nil {
		return
	}
	opts = opts.normalized()
	cookie := http.Cookie{
		Name:     opts.Name,
		Value:    value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		HttpOnly: opts.HttpOnly,
		Secure:   opts.Secure,
		SameSite: opts.SameSite,
	}
	if opts.MaxAge != 0 {
		cookie.MaxAge = opts.MaxAge
		if opts.MaxAge < 0 {
			cookie.Expires = time.Unix(0, 0)
			cookie.Value = ""
		}
	}
	http.SetCookie(res, &cookie)
}

// ClearCookie expires the cookie identified by opts.Name (and Path/Domain).
func ClearCookie(res http.ResponseWriter, opts CookieOptions) {
	opts = opts.normalized()
	opts.MaxAge = -1
	WriteCookie(res, "", opts)
}

// ReadSessionTokenFromCookie reads the default session cookie value.
// Prefer ReadCookieValue with explicit CookieOptions when product policy is known.
func ReadSessionTokenFromCookie(req *http.Request) string {
	return ReadCookieValue(req, DefaultCookieOptions().Name)
}

// WriteSessionTokenToCookie writes the default opaque session cookie attributes.
func WriteSessionTokenToCookie(res http.ResponseWriter, sessionToken string) {
	WriteCookie(res, sessionToken, DefaultCookieOptions())
}

// ClearSessionTokenCookie clears the default session cookie.
func ClearSessionTokenCookie(res http.ResponseWriter) {
	ClearCookie(res, DefaultCookieOptions())
}
