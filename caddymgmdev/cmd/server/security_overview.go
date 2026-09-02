package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"sort"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang/v2"
)

const securityOverviewLogLimit = 5000

type securityOverview struct {
	Requests        int                        `json:"requests"`
	ManagedBlocks   int                        `json:"managedBlocks"`
	GeoBlocks       int                        `json:"geoBlocks"`
	ManualIPBlocks  int                        `json:"manualIPBlocks"`
	ExternalBlocks  int                        `json:"externalBlocks"`
	Unclassified403 int                        `json:"unclassified403"`
	ClientErrors    int                        `json:"clientErrors"`
	ServerErrors    int                        `json:"serverErrors"`
	Events          []securityOverviewEvent    `json:"events"`
	RuleCounts      securityOverviewRuleCounts `json:"ruleCounts"`
}

type securityOverviewEvent struct {
	Time    string `json:"time"`
	Site    string `json:"site"`
	Address string `json:"address"`
	Country string `json:"country,omitempty"`
	Reason  string `json:"reason"`
}

type securityOverviewRuleCounts struct {
	SelectedCountries  int `json:"selectedCountries"`
	ManualBlockedIPs   int `json:"manualBlockedIPs"`
	AllowedIPs         int `json:"allowedIPs"`
	ExternalBlockedIPs int `json:"externalBlockedIPs"`
}

type securityLogRecord struct {
	Timestamp float64 `json:"ts"`
	Status    int     `json:"status"`
	Request   struct {
		ClientIP string `json:"client_ip"`
		RemoteIP string `json:"remote_ip"`
	} `json:"request"`
}

type protectionPrefixSet map[int]map[netip.Addr]struct{}

func newProtectionPrefixSet(values []string) protectionPrefixSet {
	set := make(protectionPrefixSet)
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			if address, addressErr := netip.ParseAddr(value); addressErr == nil {
				prefix = netip.PrefixFrom(address, address.BitLen())
			} else {
				continue
			}
		}
		prefix = prefix.Masked()
		addresses := set[prefix.Bits()]
		if addresses == nil {
			addresses = make(map[netip.Addr]struct{})
			set[prefix.Bits()] = addresses
		}
		addresses[prefix.Addr().Unmap()] = struct{}{}
	}
	return set
}

func (set protectionPrefixSet) contains(address netip.Addr) bool {
	address = address.Unmap()
	for bits, addresses := range set {
		if bits < 0 || bits > address.BitLen() {
			continue
		}
		if _, found := addresses[netip.PrefixFrom(address, bits).Masked().Addr()]; found {
			return true
		}
	}
	return false
}

func (a *App) handleSecurityOverview(w http.ResponseWriter, r *http.Request) {
	includeAllEvents := r.URL.Query().Get("events") == "all"
	period := r.URL.Query().Get("period")
	window := 24 * time.Hour
	switch period {
	case "7d":
		window = 7 * 24 * time.Hour
	case "30d":
		window = 30 * 24 * time.Hour
	case "", "1d":
		period = "1d"
	default:
		writeError(w, http.StatusBadRequest, errors.New("period must be 1d, 7d, or 30d"))
		return
	}
	cutoff := time.Now().Add(-window).Unix()
	a.mu.Lock()
	sites, _, _, err := a.load()
	defaults := a.settings.WebProtection
	external := append([]string(nil), a.settings.ExternalBlockedIPs...)
	retention := a.settings.LogRetention
	hideLocalIPs := a.settings.HideLocalDashboardIPs
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if retention <= 0 {
		retention = 100
	}
	if retention > securityOverviewLogLimit {
		retention = securityOverviewLogLimit
	}

	database, _ := maxminddb.Open(a.geoIPDBPath)
	if database != nil {
		defer database.Close()
	}
	externalSet := newProtectionPrefixSet(external)
	overview := securityOverview{RuleCounts: securityOverviewRuleCounts{
		SelectedCountries: len(defaults.BlockedCountries), ManualBlockedIPs: len(defaults.BlockedIPs), AllowedIPs: len(defaults.AllowedIPs), ExternalBlockedIPs: len(external),
	}}
	for _, site := range sites {
		if !site.LogsEnabled {
			continue
		}
		lines, readErr := readLastLines(accessLogPath(a.accessLogDir, site.ID), retention)
		if readErr != nil {
			continue
		}
		for _, line := range lines {
			var record securityLogRecord
			if json.Unmarshal([]byte(line), &record) != nil || record.Status == 0 {
				continue
			}
			if int64(record.Timestamp) < cutoff {
				continue
			}
			address, hasAddress := securityRecordAddress(record)
			if hideLocalIPs && hasAddress && isLocalDashboardAddress(address) {
				continue
			}
			overview.Requests++
			if record.Status >= 400 && record.Status < 500 {
				overview.ClientErrors++
			}
			if record.Status >= 500 && record.Status < 600 {
				overview.ServerErrors++
			}
			if record.Status != http.StatusForbidden {
				continue
			}
			if !hasAddress {
				overview.Unclassified403++
				continue
			}
			policy, usesGlobal := defaults, true
			if site.ProtectionOverride {
				policy, usesGlobal = site.WebProtection, false
			}
			reason, country := classifyProtectionBlock(address, policy, usesGlobal, externalSet, database)
			if reason == "" {
				overview.Unclassified403++
				continue
			}
			overview.ManagedBlocks++
			switch reason {
			case "GEO IP rule":
				overview.GeoBlocks++
			case "Manual blocked IP":
				overview.ManualIPBlocks++
			case "External blocklist":
				overview.ExternalBlocks++
			}
			overview.Events = append(overview.Events, securityOverviewEvent{Time: time.Unix(int64(record.Timestamp), int64((record.Timestamp-float64(int64(record.Timestamp)))*1e9)).Format(time.RFC3339), Site: site.Address, Address: address.String(), Country: country, Reason: reason})
		}
	}
	sort.Slice(overview.Events, func(i, j int) bool { return overview.Events[i].Time > overview.Events[j].Time })
	if !includeAllEvents && len(overview.Events) > 20 {
		overview.Events = overview.Events[:20]
	}
	writeJSON(w, http.StatusOK, overview)
}

func securityRecordAddress(record securityLogRecord) (netip.Addr, bool) {
	value := strings.TrimSpace(record.Request.ClientIP)
	if value == "" {
		value = strings.TrimSpace(record.Request.RemoteIP)
	}
	address, err := netip.ParseAddr(value)
	return address.Unmap(), err == nil
}

func classifyProtectionBlock(address netip.Addr, policy WebProtection, usesGlobal bool, external protectionPrefixSet, database *maxminddb.Reader) (string, string) {
	if !policy.Enabled || newProtectionPrefixSet(policy.AllowedIPs).contains(address) {
		return "", ""
	}
	countryCode, country := "", ""
	if database != nil && isPublicAddress(address) {
		var record geoCityRecord
		if database.Lookup(address).Decode(&record) == nil {
			countryCode = strings.ToUpper(geoFirstNonEmpty(record.Country.ISOCode, record.RegisteredCountry.ISOCode))
			country = geoFirstNonEmpty(
				record.Country.Names["en"],
				record.Country.Names["de"],
				record.RegisteredCountry.Names["en"],
				record.RegisteredCountry.Names["de"],
				countryCode,
			)
		}
	}
	if usesGlobal && external.contains(address) {
		return "External blocklist", country
	}
	if newProtectionPrefixSet(policy.BlockedIPs).contains(address) {
		return "Manual blocked IP", country
	}
	if len(policy.BlockedCountries) == 0 {
		return "", ""
	}
	selected := false
	for _, code := range policy.BlockedCountries {
		if code == countryCode {
			selected = true
			break
		}
	}
	if (policy.CountryMode == "block" && selected) || (policy.CountryMode == "allow" && !selected) {
		return "GEO IP rule", country
	}
	return "", ""
}
