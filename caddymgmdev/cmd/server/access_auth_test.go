package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateAccessProvider(t *testing.T) {
	valid := AccessOIDCProvider{Enabled: true, IssuerURL: "https://id.example.com", ClientID: "client", ClientSecret: "secret", GatewayURL: "https://sso.example.com", ACMEIssuerID: "step"}
	if err := validateAccessProvider(valid); err != nil {
		t.Fatalf("valid provider rejected: %v", err)
	}
	valid.IssuerURL = "http://id.example.com"
	if err := validateAccessProvider(valid); err == nil {
		t.Fatal("insecure issuer URL accepted")
	}
	valid.IssuerURL = "https://id.example.com"
	valid.GatewayURL = "https://sso.example.com/nested"
	if err := validateAccessProvider(valid); err == nil {
		t.Fatal("SSO base URL with a path accepted")
	}
	if err := validateAccessProvider(AccessOIDCProvider{}); err != nil {
		t.Fatalf("disabled provider rejected: %v", err)
	}
}

func TestAuthProviderFileLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-providers.json")
	app := App{authProvidersPath: path, authProviders: AuthProvidersConfig{OIDC: AccessOIDCProvider{Enabled: true, ID: "oidc"}}}
	if err := app.saveAuthProvidersLocked(); err != nil {
		t.Fatalf("save enabled provider: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat enabled provider file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("provider file mode = %o, want 600", info.Mode().Perm())
	}
	app.authProviders.OIDC.Enabled = false
	if err := app.saveAuthProvidersLocked(); err != nil {
		t.Fatalf("disable provider: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("disabled provider file still exists: %v", err)
	}
}

func TestAccessCookieSecurityAttributes(t *testing.T) {
	cookie := accessCookie("name", "value", time.Hour)
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("unsafe access cookie: %#v", cookie)
	}
}

func TestSafeAccessReturnPath(t *testing.T) {
	for input, want := range map[string]string{"": "/", "https://evil.example": "/", "//evil.example": "/", "/projects?id=1": "/projects?id=1"} {
		if got := safeAccessReturnPath(input); got != want {
			t.Errorf("safeAccessReturnPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRenderAccessGateway(t *testing.T) {
	provider := AccessOIDCProvider{Enabled: true, GatewayURL: "https://sso.example.com", ACMEIssuerID: "step"}
	issuers := []ACMEIssuer{{ID: "step", DirectoryURL: "https://ca.example.com/acme/directory"}}
	rendered := renderAccessGateway(provider, issuers, "caddymgm:8080", nil)
	for _, want := range []string{"https://sso.example.com", "bind 0.0.0.0", "reverse_proxy /.caddymgm/auth/* caddymgm:8080", "dir https://ca.example.com/acme/directory"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("gateway missing %q:\n%s", want, rendered)
		}
	}
}

func TestRenderAccessGatewayAddsIPv6ListenerForEnabledIPv6Host(t *testing.T) {
	provider := AccessOIDCProvider{Enabled: true, GatewayURL: "https://sso.example.com"}
	rendered := renderAccessGateway(provider, nil, "caddymgm:8080", []Site{
		{ID: "ipv4", Enabled: true},
		{ID: "ipv6", Enabled: true, IPv6Enabled: true},
		{ID: "disabled-ipv6", Enabled: false, IPv6Enabled: true},
	})
	if strings.Count(rendered, "https://sso.example.com {") != 2 || !strings.Contains(rendered, "bind 0.0.0.0") || !strings.Contains(rendered, "bind [::]") {
		t.Fatalf("gateway listeners are incomplete:\n%s", rendered)
	}
}

func TestRenderProtectedSite(t *testing.T) {
	site := Site{ID: "protected", Address: "app.example.com", Enabled: true, TLSMode: "public", Upstream: "app:8080", AuthEnabled: true, AuthProviderID: "oidc"}
	rendered := renderSite(site, nil, "/logs", "caddymgm:8080")
	required := []string{"# caddymgm:auth-provider oidc", "forward_auth @caddymgmProtected caddymgm:8080", "uri /.caddymgm/auth/check", "copy_headers X-Auth-User X-Auth-Email", "reverse_proxy /.caddymgm/auth/* caddymgm:8080"}
	for _, want := range required {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered site missing %q:\n%s", want, rendered)
		}
	}
}
