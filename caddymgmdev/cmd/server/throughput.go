package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type throughputPoint struct {
	Label   string `json:"label"`
	Start   string `json:"start"`
	Ingress int64  `json:"ingress"`
	Egress  int64  `json:"egress"`
}

type throughputResponse struct {
	Ingress     int64             `json:"ingress"`
	Egress      int64             `json:"egress"`
	IngressRate int64             `json:"ingressRate"`
	EgressRate  int64             `json:"egressRate"`
	Interval    string            `json:"interval"`
	Points      []throughputPoint `json:"points"`
}

func (a *App) handleThroughput(w http.ResponseWriter, r *http.Request) {
	_, spec, err := securityTrendSpecForPeriod(strings.TrimSpace(r.URL.Query().Get("period")))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	siteID := strings.TrimSpace(r.URL.Query().Get("site"))
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.demoData {
		writeJSON(w, http.StatusOK, demoThroughput(spec, siteID))
		return
	}
	if a.throughputDB != nil {
		if data, queryErr := a.queryThroughput(spec, siteID); queryErr == nil {
			writeJSON(w, http.StatusOK, data)
			return
		}
	}
	sites, _, _, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	now := time.Now().UTC()
	start := now.Add(-spec.window)
	bucketStart := start.Truncate(spec.bucketDuration)
	count := int(now.Sub(bucketStart)/spec.bucketDuration) + 1
	points := make([]throughputPoint, count)
	for i := range points {
		t := bucketStart.Add(time.Duration(i) * spec.bucketDuration)
		points[i] = throughputPoint{Label: t.Format(spec.labelFormat), Start: t.Format(time.RFC3339)}
	}
	for _, site := range sites {
		if siteID != "" && site.ID != siteID {
			continue
		}
		if !site.LogsEnabled {
			continue
		}
		lines, readErr := readLastLines(accessLogPath(a.accessLogDir, site.ID), a.settings.LogRetention)
		if readErr != nil {
			continue
		}
		for _, line := range lines {
			var payload struct {
				TS      float64 `json:"ts"`
				Size    int64   `json:"size"`
				Request struct {
					Headers map[string][]string `json:"headers"`
				} `json:"request"`
			}
			if json.Unmarshal([]byte(line), &payload) != nil || payload.TS == 0 {
				continue
			}
			t := time.Unix(int64(payload.TS), int64((payload.TS-float64(int64(payload.TS)))*1e9)).UTC()
			if t.Before(start) {
				continue
			}
			i := int(t.Sub(bucketStart) / spec.bucketDuration)
			if i < 0 || i >= len(points) {
				continue
			}
			points[i].Egress += maxInt64(payload.Size)
			if values := payload.Request.Headers["Content-Length"]; len(values) > 0 {
				if n, parseErr := strconv.ParseInt(values[0], 10, 64); parseErr == nil {
					points[i].Ingress += maxInt64(n)
				}
			}
		}
	}
	returnThroughput(w, points, spec)
}

func returnThroughput(w http.ResponseWriter, points []throughputPoint, spec securityTrendSpec) {
	var ingress, egress int64
	for _, point := range points {
		ingress += point.Ingress
		egress += point.Egress
	}
	seconds := int64(spec.window / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	writeJSON(w, http.StatusOK, throughputResponse{Ingress: ingress, Egress: egress, IngressRate: ingress / seconds, EgressRate: egress / seconds, Interval: spec.intervalLabel, Points: points})
}

func maxInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
