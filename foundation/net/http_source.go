package net

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

// HTTPSource preserves the transport peer separately from the resolved caller.
// Kind is direct, trusted-proxy, or invalid-forwarded (peer fallback).
type HTTPSource struct {
	Address     string
	PeerAddress string
	Kind        string
}

type httpSourceKey struct{}

// HTTPSourceResolver trusts only explicitly configured prefixes or DNS peers.
// DNS names must belong to an operator-controlled service discovery namespace.
// Addresses are resolved for each request so replacement proxies do not leave
// stale IP grants behind. Resolution failure never grants trust.
type HTTPSourceResolver struct {
	prefixes []netip.Prefix
	hosts    []string
	lookup   func(context.Context, string, string) ([]netip.Addr, error)
}

func NewHTTPSourceResolver(cidrs, hosts string) (*HTTPSourceResolver, error) {
	r := &HTTPSourceResolver{lookup: net.DefaultResolver.LookupNetIP}
	for _, value := range strings.Split(cidrs, ",") {
		if value = strings.TrimSpace(value); value != "" {
			prefix, err := netip.ParsePrefix(value)
			if err != nil || prefix.Bits() == 0 {
				return nil, fmt.Errorf("invalid trusted proxy prefix %q", value)
			}
			r.prefixes = append(r.prefixes, prefix.Masked())
		}
	}
	for _, value := range strings.Split(hosts, ",") {
		if value = strings.TrimSpace(value); value == "" {
			continue
		}
		if len(value) > 253 || strings.ContainsAny(value, ":/\\ \t\r\n*") {
			return nil, fmt.Errorf("invalid trusted proxy hostname %q", value)
		}
		for _, label := range strings.Split(value, ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return nil, fmt.Errorf("invalid trusted proxy hostname %q", value)
			}
			for _, ch := range strings.ToLower(label) {
				if ch != '-' && (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') {
					return nil, fmt.Errorf("invalid trusted proxy hostname %q", value)
				}
			}
		}
		r.hosts = append(r.hosts, value)
	}
	return r, nil
}

func peerIP(value string) (netip.Addr, bool) {
	value = strings.TrimSpace(value)
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	ip, err := netip.ParseAddr(value)
	return ip.Unmap(), err == nil && ip.Zone() == ""
}

func directHTTPSource(req *http.Request) HTTPSource {
	if req == nil {
		return HTTPSource{}
	}
	source := HTTPSource{PeerAddress: strings.TrimSpace(req.RemoteAddr), Kind: "direct"}
	if ip, ok := peerIP(source.PeerAddress); ok {
		source.Address = ip.String()
	}
	return source
}

func (r *HTTPSourceResolver) Resolve(req *http.Request) HTTPSource {
	source := directHTTPSource(req)
	if r == nil || source.Address == "" {
		return source
	}
	values := req.Header.Values("X-Forwarded-For")
	if len(values) == 0 {
		return source
	}
	peer := netip.MustParseAddr(source.Address)
	// Resolve names lazily, at most once per request, under one bounded budget.
	var resolved []netip.Addr
	lookedUp := false
	trusted := func(ip netip.Addr) bool {
		for _, prefix := range r.prefixes {
			if prefix.Contains(ip) {
				return true
			}
		}
		if !lookedUp {
			lookedUp = true
			ctx, cancel := context.WithTimeout(req.Context(), 500*time.Millisecond)
			defer cancel()
			for _, host := range r.hosts {
				addresses, err := r.lookup(ctx, "ip", host)
				if err == nil {
					resolved = append(resolved, addresses...)
				}
			}
		}
		for _, address := range resolved {
			if address.Unmap() == ip {
				return true
			}
		}
		return false
	}
	if !trusted(peer) {
		return source
	}
	chain := strings.Join(values, ",")
	parts := strings.Split(chain, ",")
	invalid := func() HTTPSource { source.Kind = "invalid-forwarded"; return source }
	if len(chain) > 4096 || len(parts) > 32 {
		return invalid()
	}
	addresses := make([]netip.Addr, len(parts))
	for i, part := range parts {
		ip, err := netip.ParseAddr(strings.TrimSpace(part))
		if err != nil || ip.Zone() != "" {
			return invalid()
		}
		addresses[i] = ip.Unmap()
	}
	for i := len(addresses) - 1; i >= 0; i-- {
		source.Address = addresses[i].String()
		if !trusted(addresses[i]) {
			break
		}
	}
	source.Kind = "trusted-proxy"
	return source
}

func HTTPSourceFromContext(ctx context.Context) (HTTPSource, bool) {
	if ctx == nil {
		return HTTPSource{}, false
	}
	source, ok := ctx.Value(httpSourceKey{}).(HTTPSource)
	return source, ok
}

// RequestHTTPSource uses only server-bound metadata, otherwise the socket peer.
func RequestHTTPSource(req *http.Request) HTTPSource {
	if req == nil {
		return HTTPSource{}
	}
	if source, ok := HTTPSourceFromContext(req.Context()); ok {
		return source
	}
	return directHTTPSource(req)
}

func (r *HTTPSourceResolver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		source := r.Resolve(req)
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), httpSourceKey{}, source)))
	})
}

// SetHTTPSourceHeaders discards caller-provided alternatives and propagates
// only metadata bound by an ingress resolver, never a raw request header.
func SetHTTPSourceHeaders(headers http.Header, ctx context.Context) {
	for _, name := range []string{"Forwarded", "X-Forwarded-For", "X-Real-IP"} {
		headers.Del(name)
	}
	if source, ok := HTTPSourceFromContext(ctx); ok && source.Address != "" {
		headers.Set("X-Forwarded-For", source.Address)
	}
}
