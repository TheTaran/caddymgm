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

func TestNormalizeManualIPListsPreservesEnteredAddressesAndCIDRs(t *testing.T) {
	got, err := normalizeManualIPLists(ManualIPLists{
		{Name: "Office", Mode: "allow", Entries: []string{"10.0.100.36", "10.0.100.0/23"}},
		{Name: "Threats", Mode: "block", Entries: []string{"203.0.113.9", "198.51.100.5/24"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]ManualIPList{}
	for _, list := range got {
		byName[list.Name] = list
	}
	if want := []string{"10.0.100.36", "10.0.100.0/23"}; !reflect.DeepEqual(byName["Office"].Entries, want) {
		t.Fatalf("allow entries = %#v, want %#v", byName["Office"].Entries, want)
	}
	if want := []string{"203.0.113.9", "198.51.100.5/24"}; !reflect.DeepEqual(byName["Threats"].Entries, want) {
		t.Fatalf("block entries = %#v, want %#v", byName["Threats"].Entries, want)
	}
}

func TestNormalizeExternalBlocklistsConvertsGitHubBlobURLs(t *testing.T) {
	got, err := normalizeExternalBlocklists(context.Background(), ExternalBlocklists{{
		Name: "FireHOL Level 1",
		URL:  "https://github.com/firehol/blocklist-ipsets/blob/master/firehol_level1.netset",
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := "https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset"
	if len(got) != 1 || got[0].URL != want {
		t.Fatalf("normalized GitHub URL = %#v, want %q", got, want)
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
	err := collectBlocklistEntries(strings.NewReader("203.0.113.9\n203.0.113.9 # duplicate\n198.51.100.0/24\n1.10.16.0/20 ; SBL256894\n10.0.0.1\ninvalid\n"), entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 || !entries["203.0.113.9/32"] || !entries["198.51.100.0/24"] || !entries["1.10.16.0/20"] {
		t.Fatalf("collected entries = %#v", entries)
	}
}

func TestCollectBlocklistEntriesSupportsJSONAndNDJSON(t *testing.T) {
	entries := map[string]bool{}
	content := "{\"cidr\":\"1.10.16.0/20\",\"sblid\":\"SBL256894\"}\n[\"8.8.8.8\", {\"network\":\"9.9.9.0/24\"}]"
	if err := collectBlocklistEntries(strings.NewReader(content), entries); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"1.10.16.0/20": true, "8.8.8.8/32": true, "9.9.9.0/24": true}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("collected JSON entries = %#v, want %#v", entries, want)
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
