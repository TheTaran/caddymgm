package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func openThroughputDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; CREATE TABLE IF NOT EXISTS throughput_requests (fingerprint TEXT PRIMARY KEY, ts REAL NOT NULL, site_id TEXT NOT NULL, ingress INTEGER NOT NULL DEFAULT 0, egress INTEGER NOT NULL DEFAULT 0); CREATE INDEX IF NOT EXISTS idx_throughput_time_site ON throughput_requests(ts, site_id);`)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (a *App) startThroughputIngestor() {
	if a.demoData {
		return
	}
	dbPath := env("CADDYMGM_THROUGHPUT_DB_PATH", filepath.Join(filepath.Dir(a.settingsPath), "throughput.db"))
	db, err := openThroughputDB(dbPath)
	if err != nil {
		log.Printf("throughput database unavailable: %v", err)
		return
	}
	a.throughputDB = db
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			a.ingestThroughput()
		}
	}()
	a.ingestThroughput()
}

func (a *App) ingestThroughput() {
	db := a.throughputDB
	if db == nil {
		return
	}
	sites, _, _, err := a.load()
	if err != nil {
		return
	}
	tx, err := db.Begin()
	if err != nil {
		return
	}
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO throughput_requests (fingerprint, ts, site_id, ingress, egress) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return
	}
	defer stmt.Close()
	for _, site := range sites {
		if !site.LogsEnabled {
			continue
		}
		lines, readErr := readLastLines(accessLogPath(a.accessLogDir, site.ID), 5000)
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
			var ingress int64
			if values := payload.Request.Headers["Content-Length"]; len(values) > 0 {
				_, _ = fmt.Sscan(values[0], &ingress)
			}
			hash := sha256.Sum256([]byte(site.ID + "\x00" + line))
			_, _ = stmt.Exec(hex.EncodeToString(hash[:]), payload.TS, site.ID, maxInt64(ingress), maxInt64(payload.Size))
		}
	}
	_ = tx.Commit()
	_, _ = db.Exec("DELETE FROM throughput_requests WHERE ts < ?", float64(time.Now().Add(-time.Hour*24*366).Unix()))
}

func (a *App) queryThroughput(spec securityTrendSpec, siteID string) (throughputResponse, error) {
	now := time.Now().UTC()
	start := now.Add(-spec.window)
	bucketStart := start.Truncate(spec.bucketDuration)
	count := int(now.Sub(bucketStart)/spec.bucketDuration) + 1
	points := make([]throughputPoint, count)
	for i := range points {
		t := bucketStart.Add(time.Duration(i) * spec.bucketDuration)
		points[i] = throughputPoint{Label: t.Format(spec.labelFormat), Start: t.Format(time.RFC3339)}
	}
	query := "SELECT CAST((ts-?)/? AS INTEGER), COALESCE(SUM(ingress),0), COALESCE(SUM(egress),0) FROM throughput_requests WHERE ts>=?"
	args := []any{float64(bucketStart.Unix()), spec.bucketDuration.Seconds(), float64(start.Unix())}
	if siteID != "" {
		query += " AND site_id=?"
		args = append(args, siteID)
	}
	query += " GROUP BY 1"
	rows, err := a.throughputDB.Query(query, args...)
	if err != nil {
		return throughputResponse{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var index int
		var ingress, egress int64
		if err := rows.Scan(&index, &ingress, &egress); err != nil {
			return throughputResponse{}, err
		}
		if index >= 0 && index < len(points) {
			points[index].Ingress += ingress
			points[index].Egress += egress
		}
	}
	if err := rows.Err(); err != nil {
		return throughputResponse{}, err
	}
	var ingress, egress int64
	for _, point := range points {
		ingress += point.Ingress
		egress += point.Egress
	}
	seconds := int64(spec.window / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return throughputResponse{Ingress: ingress, Egress: egress, IngressRate: ingress / seconds, EgressRate: egress / seconds, Interval: spec.intervalLabel, Points: points}, nil
}
