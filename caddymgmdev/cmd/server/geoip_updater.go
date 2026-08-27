package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	geoIPUpdateInterval = 72 * time.Hour
	geoIPRetryInterval  = time.Hour
	geoIPUpdateTimeout  = 15 * time.Minute
)

func (a *App) startGeoIPUpdater() {
	accountID := strings.TrimSpace(os.Getenv("MAXMIND_ACCOUNT_ID"))
	licenseKey := strings.TrimSpace(os.Getenv("MAXMIND_LICENSE_KEY"))
	if accountID == "" || licenseKey == "" {
		log.Printf("GeoLite2 updates disabled: MAXMIND_ACCOUNT_ID or MAXMIND_LICENSE_KEY is missing")
		return
	}

	databaseDirectory := filepath.Dir(a.geoIPDBPath)
	if err := os.MkdirAll(databaseDirectory, 0o750); err != nil {
		log.Printf("GeoLite2 updates disabled: prepare database directory: %v", err)
		return
	}
	go func() {
		for {
			next := geoIPUpdateInterval
			ctx, cancel := context.WithTimeout(context.Background(), geoIPUpdateTimeout)
			cmd := exec.CommandContext(ctx, "/usr/local/bin/geoipupdate", "-d", databaseDirectory)
			cmd.Env = append(os.Environ(),
				"GEOIPUPDATE_ACCOUNT_ID="+accountID,
				"GEOIPUPDATE_LICENSE_KEY="+licenseKey,
				"GEOIPUPDATE_EDITION_IDS=GeoLite2-City",
				"GEOIPUPDATE_FREQUENCY=0",
			)
			output, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				next = geoIPRetryInterval
				log.Printf("GeoLite2 update failed; retrying in %s: %v: %s", next, err, sanitizeUpdaterOutput(string(output), accountID, licenseKey))
			} else {
				log.Printf("GeoLite2 database update completed; next check in %s", next)
			}
			time.Sleep(next)
		}
	}()
}

func sanitizeUpdaterOutput(value string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[redacted]")
		}
	}
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value)
	value = strings.TrimSpace(value)
	if len(value) > 512 {
		value = value[:512] + "..."
	}
	return value
}
