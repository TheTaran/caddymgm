package main

import "time"

// demoGeoMapResponse is deliberately static so dashboard work can be reviewed
// in a development instance without generating traffic or modifying access logs.
func demoGeoMapResponse() geoMapResponse {
	ips := []geoMapIP{
		{Address: "203.0.113.24", Scope: "external", Country: "Switzerland", CountryCode: "CH", Count: 284, Sites: []string{"caddydev.alonso.lds", "devsso.alonso.lds"}, SiteCounts: map[string]int{"caddydev.alonso.lds": 166, "devsso.alonso.lds": 118}},
		{Address: "198.51.100.42", Scope: "external", Country: "United States", CountryCode: "US", Count: 217, Sites: []string{"devgit.alonso.lds"}, SiteCounts: map[string]int{"devgit.alonso.lds": 217}},
		{Address: "192.0.2.18", Scope: "external", Country: "Singapore", CountryCode: "SG", Count: 164, Sites: []string{"caddydev.alonso.lds"}, SiteCounts: map[string]int{"caddydev.alonso.lds": 164}},
		{Address: "203.0.113.89", Scope: "external", Country: "Germany", CountryCode: "DE", Count: 128, Sites: []string{"devsso.alonso.lds"}, SiteCounts: map[string]int{"devsso.alonso.lds": 128}},
		{Address: "198.51.100.77", Scope: "external", Country: "Brazil", CountryCode: "BR", Count: 96, Sites: []string{"devgit.alonso.lds"}, SiteCounts: map[string]int{"devgit.alonso.lds": 96}},
	}
	return geoMapResponse{
		Available: true,
		Requests:  1_146,
		Locations: []geoMapLocation{
			{Latitude: 47.3769, Longitude: 8.5417, City: "Zurich", Country: "Switzerland", CountryCode: "CH", Count: 284, IPs: []geoMapIP{ips[0]}},
			{Latitude: 40.7128, Longitude: -74.0060, City: "New York", Country: "United States", CountryCode: "US", Count: 217, IPs: []geoMapIP{ips[1]}},
			{Latitude: 1.3521, Longitude: 103.8198, City: "Singapore", Country: "Singapore", CountryCode: "SG", Count: 164, IPs: []geoMapIP{ips[2]}},
			{Latitude: 52.5200, Longitude: 13.4050, City: "Berlin", Country: "Germany", CountryCode: "DE", Count: 128, IPs: []geoMapIP{ips[3]}},
			{Latitude: -23.5505, Longitude: -46.6333, City: "Sao Paulo", Country: "Brazil", CountryCode: "BR", Count: 96, IPs: []geoMapIP{ips[4]}},
			{Latitude: 35.6762, Longitude: 139.6503, City: "Tokyo", Country: "Japan", CountryCode: "JP", Count: 143, IPs: []geoMapIP{{Address: "192.0.2.91", Scope: "external", Country: "Japan", CountryCode: "JP", Count: 143, Sites: []string{"devsso.alonso.lds"}, SiteCounts: map[string]int{"devsso.alonso.lds": 143}}}},
			{Latitude: -33.8688, Longitude: 151.2093, City: "Sydney", Country: "Australia", CountryCode: "AU", Count: 114, IPs: []geoMapIP{{Address: "198.51.100.131", Scope: "external", Country: "Australia", CountryCode: "AU", Count: 114, Sites: []string{"caddydev.alonso.lds"}, SiteCounts: map[string]int{"caddydev.alonso.lds": 114}}}},
		},
		TopIPs: ips,
	}
}

func demoSecurityOverview(now time.Time, spec securityTrendSpec, includeAllEvents bool) securityOverview {
	bucketStart := now.Add(-spec.window).Truncate(spec.bucketDuration)
	bucketCount := int(now.Sub(bucketStart)/spec.bucketDuration) + 1
	overview := securityOverview{
		TrendInterval: spec.intervalLabel,
		RuleCounts:    securityOverviewRuleCounts{SelectedCountries: 3, ManualBlockedIPs: 4, AllowedIPs: 6, ExternalBlockedIPs: 38_520},
		Trends:        make([]securityTrendPoint, bucketCount),
		TopIPs: []securityTopIP{
			{Address: "203.0.113.24", Country: "Switzerland", Count: 47},
			{Address: "198.51.100.42", Country: "United States", Count: 35},
			{Address: "192.0.2.18", Country: "Singapore", Count: 29},
			{Address: "203.0.113.89", Country: "Germany", Count: 22},
			{Address: "198.51.100.77", Country: "Brazil", Count: 18},
		},
	}
	for index := range overview.Trends {
		start := bucketStart.Add(time.Duration(index) * spec.bucketDuration)
		requests := 12 + (index*11)%31
		blocks := (index*3 + 1) % 5
		if index%7 == 3 {
			requests += 38
			blocks += 11
		}
		overview.Trends[index] = securityTrendPoint{
			Label: start.Format(spec.labelFormat), Start: start.Format(time.RFC3339), End: start.Add(spec.bucketDuration).Format(time.RFC3339), Requests: requests, Blocks: blocks,
		}
		overview.Requests += requests
		overview.ManagedBlocks += blocks
	}
	overview.GeoBlocks = overview.ManagedBlocks / 2
	overview.ManualIPBlocks = overview.ManagedBlocks / 5
	overview.ExternalBlocks = overview.ManagedBlocks - overview.GeoBlocks - overview.ManualIPBlocks
	overview.ClientErrors = overview.ManagedBlocks + 31
	overview.ServerErrors = 7
	overview.Unclassified403 = 4
	reasons := []string{"GEO IP rule", "External blocklist", "Manual blocked IP"}
	addresses := []string{"203.0.113.24", "198.51.100.42", "192.0.2.18"}
	countries := []string{"Switzerland", "United States", "Singapore"}
	sites := []string{"caddydev.alonso.lds", "devgit.alonso.lds", "devsso.alonso.lds"}
	for index := 0; index < 12; index++ {
		timestamp := now.Add(-time.Duration(index*7) * time.Minute)
		overview.Events = append(overview.Events, securityOverviewEvent{Time: timestamp.Format(time.RFC3339), Site: sites[index%len(sites)], Address: addresses[index%len(addresses)], Country: countries[index%len(countries)], Reason: reasons[index%len(reasons)]})
	}
	if !includeAllEvents {
		overview.Events = overview.Events[:8]
	}
	return overview
}

func demoAccessLogs(siteID string) []LogEntry {
	methods := []string{"GET", "GET", "POST", "GET", "GET"}
	paths := []string{"/", "/api/status", "/api/session", "/assets/app.js", "/health"}
	statuses := []string{"200", "200", "201", "304", "403", "404", "502"}
	addresses := []string{"203.0.113.24", "198.51.100.42", "192.0.2.18", "203.0.113.89", "198.51.100.77"}
	now := time.Now().UTC()
	entries := make([]LogEntry, 150)
	for index := range entries {
		entries[index] = LogEntry{
			Time: now.Add(-time.Duration(index*45) * time.Second).Format(time.RFC3339), SiteID: siteID, Site: "Demo web host", Action: "access",
			Message: "Synthetic development access log entry", Method: methods[index%len(methods)], Path: paths[index%len(paths)], Status: statuses[index%len(statuses)], IP: addresses[index%len(addresses)],
		}
	}
	return entries
}
