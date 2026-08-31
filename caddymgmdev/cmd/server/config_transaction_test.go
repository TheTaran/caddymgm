package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
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

func TestReverseProxyForwardsPublicRequestIdentity(t *testing.T) {
	site := Site{ID: "mail", Address: "mail.homedc.net", Mode: "proxy", Upstream: "https://10.0.100.102", TLSMode: "acme", Enabled: true, RewriteRedirects: true}
	rendered := renderSite(site, nil, "/logs", "caddymgm:8080")
	for _, want := range []string{"header_up Host {host}", "header_down Location"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered proxy is missing %q:\n%s", want, rendered)
		}
	}
	parsed, err := parseSite(site.ID, strings.Split(rendered, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.ExtraDirectives != "" {
		t.Fatalf("managed forwarding headers leaked into additional settings: %q", parsed.ExtraDirectives)
	}
}

func TestAdditionalRedirectOriginsRenderParseAndValidate(t *testing.T) {
	site := Site{
		ID: "mail", Address: "mail.homedc.net", Mode: "proxy",
		Upstream: "https://10.0.100.102", TLSMode: "acme", Enabled: true,
		RewriteRedirects: true,
		RedirectOrigins:  []string{" https://mail.alonso.lds:443/ ", "https://MAIL.alonso.lds:443"},
	}
	if err := normalizeSite(&site); err != nil {
		t.Fatal(err)
	}
	if len(site.RedirectOrigins) != 1 || site.RedirectOrigins[0] != "https://mail.alonso.lds:443" {
		t.Fatalf("normalized redirect origins = %#v", site.RedirectOrigins)
	}
	rendered := renderSite(site, nil, "/logs", "caddymgm:8080")
	for _, want := range []string{
		`# caddymgm:redirect-origin "https://mail.alonso.lds:443"`,
		`(?i)^https://mail\\.alonso\\.lds:443([/?#].*)?$`,
		`https://mail.homedc.net${1}`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered site is missing %q:\n%s", want, rendered)
		}
	}
	parsed, err := parseSite(site.ID, strings.Split(rendered, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.RedirectOrigins) != 1 || parsed.RedirectOrigins[0] != site.RedirectOrigins[0] {
		t.Fatalf("parsed redirect origins = %#v", parsed.RedirectOrigins)
	}

	invalid := site
	invalid.RedirectOrigins = []string{"https://mail.alonso.lds/webmail"}
	if err := normalizeSite(&invalid); err == nil || !strings.Contains(err.Error(), "redirect origins") {
		t.Fatalf("normalizeSite() error = %v, want redirect origin validation error", err)
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

func TestTLSControlsRenderParseAndValidate(t *testing.T) {
	issuer := ACMEIssuer{ID: "letsencrypt", DirectoryURL: "https://acme.example.test/directory"}
	site := Site{
		ID: "secure", Address: "secure.example.test", Mode: "proxy", Upstream: "http://app:8080",
		TLSMode: "acme", ACMEIssuerID: issuer.ID, Enabled: true,
		TLSMinVersion: "tls1.2", TLSMaxVersion: "tls1.3",
	}
	if err := normalizeSite(&site); err != nil {
		t.Fatal(err)
	}
	rendered := renderSite(site, []ACMEIssuer{issuer}, "/logs", "caddymgm:8080")
	for _, want := range []string{
		"protocols tls1.2 tls1.3",
		"issuer acme {",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered site is missing %q:\n%s", want, rendered)
		}
	}
	parsed, err := parseSite(site.ID, strings.Split(rendered, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.TLSMinVersion != site.TLSMinVersion || parsed.TLSMaxVersion != site.TLSMaxVersion {
		t.Fatalf("TLS controls did not survive render/parse: %+v", parsed)
	}

	legacyRendered := strings.Replace(rendered, "protocols tls1.2 tls1.3", "protocols tls1.2 tls1.3\n\t\tciphers TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", 1)
	legacyParsed, err := parseSite(site.ID, strings.Split(legacyRendered, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(legacyParsed.ExtraDirectives, "ciphers") {
		t.Fatalf("legacy managed cipher directive leaked into extra settings: %q", legacyParsed.ExtraDirectives)
	}
	if migrated := renderSite(legacyParsed, []ACMEIssuer{issuer}, "/logs", "caddymgm:8080"); strings.Contains(migrated, "ciphers ") {
		t.Fatalf("legacy cipher directive survived migration:\n%s", migrated)
	}

	invalidRange := site
	invalidRange.TLSMinVersion = "tls1.3"
	invalidRange.TLSMaxVersion = "tls1.2"
	if err := normalizeSite(&invalidRange); err == nil || !strings.Contains(err.Error(), "cannot exceed") {
		t.Fatalf("invalid TLS range error = %v", err)
	}
}

func TestCompressionProfilesRenderParseAndValidate(t *testing.T) {
	for _, test := range []struct {
		profile   string
		directive string
	}{
		{profile: "gzip", directive: "encode gzip"},
		{profile: "zstd-gzip", directive: "encode zstd gzip"},
	} {
		t.Run(test.profile, func(t *testing.T) {
			site := Site{
				ID: "compressed", Address: "compressed.example.test", Mode: "proxy",
				Upstream: "http://app:8080", TLSMode: "off", Enabled: true,
				CompressionProfile: test.profile,
			}
			if err := normalizeSite(&site); err != nil {
				t.Fatal(err)
			}
			rendered := renderSite(site, nil, "/logs", "caddymgm:8080")
			for _, want := range []string{"# caddymgm:compression " + test.profile, test.directive} {
				if !strings.Contains(rendered, want) {
					t.Fatalf("rendered site is missing %q:\n%s", want, rendered)
				}
			}
			parsed, err := parseSite(site.ID, strings.Split(rendered, "\n"))
			if err != nil {
				t.Fatal(err)
			}
			if parsed.CompressionProfile != test.profile || strings.Contains(parsed.ExtraDirectives, "encode ") {
				t.Fatalf("compression profile did not survive render/parse: %+v", parsed)
			}
		})
	}

	invalid := Site{Address: "invalid.example.test", Mode: "proxy", Upstream: "http://app:8080", CompressionProfile: "brotli"}
	if err := normalizeSite(&invalid); err == nil || !strings.Contains(err.Error(), "compression profile") {
		t.Fatalf("invalid compression profile error = %v", err)
	}
	conflict := Site{Address: "conflict.example.test", Mode: "proxy", Upstream: "http://app:8080", CompressionProfile: "gzip", ExtraDirectives: "encode zstd gzip"}
	if err := normalizeSite(&conflict); err == nil || !strings.Contains(err.Error(), "manual encode") {
		t.Fatalf("manual compression conflict error = %v", err)
	}
}

func TestWebProtectionRenderAndParsePreservesHostOverride(t *testing.T) {
	site := Site{
		ID: "protected", Address: "protected.example.test", Mode: "proxy",
		Upstream: "http://app:8080", TLSMode: "off", Enabled: true,
		ProtectionOverride: true,
		WebProtection: WebProtection{
			Enabled: true, CountryMode: "block", BlockedCountries: []string{"CN", "RU"},
			BlockedIPs: []string{"203.0.113.9/32"}, AllowedIPs: []string{"198.51.100.4/32"},
		},
	}
	rendered := renderManagedWithProtection([]Site{site}, nil, "/logs", WebInterface{}, WebProtection{}, AccessOIDCProvider{}, "file", "8080")
	for _, want := range []string{
		"order geo_ip first",
		"# caddymgm:protection-override true",
		"# caddymgm:protection-countries CN RU",
		"geo_ip {",
		"@caddymgmProtectionCountry {",
		"remote_ip 203.0.113.9/32",
		"@caddymgmProtectionAllow remote_ip 198.51.100.4/32",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered managed config is missing %q:\n%s", want, rendered)
		}
	}

	parsed, err := parseManaged(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || !parsed[0].ProtectionOverride || !reflect.DeepEqual(parsed[0].WebProtection, site.WebProtection) {
		t.Fatalf("protection override did not survive render/parse: %#v", parsed)
	}
}

func TestWebProtectionAllowModeBlocksCountriesOutsideSelection(t *testing.T) {
	policy := WebProtection{Enabled: true, CountryMode: "allow", BlockedCountries: []string{"CH", "DE"}}
	if err := normalizeWebProtection(&policy); err != nil {
		t.Fatal(err)
	}
	var rendered strings.Builder
	writeWebProtection(&rendered, "", policy)
	for _, want := range []string{"# caddymgm:web-protection allow-countries CH,DE", `expression "!{geoip.country_code}.matches('^(CH|DE)$')"`} {
		if !strings.Contains(rendered.String(), want) {
			t.Fatalf("allow-mode config is missing %q:\n%s", want, rendered.String())
		}
	}
}

func TestBasicAuthRenderParseAndPasswordRetention(t *testing.T) {
	site := Site{Address: "basic.example.test", Mode: "proxy", Upstream: "http://127.0.0.1:8080", TLSMode: "acme", BasicAuthEnabled: true, BasicAuthUsername: "web-user", BasicAuthPassword: "correct-horse"}
	if err := prepareBasicAuth(&site, nil); err != nil {
		t.Fatal(err)
	}
	if site.BasicAuthPassword != "" || site.BasicAuthPasswordHash == "" {
		t.Fatalf("password was not converted to a write-only hash: %+v", site)
	}
	rendered := renderSite(site, nil, "/logs", "caddymgm:8080")
	if !strings.Contains(rendered, "basic_auth {") || !strings.Contains(rendered, site.BasicAuthPasswordHash) || strings.Contains(rendered, "correct-horse") {
		t.Fatalf("unexpected rendered basic auth block:\n%s", rendered)
	}
	parsed, err := parseSite("basic", strings.Split(rendered, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.BasicAuthEnabled || parsed.BasicAuthUsername != "web-user" || parsed.BasicAuthPasswordHash != site.BasicAuthPasswordHash {
		t.Fatalf("parsed basic auth = %+v", parsed)
	}
	if strings.Contains(parsed.ExtraDirectives, "basic_auth") {
		t.Fatalf("managed basic auth leaked into extra settings: %q", parsed.ExtraDirectives)
	}
	encoded, err := json.Marshal(parsed)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), site.BasicAuthPasswordHash) || strings.Contains(string(encoded), "basicAuthPassword") {
		t.Fatalf("API JSON exposed basic auth password material: %s", encoded)
	}
	updated := Site{TLSMode: "acme", BasicAuthEnabled: true, BasicAuthUsername: "web-user"}
	if err := prepareBasicAuth(&updated, &site); err != nil || updated.BasicAuthPasswordHash != site.BasicAuthPasswordHash {
		t.Fatalf("existing password hash was not retained: hash=%q err=%v", updated.BasicAuthPasswordHash, err)
	}
}

func TestBasicAuthRequiresTLSAndExcludesOIDC(t *testing.T) {
	for _, site := range []Site{
		{TLSMode: "off", BasicAuthEnabled: true, BasicAuthUsername: "user", BasicAuthPassword: "password-123"},
		{TLSMode: "acme", AuthEnabled: true, BasicAuthEnabled: true, BasicAuthUsername: "user", BasicAuthPassword: "password-123"},
	} {
		if err := prepareBasicAuth(&site, nil); err == nil {
			t.Fatalf("prepareBasicAuth(%+v) succeeded, want error", site)
		}
	}
}

func TestDisabledSiteRendersUnavailablePageAndPreservesConfiguration(t *testing.T) {
	site := Site{
		ID: "disabled", Address: "disabled.example.test", Mode: "proxy",
		Upstream: "http://app:3000", TLSMode: "internal", Enabled: false,
		LogsEnabled: true, HSTSEnabled: true, SecurityHeaderProfile: "standard",
		TLSMinVersion: "tls1.2", TLSMaxVersion: "tls1.3",
	}
	rendered := renderManaged([]Site{site}, nil, "/logs", WebInterface{}, AccessOIDCProvider{}, "file", "8080")
	for _, want := range []string{
		"(caddymgm_unavailable) {",
		"respond \"<!doctype html>",
		" 503\n",
		"# disabled.example.test {",
		"# \treverse_proxy http://app:3000",
		"# caddymgm:unavailable-site disabled",
		"disabled.example.test {\n\timport caddymgm_unavailable",
		"\t\tprotocols tls1.2 tls1.3",
		"\t\tissuer internal",
		"\tlog {",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered managed config is missing %q:\n%s", want, rendered)
		}
	}

	parsed, err := parseManaged(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 {
		t.Fatalf("parsed sites = %d, want 1: %+v", len(parsed), parsed)
	}
	got := parsed[0]
	if got.Enabled || got.Address != site.Address || got.Upstream != site.Upstream || got.TLSMode != site.TLSMode || got.TLSMinVersion != site.TLSMinVersion || got.TLSMaxVersion != site.TLSMaxVersion || !got.LogsEnabled || !got.HSTSEnabled || got.SecurityHeaderProfile != site.SecurityHeaderProfile {
		t.Fatalf("disabled site did not survive render/parse round trip: %+v", got)
	}
}

func TestEnabledSiteDoesNotRenderUnavailableVirtualHost(t *testing.T) {
	site := Site{ID: "enabled", Address: "enabled.example.test", Mode: "proxy", Upstream: "http://app:3000", TLSMode: "off", Enabled: true}
	rendered := renderManaged([]Site{site}, nil, "/logs", WebInterface{}, AccessOIDCProvider{}, "file", "8080")
	if strings.Contains(rendered, "caddymgm:unavailable-site enabled") || strings.Contains(rendered, "enabled.example.test {\n\timport caddymgm_unavailable") {
		t.Fatalf("enabled site received an unavailable virtual host:\n%s", rendered)
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
