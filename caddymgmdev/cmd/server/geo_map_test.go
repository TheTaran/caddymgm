package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"strings"
	"testing"
)

func TestAccessLogClientAddress(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
		ok   bool
	}{
		{name: "client ip preferred", line: `{"request":{"client_ip":"203.0.113.10","remote_ip":"198.51.100.2"}}`, want: "203.0.113.10", ok: true},
		{name: "remote ip fallback", line: `{"request":{"remote_ip":"2001:db8::1"}}`, want: "2001:db8::1", ok: true},
		{name: "invalid json", line: `{`, ok: false},
		{name: "missing address", line: `{}`, ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := accessLogClientAddress(test.line)
			if ok != test.ok || (ok && got.String() != test.want) {
				t.Fatalf("accessLogClientAddress() = (%q, %v), want (%q, %v)", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestIsPublicAddress(t *testing.T) {
	for value, want := range map[string]bool{
		"8.8.8.8":     true,
		"1.1.1.1":     true,
		"127.0.0.1":   false,
		"10.0.0.1":    false,
		"192.168.1.1": false,
		"::1":         false,
		"fc00::1":     false,
	} {
		if got := isPublicAddress(netip.MustParseAddr(value)); got != want {
			t.Errorf("isPublicAddress(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestHandleGeoMapWithoutDatabase(t *testing.T) {
	directory := t.TempDir()
	configPath := directory + "/Caddyfile"
	if err := os.WriteFile(configPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	app := &App{geoIPDBPath: directory + "/missing.mmdb", configPath: configPath}
	request := httptest.NewRequest(http.MethodGet, "/api/geo-map", nil)
	response := httptest.NewRecorder()

	app.handleGeoMap(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var payload geoMapResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Available || payload.Message == "" || payload.Locations == nil {
		t.Fatalf("unexpected response: %+v", payload)
	}
}

func TestSortedGeoIPsRanksPrivateAndPublicAddresses(t *testing.T) {
	aggregates := make(map[string]*geoIPAggregate)
	addGeoIPAggregate(aggregates, "192.168.1.10", "internal.example")
	addGeoIPAggregate(aggregates, "192.168.1.10", "internal.example")
	addGeoIPAggregate(aggregates, "8.8.8.8", "public.example")

	got := sortedGeoIPs(aggregates, geoTopIPLimit)
	if len(got) != 2 {
		t.Fatalf("top IP count = %d, want 2", len(got))
	}
	if got[0].Address != "192.168.1.10" || got[0].Count != 2 || len(got[0].Sites) != 1 {
		t.Fatalf("first top IP = %+v", got[0])
	}
}

func TestSanitizeUpdaterOutputRedactsSecrets(t *testing.T) {
	got := sanitizeUpdaterOutput("failure for account-123 using secret-key\nnext line", "account-123", "secret-key")
	if strings.Contains(got, "account-123") || strings.Contains(got, "secret-key") || strings.ContainsAny(got, "\n\r") {
		t.Fatalf("sensitive or multiline updater output was not sanitized: %q", got)
	}
}

func TestGeoIPScope(t *testing.T) {
	for address, want := range map[string]string{
		"192.168.1.10": "internal",
		"10.0.0.1":     "internal",
		"127.0.0.1":    "internal",
		"fe80::1":      "internal",
		"8.8.8.8":      "external",
	} {
		if got := geoIPScope(address); got != want {
			t.Errorf("geoIPScope(%q) = %q, want %q", address, got, want)
		}
	}
}

func TestGeoIPSiteCounts(t *testing.T) {
	aggregates := make(map[string]*geoIPAggregate)
	addGeoIPAggregate(aggregates, "192.168.1.10", "one.example")
	addGeoIPAggregate(aggregates, "192.168.1.10", "one.example")
	addGeoIPAggregate(aggregates, "192.168.1.10", "two.example")

	got := sortedGeoIPs(aggregates, 10)
	if len(got) != 1 || got[0].SiteCounts["one.example"] != 2 || got[0].SiteCounts["two.example"] != 1 {
		t.Fatalf("site counts = %+v", got)
	}
}
