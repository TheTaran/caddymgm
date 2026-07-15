package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureSettingsRequiresInitialAdministratorPassword(t *testing.T) {
	t.Setenv("CADDYMGM_ADMIN_PASSWORD", "")
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	app := &App{
		settingsPath: settingsPath,
		configPath:   "/config/Caddyfile",
		caddyMode:    "file",
		caddyAPIURL:  "",
		caddyDataDir: "/caddy-data",
		caCertDir:    "/ca-certificates",
		webPort:      "8080",
		oidcCache:    make(map[string]*oidcRuntime),
	}

	err := app.ensureSettings()
	if err == nil || !strings.Contains(err.Error(), "CADDYMGM_ADMIN_PASSWORD") {
		t.Fatalf("ensureSettings error = %v, want missing-password error", err)
	}
	if _, statErr := os.Stat(settingsPath); !os.IsNotExist(statErr) {
		t.Fatalf("settings file exists after rejected first startup: %v", statErr)
	}
}

func TestEnsureSettingsAcceptsConfiguredAdministratorPassword(t *testing.T) {
	password := "a-strong-test-password-123"
	t.Setenv("CADDYMGM_ADMIN_PASSWORD", password)
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	app := &App{
		settingsPath: settingsPath,
		configPath:   "/config/Caddyfile",
		caddyMode:    "file",
		caddyDataDir: "/caddy-data",
		caCertDir:    "/ca-certificates",
		webPort:      "8080",
		oidcCache:    make(map[string]*oidcRuntime),
	}

	if err := app.ensureSettings(); err != nil {
		t.Fatalf("ensureSettings: %v", err)
	}
	if !passwordMatchesHash(password, app.settings.PasswordHash) {
		t.Fatal("configured administrator password was not stored as a matching hash")
	}
}

func TestEnsureSettingsRejectsExistingSettingsWithoutPasswordHash(t *testing.T) {
	t.Setenv("CADDYMGM_ADMIN_PASSWORD", "")
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"username":"admin"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	app := &App{
		settingsPath: settingsPath,
		configPath:   "/config/Caddyfile",
		caddyMode:    "file",
		caddyDataDir: "/caddy-data",
		caCertDir:    "/ca-certificates",
		webPort:      "8080",
		oidcCache:    make(map[string]*oidcRuntime),
	}

	if err := app.ensureSettings(); err == nil {
		t.Fatal("settings without a password hash were accepted")
	}
}
