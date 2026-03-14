package net

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewDNSCacheHttpClientDoesNotMutateDefaultTransport(t *testing.T) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		t.Fatalf("expected default transport to be *http.Transport")
	}

	originalDial := defaultTransport.DialContext
	client := NewDNSCacheHttpClient()

	if client.Transport == nil {
		t.Fatalf("expected client transport to be initialized")
	}

	clientTransport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected client transport to be *http.Transport")
	}
	if clientTransport == defaultTransport {
		t.Fatalf("expected a cloned transport, got shared default transport")
	}
	if reflect.ValueOf(defaultTransport.DialContext).Pointer() != reflect.ValueOf(originalDial).Pointer() {
		t.Fatalf("expected default transport DialContext to remain unchanged")
	}
	if clientTransport.DialContext == nil {
		t.Fatalf("expected cloned transport to install custom DialContext")
	}
}
