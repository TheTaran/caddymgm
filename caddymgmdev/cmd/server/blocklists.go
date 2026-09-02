package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	externalBlocklistUpdateInterval = 24 * time.Hour
	externalBlocklistUpdateTimeout  = 10 * time.Minute
)

type ExternalBlocklist struct {
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Count     int      `json:"count,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
	Entries   []string `json:"entries,omitempty"`
}

type ExternalBlocklists []ExternalBlocklist

type ManualIPList struct {
	Name      string   `json:"name"`
	Reference string   `json:"reference,omitempty"`
	Mode      string   `json:"mode"`
	Entries   []string `json:"entries"`
}

type ManualIPLists []ManualIPList

func normalizeManualIPLists(values ManualIPLists) (ManualIPLists, error) {
	if len(values) > 50 {
		return nil, errors.New("at most 50 manual IP lists are supported")
	}
	result, seen := make(ManualIPLists, 0, len(values)), map[string]bool{}
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" || len(name) > 80 {
			return nil, errors.New("every manual IP list requires a name of at most 80 characters")
		}
		mode := strings.ToLower(strings.TrimSpace(value.Mode))
		if mode != "allow" && mode != "block" {
			return nil, errors.New("manual IP list mode must be allow or block")
		}
		key := strings.ToLower(name)
		if seen[key] {
			return nil, errors.New("manual IP list names must be unique")
		}
		entries, err := normalizeProtectionPrefixes(value.Entries, mode == "allow")
		if err != nil {
			return nil, err
		}
		seen[key] = true
		result = append(result, ManualIPList{Name: name, Reference: strings.TrimSpace(value.Reference), Mode: mode, Entries: entries})
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name) })
	return result, nil
}

func manualIPListEntries(values ManualIPLists, mode string) []string {
	entries, seen := make([]string, 0), map[string]bool{}
	for _, value := range values {
		if value.Mode != mode {
			continue
		}
		for _, entry := range value.Entries {
			if !seen[entry] {
				seen[entry] = true
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

func (values *ExternalBlocklists) UnmarshalJSON(content []byte) error {
	var entries []json.RawMessage
	if err := json.Unmarshal(content, &entries); err != nil {
		return err
	}
	result := make(ExternalBlocklists, 0, len(entries))
	for _, entry := range entries {
		var source ExternalBlocklist
		if len(entry) > 0 && entry[0] == '"' {
			if err := json.Unmarshal(entry, &source.URL); err != nil {
				return err
			}
			source.Name = defaultExternalBlocklistName(source.URL)
		} else if err := json.Unmarshal(entry, &source); err != nil {
			return err
		}
		result = append(result, source)
	}
	*values = result
	return nil
}

func defaultExternalBlocklistName(value string) string {
	if parsed, err := url.Parse(value); err == nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return "Imported blocklist"
}

func normalizeExternalBlocklists(ctx context.Context, values ExternalBlocklists) (ExternalBlocklists, error) {
	result, seenURLs, seenNames := make(ExternalBlocklists, 0, len(values)), map[string]bool{}, map[string]bool{}
	if len(values) > 20 {
		return nil, errors.New("at most 20 external blocklists are supported")
	}
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" || len(name) > 80 {
			return nil, errors.New("every external blocklist requires a name of at most 80 characters")
		}
		normalized, err := validateExternalBlocklistURL(ctx, normalizeGitHubBlobURL(strings.TrimSpace(value.URL)))
		if err != nil {
			return nil, err
		}
		nameKey := strings.ToLower(name)
		if seenNames[nameKey] {
			return nil, errors.New("external blocklist names must be unique")
		}
		if seenURLs[normalized] {
			return nil, errors.New("external blocklist URLs must be unique")
		}
		seenURLs[normalized] = true
		seenNames[nameKey] = true
		result = append(result, ExternalBlocklist{Name: name, URL: normalized})
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name) })
	return result, nil
}

// normalizeGitHubBlobURL converts GitHub's HTML file view into its raw text endpoint.
// This keeps pasted GitHub links usable while the stored URL remains a direct feed URL.
func normalizeGitHubBlobURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || !strings.EqualFold(parsed.Hostname(), "github.com") {
		return value
	}
	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) < 5 || parts[2] != "blob" || parts[0] == "" || parts[1] == "" || parts[3] == "" {
		return value
	}
	parsed.Host = "raw.githubusercontent.com"
	parsed.Path = "/" + strings.Join(append(parts[:2], parts[3:]...), "/")
	parsed.RawPath = ""
	return parsed.String()
}

func externalBlocklistsEqual(left, right ExternalBlocklists) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Name != right[index].Name || left[index].URL != right[index].URL {
			return false
		}
	}
	return true
}

func (a *App) prepareExternalBlocklists(ctx context.Context, feeds, previous ExternalBlocklists, refreshAll bool, refreshURL string) (ExternalBlocklists, []string, error) {
	previousByURL := make(map[string]ExternalBlocklist, len(previous))
	for _, feed := range previous {
		previousByURL[feed.URL] = feed
	}
	refreshURL = strings.TrimSpace(refreshURL)
	refreshMatched := refreshURL == ""
	aggregate := map[string]bool{}
	for index := range feeds {
		old, exists := previousByURL[feeds[index].URL]
		refresh := refreshAll || !exists || len(old.Entries) == 0 || feeds[index].URL == refreshURL
		if feeds[index].URL == refreshURL {
			refreshMatched = true
		}
		if refresh {
			entries, err := a.downloadExternalBlocklists(ctx, ExternalBlocklists{feeds[index]})
			if err != nil {
				return nil, nil, err
			}
			feeds[index].Entries = entries
			feeds[index].Count = len(entries)
			feeds[index].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		} else {
			feeds[index].Entries = append([]string(nil), old.Entries...)
			feeds[index].Count = len(old.Entries)
			feeds[index].UpdatedAt = old.UpdatedAt
		}
		for _, entry := range feeds[index].Entries {
			aggregate[entry] = true
		}
	}
	if !refreshMatched {
		return nil, nil, errors.New("external blocklist to update was not found")
	}
	blocked := make([]string, 0, len(aggregate))
	for entry := range aggregate {
		blocked = append(blocked, entry)
	}
	sort.Strings(blocked)
	return feeds, blocked, nil
}

// startExternalBlocklistUpdater refreshes every configured public list on
// startup and then every 24 hours. The refresh uses the same settings and
// Caddy configuration transaction as a manual update, so a failed Caddy load
// leaves the active configuration untouched.
func (a *App) startExternalBlocklistUpdater() {
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), externalBlocklistUpdateTimeout)
			count, err := a.refreshExternalBlocklists(ctx)
			cancel()
			if err != nil {
				log.Printf("external blocklist refresh failed; retrying in %s: %v", externalBlocklistUpdateInterval, err)
			} else if count > 0 {
				log.Printf("external blocklists refreshed: %d blocked IPs; next refresh in %s", count, externalBlocklistUpdateInterval)
			}
			time.Sleep(externalBlocklistUpdateInterval)
		}
	}()
}

func (a *App) refreshExternalBlocklists(ctx context.Context) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.settings.ExternalBlocklists) == 0 {
		return 0, nil
	}
	feeds, err := normalizeExternalBlocklists(ctx, a.settings.ExternalBlocklists)
	if err != nil {
		return 0, err
	}
	feeds, blocked, err := a.prepareExternalBlocklists(ctx, feeds, a.settings.ExternalBlocklists, true, "")
	if err != nil {
		return 0, err
	}

	sites, head, tail, err := a.load()
	if err != nil {
		return 0, err
	}
	previousSettings := a.settings
	previousSettingsFile, err := os.ReadFile(a.settingsPath)
	if err != nil {
		return 0, err
	}
	next := a.settings
	next.ExternalBlocklists = feeds
	next.ExternalBlockedIPs = blocked
	next.ExternalBlockedIPCount = len(blocked)
	next.RefreshBlocklists = false
	next.RefreshBlocklistURL = ""
	a.settings = next
	if err := a.saveSettingsLocked(); err != nil {
		a.settings = previousSettings
		return 0, err
	}
	if err := a.saveAndApplyCaddyConfigLocked(head, sites, tail); err != nil {
		a.settings = previousSettings
		if restoreErr := writeFileAtomically(a.settingsPath, previousSettingsFile, 0o600); restoreErr != nil {
			err = fmt.Errorf("%w; restoring previous settings failed: %v", err, restoreErr)
		}
		return 0, fmt.Errorf("caddy rejected the external blocklist refresh; previous config restored: %w", err)
	}
	return len(blocked), nil
}

func validateExternalBlocklistURL(ctx context.Context, value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return "", errors.New("external blocklists must use a valid HTTPS URL without credentials or fragments")
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return "", errors.New("external blocklist URLs may only use the standard HTTPS port")
	}
	host := parsed.Hostname()
	if address, err := netip.ParseAddr(host); err == nil {
		if !isSafeExternalAddress(address) {
			return "", errors.New("external blocklist URLs must not target local or private addresses")
		}
	} else {
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if err != nil || len(addresses) == 0 {
			return "", fmt.Errorf("resolve external blocklist host %q", host)
		}
		for _, address := range addresses {
			if !isSafeExternalAddress(address) {
				return "", errors.New("external blocklist URLs must not resolve to local or private addresses")
			}
		}
	}
	return parsed.String(), nil
}

func (a *App) downloadExternalBlocklists(ctx context.Context, feeds ExternalBlocklists) ([]string, error) {
	entries := map[string]bool{}
	client := *a.httpClient
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	dialer := &net.Dialer{}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, err
		}
		for _, resolved := range addresses {
			if !isSafeExternalAddress(resolved) {
				return nil, errors.New("external blocklist resolved to a local, private, or reserved address")
			}
		}
		if len(addresses) == 0 {
			return nil, errors.New("external blocklist host has no addresses")
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(addresses[0].String(), port))
	}
	client.Transport = transport
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("external blocklist has too many redirects")
		}
		_, err := validateExternalBlocklistURL(req.Context(), req.URL.String())
		return err
	}
	for _, feed := range feeds {
		if _, err := validateExternalBlocklistURL(ctx, feed.URL); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("download blocklist %q: %w", feed.Name, err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("download blocklist %q: HTTP %d", feed.Name, resp.StatusCode)
		}
		err = collectBlocklistEntries(io.LimitReader(resp.Body, 2<<20), entries)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read blocklist %q: %w", feed.Name, err)
		}
	}
	result := make([]string, 0, len(entries))
	for entry := range entries {
		result = append(result, entry)
	}
	sort.Strings(result)
	return result, nil
}

var unsafeExternalPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"), netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("ff00::/8"),
}

func isSafeExternalAddress(address netip.Addr) bool {
	if !address.IsValid() || !address.IsGlobalUnicast() {
		return false
	}
	address = address.Unmap()
	for _, prefix := range unsafeExternalPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func collectBlocklistEntries(reader io.Reader, entries map[string]bool) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		value := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if value == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			if addr, addrErr := netip.ParseAddr(value); addrErr == nil {
				prefix = netip.PrefixFrom(addr, addr.BitLen())
			} else {
				continue
			}
		}
		if !isPublicAddress(prefix.Addr()) {
			continue
		}
		entries[prefix.Masked().String()] = true
		if len(entries) > 50000 {
			return errors.New("external blocklists exceed 50000 entries")
		}
	}
	return scanner.Err()
}
