package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const oidcAuditFieldLimit = 512

func auditField(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > oidcAuditFieldLimit {
		return value[:oidcAuditFieldLimit]
	}
	return value
}

func (a *App) requestClientIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		remoteHost = strings.TrimSpace(r.RemoteAddr)
	}
	if a.trustedProxies != nil && a.trustedProxies.contains(r.Context(), r.RemoteAddr) {
		for _, candidate := range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
			if ip, err := netip.ParseAddr(strings.TrimSpace(candidate)); err == nil {
				return ip.Unmap().String()
			}
		}
	}
	if ip, err := netip.ParseAddr(remoteHost); err == nil {
		return ip.Unmap().String()
	}
	return "unknown"
}

func (a *App) recordOIDCAudit(r *http.Request, action, status, username, email, site, message string) {
	a.mu.Lock()
	limit := a.settings.LogRetention
	a.mu.Unlock()
	if limit <= 0 {
		limit = 100
	}
	entry := LogEntry{
		Time: time.Now().UTC().Format(time.RFC3339), Action: auditField(action), Status: auditField(status),
		Username: auditField(username), Email: auditField(email), IP: a.requestClientIP(r),
		Site: auditField(site), Message: auditField(message),
	}
	a.auditMu.Lock()
	defer a.auditMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(a.oidcAuditLog), 0o750); err != nil {
		log.Printf("write OIDC audit log: %v", err)
		return
	}
	file, err := os.OpenFile(a.oidcAuditLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		log.Printf("write OIDC audit log: %v", err)
		return
	}
	_ = file.Chmod(0o600)
	if err := json.NewEncoder(file).Encode(entry); err != nil {
		_ = file.Close()
		log.Printf("write OIDC audit log: %v", err)
		return
	}
	if err := file.Close(); err != nil {
		log.Printf("write OIDC audit log: %v", err)
		return
	}
	if err := trimOIDCAuditFile(a.oidcAuditLog, limit); err != nil {
		log.Printf("trim OIDC audit log: %v", err)
	}
}

func trimOIDCAuditFile(path string, limit int) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if len(lines) <= limit {
		return nil
	}
	return writeFileAtomically(path, []byte(strings.Join(lines[len(lines)-limit:], "\n")+"\n"), 0o600)
}

func (a *App) readOIDCAuditLogsLocked() []LogEntry {
	a.auditMu.Lock()
	defer a.auditMu.Unlock()
	limit := a.settings.LogRetention
	if limit <= 0 {
		limit = 100
	}
	lines, err := readLastLines(a.oidcAuditLog, limit)
	if err != nil {
		return nil
	}
	entries := make([]LogEntry, 0, len(lines))
	for _, line := range lines {
		var entry LogEntry
		if json.Unmarshal([]byte(line), &entry) == nil && entry.Time != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}
