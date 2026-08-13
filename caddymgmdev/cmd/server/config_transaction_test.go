package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
