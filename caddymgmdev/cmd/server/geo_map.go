package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang/v2"
)

func (a *App) handleGeoFlag(w http.ResponseWriter, r *http.Request) {
	code := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/geo-flag/")))
	if !countryCodePattern.MatchString(strings.ToUpper(code)) {
		writeError(w, http.StatusBadRequest, errors.New("invalid country code"))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://flagcdn.com/32x24/"+code+".png", nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, errors.New("country flag unavailable"))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		writeError(w, http.StatusBadGateway, errors.New("country flag unavailable"))
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=604800")
	_, _ = io.Copy(w, io.LimitReader(resp.Body, 64<<10))
}

const geoTopIPLimit = 1_000

type geoMapIP struct {
	Address     string         `json:"address"`
	Scope       string         `json:"scope"`
	Country     string         `json:"country,omitempty"`
	CountryCode string         `json:"countryCode,omitempty"`
	Count       int            `json:"count"`
	Sites       []string       `json:"sites"`
	SiteCounts  map[string]int `json:"siteCounts"`
}

type geoMapLocation struct {
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	City        string     `json:"city,omitempty"`
	Country     string     `json:"country,omitempty"`
	CountryCode string     `json:"countryCode,omitempty"`
	Count       int        `json:"count"`
	IPs         []geoMapIP `json:"ips"`
}

type geoCityRecord struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

type geoMapResponse struct {
	Available bool             `json:"available"`
	Message   string           `json:"message,omitempty"`
	Requests  int              `json:"requests"`
	Locations []geoMapLocation `json:"locations"`
	TopIPs    []geoMapIP       `json:"topIps"`
}

type geoCountry struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type geoCountryCache struct {
	DatabaseModified int64        `json:"databaseModified"`
	DatabaseSize     int64        `json:"databaseSize"`
	Countries        []geoCountry `json:"countries"`
}

func (a *App) handleGeoCountries(w http.ResponseWriter, _ *http.Request) {
	countries, err := a.loadGeoCountries()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"countries": countries})
}

func (a *App) loadGeoCountries() ([]geoCountry, error) {
	info, err := os.Stat(a.geoIPDBPath)
	if err != nil {
		return nil, errors.New("GeoLite2 City database is unavailable")
	}
	a.geoCountriesMu.Lock()
	defer a.geoCountriesMu.Unlock()
	if len(a.geoCountries) > 0 && info.ModTime().Equal(a.geoCountriesMTime) {
		return append([]geoCountry(nil), a.geoCountries...), nil
	}
	if countries := readGeoCountryCache(a.geoIPDBPath, info); len(countries) > 0 {
		a.geoCountries = countries
		a.geoCountriesMTime = info.ModTime()
		return append([]geoCountry(nil), countries...), nil
	}

	database, err := maxminddb.Open(a.geoIPDBPath)
	if err != nil {
		return nil, errors.New("GeoLite2 City database is unavailable")
	}
	defer database.Close()

	countriesByCode := make(map[string]string)
	for result := range database.Networks() {
		var record geoCityRecord
		if err := result.Decode(&record); err != nil {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(record.Country.ISOCode))
		if !countryCodePattern.MatchString(code) {
			continue
		}
		name := geoFirstNonEmpty(record.Country.Names["en"], record.Country.Names["de"], code)
		if _, exists := countriesByCode[code]; !exists {
			countriesByCode[code] = name
		}
	}
	countries := make([]geoCountry, 0, len(countriesByCode))
	for code, name := range countriesByCode {
		countries = append(countries, geoCountry{Code: code, Name: name})
	}
	sort.Slice(countries, func(i, j int) bool { return countries[i].Name < countries[j].Name })
	a.geoCountries = append([]geoCountry(nil), countries...)
	a.geoCountriesMTime = info.ModTime()
	writeGeoCountryCache(a.geoIPDBPath, info, countries)
	return countries, nil
}

func readGeoCountryCache(databasePath string, info os.FileInfo) []geoCountry {
	content, err := os.ReadFile(databasePath + ".countries.json")
	if err != nil {
		return nil
	}
	var cache geoCountryCache
	if json.Unmarshal(content, &cache) != nil || cache.DatabaseModified != info.ModTime().UnixNano() || cache.DatabaseSize != info.Size() {
		return nil
	}
	return cache.Countries
}

func writeGeoCountryCache(databasePath string, info os.FileInfo, countries []geoCountry) {
	content, err := json.Marshal(geoCountryCache{DatabaseModified: info.ModTime().UnixNano(), DatabaseSize: info.Size(), Countries: countries})
	if err != nil {
		return
	}
	cachePath := databasePath + ".countries.json"
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o750); err != nil {
		return
	}
	_ = writeFileAtomically(cachePath, append(content, '\n'), 0o600)
}

type geoIPAggregate struct {
	entry geoMapIP
	sites map[string]struct{}
}

type geoLocationAggregate struct {
	entry geoMapLocation
	ips   map[string]*geoIPAggregate
}

func (a *App) handleGeoMap(w http.ResponseWriter, _ *http.Request) {
	database, databaseErr := maxminddb.Open(a.geoIPDBPath)
	if databaseErr == nil {
		defer database.Close()
	}

	a.mu.Lock()
	sites, _, _, err := a.load()
	retention := a.settings.LogRetention
	a.mu.Unlock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if retention <= 0 {
		retention = 100
	}
	if retention > geoMapLogLimit {
		retention = geoMapLogLimit
	}

	locationsByKey := make(map[string]*geoLocationAggregate)
	topByAddress := make(map[string]*geoIPAggregate)
	mappedRequests := 0
	for _, site := range sites {
		if !site.LogsEnabled {
			continue
		}
		lines, readErr := readLastLines(accessLogPath(a.accessLogDir, site.ID), retention)
		if readErr != nil {
			continue
		}
		for _, line := range lines {
			address, ok := accessLogClientAddress(line)
			if !ok || !address.IsValid() {
				continue
			}
			topIP := addGeoIPAggregate(topByAddress, address.String(), site.Address)
			if databaseErr != nil || !isPublicAddress(address) {
				continue
			}

			var record geoCityRecord
			if lookupErr := database.Lookup(address).Decode(&record); lookupErr != nil {
				continue
			}
			topIP.entry.Country = geoFirstNonEmpty(record.Country.Names["en"], record.Country.Names["de"])
			topIP.entry.CountryCode = record.Country.ISOCode
			if record.Location.Latitude == 0 && record.Location.Longitude == 0 {
				continue
			}
			key := record.Country.ISOCode + "|" + record.City.Names["en"] + "|" + coordinateKey(record.Location.Latitude) + "|" + coordinateKey(record.Location.Longitude)
			location := locationsByKey[key]
			if location == nil {
				country := geoFirstNonEmpty(record.Country.Names["en"], record.Country.Names["de"])
				city := geoFirstNonEmpty(record.City.Names["en"], record.City.Names["de"])
				location = &geoLocationAggregate{entry: geoMapLocation{
					Latitude: record.Location.Latitude, Longitude: record.Location.Longitude,
					City: city, Country: country, CountryCode: record.Country.ISOCode,
				}, ips: make(map[string]*geoIPAggregate)}
				locationsByKey[key] = location
			}
			location.entry.Count++
			addGeoIPAggregate(location.ips, address.String(), site.Address)
			mappedRequests++
		}
	}

	locations := make([]geoMapLocation, 0, len(locationsByKey))
	for _, location := range locationsByKey {
		location.entry.IPs = sortedGeoIPs(location.ips, 0)
		locations = append(locations, location.entry)
	}
	sort.Slice(locations, func(i, j int) bool { return locations[i].Count > locations[j].Count })

	response := geoMapResponse{
		Available: databaseErr == nil,
		Requests:  mappedRequests,
		Locations: locations,
		TopIPs:    sortedGeoIPs(topByAddress, geoTopIPLimit),
	}
	if databaseErr != nil {
		response.Message = "GeoLite2 City database is not configured"
		if !errors.Is(databaseErr, os.ErrNotExist) {
			response.Message = "GeoLite2 City database is unavailable"
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func addGeoIPAggregate(aggregates map[string]*geoIPAggregate, address, site string) *geoIPAggregate {
	item := aggregates[address]
	if item == nil {
		item = &geoIPAggregate{entry: geoMapIP{Address: address, Scope: geoIPScope(address), SiteCounts: make(map[string]int)}, sites: make(map[string]struct{})}
		aggregates[address] = item
	}
	item.entry.Count++
	item.entry.SiteCounts[site]++
	item.sites[site] = struct{}{}
	return item
}

func sortedGeoIPs(aggregates map[string]*geoIPAggregate, limit int) []geoMapIP {
	entries := make([]geoMapIP, 0, len(aggregates))
	for _, item := range aggregates {
		for site := range item.sites {
			item.entry.Sites = append(item.entry.Sites, site)
		}
		sort.Strings(item.entry.Sites)
		entries = append(entries, item.entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count == entries[j].Count {
			return entries[i].Address < entries[j].Address
		}
		return entries[i].Count > entries[j].Count
	})
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return entries
}

func geoIPScope(value string) string {
	address, err := netip.ParseAddr(value)
	if err != nil {
		return "external"
	}
	address = address.Unmap()
	if address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() {
		return "internal"
	}
	return "external"
}

func accessLogClientAddress(line string) (netip.Addr, bool) {
	var payload struct {
		Request struct {
			ClientIP string `json:"client_ip"`
			RemoteIP string `json:"remote_ip"`
		} `json:"request"`
	}
	if json.Unmarshal([]byte(line), &payload) != nil {
		return netip.Addr{}, false
	}
	value := strings.TrimSpace(payload.Request.ClientIP)
	if value == "" {
		value = strings.TrimSpace(payload.Request.RemoteIP)
	}
	address, err := netip.ParseAddr(value)
	return address.Unmap(), err == nil
}

func isPublicAddress(address netip.Addr) bool {
	return address.IsValid() && address.IsGlobalUnicast() && !address.IsPrivate() && !address.IsLoopback()
}

func coordinateKey(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}

func geoFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
