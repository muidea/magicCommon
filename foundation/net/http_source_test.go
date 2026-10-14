package net

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func TestHTTPSourceTrustBoundary(t *testing.T) {
	r, err := NewHTTPSourceResolver("192.0.2.0/24,::1/128", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, peer, chain, address, kind string }{
		{"direct spoof", "198.51.100.4:82", "203.0.113.1", "198.51.100.4", "direct"},
		{"multi hop and left spoof", "192.0.2.2:82", "198.51.100.99, 203.0.113.4,192.0.2.3", "203.0.113.4", "trusted-proxy"},
		{"IPv6", "[::1]:82", "2001:db8::4", "2001:db8::4", "trusted-proxy"},
		{"mapped", "[::ffff:192.0.2.2]:82", "::ffff:203.0.113.4", "203.0.113.4", "trusted-proxy"},
		{"invalid", "192.0.2.2:82", "garbage", "192.0.2.2", "invalid-forwarded"},
		{"oversized", "192.0.2.2:82", strings.Repeat("203.0.113.4,", 33), "192.0.2.2", "invalid-forwarded"},
		{"missing", "192.0.2.2:82", "", "192.0.2.2", "direct"},
		{"zoned", "192.0.2.2:82", "fe80::1%eth0", "192.0.2.2", "invalid-forwarded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.peer
			if tc.chain != "" {
				req.Header.Set("X-Forwarded-For", tc.chain)
			}
			req.Header.Set("X-Real-IP", "203.0.113.99")
			source := r.Resolve(req)
			if source.Address != tc.address || source.Kind != tc.kind || source.PeerAddress != tc.peer {
				t.Fatalf("source=%+v", source)
			}
		})
	}
}

func TestHTTPSourceDNSReplacementAndFailureDoNotLeaveStaleTrust(t *testing.T) {
	r, err := NewHTTPSourceResolver("", "proxy.internal")
	if err != nil {
		t.Fatal(err)
	}
	current := "192.0.2.1"
	r.lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		if current == "" {
			return nil, errors.New("unavailable")
		}
		return []netip.Addr{netip.MustParseAddr(current)}, nil
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:82"
	req.Header.Set("X-Forwarded-For", "203.0.113.4")
	if r.Resolve(req).Address != "203.0.113.4" {
		t.Fatal("current proxy not trusted")
	}
	for _, replacement := range []string{"192.0.2.2", ""} {
		current = replacement
		if r.Resolve(req).Address != "192.0.2.1" {
			t.Fatal("stale or failed DNS granted trust")
		}
	}
}

func TestHTTPSourceMiddlewareAndPropagationIgnoreUnboundHeaders(t *testing.T) {
	r, _ := NewHTTPSourceResolver("192.0.2.1/32", "")
	for _, address := range []string{"203.0.113.4", "198.51.100.4"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.0.2.1:82"
		req.Header.Set("X-Forwarded-For", address)
		r.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, bound *http.Request) {
			headers := http.Header{"Forwarded": {"for=forged"}, "X-Real-Ip": {"forged"}, "X-Forwarded-For": {"forged"}}
			SetHTTPSourceHeaders(headers, bound.Context())
			if GetHTTPRemoteAddress(bound) != address || headers.Get("X-Forwarded-For") != address || headers.Get("Forwarded") != "" || headers.Get("X-Real-IP") != "" {
				t.Fatal("source changed or spoof retained")
			}
		})).ServeHTTP(httptest.NewRecorder(), req)
		if GetHTTPRemoteAddress(req) != "192.0.2.1" {
			t.Fatal("unbound request trusted forwarding")
		}
	}
	headers := http.Header{"X-Forwarded-For": {"forged"}}
	SetHTTPSourceHeaders(headers, context.Background())
	if headers.Get("X-Forwarded-For") != "" {
		t.Fatal("background call inherited origin")
	}
}

func TestHTTPSourceRejectsUnsafeProxyConfiguration(t *testing.T) {
	for _, tc := range [][2]string{{"0.0.0.0/0", ""}, {"::/0", ""}, {"garbage", ""}, {"", "*"}, {"", "proxy:8080"}, {"", "proxy/evil"}} {
		if _, err := NewHTTPSourceResolver(tc[0], tc[1]); err == nil {
			t.Fatalf("accepted %v", tc)
		}
	}
}
