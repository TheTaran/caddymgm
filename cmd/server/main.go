package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	managedStart = "# caddymgm:start"
	managedEnd   = "# caddymgm:end"
)

//go:embed web/*
var webFS embed.FS

type Site struct {
	ID              string `json:"id"`
	Address         string `json:"address"`
	Mode            string `json:"mode"`
	Upstream        string `json:"upstream,omitempty"`
	Root            string `json:"root,omitempty"`
	ExtraDirectives string `json:"extraDirectives,omitempty"`
	LogsEnabled     bool   `json:"logsEnabled"`
	Enabled         bool   `json:"enabled"`
}

type sitePayload struct {
	Address         string `json:"address"`
	Mode            string `json:"mode"`
	Upstream        string `json:"upstream,omitempty"`
	Root            string `json:"root,omitempty"`
	ExtraDirectives string `json:"extraDirectives,omitempty"`
	LogsEnabled     *bool  `json:"logsEnabled,omitempty"`
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
		Enabled:         p.Enabled,
	}
}

type Settings struct {
	AppName      string `json:"appName"`
	AuthEnabled  bool   `json:"authEnabled"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	PasswordHash string `json:"passwordHash,omitempty"`
	ConfigPath   string `json:"configPath"`
	LogRetention int    `json:"logRetention"`
}

type LogEntry struct {
	Time    string `json:"time"`
	SiteID  string `json:"siteId,omitempty"`
	Site    string `json:"site,omitempty"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

type App struct {
	mu           sync.Mutex
	configPath   string
	settingsPath string
	settings     Settings
	logs         []LogEntry
	sessions     map[string]time.Time
}

func main() {
	configPath := env("CADDY_CONFIG_PATH", "/config/Caddyfile")
	settingsPath := env("CADDYMGM_SETTINGS_PATH", "/config/caddymgm-settings.json")
	addr := env("CADDYMGM_LISTEN", ":8080")

	app := &App{configPath: configPath, settingsPath: settingsPath, sessions: make(map[string]time.Time)}
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

	log.Printf("caddymgm listening on %s, config=%s, settings=%s", addr, configPath, settingsPath)
	log.Fatal(http.ListenAndServe(addr, logRequest(app.requireAuth(mux))))
}

func (a *App) handleListSites(w http.ResponseWriter, r *http.Request) {
	sites, err := a.readSites()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
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
	sites = append(sites, site)
	if err := a.save(head, sites, tail); err != nil {
		writeError(w, http.StatusInternalServerError, err)
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
	for i := range sites {
		if sites[i].ID == id {
			sites[i] = updated
			if err := a.save(head, sites, tail); err != nil {
				writeError(w, http.StatusInternalServerError, err)
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

	next.ConfigPath = a.configPath
	next.AuthEnabled = authEnabledFromEnv()
	next.PasswordHash = a.settings.PasswordHash
	if strings.TrimSpace(next.Password) != "" {
		next.PasswordHash = hashPassword(next.Password)
	}
	next.Password = ""
	a.settings = next
	if err := a.saveSettingsLocked(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.trimLogsLocked()
	writeJSON(w, http.StatusOK, a.publicSettingsLocked())
}

func (a *App) handleLogs(w http.ResponseWriter, r *http.Request) {
	siteID := strings.TrimSpace(r.URL.Query().Get("siteId"))

	a.mu.Lock()
	defer a.mu.Unlock()

	entries := make([]LogEntry, 0, len(a.logs))
	for i := len(a.logs) - 1; i >= 0; i-- {
		entry := a.logs[i]
		if siteID != "" && entry.SiteID != siteID {
			continue
		}
		entries = append(entries, entry)
	}
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
	var out bytes.Buffer
	out.WriteString(strings.TrimRight(head, "\n"))
	if out.Len() > 0 {
		out.WriteString("\n\n")
	}
	out.WriteString(renderManaged(sites))
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
	site := Site{ID: id, Enabled: true}
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

	var extra []string
	for _, raw := range lines[1:] {
		line := strings.TrimSpace(strings.TrimPrefix(raw, "#"))
		line = strings.TrimSpace(line)
		switch {
		case line == "" || line == "}":
			continue
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

func renderManaged(sites []Site) string {
	var out strings.Builder
	out.WriteString(managedStart + "\n")
	for _, site := range sites {
		out.WriteString("# caddymgm:site " + site.ID + "\n")
		out.WriteString(renderSite(site))
		out.WriteString("# caddymgm:end-site\n")
	}
	out.WriteString(managedEnd + "\n")
	return out.String()
}

func renderSite(site Site) string {
	var out strings.Builder
	prefix := ""
	if !site.Enabled {
		prefix = "# "
	}
	out.WriteString(prefix + site.Address + " {\n")
	if site.Mode == "static" {
		out.WriteString(prefix + "\troot * " + site.Root + "\n")
		out.WriteString(prefix + "\tfile_server\n")
	} else {
		out.WriteString(prefix + "\treverse_proxy " + site.Upstream + "\n")
	}
	if site.LogsEnabled {
		out.WriteString(prefix + "\tlog\n")
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
	site.ExtraDirectives = strings.TrimSpace(site.ExtraDirectives)
	if site.Address == "" {
		return errors.New("address is required")
	}
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
	return nil
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
