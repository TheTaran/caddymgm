package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDemoGeoMapResponseIsAvailableWithoutDatabase(t *testing.T) {
	app := &App{demoData: true}
	response := httptest.NewRecorder()
	app.handleGeoMap(response, httptest.NewRequest(http.MethodGet, "/api/geo-map", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !containsJSONField(response.Body.String(), `"available":true`) || !containsJSONField(response.Body.String(), `"locations"`) {
		t.Fatalf("unexpected demo map response: %s", response.Body.String())
	}
}

func TestDemoSecurityOverviewProvidesTrends(t *testing.T) {
	app := &App{demoData: true}
	response := httptest.NewRecorder()
	app.handleSecurityOverview(response, httptest.NewRequest(http.MethodGet, "/api/security-overview?period=1h", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !containsJSONField(response.Body.String(), `"trendInterval":"5 minutes"`) || !containsJSONField(response.Body.String(), `"managedBlocks"`) {
		t.Fatalf("unexpected demo security response: %s", response.Body.String())
	}
}

func TestDemoAccessLogsFillFirstPage(t *testing.T) {
	entries := demoAccessLogs("demo-site")
	if len(entries) != 150 {
		t.Fatalf("entry count = %d, want 150", len(entries))
	}
	if entries[0].SiteID != "demo-site" || entries[0].IP == "" || entries[0].Status == "" {
		t.Fatalf("unexpected demo log: %#v", entries[0])
	}
}

func TestDemoThroughputContainsTraffic(t *testing.T) {
	_, spec, err := securityTrendSpecForPeriod("1h")
	if err != nil {
		t.Fatal(err)
	}
	data := demoThroughput(spec, "")
	if len(data.Points) == 0 || data.Egress == 0 || data.EgressRate == 0 {
		t.Fatalf("demo throughput is empty: %#v", data)
	}
}

func containsJSONField(value, field string) bool {
	return strings.Contains(value, field)
}
