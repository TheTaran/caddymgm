package main

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeExternalBlocklistsAcceptsCustomPublicHTTPSURLs(t *testing.T) {
	got, err := normalizeExternalBlocklists(context.Background(), ExternalBlocklists{{Name: "Primary", URL: "https://1.1.1.1/list.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	want := ExternalBlocklists{{Name: "Primary", URL: "https://1.1.1.1/list.txt"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeExternalBlocklists() = %#v, want %#v", got, want)
	}
	for _, value := range []string{"http://example.com/list.txt", "https://127.0.0.1/list.txt", "https://10.0.0.1/list.txt"} {
		if _, err := normalizeExternalBlocklists(context.Background(), ExternalBlocklists{{Name: "Unsafe", URL: value}}); err == nil {
			t.Fatalf("unsafe blocklist URL %q was accepted", value)
		}
	}
}

func TestExternalBlocklistsMigratesLegacyURLValues(t *testing.T) {
	var values ExternalBlocklists
	if err := json.Unmarshal([]byte(`["https://example.com/list.txt"]`), &values); err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || values[0].Name != "example.com" || values[0].URL != "https://example.com/list.txt" {
		t.Fatalf("legacy values were not migrated: %#v", values)
	}
}

func TestCollectBlocklistEntriesValidatesAndDeduplicates(t *testing.T) {
	entries := map[string]bool{}
	err := collectBlocklistEntries(strings.NewReader("203.0.113.9\n203.0.113.9 # duplicate\n198.51.100.0/24\n10.0.0.1\ninvalid\n"), entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || !entries["203.0.113.9/32"] || !entries["198.51.100.0/24"] {
		t.Fatalf("collected entries = %#v", entries)
	}
}

func TestPrepareExternalBlocklistsPreservesListMetadataAndDeduplicatesTotal(t *testing.T) {
	previous := ExternalBlocklists{
		{Name: "First", URL: "https://1.1.1.1/first.txt", Entries: []string{"1.1.1.1/32", "8.8.8.8/32"}, Count: 2, UpdatedAt: "2026-08-31T12:00:00Z"},
		{Name: "Second", URL: "https://8.8.8.8/second.txt", Entries: []string{"8.8.8.8/32", "9.9.9.9/32"}, Count: 2, UpdatedAt: "2026-08-31T12:01:00Z"},
	}
	feeds, total, err := (&App{}).prepareExternalBlocklists(context.Background(), ExternalBlocklists{
		{Name: "First", URL: "https://1.1.1.1/first.txt"},
		{Name: "Second", URL: "https://8.8.8.8/second.txt"},
	}, previous, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(feeds) != 2 || feeds[0].Count != 2 || feeds[0].UpdatedAt == "" {
		t.Fatalf("list metadata was not preserved: %#v", feeds)
	}
	want := []string{"1.1.1.1/32", "8.8.8.8/32", "9.9.9.9/32"}
	if !reflect.DeepEqual(total, want) {
		t.Fatalf("aggregate entries = %#v, want %#v", total, want)
	}
}
