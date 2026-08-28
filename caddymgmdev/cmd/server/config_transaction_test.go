package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSaveAndApplyRestoresPreviousConfigWhenCaddyRejectsChange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/load" {
			t.Fatalf("request path = %q, want /load", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "new.example.test") {
			t.Fatalf("candidate config was not sent to Caddy: %s", body)
		}
		http.Error(w, "invalid caddyfile", http.StatusBadRequest)
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "Caddyfile")
	previous := []byte("# existing config\n")
	if err := os.WriteFile(configPath, previous, 0o640); err != nil {
		t.Fatal(err)
	}
	app := &App{
		configPath:  configPath,
		caddyMode:   "api",
		caddyAPIURL: server.URL,
		httpClient:  server.Client(),
		webPort:     "8080",
	}

	err := app.saveAndApplyCaddyConfigLocked("", []Site{{
		ID: "new.example.test", Address: "new.example.test", Mode: "proxy",
		Upstream: "http://app:3000", TLSMode: "off", Enabled: true,
	}}, "")
	if err == nil {
		t.Fatal("rejected Caddy config was reported as successful")
	}
	got, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != string(previous) {
		t.Fatalf("config after rejected reload = %q, want previous %q", got, previous)
	}
}

func TestNormalizeSiteRejectsTLSVerificationSkipForHTTPUpstream(t *testing.T) {
	site := Site{Address: "app.example.test", Mode: "proxy", Upstream: "http://app:3000", SkipTLSVerify: true}
	if err := normalizeSite(&site); err == nil || !strings.Contains(err.Error(), "requires an HTTPS upstream") {
		t.Fatalf("normalizeSite error = %v, want HTTPS-upstream error", err)
	}
}

func TestSaveRepairsLegacyInvalidAndDuplicateSites(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "Caddyfile")
	if err := os.WriteFile(configPath, nil, 0o640); err != nil {
		t.Fatal(err)
	}
	app := &App{configPath: configPath, caddyMode: "file", webPort: "8080"}
	sites := []Site{
		{ID: "app", Address: "app.example.test", Mode: "proxy", Upstream: "http://app:3000", SkipTLSVerify: true, TLSMode: "off", Enabled: true},
		{ID: "app-2", Address: "APP.example.test", Mode: "proxy", Upstream: "http://other:3000", TLSMode: "off", Enabled: true},
	}
	if err := app.save("", sites, ""); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "app.example.test {") != 1 {
		t.Fatalf("duplicate host was not removed:\n%s", content)
	}
	if strings.Contains(string(content), "tls_insecure_skip_verify") {
		t.Fatalf("invalid HTTP upstream TLS transport was not removed:\n%s", content)
	}
}

func TestRedirectRewriteRuleMatchesOnlyConfiguredUpstreamOrigin(t *testing.T) {
	site := Site{Address: "public.example.test", Upstream: "http://10.0.100.102:8080", TLSMode: "acme"}
	pattern, replacement, ok := redirectRewriteRule(site)
	if !ok {
		t.Fatal("redirect rewrite rule was not generated")
	}
	matcher := regexp.MustCompile(pattern)
	if !matcher.MatchString("http://10.0.100.102:8080/login?next=/") {
		t.Fatalf("pattern %q does not match configured upstream redirect", pattern)
	}
	if matcher.MatchString("https://login.example.com/oauth") || matcher.MatchString("http://10.0.100.102.evil.test/") {
		t.Fatalf("pattern %q matches an external redirect", pattern)
	}
	if !strings.HasPrefix(replacement, "https://public.example.test") {
		t.Fatalf("replacement = %q, want public HTTPS origin", replacement)
	}
}

func TestRenderAndParseSitePreservesManagedSecurityOptions(t *testing.T) {
	site := Site{
		ID: "public", Address: "public.example.test", Mode: "proxy",
		Upstream: "http://10.0.100.102", TLSMode: "acme", Enabled: true,
		RewriteRedirects: true, HSTSEnabled: true,
		SecurityHeaderProfile: "standard",
	}
	rendered := renderSite(site, nil, "/logs", "caddymgm:8080")
	for _, want := range []string{"header_down Location", "# caddymgm:hsts", "header Strict-Transport-Security", "max-age=31536000", "# caddymgm:security-header-profile standard", "X-Content-Type-Options", "X-Frame-Options"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered site is missing %q:\n%s", want, rendered)
		}
	}
	parsed, err := parseSite(site.ID, strings.Split(rendered, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.RewriteRedirects || !parsed.HSTSEnabled || parsed.SecurityHeaderProfile != "standard" {
		t.Fatalf("parsed security options = rewrite:%v hsts:%v headers:%q", parsed.RewriteRedirects, parsed.HSTSEnabled, parsed.SecurityHeaderProfile)
	}
	if strings.Contains(parsed.ExtraDirectives, "header_down Location") || strings.Contains(parsed.ExtraDirectives, "Strict-Transport-Security") || strings.Contains(parsed.ExtraDirectives, "X-Content-Type-Options") {
		t.Fatalf("managed security directives leaked into extra settings: %q", parsed.ExtraDirectives)
	}
}

func TestStrictSecurityHeaderProfileAndValidation(t *testing.T) {
	directives := strings.Join(securityHeaderProfileDirectives("strict"), "\n")
	for _, want := range []string{"Content-Security-Policy", "Permissions-Policy", `X-Frame-Options "DENY"`, `Referrer-Policy "no-referrer"`} {
		if !strings.Contains(directives, want) {
			t.Fatalf("strict security headers are missing %q: %s", want, directives)
		}
	}
	site := Site{Address: "security.example.test", Mode: "proxy", Upstream: "http://127.0.0.1:8080", TLSMode: "off", SecurityHeaderProfile: "unknown"}
	if err := normalizeSite(&site); err == nil || !strings.Contains(err.Error(), "security header profile") {
		t.Fatalf("validateSite() error = %v, want security header profile error", err)
	}
}
