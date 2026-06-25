package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"embed"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	managedStart = "# caddymgm:start"
	managedEnd   = "# caddymgm:end"
)

//go:embed web/*
var webFS embed.FS

type Site struct {
	ID                   string `json:"id"`
	Address              string `json:"address"`
	Mode                 string `json:"mode"`
	Upstream             string `json:"upstream,omitempty"`
	Root                 string `json:"root,omitempty"`
	ExtraDirectives      string `json:"extraDirectives,omitempty"`
	LogsEnabled          bool   `json:"logsEnabled"`
	TLSMode              string `json:"tlsMode"`
	ACMEIssuerID         string `json:"acmeIssuerId,omitempty"`
	CertificateExpiresAt string `json:"certificateExpiresAt,omitempty"`
	Enabled              bool   `json:"enabled"`
}

type sitePayload struct {
	Address         string `json:"address"`
	Mode            string `json:"mode"`
	Upstream        string `json:"upstream,omitempty"`
	Root            string `json:"root,omitempty"`
	ExtraDirectives string `json:"extraDirectives,omitempty"`
	LogsEnabled     *bool  `json:"logsEnabled,omitempty"`
	TLSMode         string `json:"tlsMode,omitempty"`
	ACMEIssuerID    string `json:"acmeIssuerId,omitempty"`
	Enabled         bool   `json:"enabled"`
}

func (p sitePayload) site(id string, defaultLogsEnabled bool) Site {
	logsEnabled := defaultLogsEnabled
	if p.LogsEnabled != nil {
		logsEnabled = *p.LogsEnabled
	}
	return Site{
		ID:              id,
		Address:         p.Address,
		Mode:            p.Mode,
		Upstream:        p.Upstream,
		Root:            p.Root,
		ExtraDirectives: p.ExtraDirectives,
		LogsEnabled:     logsEnabled,
		TLSMode:         p.TLSMode,
		ACMEIssuerID:    p.ACMEIssuerID,
		Enabled:         p.Enabled,
	}
}

type ACMEIssuer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DirectoryURL string `json:"directoryUrl"`
	Email        string `json:"email,omitempty"`
	BuiltIn      bool   `json:"builtIn,omitempty"`
}

type Settings struct {
	AppName      string       `json:"appName"`
	AuthEnabled  bool         `json:"authEnabled"`
	Username     string       `json:"username"`
	Password     string       `json:"password,omitempty"`
	PasswordHash string       `json:"passwordHash,omitempty"`
	ConfigPath   string       `json:"configPath"`
	LogRetention int          `json:"logRetention"`
	ACMEIssuers  []ACMEIssuer `json:"acmeIssuers,omitempty"`
	CaddyMode    string       `json:"caddyMode"`
	CaddyAPIURL  string       `json:"caddyApiUrl"`
}

type LogEntry struct {
	Time    string `json:"time"`
	SiteID  string `json:"siteId,omitempty"`
	Site    string `json:"site,omitempty"`
	Action  string `json:"action"`
	Message string `json:"message"`
	Method  string `json:"method,omitempty"`
	Path    string `json:"path,omitempty"`
	Status  string `json:"status,omitempty"`
}

type App struct {
	mu           sync.Mutex
	configPath   string
	settingsPath string
	caddyMode    string
	caddyAPIURL  string
	accessLogDir string
	caddyLogDir  string
	caddyDataDir string
	httpClient   *http.Client
	settings     Settings
	logs         []LogEntry
	sessions     map[string]time.Time
}

func main() {
	configPath := env("CADDY_CONFIG_PATH", "/config/Caddyfile")
	settingsPath := env("CADDYMGM_SETTINGS_PATH", "/config/caddymgm-settings.json")
	addr := env("CADDYMGM_LISTEN", ":8080")
	caddyMode := normalizeCaddyMode(env("CADDYMGM_CADDY_MODE", "file"))
	caddyAPIURL := env("CADDYMGM_CADDY_API_URL", defaultCaddyAPIURL(caddyMode))
	accessLogDir := env("CADDYMGM_ACCESS_LOG_DIR", "/logs")
	caddyLogDir := env("CADDY_ACCESS_LOG_DIR", accessLogDir)
	caddyDataDir := env("CADDYMGM_CADDY_DATA_DIR", "/caddy-data")

	app := &App{
		configPath:   configPath,
		settingsPath: settingsPath,
		caddyMode:    caddyMode,
		caddyAPIURL:  strings.TrimRight(caddyAPIURL, "/"),
		accessLogDir: accessLogDir,
		caddyLogDir:  caddyLogDir,
		caddyDataDir: caddyDataDir,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		sessions:     make(map[string]time.Time),
	}
	if err := app.ensureConfig(); err != nil {
		log.Fatalf("prepare config: %v", err)
	}
	if err := app.ensureSettings(); err != nil {
		log.Fatalf("prepare settings: %v", err)
	}

	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(webRoot)))
	mux.HandleFunc("GET /api/sites", app.handleListSites)
	mux.HandleFunc("POST /api/sites", app.handleCreateSite)
	mux.HandleFunc("PUT /api/sites/", app.handleUpdateSite)
	mux.HandleFunc("DELETE /api/sites/", app.handleDeleteSite)
	mux.HandleFunc("GET /api/config", app.handleConfig)
	mux.HandleFunc("GET /api/settings", app.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", app.handleUpdateSettings)
	mux.HandleFunc("GET /api/logs", app.handleLogs)
	mux.HandleFunc("POST /api/auth/login", app.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", app.handleLogout)
	mux.HandleFunc("GET /api/auth/me", app.handleMe)

	log.Printf("caddymgm listening on %s, config=%s, settings=%s, caddy_mode=%s, caddy_api=%s, access_logs=%s", addr, configPath, settingsPath, app.caddyMode, app.caddyAPIURL, app.accessLogDir)
	log.Fatal(http.ListenAndServe(addr, logRequest(app.requireAuth(mux))))
}

func (a *App) handleListSites(w http.ResponseWriter, r *http.Request) {
	sites, err := a.readSites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.populateCertificateMetadata(sites)
	writeJSON(w, http.StatusOK, map[string]any{"sites": sites})
}

func (a *App) handleCreateSite(w http.ResponseWriter, r *http.Request) {
	var payload sitePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	site := payload.site(newID(), true)
	if err := normalizeSite(&site); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, head, tail, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.validateSiteTLSLocked(site); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	site.ID = uniqueSiteID(site.Address, sites, "")
	sites = append(sites, site)
	if err := a.save(head, sites, tail); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.applyCaddyConfigLocked(); err != nil {
		a.addLogLocked(site, "reload failed", err.Error())
		writeError(w, http.StatusBadGateway, fmt.Errorf("caddy config saved but reload failed: %w", err))
		return
	}
	a.addLogLocked(site, "created", "Proxy host created")
	writeJSON(w, http.StatusCreated, site)
}

func (a *App) handleUpdateSite(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing site id"))
		return
	}

	var payload sitePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	updated := payload.site(id, false)
	if err := normalizeSite(&updated); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, head, tail, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.validateSiteTLSLocked(updated); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	for i := range sites {
		if sites[i].ID == id {
			oldID := sites[i].ID
			updated.ID = uniqueSiteID(updated.Address, sites, oldID)
			sites[i] = updated
			if err := a.renameAccessLog(oldID, updated.ID); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			if err := a.save(head, sites, tail); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			if err := a.applyCaddyConfigLocked(); err != nil {
				a.addLogLocked(updated, "reload failed", err.Error())
				writeError(w, http.StatusBadGateway, fmt.Errorf("caddy config saved but reload failed: %w", err))
				return
			}
			a.addLogLocked(updated, "updated", "Proxy host updated")
			writeJSON(w, http.StatusOK, updated)
			return
		}
	}
	writeError(w, http.StatusNotFound, errors.New("site not found"))
}

func (a *App) handleDeleteSite(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing site id"))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, head, tail, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	next := sites[:0]
	found := false
	var removed Site
	for _, site := range sites {
		if site.ID == id {
			found = true
			removed = site
			continue
		}
		next = append(next, site)
	}
	if !found {
		writeError(w, http.StatusNotFound, errors.New("site not found"))
		return
	}
	if err := a.save(head, next, tail); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.applyCaddyConfigLocked(); err != nil {
		a.addLogLocked(removed, "reload failed", err.Error())
		writeError(w, http.StatusBadGateway, fmt.Errorf("caddy config saved but reload failed: %w", err))
		return
	}
	if err := a.deleteSiteArtifacts(removed); err != nil {
		a.addLogLocked(removed, "cleanup warning", err.Error())
	}
	a.addLogLocked(removed, "deleted", "Proxy host deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleConfig(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(a.configPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(content)
}

func (a *App) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	settings := a.publicSettingsLocked()
	writeJSON(w, http.StatusOK, settings)
}

func (a *App) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var next Settings
	if err := json.NewDecoder(r.Body).Decode(&next); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if strings.TrimSpace(next.AppName) == "" {
		next.AppName = "CaddyMGM"
	}
	if strings.TrimSpace(next.Username) == "" {
		writeError(w, http.StatusBadRequest, errors.New("username is required"))
		return
	}
	if next.LogRetention < 25 {
		next.LogRetention = 100
	}
	if next.ACMEIssuers == nil {
		next.ACMEIssuers = a.settings.ACMEIssuers
	}
	issuers, err := normalizeACMEIssuers(next.ACMEIssuers)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	next.ACMEIssuers = issuers

	next.ConfigPath = a.configPath
	next.AuthEnabled = authEnabledFromEnv()
	next.CaddyMode = a.caddyMode
	next.CaddyAPIURL = a.caddyAPIURL
	next.PasswordHash = a.settings.PasswordHash
	if strings.TrimSpace(next.Password) != "" {
		next.PasswordHash = hashPassword(next.Password)
	}
	next.Password = ""
	sites, head, tail, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.settings = next
	if err := a.saveSettingsLocked(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.save(head, sites, tail); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.applyCaddyConfigLocked(); err != nil {
		writeError(w, http.StatusBadGateway, fmt.Errorf("caddy config saved but reload failed: %w", err))
		return
	}
	a.trimLogsLocked()
	writeJSON(w, http.StatusOK, a.publicSettingsLocked())
}

func (a *App) handleLogs(w http.ResponseWriter, r *http.Request) {
	siteID := strings.TrimSpace(r.URL.Query().Get("siteId"))
	if siteID == "" {
		writeJSON(w, http.StatusOK, map[string]any{"logs": []LogEntry{}})
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, _, _, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	entries := a.readAccessLogsLocked(siteID, sites)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Time > entries[j].Time
	})
	writeJSON(w, http.StatusOK, map[string]any{"logs": entries})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.validCredentialsLocked(payload.Username, payload.Password) {
		token := newSessionToken()
		a.sessions[token] = time.Now().Add(12 * time.Hour)
		http.SetCookie(w, sessionCookie(token, 12*time.Hour))
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"username":      a.settings.Username,
		})
		return
	}
	writeError(w, http.StatusUnauthorized, errors.New("invalid username or password"))
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	if cookie, err := r.Cookie("caddymgm_session"); err == nil {
		delete(a.sessions, cookie.Value)
	}
	a.mu.Unlock()

	expired := sessionCookie("", -time.Hour)
	http.SetCookie(w, expired)
	writeJSON(w, http.StatusOK, map[string]bool{"loggedOut": true})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"username":      a.settings.Username,
		"appName":       a.settings.AppName,
	})
}

func (a *App) ensureConfig() error {
	if err := os.MkdirAll(filepath.Dir(a.configPath), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(a.accessLogDir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(a.configPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(a.configPath, []byte(managedStart+"\n"+managedEnd+"\n"), 0o644)
}

func (a *App) ensureSettings() error {
	if err := os.MkdirAll(filepath.Dir(a.settingsPath), 0o755); err != nil {
		return err
	}
	content, err := os.ReadFile(a.settingsPath)
	if err == nil {
		if err := json.Unmarshal(content, &a.settings); err != nil {
			return err
		}
		a.settings.AuthEnabled = authEnabledFromEnv()
		a.settings.ConfigPath = a.configPath
		if a.settings.AppName == "" {
			a.settings.AppName = "CaddyMGM"
		}
		if a.settings.Username == "" {
			a.settings.Username = "admin"
		}
		if a.settings.PasswordHash == "" {
			a.settings.PasswordHash = hashPassword(env("CADDYMGM_ADMIN_PASSWORD", "changeme"))
		}
		if a.settings.LogRetention == 0 {
			a.settings.LogRetention = 100
		}
		a.settings.ACMEIssuers = ensureBuiltInACMEIssuers(a.settings.ACMEIssuers)
		return a.saveSettingsLocked()
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	a.settings = Settings{
		AppName:      "CaddyMGM",
		AuthEnabled:  authEnabledFromEnv(),
		Username:     env("CADDYMGM_ADMIN_USER", "admin"),
		PasswordHash: hashPassword(env("CADDYMGM_ADMIN_PASSWORD", "changeme")),
		ConfigPath:   a.configPath,
		LogRetention: 100,
		ACMEIssuers:  ensureBuiltInACMEIssuers(nil),
	}
	return a.saveSettingsLocked()
}

func (a *App) saveSettingsLocked() error {
	content, err := json.MarshalIndent(a.settings, "", "  ")
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", a.settingsPath, time.Now().UnixNano())
	if err := os.WriteFile(tmp, append(content, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.settingsPath)
}

func (a *App) publicSettingsLocked() Settings {
	settings := a.settings
	settings.Password = ""
	settings.PasswordHash = ""
	settings.ConfigPath = a.configPath
	settings.CaddyMode = a.caddyMode
	settings.CaddyAPIURL = a.caddyAPIURL
	return settings
}

func (a *App) readSites() ([]Site, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, _, _, err := a.load()
	return sites, err
}

func (a *App) load() ([]Site, string, string, error) {
	content, err := os.ReadFile(a.configPath)
	if err != nil {
		return nil, "", "", err
	}
	head, managed, tail := splitManaged(string(content))
	sites, err := parseManaged(managed)
	if err != nil {
		return nil, "", "", err
	}
	sort.SliceStable(sites, func(i, j int) bool {
		return sites[i].Address < sites[j].Address
	})
	return sites, head, tail, nil
}

func (a *App) save(head string, sites []Site, tail string) error {
	normalizeSiteIDs(sites)
	var out bytes.Buffer
	out.WriteString(strings.TrimRight(head, "\n"))
	if out.Len() > 0 {
		out.WriteString("\n\n")
	}
	out.WriteString(renderManaged(sites, a.settings.ACMEIssuers, a.caddyLogDir))
	if strings.TrimSpace(tail) != "" {
		out.WriteString("\n")
		out.WriteString(strings.TrimLeft(tail, "\n"))
	}

	tmp := fmt.Sprintf("%s.tmp.%d", a.configPath, time.Now().UnixNano())
	if err := os.WriteFile(tmp, out.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, a.configPath)
}

func (a *App) applyCaddyConfigLocked() error {
	switch a.caddyMode {
	case "file", "none":
		return nil
	case "native", "docker", "api":
	default:
		return fmt.Errorf("unsupported caddy mode %q", a.caddyMode)
	}
	if a.caddyAPIURL == "" {
		return errors.New("caddy api url is required")
	}

	content, err := os.ReadFile(a.configPath)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, a.caddyAPIURL+"/load", bytes.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/caddyfile")
	req.Header.Set("Cache-Control", "must-revalidate")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("caddy api %s: %s", resp.Status, msg)
	}
	return nil
}

func splitManaged(content string) (head, managed, tail string) {
	start, _, okStart := markerLineRange(content, managedStart)
	endStart, end, okEnd := markerLineRange(content, managedEnd)
	if !okStart || !okEnd || endStart < start {
		return content, "", ""
	}
	return content[:start], content[start:end], content[end:]
}

func markerLineRange(content, marker string) (start, end int, ok bool) {
	pos := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		if strings.TrimSpace(line) == marker {
			return pos, pos + len(line), true
		}
		pos += len(line)
	}
	return 0, 0, false
}

func parseManaged(content string) ([]Site, error) {
	sites := make([]Site, 0)
	scanner := bufio.NewScanner(strings.NewReader(content))
	var block []string
	var id string
	inSite := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# caddymgm:site ") {
			if inSite {
				site, err := parseSite(id, block)
				if err != nil {
					return nil, err
				}
				sites = append(sites, site)
			}
			id = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:site "))
			block = nil
			inSite = true
			continue
		}
		if strings.HasPrefix(line, "# caddymgm:end-site") {
			if inSite {
				site, err := parseSite(id, block)
				if err != nil {
					return nil, err
				}
				sites = append(sites, site)
			}
			inSite = false
			id = ""
			block = nil
			continue
		}
		if inSite {
			block = append(block, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sites, nil
}

func parseSite(id string, lines []string) (Site, error) {
	site := Site{ID: id, Enabled: true, TLSMode: "off"}
	if len(lines) == 0 {
		return site, nil
	}
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) == 0 {
		return site, nil
	}

	first := strings.TrimSpace(strings.TrimPrefix(lines[0], "#"))
	site.Enabled = !strings.HasPrefix(strings.TrimSpace(lines[0]), "#")
	site.Address = strings.TrimSpace(strings.TrimSuffix(first, "{"))
	if strings.HasPrefix(site.Address, "http://") {
		site.Address = strings.TrimPrefix(site.Address, "http://")
		site.TLSMode = "off"
	} else if strings.HasPrefix(site.Address, "https://") {
		site.Address = strings.TrimPrefix(site.Address, "https://")
		site.TLSMode = "auto"
	}

	var extra []string
	inTLS := false
	inLog := false
	logDepth := 0
	for _, raw := range lines[1:] {
		line := strings.TrimSpace(strings.TrimPrefix(raw, "#"))
		line = strings.TrimSpace(line)
		if inLog {
			logDepth += braceDelta(line)
			if logDepth <= 0 {
				inLog = false
			}
			continue
		}
		if inTLS {
			switch {
			case strings.HasPrefix(line, "dir "):
				site.TLSMode = "acme"
			case strings.HasPrefix(line, "# caddymgm:tls-issuer "):
				site.ACMEIssuerID = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:tls-issuer "))
			case line == "}":
				inTLS = false
			}
			continue
		}
		switch {
		case line == "" || line == "}":
			continue
		case strings.HasPrefix(line, "# caddymgm:tls-issuer "):
			site.ACMEIssuerID = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:tls-issuer "))
		case line == "tls internal":
			site.TLSMode = "internal"
		case line == "tls {" || strings.HasPrefix(line, "tls {"):
			site.TLSMode = "acme"
			inTLS = true
		case strings.HasPrefix(line, "reverse_proxy "):
			site.Mode = "proxy"
			site.Upstream = strings.TrimSpace(strings.TrimPrefix(line, "reverse_proxy "))
		case strings.HasPrefix(line, "root * "):
			site.Mode = "static"
			site.Root = strings.TrimSpace(strings.TrimPrefix(line, "root * "))
		case line == "file_server":
			if site.Mode == "" {
				site.Mode = "static"
			}
		case line == "log":
			site.LogsEnabled = true
		case line == "log {" || strings.HasPrefix(line, "log {"):
			site.LogsEnabled = true
			inLog = true
			logDepth = braceDelta(line)
		default:
			extra = append(extra, line)
		}
	}
	site.ExtraDirectives = strings.Join(extra, "\n")
	if site.Mode == "" {
		site.Mode = "proxy"
	}
	return site, nil
}

func renderManaged(sites []Site, issuers []ACMEIssuer, logDir string) string {
	var out strings.Builder
	out.WriteString(managedStart + "\n")
	for _, site := range sites {
		out.WriteString("# caddymgm:site " + site.ID + "\n")
		out.WriteString(renderSite(site, issuers, logDir))
		out.WriteString("# caddymgm:end-site\n")
	}
	out.WriteString(managedEnd + "\n")
	return out.String()
}

func renderSite(site Site, issuers []ACMEIssuer, logDir string) string {
	var out strings.Builder
	prefix := ""
	if !site.Enabled {
		prefix = "# "
	}
	address := site.Address
	if site.TLSMode == "" || site.TLSMode == "off" {
		address = "http://" + strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://")
	}
	out.WriteString(prefix + address + " {\n")
	if site.Mode == "static" {
		out.WriteString(prefix + "\troot * " + site.Root + "\n")
		out.WriteString(prefix + "\tfile_server\n")
	} else {
		out.WriteString(prefix + "\treverse_proxy " + site.Upstream + "\n")
	}
	switch site.TLSMode {
	case "internal":
		out.WriteString(prefix + "\ttls internal\n")
	case "acme":
		if issuer, ok := findACMEIssuer(issuers, site.ACMEIssuerID); ok {
			out.WriteString(prefix + "\t# caddymgm:tls-issuer " + issuer.ID + "\n")
			out.WriteString(prefix + "\ttls {\n")
			out.WriteString(prefix + "\t\tissuer acme {\n")
			out.WriteString(prefix + "\t\t\tdir " + issuer.DirectoryURL + "\n")
			if issuer.Email != "" {
				out.WriteString(prefix + "\t\t\temail " + issuer.Email + "\n")
			}
			out.WriteString(prefix + "\t\t}\n")
			out.WriteString(prefix + "\t}\n")
		}
	}
	if site.LogsEnabled {
		out.WriteString(prefix + "\tlog {\n")
		out.WriteString(prefix + "\t\toutput file " + accessLogPath(logDir, site.ID) + " {\n")
		out.WriteString(prefix + "\t\t\tmode 0644\n")
		out.WriteString(prefix + "\t\t}\n")
		out.WriteString(prefix + "\t\tformat json\n")
		out.WriteString(prefix + "\t}\n")
	}
	for _, line := range strings.Split(site.ExtraDirectives, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out.WriteString(prefix + "\t" + line + "\n")
	}
	out.WriteString(prefix + "}\n")
	return out.String()
}

var addressPattern = regexp.MustCompile(`^[A-Za-z0-9*._:-]+$`)

func normalizeSite(site *Site) error {
	site.Address = strings.TrimSpace(site.Address)
	site.Mode = strings.TrimSpace(site.Mode)
	site.Upstream = strings.TrimSpace(site.Upstream)
	site.Root = strings.TrimSpace(site.Root)
	site.ExtraDirectives = cleanExtraDirectives(site.ExtraDirectives)
	site.TLSMode = strings.TrimSpace(site.TLSMode)
	site.ACMEIssuerID = strings.TrimSpace(site.ACMEIssuerID)
	if site.Address == "" {
		return errors.New("address is required")
	}
	site.Address = strings.TrimPrefix(strings.TrimPrefix(site.Address, "http://"), "https://")
	if !addressPattern.MatchString(site.Address) {
		return errors.New("address contains unsupported characters")
	}
	if site.Mode == "" {
		site.Mode = "proxy"
	}
	switch site.Mode {
	case "proxy":
		if site.Upstream == "" {
			return errors.New("upstream is required")
		}
		upstream, err := normalizeProxyUpstream(site.Upstream)
		if err != nil {
			return err
		}
		site.Upstream = upstream
	case "static":
		if site.Root == "" {
			return errors.New("root path is required")
		}
	default:
		return errors.New("mode must be proxy or static")
	}
	if site.ID == "" {
		site.ID = newID()
	}
	if site.TLSMode == "" {
		site.TLSMode = "off"
	}
	switch site.TLSMode {
	case "off", "internal", "acme":
	default:
		return errors.New("tls mode must be off, internal or acme")
	}
	if site.TLSMode != "acme" {
		site.ACMEIssuerID = ""
	}
	return nil
}

func (a *App) validateSiteTLSLocked(site Site) error {
	if site.TLSMode != "acme" {
		return nil
	}
	if site.ACMEIssuerID == "" {
		return errors.New("acme certificate authority is required")
	}
	if _, ok := findACMEIssuer(a.settings.ACMEIssuers, site.ACMEIssuerID); !ok {
		return errors.New("selected acme certificate authority was not found")
	}
	return nil
}

func normalizeACMEIssuers(issuers []ACMEIssuer) ([]ACMEIssuer, error) {
	issuers = ensureBuiltInACMEIssuers(issuers)
	for i := range issuers {
		issuers[i].ID = strings.TrimSpace(issuers[i].ID)
		issuers[i].Name = strings.TrimSpace(issuers[i].Name)
		issuers[i].DirectoryURL = strings.TrimSpace(issuers[i].DirectoryURL)
		issuers[i].Email = strings.TrimSpace(issuers[i].Email)
		if issuers[i].ID == "letsencrypt" {
			issuers[i].Name = "Let's Encrypt"
			issuers[i].DirectoryURL = "https://acme-v02.api.letsencrypt.org/directory"
			issuers[i].BuiltIn = true
		}
		if issuers[i].ID == "" {
			issuers[i].ID = newID()
		}
		if issuers[i].Name == "" {
			return nil, errors.New("certificate authority name is required")
		}
		if !strings.HasPrefix(issuers[i].DirectoryURL, "https://") {
			return nil, errors.New("certificate authority directory URL must start with https://")
		}
	}
	return issuers, nil
}

func ensureBuiltInACMEIssuers(issuers []ACMEIssuer) []ACMEIssuer {
	for i := range issuers {
		if issuers[i].ID == "letsencrypt" {
			issuers[i].Name = "Let's Encrypt"
			issuers[i].DirectoryURL = "https://acme-v02.api.letsencrypt.org/directory"
			issuers[i].BuiltIn = true
			return issuers
		}
	}
	return append([]ACMEIssuer{{
		ID:           "letsencrypt",
		Name:         "Let's Encrypt",
		DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
		BuiltIn:      true,
	}}, issuers...)
}

func findACMEIssuer(issuers []ACMEIssuer, id string) (ACMEIssuer, bool) {
	for _, issuer := range issuers {
		if issuer.ID == id {
			return issuer, true
		}
	}
	return ACMEIssuer{}, false
}

func (a *App) readAccessLogsLocked(siteID string, sites []Site) []LogEntry {
	limit := a.settings.LogRetention
	if limit <= 0 {
		limit = 100
	}
	entries := make([]LogEntry, 0)
	for _, site := range sites {
		if !site.LogsEnabled {
			continue
		}
		if siteID != "" && site.ID != siteID {
			continue
		}
		lines, err := readLastLines(accessLogPath(a.accessLogDir, site.ID), limit)
		if err != nil {
			continue
		}
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			entries = append(entries, accessLogEntry(site, line))
		}
	}
	return entries
}

func readLastLines(path string, limit int) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(content), "\n"), "\n")
	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines, nil
}

func accessLogEntry(site Site, line string) LogEntry {
	entry := LogEntry{
		Time:    time.Now().Format(time.RFC3339),
		SiteID:  site.ID,
		Site:    site.Address,
		Action:  "access",
		Message: line,
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return entry
	}
	if ts, ok := payload["ts"].(float64); ok {
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1_000_000_000)
		entry.Time = time.Unix(sec, nsec).Format(time.RFC3339)
	}
	req, _ := payload["request"].(map[string]any)
	method, _ := req["method"].(string)
	uri, _ := req["uri"].(string)
	status := ""
	switch value := payload["status"].(type) {
	case float64:
		status = fmt.Sprintf("%d", int(value))
	case string:
		status = value
	}
	parts := make([]string, 0, 3)
	if method != "" || uri != "" {
		parts = append(parts, strings.TrimSpace(method+" "+uri))
	}
	if status != "" {
		parts = append(parts, "status "+status)
	}
	entry.Method = method
	entry.Path = uri
	entry.Status = status
	if len(parts) > 0 {
		entry.Message = strings.Join(parts, " - ")
	}
	return entry
}

func accessLogPath(dir, siteID string) string {
	return filepath.Join(dir, siteID+".access.log")
}

func (a *App) populateCertificateMetadata(sites []Site) {
	for i := range sites {
		expiresAt, err := a.certificateExpiresAt(sites[i].Address)
		if err == nil && !expiresAt.IsZero() {
			sites[i].CertificateExpiresAt = expiresAt.Format(time.RFC3339)
		}
	}
}

func (a *App) certificateExpiresAt(domain string) (time.Time, error) {
	domain = strings.ToLower(siteIDFromAddress(domain))
	if domain == "" || a.caddyDataDir == "" {
		return time.Time{}, os.ErrNotExist
	}
	var newest time.Time
	err := filepath.WalkDir(filepath.Join(a.caddyDataDir, "caddy", "certificates"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".crt") {
			return nil
		}
		if !strings.EqualFold(strings.TrimSuffix(entry.Name(), ".crt"), domain) {
			return nil
		}
		expiresAt, err := certificateFileExpiresAt(path)
		if err != nil {
			return nil
		}
		if expiresAt.After(newest) {
			newest = expiresAt
		}
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	if newest.IsZero() {
		return time.Time{}, os.ErrNotExist
	}
	return newest, nil
}

func certificateFileExpiresAt(path string) (time.Time, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, err
	}
	for {
		block, rest := pem.Decode(content)
		if block == nil {
			break
		}
		content = rest
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		return cert.NotAfter, nil
	}
	return time.Time{}, errors.New("certificate not found")
}

func (a *App) renameAccessLog(oldID, newID string) error {
	if oldID == "" || newID == "" || oldID == newID {
		return nil
	}
	oldPath := accessLogPath(a.accessLogDir, oldID)
	newPath := accessLogPath(a.accessLogDir, newID)
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(newPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(oldPath, newPath)
}

func (a *App) deleteSiteArtifacts(site Site) error {
	var failures []string
	if err := removeIfExists(accessLogPath(a.accessLogDir, site.ID)); err != nil {
		failures = append(failures, err.Error())
	}
	certificateFiles := a.certificateFiles(site.Address)
	certificateDirs := make([]string, 0, len(certificateFiles))
	for _, path := range certificateFiles {
		if err := removeIfExists(path); err != nil {
			failures = append(failures, err.Error())
		}
		certificateDirs = append(certificateDirs, filepath.Dir(path))
	}
	for _, dir := range uniqueStrings(certificateDirs) {
		if err := removeEmptyDir(dir); err != nil {
			failures = append(failures, err.Error())
		}
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func (a *App) certificateFiles(domain string) []string {
	domain = strings.ToLower(siteIDFromAddress(domain))
	if domain == "" || a.caddyDataDir == "" {
		return nil
	}
	files := make([]string, 0)
	_ = filepath.WalkDir(filepath.Join(a.caddyDataDir, "caddy", "certificates"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		name := entry.Name()
		base := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".crt"), ".key"), ".json")
		if strings.EqualFold(base, domain) && (strings.HasSuffix(name, ".crt") || strings.HasSuffix(name, ".key") || strings.HasSuffix(name, ".json")) {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func removeIfExists(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

func removeEmptyDir(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		if errors.Is(err, syscall.ENOTEMPTY) {
			return nil
		}
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func normalizeSiteIDs(sites []Site) {
	used := make(map[string]int)
	for i := range sites {
		base := siteIDFromAddress(sites[i].Address)
		if base == "" {
			base = sites[i].ID
		}
		if base == "" {
			base = newID()
		}
		used[base]++
		if used[base] == 1 {
			sites[i].ID = base
			continue
		}
		sites[i].ID = fmt.Sprintf("%s-%d", base, used[base])
	}
}

func uniqueSiteID(address string, sites []Site, currentID string) string {
	base := siteIDFromAddress(address)
	if base == "" {
		base = currentID
	}
	if base == "" {
		base = newID()
	}
	id := base
	next := 2
	for siteIDExists(id, sites, currentID) {
		id = fmt.Sprintf("%s-%d", base, next)
		next++
	}
	return id
}

func siteIDExists(id string, sites []Site, currentID string) bool {
	for _, site := range sites {
		if site.ID == currentID {
			continue
		}
		if site.ID == id {
			return true
		}
	}
	return false
}

func siteIDFromAddress(address string) string {
	address = strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(address), "http://"), "https://")
	address = strings.Trim(address, "[]")
	if host, _, ok := strings.Cut(address, ":"); ok {
		address = host
	}
	address = strings.Trim(address, "*.")
	address = strings.ToLower(address)
	var out strings.Builder
	lastDash := false
	for _, r := range address {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-'
		if valid {
			out.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			out.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(out.String(), ".-")
}

func cleanExtraDirectives(input string) string {
	lines := make([]string, 0)
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || isManagedLogDirective(line) {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func braceDelta(line string) int {
	return strings.Count(line, "{") - strings.Count(line, "}")
}

func isManagedLogDirective(line string) bool {
	if line == "format json" || strings.HasPrefix(line, "output file ") {
		return true
	}
	if strings.HasPrefix(line, "mode ") || strings.HasPrefix(line, "roll_") {
		return true
	}
	return false
}

func normalizeProxyUpstream(upstream string) (string, error) {
	parsed, err := url.Parse(upstream)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return upstream, nil
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("upstream URL must not include query or fragment")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", errors.New("upstream URL must only include scheme, host and port")
	}
	parsed.Path = ""
	return parsed.String(), nil
}

func normalizeCaddyMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "native", "docker", "api", "none":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return "file"
	}
}

func defaultCaddyAPIURL(mode string) string {
	switch mode {
	case "docker":
		return "http://caddy:2019"
	case "native", "api":
		return "http://host.docker.internal:2019"
	default:
		return ""
	}
}

func newID() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		authEnabled := a.settings.AuthEnabled
		authenticated := a.hasValidSessionLocked(r)
		a.mu.Unlock()

		if !authEnabled || authenticated || isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusUnauthorized, errors.New("authentication required"))
			return
		}

		http.Redirect(w, r, "/login.html", http.StatusFound)
	})
}

func (a *App) validCredentialsLocked(username, password string) bool {
	if subtle.ConstantTimeCompare([]byte(username), []byte(a.settings.Username)) != 1 {
		return false
	}
	passHash := hashPassword(password)
	return subtle.ConstantTimeCompare([]byte(passHash), []byte(a.settings.PasswordHash)) == 1
}

func (a *App) hasValidSessionLocked(r *http.Request) bool {
	cookie, err := r.Cookie("caddymgm_session")
	if err != nil || cookie.Value == "" {
		return false
	}
	expires, ok := a.sessions[cookie.Value]
	if !ok {
		return false
	}
	if time.Now().After(expires) {
		delete(a.sessions, cookie.Value)
		return false
	}
	return true
}

func isPublicPath(path string) bool {
	switch path {
	case "/login.html", "/login.css", "/login.js", "/styles.css":
		return true
	case "/api/auth/login":
		return true
	default:
		return false
	}
}

func newSessionToken() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func sessionCookie(value string, maxAge time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     "caddymgm_session",
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (a *App) addLogLocked(site Site, action, message string) {
	a.logs = append(a.logs, LogEntry{
		Time:    time.Now().Format(time.RFC3339),
		SiteID:  site.ID,
		Site:    site.Address,
		Action:  action,
		Message: message,
	})
	a.trimLogsLocked()
}

func (a *App) trimLogsLocked() {
	limit := a.settings.LogRetention
	if limit <= 0 {
		limit = 100
	}
	if len(a.logs) > limit {
		a.logs = a.logs[len(a.logs)-limit:]
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func authEnabledFromEnv() bool {
	value := strings.ToLower(env("CADDYMGM_AUTH_ENABLED", "true"))
	return value != "0" && value != "false" && value != "no" && value != "off"
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
