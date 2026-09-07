package main

import (
	"bufio"
	"bytes"
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
	dnsAllowlistUpdateInterval      = 12 * time.Hour
	dnsAllowlistUpdateTimeout       = time.Minute
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
	Name            string   `json:"name"`
	Reference       string   `json:"reference,omitempty"`
	Mode            string   `json:"mode"`
	Entries         []string `json:"entries"`
	ResolvedEntries []string `json:"resolvedEntries,omitempty"`
	ResolvedAt      string   `json:"resolvedAt,omitempty"`
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
		entries, err := normalizeManualIPListEntries(value.Entries, mode == "allow")
		if err != nil {
			return nil, err
		}
		seen[key] = true
		resolved, err := normalizeManualResolvedEntries(value.ResolvedEntries, mode == "allow")
		if err != nil {
			return nil, err
		}
		result = append(result, ManualIPList{Name: name, Reference: strings.TrimSpace(value.Reference), Mode: mode, Entries: entries, ResolvedEntries: resolved, ResolvedAt: strings.TrimSpace(value.ResolvedAt)})
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name) })
	return result, nil
}

// normalizeManualIPListEntries validates manual values without changing their
// notation. Named manual lists are administrator-managed records, so a host
// address remains a host address and an entered CIDR remains exactly that CIDR.
func normalizeManualIPListEntries(values []string, allowPrivate bool) ([]string, error) {
	if len(values) > 5000 {
		return nil, errors.New("a protection list may contain at most 5000 entries")
	}
	result, seen := make([]string, 0, len(values)), map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		address, err := netip.ParseAddr(value)
		if err != nil {
			prefix, prefixErr := netip.ParsePrefix(value)
			if prefixErr != nil {
				if isDNSName(value) {
					value = strings.ToLower(value)
					if !seen[value] {
						seen[value] = true
						result = append(result, value)
					}
					continue
				}
				return nil, errors.New("IP protection entries must be valid IP addresses, CIDR ranges, or DNS names")
			}
			address = prefix.Addr()
		}
		address = address.Unmap()
		if !allowPrivate && (address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsMulticast() || address.IsUnspecified()) {
			return nil, errors.New("block lists cannot contain private, loopback, link-local, multicast, or unspecified addresses")
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result, nil
}

func normalizeManualResolvedEntries(values []string, allowPrivate bool) ([]string, error) {
	result, seen := make([]string, 0, len(values)), map[string]bool{}
	for _, value := range values {
		address, err := netip.ParseAddr(strings.TrimSpace(value))
		if err != nil {
			return nil, errors.New("resolved DNS entries must be IP addresses")
		}
		address = address.Unmap()
		if !allowPrivate && (address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsMulticast() || address.IsUnspecified()) {
			return nil, errors.New("resolved DNS block entries cannot be private, loopback, link-local, multicast, or unspecified addresses")
		}
		value = address.String()
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result, nil
}

func isDNSName(value string) bool {
	value = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	if len(value) == 0 || len(value) > 253 || !strings.Contains(value, ".") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if !(character == '-' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

func manualIPListEntries(values ManualIPLists, mode string) []string {
	entries, seen := make([]string, 0), map[string]bool{}
	for _, value := range values {
		if value.Mode != mode {
			continue
		}
		for _, entry := range value.Entries {
			if _, err := netip.ParseAddr(entry); err != nil {
				if _, prefixErr := netip.ParsePrefix(entry); prefixErr != nil {
					continue
				}
			}
			if !seen[entry] {
				seen[entry] = true
				entries = append(entries, entry)
			}
		}
		for _, entry := range value.ResolvedEntries {
			if !seen[entry] {
				seen[entry] = true
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

// resolveManualAllowlistDNS updates DNS-name entries. If a later refresh fails,
// the last successful addresses remain active instead of weakening an existing rule.
func resolveManualAllowlistDNS(ctx context.Context, lists, previous ManualIPLists) (ManualIPLists, bool, error) {
	prior := make(map[string]ManualIPList, len(previous))
	for _, list := range previous {
		prior[strings.ToLower(list.Name)] = list
	}
	changed := false
	for index := range lists {
		list := &lists[index]
		names := make([]string, 0)
		for _, entry := range list.Entries {
			if isDNSName(entry) {
				names = append(names, entry)
			}
		}
		if len(names) == 0 {
			continue
		}
		resolved, seen := make([]string, 0), map[string]bool{}
		for _, name := range names {
			addresses, err := net.DefaultResolver.LookupIP(ctx, "ip", name)
			if err != nil || len(addresses) == 0 {
				continue
			}
			for _, address := range addresses {
				parsed, ok := netip.AddrFromSlice(address)
				if !ok {
					continue
				}
				parsed = parsed.Unmap()
				if list.Mode == "block" && (parsed.IsPrivate() || parsed.IsLoopback() || parsed.IsLinkLocalUnicast() || parsed.IsMulticast() || parsed.IsUnspecified()) {
					continue
				}
				value := parsed.String()
				if !seen[value] {
					seen[value] = true
					resolved = append(resolved, value)
				}
			}
		}
		if len(resolved) == 0 {
			if cached, ok := prior[strings.ToLower(list.Name)]; ok && len(cached.ResolvedEntries) > 0 {
				resolved = append([]string(nil), cached.ResolvedEntries...)
			} else if len(list.ResolvedEntries) > 0 {
				resolved = append([]string(nil), list.ResolvedEntries...)
			} else {
				return nil, false, fmt.Errorf("could not resolve DNS allowlist entries for %q", list.Name)
			}
		}
		sort.Strings(resolved)
		if !slicesEqual(list.ResolvedEntries, resolved) {
			list.ResolvedEntries = resolved
			changed = true
		}
		list.ResolvedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return lists, changed, nil
}

func slicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
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

func (a *App) startManualAllowlistDNSUpdater() {
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), dnsAllowlistUpdateTimeout)
			changed, err := a.refreshManualAllowlistDNS(ctx)
			cancel()
			if err != nil {
				log.Printf("DNS allowlist refresh failed; retrying in %s: %v", dnsAllowlistUpdateInterval, err)
			} else if changed {
				log.Printf("DNS allowlist addresses refreshed; next refresh in %s", dnsAllowlistUpdateInterval)
			}
			time.Sleep(dnsAllowlistUpdateInterval)
		}
	}()
}

func (a *App) refreshManualAllowlistDNS(ctx context.Context) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.settings.ManualIPLists) == 0 {
		return false, nil
	}
	nextLists, changed, err := resolveManualAllowlistDNS(ctx, append(ManualIPLists(nil), a.settings.ManualIPLists...), a.settings.ManualIPLists)
	if err != nil || !changed {
		return changed, err
	}
	sites, head, tail, err := a.load()
	if err != nil {
		return false, err
	}
	previousSettings := a.settings
	previousSettingsFile, err := os.ReadFile(a.settingsPath)
	if err != nil {
		return false, err
	}
	a.settings.ManualIPLists = nextLists
	a.settings.WebProtection.AllowedIPs = manualIPListEntries(nextLists, "allow")
	if err := a.saveSettingsLocked(); err != nil {
		a.settings = previousSettings
		return false, err
	}
	if err := a.saveAndApplyCaddyConfigLocked(head, sites, tail); err != nil {
		a.settings = previousSettings
		if restoreErr := writeFileAtomically(a.settingsPath, previousSettingsFile, 0o600); restoreErr != nil {
			return false, fmt.Errorf("%w; restoring DNS allowlist settings failed: %v", err, restoreErr)
		}
		return false, err
	}
	return true, nil
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
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return collectJSONBlocklistEntries(trimmed, entries)
	}
	return collectTextBlocklistEntries(bytes.NewReader(content), entries)
}

// collectJSONBlocklistEntries accepts both regular JSON documents and newline-
// delimited JSON feeds. Every string value is validated as an IP or CIDR, so
// feeds may use fields such as cidr, ip, address, network, or prefix without
// requiring source-specific parsers.
func collectJSONBlocklistEntries(content []byte, entries map[string]bool) error {
	decoder := json.NewDecoder(bytes.NewReader(content))
	for {
		var value any
		err := decoder.Decode(&value)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parse JSON blocklist: %w", err)
		}
		if err := collectJSONBlocklistValue(value, entries); err != nil {
			return err
		}
	}
}

func collectJSONBlocklistValue(value any, entries map[string]bool) error {
	switch item := value.(type) {
	case string:
		return collectBlocklistEntry(item, entries)
	case []any:
		for _, value := range item {
			if err := collectJSONBlocklistValue(value, entries); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, value := range item {
			if err := collectJSONBlocklistValue(value, entries); err != nil {
				return err
			}
		}
	}
	return nil
}

func collectTextBlocklistEntries(reader io.Reader, entries map[string]bool) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		value := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if value == "" {
			continue
		}
		// Threat feeds commonly append a classification after whitespace or a
		// semicolon (for example: "1.2.3.0/24 ; SBL123"). Only the first token
		// is the address/CIDR and must be passed to the IP parser.
		if fields := strings.Fields(value); len(fields) > 0 {
			value = fields[0]
		}
		if err := collectBlocklistEntry(value, entries); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func collectBlocklistEntry(value string, entries map[string]bool) error {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
	if err != nil {
		address, addressErr := netip.ParseAddr(strings.TrimSpace(value))
		if addressErr != nil {
			return nil
		}
		prefix = netip.PrefixFrom(address, address.BitLen())
	}
	if !isPublicAddress(prefix.Addr()) {
		return nil
	}
	entries[prefix.Masked().String()] = true
	if len(entries) > 50000 {
		return errors.New("external blocklists exceed 50000 entries")
	}
	return nil
}
