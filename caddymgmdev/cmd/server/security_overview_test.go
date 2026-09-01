package main

import (
	"net/netip"
	"testing"
)

func TestClassifyProtectionBlockPriorities(t *testing.T) {
	address := netip.MustParseAddr("203.0.113.10")
	external := newProtectionPrefixSet([]string{"203.0.113.0/24"})

	if reason, _ := classifyProtectionBlock(address, WebProtection{Enabled: true, AllowedIPs: []string{"203.0.113.10"}, BlockedIPs: []string{"203.0.113.10"}}, true, external, nil); reason != "" {
		t.Fatalf("allowlist must take priority, got %q", reason)
	}
	if reason, _ := classifyProtectionBlock(address, WebProtection{Enabled: true}, true, external, nil); reason != "External blocklist" {
		t.Fatalf("expected external blocklist, got %q", reason)
	}
	if reason, _ := classifyProtectionBlock(address, WebProtection{Enabled: true, BlockedIPs: []string{"203.0.113.10"}}, false, external, nil); reason != "Manual blocked IP" {
		t.Fatalf("expected manual blocked IP, got %q", reason)
	}
}

func TestProtectionPrefixSetMatchesCIDR(t *testing.T) {
	set := newProtectionPrefixSet([]string{"198.51.100.0/24", "2001:db8::/32"})
	for _, value := range []string{"198.51.100.99", "2001:db8:1::1"} {
		if !set.contains(netip.MustParseAddr(value)) {
			t.Fatalf("expected %s to match", value)
		}
	}
	if set.contains(netip.MustParseAddr("203.0.113.1")) {
		t.Fatal("unexpected prefix match")
	}
}
