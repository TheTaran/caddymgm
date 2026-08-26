package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOIDCAuditLogPersistsWithoutSecrets(t *testing.T) {
	dir := t.TempDir()
	app := &App{oidcAuditLog: filepath.Join(dir, "oidc-audit.log"), settings: Settings{LogRetention: 10}}
	req := httptest.NewRequest("GET", "https://sso.example.test/.caddymgm/auth/callback", nil)
	req.RemoteAddr = "192.0.2.44:4123"
	app.recordOIDCAudit(req, "login_success", "success", "alice", "alice@example.test", "git.example.test", "Website SSO login")

	info, err := os.Stat(app.oidcAuditLog)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	entries := app.readOIDCAuditLogsLocked()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	if entries[0].IP != "192.0.2.44" || entries[0].Username != "alice" {
		t.Fatalf("unexpected entry: %#v", entries[0])
	}
	content, err := os.ReadFile(app.oidcAuditLog)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "" {
		t.Fatal("audit log is empty")
	}
}

func TestOIDCAuditIgnoresForwardedIPFromUntrustedClient(t *testing.T) {
	app := &App{}
	req := httptest.NewRequest("GET", "https://sso.example.test/", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	if got := app.requestClientIP(req); got != "192.0.2.10" {
		t.Fatalf("IP = %q", got)
	}
}
