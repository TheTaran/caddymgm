package main

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestCleanupSessionsRemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	app := &App{sessions: map[string]Session{
		"expired": {ExpiresAt: now.Add(-time.Second)},
		"active":  {ExpiresAt: now.Add(time.Hour)},
	}}
	app.cleanupSessionsLocked(now)
	if _, ok := app.sessions["expired"]; ok {
		t.Fatal("expired session was retained")
	}
	if _, ok := app.sessions["active"]; !ok {
		t.Fatal("active session was removed")
	}
}

func TestCreateSessionKeepsBoundedStorage(t *testing.T) {
	app := &App{sessions: make(map[string]Session)}
	now := time.Now()
	for i := 0; i < sessionLimit; i++ {
		app.sessions[newSessionToken()] = Session{ExpiresAt: now.Add(time.Duration(i+1) * time.Second)}
	}
	request := httptest.NewRequest("GET", "http://localhost/", nil)
	request.RemoteAddr = "127.0.0.1:1234"
	app.createSessionLocked(httptest.NewRecorder(), request, "admin", "local")
	if len(app.sessions) != sessionLimit {
		t.Fatalf("session count = %d, want %d", len(app.sessions), sessionLimit)
	}
}

func TestValidateOIDCSettingsRequiresSecureTransport(t *testing.T) {
	base := OIDCSettings{
		Enabled: true, IssuerURL: "https://id.example.test", ClientID: "client",
		ClientSecret: "secret", RedirectURL: "https://admin.example.test/auth/oidc/callback",
	}
	if err := validateOIDCSettings(base); err != nil {
		t.Fatalf("secure OIDC settings rejected: %v", err)
	}
	insecureIssuer := base
	insecureIssuer.IssuerURL = "http://id.example.test"
	if err := validateOIDCSettings(insecureIssuer); err == nil {
		t.Fatal("insecure non-loopback issuer URL was accepted")
	}
	insecureRedirect := base
	insecureRedirect.RedirectURL = "http://admin.example.test/auth/oidc/callback"
	if err := validateOIDCSettings(insecureRedirect); err == nil {
		t.Fatal("insecure non-loopback redirect URL was accepted")
	}
	loopback := base
	loopback.IssuerURL = "http://127.0.0.1:1411"
	loopback.RedirectURL = "http://localhost:8080/auth/oidc/callback"
	if err := validateOIDCSettings(loopback); err != nil {
		t.Fatalf("loopback development URLs rejected: %v", err)
	}
}
