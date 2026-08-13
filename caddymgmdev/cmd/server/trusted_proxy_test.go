package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"
)

func mustTrustedProxySet(t *testing.T, value string) *trustedProxySet {
	t.Helper()
	set, err := newTrustedProxySet(value)
	if err != nil {
		t.Fatalf("newTrustedProxySet(%q): %v", value, err)
	}
	return set
}

func forwardedRequest(remoteAddr string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "http://admin.example.test/", nil)
	request.RemoteAddr = remoteAddr
	request.Header.Set("X-Forwarded-Proto", "https")
	return request
}

func TestSecureRequestRejectsForwardedProtoFromUntrustedPeer(t *testing.T) {
	app := &App{trustedProxies: mustTrustedProxySet(t, "10.0.0.0/8")}
	request := forwardedRequest("192.0.2.10:4321")

	if app.isSecureRequest(request) {
		t.Fatal("untrusted peer controlled the detected request scheme")
	}
	if got := app.requestScheme(request); got != "http" {
		t.Fatalf("request scheme = %q, want http", got)
	}
}

func TestSecureRequestAcceptsForwardedProtoFromTrustedIPOrCIDR(t *testing.T) {
	for _, configured := range []string{"192.0.2.10", "192.0.2.0/24"} {
		t.Run(configured, func(t *testing.T) {
			app := &App{trustedProxies: mustTrustedProxySet(t, configured)}
			if !app.isSecureRequest(forwardedRequest("192.0.2.10:4321")) {
				t.Fatal("trusted peer's forwarded HTTPS scheme was ignored")
			}
		})
	}
}

func TestSecureRequestAcceptsForwardedProtoFromTrustedHostname(t *testing.T) {
	set := mustTrustedProxySet(t, "caddy-admin")
	lookupCalls := 0
	set.lookup = func(_ context.Context, network, host string) ([]netip.Addr, error) {
		lookupCalls++
		if network != "ip" || host != "caddy-admin" {
			t.Fatalf("lookup = (%q, %q), want (ip, caddy-admin)", network, host)
		}
		return []netip.Addr{netip.MustParseAddr("172.20.0.3")}, nil
	}
	app := &App{trustedProxies: set}

	if !app.isSecureRequest(forwardedRequest("172.20.0.3:4321")) {
		t.Fatal("trusted hostname address was rejected")
	}
	if !app.isSecureRequest(forwardedRequest("172.20.0.3:9876")) {
		t.Fatal("cached trusted hostname address was rejected")
	}
	if lookupCalls != 1 {
		t.Fatalf("hostname lookups = %d, want 1 due to cache", lookupCalls)
	}
}

func TestSecureRequestAlwaysAcceptsDirectTLS(t *testing.T) {
	app := &App{trustedProxies: mustTrustedProxySet(t, "")}
	request := httptest.NewRequest(http.MethodGet, "https://admin.example.test/", nil)
	request.RemoteAddr = "192.0.2.10:4321"

	if !app.isSecureRequest(request) {
		t.Fatal("direct TLS request was not detected as secure")
	}
}

func TestInsecureTransportExceptionRequiresLoopbackPeer(t *testing.T) {
	t.Setenv("CADDYMGM_ALLOW_INSECURE_HTTP", "false")
	app := &App{trustedProxies: mustTrustedProxySet(t, "")}
	spoofedHost := httptest.NewRequest(http.MethodPost, "http://localhost/api/auth/login", nil)
	spoofedHost.RemoteAddr = "192.0.2.10:4321"
	if app.secureSessionTransportAllowed(spoofedHost) {
		t.Fatal("non-loopback peer bypassed HTTPS by supplying a localhost Host header")
	}

	loopback := httptest.NewRequest(http.MethodPost, "http://localhost/api/auth/login", nil)
	loopback.RemoteAddr = "127.0.0.1:4321"
	if !app.secureSessionTransportAllowed(loopback) {
		t.Fatal("real loopback peer was not allowed to use local HTTP")
	}
}

func TestInsecureTransportCanBeExplicitlyEnabled(t *testing.T) {
	t.Setenv("CADDYMGM_ALLOW_INSECURE_HTTP", "true")
	app := &App{trustedProxies: mustTrustedProxySet(t, "")}
	request := httptest.NewRequest(http.MethodPost, "http://admin.example.test/api/auth/login", nil)
	request.RemoteAddr = "192.0.2.10:4321"

	if !app.secureSessionTransportAllowed(request) {
		t.Fatal("explicit insecure HTTP opt-in did not allow a remote HTTP login")
	}
	if app.isSecureRequest(request) {
		t.Fatal("insecure HTTP opt-in incorrectly marked the request as HTTPS")
	}
}

func TestTrustedProxyConfigurationRejectsInvalidEntry(t *testing.T) {
	if _, err := newTrustedProxySet("192.0.2.0/24,not/a/proxy"); err == nil {
		t.Fatal("invalid trusted proxy entry was accepted")
	}
}

func TestCSRFCookieSecureFlagDependsOnTrustedTransport(t *testing.T) {
	app := &App{trustedProxies: mustTrustedProxySet(t, "192.0.2.10")}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	trustedResponse := httptest.NewRecorder()
	app.ensureCSRFCookie(next).ServeHTTP(trustedResponse, forwardedRequest("192.0.2.10:4321"))
	if cookie := trustedResponse.Header().Get("Set-Cookie"); !strings.Contains(cookie, "; Secure;") {
		t.Fatalf("trusted HTTPS cookie is not Secure: %q", cookie)
	}

	untrustedResponse := httptest.NewRecorder()
	app.ensureCSRFCookie(next).ServeHTTP(untrustedResponse, forwardedRequest("198.51.100.10:4321"))
	if cookie := untrustedResponse.Header().Get("Set-Cookie"); strings.Contains(cookie, "; Secure;") {
		t.Fatalf("untrusted forwarded header set Secure transport state: %q", cookie)
	}
}

func TestTrustedHostnameCacheExpires(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	set := mustTrustedProxySet(t, "caddy-admin")
	set.now = func() time.Time { return now }
	lookupCalls := 0
	set.lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		lookupCalls++
		return []netip.Addr{netip.MustParseAddr("172.20.0.3")}, nil
	}

	set.contains(context.Background(), "172.20.0.3:4321")
	now = now.Add(time.Minute)
	set.contains(context.Background(), "172.20.0.3:4321")
	if lookupCalls != 2 {
		t.Fatalf("hostname lookups after cache expiry = %d, want 2", lookupCalls)
	}
}

func TestTrustedHostnameLookupDoesNotHoldCacheMutex(t *testing.T) {
	set := mustTrustedProxySet(t, "caddy-admin")
	lookupStarted := make(chan struct{})
	releaseLookup := make(chan struct{})
	var calls sync.Mutex
	lookupCalls := 0
	set.lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		calls.Lock()
		lookupCalls++
		call := lookupCalls
		calls.Unlock()
		if call == 1 {
			close(lookupStarted)
			<-releaseLookup
		}
		return []netip.Addr{netip.MustParseAddr("172.20.0.3")}, nil
	}

	firstDone := make(chan struct{})
	go func() {
		set.contains(context.Background(), "172.20.0.3:4321")
		close(firstDone)
	}()
	<-lookupStarted
	secondDone := make(chan struct{})
	go func() {
		set.contains(context.Background(), "172.20.0.3:4321")
		close(secondDone)
	}()
	select {
	case <-secondDone:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second trusted-proxy check blocked behind DNS resolution")
	}
	close(releaseLookup)
	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Fatal("first trusted-proxy check did not finish")
	}
}
