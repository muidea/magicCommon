package net

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

var resolver *Resolver
var initOnce sync.Once

func init() {
	initOnce.Do(func() {
		resolver = &Resolver{}

		options := ResolverRefreshOptions{}
		options.ClearUnused = true
		options.PersistOnFailure = false
		resolver.RefreshWithOptions(options)

		go func() {
			t := time.NewTicker(1 * time.Minute)
			defer t.Stop()
			for range t.C {
				resolver.Refresh(true)
			}
		}()
	})
}

func NewDNSCacheHttpClient() *http.Client {
	dialContext := func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			var dialer net.Dialer
			conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
			if err == nil {
				break
			}
		}
		return
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
	return &http.Client{}
}
