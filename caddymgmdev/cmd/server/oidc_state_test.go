package main

import (
	"strconv"
	"testing"
	"time"
)

func TestCleanupOIDCStatesRemovesExpiredEntries(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	app := &App{oidcStates: map[string]time.Time{
		"expired": now,
		"active":  now.Add(time.Minute),
	}}

	app.cleanupOIDCStatesLocked(now)
	if _, ok := app.oidcStates["expired"]; ok {
		t.Fatal("expired OIDC state was not removed")
	}
	if _, ok := app.oidcStates["active"]; !ok {
		t.Fatal("active OIDC state was removed")
	}
}

func TestOIDCStateLimitIsBoundedAfterCleanup(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	states := make(map[string]time.Time, oidcStateLimit)
	for i := 0; i < oidcStateLimit; i++ {
		states["state-"+strconv.Itoa(i)] = now.Add(oidcStateLifetime)
	}
	app := &App{oidcStates: states}
	if state, ok := app.createOIDCStateLocked(now); ok || state != "" {
		t.Fatal("OIDC state was created beyond configured limit")
	}
	delete(app.oidcStates, "state-0")
	if state, ok := app.createOIDCStateLocked(now); !ok || state == "" {
		t.Fatal("OIDC state was not created after capacity was freed")
	}
}

func TestOIDCStateLifetimeIsTenMinutes(t *testing.T) {
	if oidcStateLifetime != 10*time.Minute {
		t.Fatalf("OIDC state lifetime = %s, want 10m", oidcStateLifetime)
	}
}
