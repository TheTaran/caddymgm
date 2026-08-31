package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
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
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

const (
	managedStart         = "# caddymgm:start"
	managedEnd           = "# caddymgm:end"
	sessionCookieName    = "caddymgm_session"
	oidcStateCookieName  = "caddymgm_oidc_state"
	csrfCookieName       = "caddymgm_csrf"
	loginMaxFailures     = 5
	loginFailureWindow   = 15 * time.Minute
	loginLockout         = 15 * time.Minute
	loginAttemptLimit    = 10_000
	sessionLimit         = 10_000
	sessionLifetime      = 12 * time.Hour
	oidcStateLimit       = 1_000
	oidcStateLifetime    = 10 * time.Minute
	loginJSONBodyLimit   = 64 << 10
	adminJSONBodyLimit   = 1 << 20
	rootCAUploadLimit    = 4 << 20
	accessSessionLimit   = 10_000
	accessStateLimit     = 1_000
	accessTicketLimit    = 1_000
	accessTicketLifetime = time.Minute
	geoMapLogLimit       = 5_000
)

var version = "dev"

//go:embed web/*
var webFS embed.FS

type Site struct {
	ID                    string   `json:"id"`
	Address               string   `json:"address"`
	Comment               string   `json:"comment,omitempty"`
	Mode                  string   `json:"mode"`
	Upstream              string   `json:"upstream,omitempty"`
	SkipTLSVerify         bool     `json:"skipTlsVerify,omitempty"`
	RewriteRedirects      bool     `json:"rewriteRedirects"`
	RedirectOrigins       []string `json:"redirectOrigins,omitempty"`
	HSTSEnabled           bool     `json:"hstsEnabled,omitempty"`
	SecurityHeaderProfile string   `json:"securityHeaderProfile,omitempty"`
	CompressionProfile    string   `json:"compressionProfile,omitempty"`
	Root                  string   `json:"root,omitempty"`
	ExtraDirectives       string   `json:"extraDirectives,omitempty"`
	LogsEnabled           bool     `json:"logsEnabled"`
	TLSMode               string   `json:"tlsMode"`
	ACMEIssuerID          string   `json:"acmeIssuerId,omitempty"`
	TLSMinVersion         string   `json:"tlsMinVersion,omitempty"`
	TLSMaxVersion         string   `json:"tlsMaxVersion,omitempty"`
	TLSCipherSuites       []string `json:"tlsCipherSuites,omitempty"`
	CertificateExpiresAt  string   `json:"certificateExpiresAt,omitempty"`
	Enabled               bool     `json:"enabled"`
	AuthEnabled           bool     `json:"authEnabled,omitempty"`
	AuthProviderID        string   `json:"authProviderId,omitempty"`
	BasicAuthEnabled      bool     `json:"basicAuthEnabled,omitempty"`
	BasicAuthUsername     string   `json:"basicAuthUsername,omitempty"`
	BasicAuthPasswordHash string   `json:"-"`
	BasicAuthPassword     string   `json:"-"`
}

type sitePayload struct {
	Address               string   `json:"address"`
	Comment               string   `json:"comment,omitempty"`
	Mode                  string   `json:"mode"`
	Upstream              string   `json:"upstream,omitempty"`
	SkipTLSVerify         bool     `json:"skipTlsVerify,omitempty"`
	RewriteRedirects      *bool    `json:"rewriteRedirects,omitempty"`
	RedirectOrigins       []string `json:"redirectOrigins,omitempty"`
	HSTSEnabled           bool     `json:"hstsEnabled,omitempty"`
	SecurityHeaderProfile string   `json:"securityHeaderProfile,omitempty"`
	CompressionProfile    string   `json:"compressionProfile,omitempty"`
	Root                  string   `json:"root,omitempty"`
	ExtraDirectives       string   `json:"extraDirectives,omitempty"`
	LogsEnabled           *bool    `json:"logsEnabled,omitempty"`
	TLSMode               string   `json:"tlsMode,omitempty"`
	ACMEIssuerID          string   `json:"acmeIssuerId,omitempty"`
	TLSMinVersion         string   `json:"tlsMinVersion,omitempty"`
	TLSMaxVersion         string   `json:"tlsMaxVersion,omitempty"`
	TLSCipherSuites       []string `json:"tlsCipherSuites,omitempty"`
	Enabled               bool     `json:"enabled"`
	AuthEnabled           bool     `json:"authEnabled,omitempty"`
	AuthProviderID        string   `json:"authProviderId,omitempty"`
	BasicAuthEnabled      bool     `json:"basicAuthEnabled,omitempty"`
	BasicAuthUsername     string   `json:"basicAuthUsername,omitempty"`
	BasicAuthPassword     string   `json:"basicAuthPassword,omitempty"`
}

func (p sitePayload) site(id string, defaultLogsEnabled bool) Site {
	logsEnabled := defaultLogsEnabled
	rewriteRedirects := true
	if p.RewriteRedirects != nil {
		rewriteRedirects = *p.RewriteRedirects
	}
	if p.LogsEnabled != nil {
		logsEnabled = *p.LogsEnabled
	}
	return Site{
		ID:                    id,
		Address:               p.Address,
		Comment:               p.Comment,
		Mode:                  p.Mode,
		Upstream:              p.Upstream,
		SkipTLSVerify:         p.SkipTLSVerify,
		RewriteRedirects:      rewriteRedirects,
		RedirectOrigins:       p.RedirectOrigins,
		HSTSEnabled:           p.HSTSEnabled,
		SecurityHeaderProfile: p.SecurityHeaderProfile,
		CompressionProfile:    p.CompressionProfile,
		Root:                  p.Root,
		ExtraDirectives:       p.ExtraDirectives,
		LogsEnabled:           logsEnabled,
		TLSMode:               p.TLSMode,
		ACMEIssuerID:          p.ACMEIssuerID,
		TLSMinVersion:         p.TLSMinVersion,
		TLSMaxVersion:         p.TLSMaxVersion,
		TLSCipherSuites:       p.TLSCipherSuites,
		Enabled:               p.Enabled,
		AuthEnabled:           p.AuthEnabled,
		AuthProviderID:        p.AuthProviderID,
		BasicAuthEnabled:      p.BasicAuthEnabled,
		BasicAuthUsername:     p.BasicAuthUsername,
		BasicAuthPassword:     p.BasicAuthPassword,
	}
}

type ACMEIssuer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DirectoryURL string `json:"directoryUrl"`
	Email        string `json:"email,omitempty"`
	RootCAFile   string `json:"rootCaFile,omitempty"`
	BuiltIn      bool   `json:"builtIn,omitempty"`
}

type Settings struct {
	AppName          string       `json:"appName"`
	AuthEnabled      bool         `json:"authEnabled"`
	LocalAuthEnabled bool         `json:"localAuthEnabled"`
	OIDCAuthEnabled  bool         `json:"oidcAuthEnabled"`
	Username         string       `json:"username"`
	Password         string       `json:"password,omitempty"`
	PasswordHash     string       `json:"passwordHash,omitempty"`
	OIDC             OIDCSettings `json:"oidc"`
	ConfigPath       string       `json:"configPath"`
	LogRetention     int          `json:"logRetention"`
	ACMEIssuers      []ACMEIssuer `json:"acmeIssuers,omitempty"`
	CaddyMode        string       `json:"caddyMode"`
	CaddyAPIURL      string       `json:"caddyApiUrl"`
	WebInterface     WebInterface `json:"webInterface"`
}

type OIDCSettings struct {
	Enabled      bool   `json:"enabled"`
	IssuerURL    string `json:"issuerUrl,omitempty"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	RedirectURL  string `json:"redirectUrl,omitempty"`
	Scopes       string `json:"scopes,omitempty"`
}

type WebInterface struct {
	Host         string `json:"host,omitempty"`
	Upstream     string `json:"upstream,omitempty"`
	TLSEnabled   bool   `json:"tlsEnabled"`
	ACMEIssuerID string `json:"acmeIssuerId,omitempty"`
}

type LogEntry struct {
	Time     string `json:"time"`
	SiteID   string `json:"siteId,omitempty"`
	Site     string `json:"site,omitempty"`
	Action   string `json:"action"`
	Message  string `json:"message"`
	Method   string `json:"method,omitempty"`
	Path     string `json:"path,omitempty"`
	Status   string `json:"status,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	IP       string `json:"ip,omitempty"`
}

type ComponentVersion struct {
	Current         string `json:"current"`
	Latest          string `json:"latest,omitempty"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseURL      string `json:"releaseUrl,omitempty"`
}

type App struct {
	mu                sync.Mutex
	configPath        string
	settingsPath      string
	authProvidersPath string
	caddyMode         string
	caddyAPIURL       string
	accessLogDir      string
	geoIPDBPath       string
	caddyLogDir       string
	serviceLog        string
	oidcAuditLog      string
	auditMu           sync.Mutex
	caddyDataDir      string
	caddyVersionFile  string
	caCertDir         string
	staticRootBase    string
	webListen         string
	webPort           string
	httpClient        *http.Client
	settings          Settings
	logs              []LogEntry
	sessions          map[string]Session
	oidcStates        map[string]time.Time
	oidcCache         map[string]*oidcRuntime
	authProviders     AuthProvidersConfig
	accessSessions    map[string]AccessSession
	accessStates      map[string]AccessState
	accessTickets     map[string]AccessTicket
	loginLimiter      *loginLimiter
	trustedProxies    *trustedProxySet
	oidcProvider      func(context.Context, string) (*oidc.Provider, error)
	versionMu         sync.Mutex
	latestVersions    map[string]ComponentVersion
	versionsChecked   time.Time
}

type Session struct {
	ExpiresAt time.Time
	Username  string
	Provider  string
}

type oidcRuntime struct {
	Provider *oidc.Provider
	Verifier *oidc.IDTokenVerifier
	Config   oauth2.Config
}

type loginAttempt struct {
	failures      int
	windowStarted time.Time
	lockedUntil   time.Time
	lastSeen      time.Time
}

type loginLimiter struct {
	mu              sync.Mutex
	attempts        map[string]loginAttempt
	maxFailures     int
	failureWindow   time.Duration
	lockoutDuration time.Duration
	maxEntries      int
	nextCleanup     time.Time
	now             func() time.Time
}

type trustedProxyCacheEntry struct {
	addresses []netip.Addr
	expiresAt time.Time
}

type trustedProxySet struct {
	mu        sync.Mutex
	prefixes  []netip.Prefix
	hostnames []string
	cache     map[string]trustedProxyCacheEntry
	lookup    func(context.Context, string, string) ([]netip.Addr, error)
	now       func() time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{
		attempts:        make(map[string]loginAttempt),
		maxFailures:     loginMaxFailures,
		failureWindow:   loginFailureWindow,
		lockoutDuration: loginLockout,
		maxEntries:      loginAttemptLimit,
		now:             time.Now,
	}
}

func main() {
	configPath := env("CADDY_CONFIG_PATH", "/config/Caddyfile")
	settingsPath := env("CADDYMGM_SETTINGS_PATH", "/caddymgm/caddymgm-settings.json")
	authProvidersPath := filepath.Join(filepath.Dir(settingsPath), "auth-providers.json")
	caddyMode := normalizeCaddyMode(env("CADDYMGM_CADDY_MODE", "file"))
	caddyAPIURL := env("CADDYMGM_CADDY_API_URL", defaultCaddyAPIURL(caddyMode))
	accessLogDir := env("CADDYMGM_ACCESS_LOG_DIR", "/logs")
	geoIPDBPath := env("CADDYMGM_GEOIP_DB_PATH", "/caddymgm/geoip/GeoLite2-City.mmdb")
	caddyLogDir := env("CADDY_ACCESS_LOG_DIR", accessLogDir)
	serviceLog := env("CADDYMGM_CADDY_SERVICE_LOG", filepath.Join(accessLogDir, "caddy-service.log"))
	oidcAuditLog := filepath.Join(filepath.Dir(settingsPath), "oidc-audit.log")
	caddyDataDir := env("CADDYMGM_CADDY_DATA_DIR", "/caddy-data")
	caddyVersionFile := env("CADDYMGM_CADDY_VERSION_FILE", filepath.Join(accessLogDir, "caddy-version"))
	caCertDir := env("CADDYMGM_CA_CERT_DIR", "/ca-certificates")
	staticRootBase := env("CADDYMGM_STATIC_ROOT_BASE", "/srv")
	webListen := env("CADDYMGM_WEB_LISTEN", ":8080")
	webPort := strings.TrimSpace(env("CADDYMGM_WEB_PORT", "8080"))
	if _, err := net.LookupPort("tcp", webPort); err != nil {
		log.Fatalf("invalid CADDYMGM_WEB_PORT %q: %v", webPort, err)
	}
	trustedProxies, err := newTrustedProxySet(env("CADDYMGM_TRUSTED_PROXIES", ""))
	if err != nil {
		log.Fatalf("invalid CADDYMGM_TRUSTED_PROXIES: %v", err)
	}

	app := &App{
		configPath:        configPath,
		settingsPath:      settingsPath,
		authProvidersPath: authProvidersPath,
		caddyMode:         caddyMode,
		caddyAPIURL:       strings.TrimRight(caddyAPIURL, "/"),
		accessLogDir:      accessLogDir,
		geoIPDBPath:       geoIPDBPath,
		caddyLogDir:       caddyLogDir,
		serviceLog:        serviceLog,
		oidcAuditLog:      oidcAuditLog,
		caddyDataDir:      caddyDataDir,
		caddyVersionFile:  caddyVersionFile,
		caCertDir:         caCertDir,
		staticRootBase:    staticRootBase,
		webListen:         webListen,
		webPort:           webPort,
		httpClient:        &http.Client{Timeout: 15 * time.Second},
		sessions:          make(map[string]Session),
		oidcStates:        make(map[string]time.Time),
		oidcCache:         make(map[string]*oidcRuntime),
		accessSessions:    make(map[string]AccessSession),
		accessStates:      make(map[string]AccessState),
		accessTickets:     make(map[string]AccessTicket),
		loginLimiter:      newLoginLimiter(),
		trustedProxies:    trustedProxies,
		oidcProvider:      oidc.NewProvider,
	}
	if err := app.ensureConfig(); err != nil {
		log.Fatalf("prepare config: %v", err)
	}
	if err := app.ensureSettings(); err != nil {
		log.Fatalf("prepare settings: %v", err)
	}
	if err := app.ensureAuthProviders(); err != nil {
		log.Fatalf("prepare auth providers: %v", err)
	}
	app.startGeoIPUpdater()
	if err := app.syncManagedConfig(); err != nil {
		log.Printf("sync managed caddy config: %v", err)
	}

	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", noCacheFileServer(http.FileServer(http.FS(webRoot))))
	mux.HandleFunc("GET /api/sites", app.handleListSites)
	mux.HandleFunc("POST /api/sites", app.handleCreateSite)
	mux.HandleFunc("PUT /api/sites/", app.handleUpdateSite)
	mux.HandleFunc("DELETE /api/sites/", app.handleDeleteSite)
	mux.HandleFunc("GET /api/config", app.handleConfig)
	mux.HandleFunc("GET /api/versions", app.handleVersions)
	mux.HandleFunc("GET /api/settings", app.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", app.handleUpdateSettings)
	mux.HandleFunc("GET /api/auth-providers", app.handleGetAuthProviders)
	mux.HandleFunc("PUT /api/auth-providers", app.handleUpdateAuthProviders)
	mux.HandleFunc("GET /api/logs", app.handleLogs)
	mux.HandleFunc("GET /api/geo-map", app.handleGeoMap)
	mux.HandleFunc("POST /api/certificates/root-ca", app.handleUploadRootCA)
	mux.HandleFunc("POST /api/certificates/renew/", app.handleRenewCertificate)
	mux.HandleFunc("POST /api/auth/login", app.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", app.handleLogout)
	mux.HandleFunc("GET /api/auth/me", app.handleMe)
	mux.HandleFunc("GET /api/auth/config", app.handleAuthConfig)
	mux.HandleFunc("GET /api/auth/oidc/start", app.handleOIDCStart)
	mux.HandleFunc("GET /api/auth/oidc/callback", app.handleOIDCCallback)
	mux.HandleFunc("GET /auth/oidc/callback", app.handleOIDCCallback)
	mux.HandleFunc("GET /.caddymgm/auth/check", app.handleAccessCheck)
	mux.HandleFunc("GET /.caddymgm/auth/portal", app.handleAccessPortal)
	mux.HandleFunc("POST /.caddymgm/auth/logout", app.handleAccessLogout)
	mux.HandleFunc("GET /.caddymgm/auth/start", app.handleAccessStart)
	mux.HandleFunc("GET /.caddymgm/auth/callback", app.handleAccessCallback)
	mux.HandleFunc("GET /.caddymgm/auth/complete", app.handleAccessComplete)

	listenAddr := app.webListen
	handler := logRequest(app.securityHeaders(app.ensureCSRFCookie(app.requireCSRF(app.requireAuth(mux)))))
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("caddymgm listening on %s, config=%s, settings=%s, caddy_mode=%s, caddy_api=%s, access_logs=%s", listenAddr, configPath, settingsPath, app.caddyMode, app.caddyAPIURL, app.accessLogDir)
	log.Fatal(server.ListenAndServe())
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
	if err := decodeJSONBody(w, r, &payload, adminJSONBodyLimit); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	site := payload.site(newID(), true)
	if err := normalizeSite(&site); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := prepareBasicAuth(&site, nil); err != nil {
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
	if err := a.validateSiteLocked(site); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if duplicateSiteAddress(site.Address, sites, "") {
		writeError(w, http.StatusConflict, errors.New("a web host with this address already exists"))
		return
	}
	site.ID = uniqueSiteID(site.Address, sites, "")
	sites = append(sites, site)
	if err := a.saveAndApplyCaddyConfigLocked(head, sites, tail); err != nil {
		a.addLogLocked(site, "reload failed", err.Error())
		writeError(w, http.StatusBadGateway, fmt.Errorf("caddy rejected the change; previous config restored: %w", err))
		return
	}
	a.populateSiteCertificateMetadata(&site)
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
	if err := decodeJSONBody(w, r, &payload, adminJSONBodyLimit); err != nil {
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
	var existing *Site
	for i := range sites {
		if sites[i].ID == id {
			existing = &sites[i]
			break
		}
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, errors.New("site not found"))
		return
	}
	if err := prepareBasicAuth(&updated, existing); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := a.validateSiteLocked(updated); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if duplicateSiteAddress(updated.Address, sites, id) {
		writeError(w, http.StatusConflict, errors.New("a web host with this address already exists"))
		return
	}
	for i := range sites {
		if sites[i].ID == id {
			oldID := sites[i].ID
			updated.ID = uniqueSiteID(updated.Address, sites, oldID)
			sites[i] = updated
			if err := a.saveAndApplyCaddyConfigLocked(head, sites, tail); err != nil {
				a.addLogLocked(updated, "reload failed", err.Error())
				writeError(w, http.StatusBadGateway, fmt.Errorf("caddy rejected the change; previous config restored: %w", err))
				return
			}
			if err := a.renameAccessLog(oldID, updated.ID); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			a.populateSiteCertificateMetadata(&updated)
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
	if err := a.saveAndApplyCaddyConfigLocked(head, next, tail); err != nil {
		a.addLogLocked(removed, "reload failed", err.Error())
		writeError(w, http.StatusBadGateway, fmt.Errorf("caddy rejected the change; previous config restored: %w", err))
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

func (a *App) handleVersions(w http.ResponseWriter, r *http.Request) {
	current := map[string]string{
		"caddymgm": version,
		"caddy":    a.currentCaddyVersion(),
		"go":       runtime.Version(),
	}
	latest := a.getLatestVersions(r.Context())
	result := make(map[string]ComponentVersion, len(current))
	for name, currentVersion := range current {
		info := latest[name]
		info.Current = currentVersion
		info.UpdateAvailable = isVersionNewer(info.Latest, currentVersion)
		result[name] = info
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) currentCaddyVersion() string {
	if file, err := os.Open(a.caddyVersionFile); err == nil {
		defer file.Close()
		if content, readErr := io.ReadAll(io.LimitReader(file, 129)); readErr == nil && len(content) <= 128 {
			if fields := strings.Fields(string(content)); len(fields) > 0 {
				return fields[0]
			}
		}
	}
	return "unknown"
}

func (a *App) getLatestVersions(ctx context.Context) map[string]ComponentVersion {
	a.versionMu.Lock()
	defer a.versionMu.Unlock()
	if len(a.latestVersions) > 0 && time.Since(a.versionsChecked) < 6*time.Hour {
		return cloneVersions(a.latestVersions)
	}

	type result struct {
		name string
		info ComponentVersion
	}
	results := make(chan result, 3)
	go func() {
		tag, _ := a.fetchLatestGitHubRelease(ctx, "caddyserver", "caddy")
		results <- result{
			name: "caddy",
			info: ComponentVersion{
				Latest:     tag,
				ReleaseURL: "https://github.com/caddyserver/caddy/releases/latest",
			},
		}
	}()
	go func() {
		tag, _ := a.fetchLatestGitHubRelease(ctx, "TheTaran", "caddymgm")
		results <- result{
			name: "caddymgm",
			info: ComponentVersion{
				Latest:     tag,
				ReleaseURL: "https://github.com/TheTaran/caddymgm/releases/latest",
			},
		}
	}()
	go func() {
		tag := a.fetchLatestGoRelease(ctx)
		results <- result{
			name: "go",
			info: ComponentVersion{
				Latest:     tag,
				ReleaseURL: "https://go.dev/dl/",
			},
		}
	}()

	latest := make(map[string]ComponentVersion, 3)
	for range 3 {
		item := <-results
		latest[item.name] = item.info
	}
	a.latestVersions = latest
	a.versionsChecked = time.Now()
	return cloneVersions(latest)
}

func (a *App) fetchLatestGoRelease(parent context.Context) string {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://go.dev/dl/?mode=json", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CaddyMGM/"+version)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var releases []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&releases); err != nil {
		return ""
	}
	for _, release := range releases {
		if release.Stable {
			return strings.TrimSpace(release.Version)
		}
	}
	return ""
}

func cloneVersions(input map[string]ComponentVersion) map[string]ComponentVersion {
	result := make(map[string]ComponentVersion, len(input))
	for name, info := range input {
		result[name] = info
	}
	return result
}

func (a *App) fetchLatestGitHubRelease(parent context.Context, owner, repository string) (string, string) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repository)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "CaddyMGM/"+version)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", ""
	}
	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return "", ""
	}
	return strings.TrimSpace(release.TagName), strings.TrimSpace(release.HTMLURL)
}

func isVersionNewer(candidate, current string) bool {
	candidateParts, candidateOK := numericVersion(candidate)
	currentParts, currentOK := numericVersion(current)
	if !candidateOK || !currentOK {
		return false
	}
	for i := range candidateParts {
		if candidateParts[i] != currentParts[i] {
			return candidateParts[i] > currentParts[i]
		}
	}
	return false
}

func numericVersion(value string) ([3]int, bool) {
	var result [3]int
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "go")
	value = strings.TrimPrefix(value, "v")
	value, _, _ = strings.Cut(value, "-")
	value, _, _ = strings.Cut(value, "+")
	parts := strings.Split(value, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return result, false
	}
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return result, false
		}
		result[index] = number
	}
	return result, true
}

func (a *App) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	settings := a.publicSettingsLocked()
	writeJSON(w, http.StatusOK, settings)
}

func (a *App) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var next Settings
	if err := decodeJSONBody(w, r, &next, adminJSONBodyLimit); err != nil {
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
	normalizeWebInterface(&next.WebInterface, a.caddyMode)
	if next.ACMEIssuers == nil {
		next.ACMEIssuers = a.settings.ACMEIssuers
	}
	issuers, err := normalizeACMEIssuers(next.ACMEIssuers)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	next.ACMEIssuers = issuers
	if err := validateWebInterface(next.WebInterface, issuers); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(next.Password) != "" && len(next.Password) > 72 {
		writeError(w, http.StatusBadRequest, errors.New("password must be 72 characters or fewer"))
		return
	}

	next.ConfigPath = a.configPath
	next.AuthEnabled = authEnabledFromEnv()
	next.CaddyMode = a.caddyMode
	next.CaddyAPIURL = a.caddyAPIURL
	next.PasswordHash = a.settings.PasswordHash
	if strings.TrimSpace(next.Password) != "" {
		next.PasswordHash = hashPassword(next.Password)
	}
	if strings.TrimSpace(next.PasswordHash) == "" {
		writeError(w, http.StatusBadRequest, errors.New("administrator password is not configured"))
		return
	}
	next.Password = ""
	next.OIDC = normalizeOIDCSettings(next.OIDC, a.settings.OIDC)
	if err := validateOIDCSettings(next.OIDC); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	a.applySettingsEnvOverridesLocked(&next)
	sites, head, tail, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	previousSettings := a.settings
	previousSettingsFile, err := os.ReadFile(a.settingsPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.settings = next
	if err := a.saveSettingsLocked(); err != nil {
		a.settings = previousSettings
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.saveAndApplyCaddyConfigLocked(head, sites, tail); err != nil {
		a.settings = previousSettings
		if restoreErr := writeFileAtomically(a.settingsPath, previousSettingsFile, 0o600); restoreErr != nil {
			err = fmt.Errorf("%w; restoring previous settings failed: %v", err, restoreErr)
		}
		writeError(w, http.StatusBadGateway, fmt.Errorf("caddy rejected the settings change; previous config restored: %w", err))
		return
	}
	a.trimLogsLocked()
	writeJSON(w, http.StatusOK, a.publicSettingsLocked())
}

func (a *App) handleLogs(w http.ResponseWriter, r *http.Request) {
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	siteID := strings.TrimSpace(r.URL.Query().Get("siteId"))

	a.mu.Lock()
	defer a.mu.Unlock()

	if source == "oidc" {
		entries := a.readOIDCAuditLogsLocked()
		sort.SliceStable(entries, func(i, j int) bool { return entries[i].Time > entries[j].Time })
		writeJSON(w, http.StatusOK, map[string]any{"logs": entries})
		return
	}

	if source == "caddy-service" {
		entries := a.readServiceLogsLocked()
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].Time > entries[j].Time
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"logs":      entries,
			"available": a.serviceLogAvailableLocked(),
		})
		return
	}

	if siteID == "" {
		writeJSON(w, http.StatusOK, map[string]any{"logs": []LogEntry{}})
		return
	}

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

func (a *App) handleUploadRootCA(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, rootCAUploadLimit)
	if err := r.ParseMultipartForm(rootCAUploadLimit); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid upload"))
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()
	file, header, err := r.FormFile("certificate")
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("certificate file is required"))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, rootCAUploadLimit))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if len(content) == 0 {
		writeError(w, http.StatusBadRequest, errors.New("certificate file is empty"))
		return
	}
	pemContent, err := normalizeCertificateBundle(content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	filename := rootCAFilename(header.Filename)
	if err := os.MkdirAll(a.caCertDir, 0o750); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	path := filepath.Join(a.caCertDir, filename)
	if err := os.WriteFile(path, pemContent, 0o640); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{
		"rootCaFile": "/ca-certificates/" + filename,
	})
}

func (a *App) handleRenewCertificate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/certificates/renew/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing site id"))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	sites, _, _, err := a.load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var site Site
	found := false
	for _, candidate := range sites {
		if candidate.ID == id {
			site = candidate
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, errors.New("site not found"))
		return
	}
	if !site.Enabled {
		writeError(w, http.StatusBadRequest, errors.New("site must be enabled before forcing renewal"))
		return
	}
	if site.TLSMode != "acme" {
		writeError(w, http.StatusBadRequest, errors.New("site does not use ACME TLS"))
		return
	}

	files := a.certificateFiles(site.Address)
	for _, path := range files {
		if err := removeIfExists(path); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if err := a.applyCaddyConfigLocked(); err != nil {
		a.addLogLocked(site, "certificate renew reload failed", err.Error())
		writeError(w, http.StatusBadGateway, fmt.Errorf("certificate files removed but caddy reload failed: %w", err))
		return
	}
	if err := a.triggerTLSHandshake(site.Address); err != nil {
		a.addLogLocked(site, "certificate renew handshake warning", err.Error())
	} else {
		a.addLogLocked(site, "certificate renew triggered", "TLS handshake started for certificate renewal")
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"site":    site,
		"message": "certificate renewal triggered",
	})
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(w, r, &payload, loginJSONBodyLimit); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	loginKeys := loginAttemptKeys(payload.Username, r.RemoteAddr)
	if retryAfter, blocked := a.loginLimiter.retryAfterAny(loginKeys); blocked {
		writeLoginRateLimit(w, retryAfter)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !localAuthEnabledFromEnv() {
		writeError(w, http.StatusForbidden, errors.New("local authentication is disabled"))
		return
	}
	if !a.secureSessionTransportAllowed(r) {
		writeError(w, http.StatusForbidden, errors.New("login requires HTTPS unless you are connecting from localhost"))
		return
	}

	if a.validCredentialsLocked(payload.Username, payload.Password) {
		a.loginLimiter.reset(loginKeys...)
		if passwordNeedsUpgrade(a.settings.PasswordHash) {
			a.settings.PasswordHash = hashPassword(payload.Password)
			if err := a.saveSettingsLocked(); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		}
		a.createSessionLocked(w, r, a.settings.Username, "local")
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": true,
			"username":      a.settings.Username,
		})
		return
	}
	if retryAfter, blocked := a.loginLimiter.recordFailure(loginKeys...); blocked {
		writeLoginRateLimit(w, retryAfter)
		return
	}
	writeError(w, http.StatusUnauthorized, errors.New("invalid username or password"))
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if !a.secureSessionTransportAllowed(r) {
		writeError(w, http.StatusForbidden, errors.New("logout requires HTTPS unless you are connecting from localhost"))
		return
	}
	a.mu.Lock()
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		delete(a.sessions, cookie.Value)
	}
	a.mu.Unlock()

	expired := sessionCookie("", -time.Hour, a.isSecureRequest(r))
	http.SetCookie(w, expired)
	writeJSON(w, http.StatusOK, map[string]bool{"loggedOut": true})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	username := a.settings.Username
	provider := "local"
	authenticated := a.hasValidSessionLocked(r)
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if session, ok := a.sessions[cookie.Value]; ok {
			username = session.Username
			provider = session.Provider
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": authenticated || !a.settings.AuthEnabled,
		"username":      username,
		"provider":      provider,
		"appName":       a.settings.AppName,
	})
}

func (a *App) handleAuthConfig(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{
		"authEnabled":         a.settings.AuthEnabled,
		"localAuthEnabled":    localAuthEnabledFromEnv(),
		"oidcAuthEnabled":     oidcAuthEnabledFromEnv() && a.settings.OIDC.Enabled,
		"insecureHTTPAllowed": insecureHTTPAllowedFromEnv(),
		"appName":             a.settings.AppName,
	})
}

func (a *App) handleOIDCStart(w http.ResponseWriter, r *http.Request) {
	runtime, err := a.oidcRuntime(r.Context())
	if err != nil {
		http.Redirect(w, r, "/login.html?error=oidc_not_available", http.StatusFound)
		return
	}

	a.mu.Lock()
	if !a.settings.AuthEnabled {
		a.mu.Unlock()
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if !a.secureSessionTransportAllowed(r) {
		a.mu.Unlock()
		http.Redirect(w, r, "/login.html?error=insecure_transport", http.StatusFound)
		return
	}
	now := time.Now()
	state, ok := a.createOIDCStateLocked(now)
	if !ok {
		a.mu.Unlock()
		w.Header().Set("Retry-After", "60")
		writeError(w, http.StatusTooManyRequests, errors.New("too many pending OIDC login requests; try again later"))
		return
	}
	a.mu.Unlock()

	http.SetCookie(w, oidcStateCookie(state, oidcStateLifetime, a.isSecureRequest(r)))
	a.recordOIDCAudit(r, "login_started", "pending", "", "", "CaddyMGM", "Administration login")
	http.Redirect(w, r, runtime.Config.AuthCodeURL(state), http.StatusFound)
}

func (a *App) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	auditSuccess := false
	defer func() {
		if !auditSuccess {
			a.recordOIDCAudit(r, "login_failed", "failed", "", "", "CaddyMGM", "Administration login")
		}
	}()
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if state == "" || code == "" {
		http.Redirect(w, r, "/login.html?error=missing_oidc_response", http.StatusFound)
		return
	}

	runtime, err := a.oidcRuntime(r.Context())
	if err != nil {
		http.SetCookie(w, oidcStateCookie("", -time.Hour, a.isSecureRequest(r)))
		http.Redirect(w, r, "/login.html?error=oidc_not_available", http.StatusFound)
		return
	}

	a.mu.Lock()
	if !a.settings.AuthEnabled {
		a.mu.Unlock()
		http.SetCookie(w, oidcStateCookie("", -time.Hour, a.isSecureRequest(r)))
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if !a.secureSessionTransportAllowed(r) {
		a.mu.Unlock()
		http.SetCookie(w, oidcStateCookie("", -time.Hour, a.isSecureRequest(r)))
		http.Redirect(w, r, "/login.html?error=insecure_transport", http.StatusFound)
		return
	}
	if err := a.consumeOIDCStateLocked(r, state); err != nil {
		a.mu.Unlock()
		http.SetCookie(w, oidcStateCookie("", -time.Hour, a.isSecureRequest(r)))
		http.Redirect(w, r, "/login.html?error=invalid_oidc_state", http.StatusFound)
		return
	}
	a.mu.Unlock()
	http.SetCookie(w, oidcStateCookie("", -time.Hour, a.isSecureRequest(r)))

	token, err := runtime.Config.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, "/login.html?error=oidc_exchange_failed", http.StatusFound)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(rawIDToken) == "" {
		http.Redirect(w, r, "/login.html?error=missing_id_token", http.StatusFound)
		return
	}
	idToken, err := runtime.Verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Redirect(w, r, "/login.html?error=invalid_id_token", http.StatusFound)
		return
	}
	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
		Subject           string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Redirect(w, r, "/login.html?error=invalid_oidc_claims", http.StatusFound)
		return
	}
	username := firstNonEmpty(claims.PreferredUsername, claims.Email, claims.Name, claims.Subject)
	if username == "" {
		username = "oidc-user"
	}

	a.mu.Lock()
	a.createSessionLocked(w, r, username, "oidc")
	a.mu.Unlock()
	auditSuccess = true
	a.recordOIDCAudit(r, "login_success", "success", username, claims.Email, "CaddyMGM", "Administration login")
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *App) ensureConfig() error {
	if err := os.MkdirAll(filepath.Dir(a.configPath), 0o750); err != nil {
		return err
	}
	if err := os.MkdirAll(a.accessLogDir, 0o750); err != nil {
		return err
	}
	if _, err := os.Stat(a.configPath); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(a.configPath, []byte(managedStart+"\n"+managedEnd+"\n"), 0o640)
}

func (a *App) ensureSettings() error {
	if err := os.MkdirAll(filepath.Dir(a.settingsPath), 0o750); err != nil {
		return err
	}
	content, err := os.ReadFile(a.settingsPath)
	if err == nil {
		if err := json.Unmarshal(content, &a.settings); err != nil {
			return err
		}
		a.applySettingsEnvOverridesLocked(&a.settings)
		if strings.TrimSpace(a.settings.PasswordHash) == "" {
			return errors.New("CADDYMGM_ADMIN_PASSWORD is required when no administrator password is configured")
		}
		return a.saveSettingsLocked()
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	adminPassword := strings.TrimSpace(env("CADDYMGM_ADMIN_PASSWORD", ""))
	if adminPassword == "" {
		return errors.New("CADDYMGM_ADMIN_PASSWORD is required for first startup")
	}
	a.settings = Settings{
		AppName:      "CaddyMGM",
		Username:     env("CADDYMGM_ADMIN_USER", "admin"),
		PasswordHash: hashPassword(adminPassword),
		ConfigPath:   a.configPath,
		LogRetention: 100,
		ACMEIssuers:  ensureBuiltInACMEIssuers(nil),
		WebInterface: WebInterface{Upstream: defaultWebInterfaceUpstream(a.caddyMode)},
	}
	a.applySettingsEnvOverridesLocked(&a.settings)
	return a.saveSettingsLocked()
}

func (a *App) saveSettingsLocked() error {
	a.oidcCache = make(map[string]*oidcRuntime)
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
	a.applySettingsEnvOverridesLocked(&settings)
	settings.Password = ""
	settings.PasswordHash = ""
	settings.OIDC.ClientSecret = ""
	settings.ConfigPath = a.configPath
	settings.CaddyMode = a.caddyMode
	settings.CaddyAPIURL = a.caddyAPIURL
	return settings
}

func (a *App) applySettingsEnvOverridesLocked(settings *Settings) {
	settings.AuthEnabled = authEnabledFromEnv()
	settings.LocalAuthEnabled = localAuthEnabledFromEnv()
	settings.OIDCAuthEnabled = oidcAuthEnabledFromEnv()
	settings.ConfigPath = a.configPath
	settings.CaddyMode = a.caddyMode
	settings.CaddyAPIURL = a.caddyAPIURL
	if settings.AppName == "" {
		settings.AppName = "CaddyMGM"
	}
	if username := strings.TrimSpace(env("CADDYMGM_ADMIN_USER", settings.Username)); username != "" {
		settings.Username = username
	}
	if password := strings.TrimSpace(env("CADDYMGM_ADMIN_PASSWORD", "")); password != "" {
		if !passwordMatchesHash(password, settings.PasswordHash) {
			settings.PasswordHash = hashPassword(password)
		}
	}
	if settings.LogRetention == 0 {
		settings.LogRetention = 100
	}
	normalizeWebInterface(&settings.WebInterface, a.caddyMode)
	settings.OIDC = normalizeOIDCSettings(settings.OIDC, OIDCSettings{})
	if !settings.OIDCAuthEnabled {
		settings.OIDC.Enabled = false
	}
	settings.ACMEIssuers = ensureBuiltInACMEIssuers(settings.ACMEIssuers)
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
	sites = deduplicateSitesByAddress(sites)
	normalizeSiteIDs(sites)
	var out bytes.Buffer
	out.WriteString(strings.TrimRight(head, "\n"))
	if out.Len() > 0 {
		out.WriteString("\n\n")
	}
	out.WriteString(renderManaged(sites, a.settings.ACMEIssuers, a.caddyLogDir, a.settings.WebInterface, a.authProviders.OIDC, a.caddyMode, a.webPort))
	if strings.TrimSpace(tail) != "" {
		out.WriteString("\n")
		out.WriteString(strings.TrimLeft(tail, "\n"))
	}

	return writeFileAtomically(a.configPath, out.Bytes(), 0o640)
}

func (a *App) saveAndApplyCaddyConfigLocked(head string, sites []Site, tail string) error {
	previous, err := os.ReadFile(a.configPath)
	if err != nil {
		return err
	}
	if err := a.save(head, sites, tail); err != nil {
		return err
	}
	if err := a.applyCaddyConfigLocked(); err != nil {
		if restoreErr := writeFileAtomically(a.configPath, previous, 0o640); restoreErr != nil {
			return fmt.Errorf("%w; restoring previous config failed: %v", err, restoreErr)
		}
		return err
	}
	return nil
}

func writeFileAtomically(path string, content []byte, mode os.FileMode) error {
	tmp := fmt.Sprintf("%s.tmp.%d", path, time.Now().UnixNano())
	if err := os.WriteFile(tmp, content, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
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
	site := Site{ID: id, Enabled: true, TLSMode: "off", RewriteRedirects: true}
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
	inReverseProxy := false
	inTransport := false
	inAuthDirective := false
	inSecurityHeaderDirective := false
	inBasicAuthDirective := false
	skipManagedEncode := false
	logDepth := 0
	reverseProxyDepth := 0
	transportDepth := 0
	for _, raw := range lines[1:] {
		line := strings.TrimSpace(strings.TrimPrefix(raw, "#"))
		line = strings.TrimSpace(line)
		if line == "# caddymgm:auth-directive" {
			inAuthDirective = true
			continue
		}
		if line == "# caddymgm:end-auth-directive" {
			inAuthDirective = false
			continue
		}
		if line == "# caddymgm:security-header-directive" {
			inSecurityHeaderDirective = true
			continue
		}
		if line == "# caddymgm:end-security-header-directive" {
			inSecurityHeaderDirective = false
			continue
		}
		if line == "# caddymgm:basic-auth-directive" {
			inBasicAuthDirective = true
			continue
		}
		if line == "# caddymgm:end-basic-auth-directive" {
			inBasicAuthDirective = false
			continue
		}
		if inAuthDirective {
			continue
		}
		if inSecurityHeaderDirective {
			continue
		}
		if inBasicAuthDirective {
			if line == "basic_auth {" || line == "}" || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) == 2 {
				site.BasicAuthUsername = fields[0]
				site.BasicAuthPasswordHash = fields[1]
			}
			continue
		}
		if inLog {
			logDepth += braceDelta(line)
			if logDepth <= 0 {
				inLog = false
			}
			continue
		}
		if inTransport {
			if line == "tls_insecure_skip_verify" {
				site.SkipTLSVerify = true
			}
			transportDepth += braceDelta(line)
			if transportDepth <= 0 {
				inTransport = false
			}
			continue
		}
		if inReverseProxy {
			if line == "# caddymgm:skip-tls-verify" {
				site.SkipTLSVerify = true
				continue
			}
			if strings.HasPrefix(line, "header_down Location ") {
				continue
			}
			if line == "header_up Host {host}" {
				continue
			}
			if line == "transport http {" {
				inTransport = true
				transportDepth = braceDelta(line)
				if transportDepth <= 0 {
					transportDepth = 1
				}
				continue
			}
			reverseProxyDepth += braceDelta(line)
			if reverseProxyDepth <= 0 {
				inReverseProxy = false
				continue
			}
			if line != "" {
				extra = append(extra, line)
			}
			continue
		}
		if inTLS {
			switch {
			case strings.HasPrefix(line, "protocols "):
				versions := strings.Fields(strings.TrimPrefix(line, "protocols "))
				if len(versions) > 0 {
					site.TLSMinVersion = versions[0]
				}
				if len(versions) > 1 {
					site.TLSMaxVersion = versions[1]
				}
			case strings.HasPrefix(line, "ciphers "):
				site.TLSCipherSuites = strings.Fields(strings.TrimPrefix(line, "ciphers "))
			case line == "issuer internal":
				site.TLSMode = "internal"
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
		case strings.HasPrefix(line, "# caddymgm:compression "):
			site.CompressionProfile = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:compression "))
			skipManagedEncode = true
		case skipManagedEncode && strings.HasPrefix(line, "encode "):
			skipManagedEncode = false
			continue
		case strings.HasPrefix(line, "# caddymgm:comment "):
			site.Comment = parseManagedComment(strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:comment ")))
		case strings.HasPrefix(line, "# caddymgm:auth-provider "):
			site.AuthEnabled = true
			site.AuthProviderID = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:auth-provider "))
		case strings.HasPrefix(line, "# caddymgm:auth-groups "):
			continue
		case line == "# caddymgm:skip-tls-verify":
			site.SkipTLSVerify = true
		case line == "# caddymgm:rewrite-redirects":
			site.RewriteRedirects = true
		case line == "# caddymgm:no-rewrite-redirects":
			site.RewriteRedirects = false
		case strings.HasPrefix(line, "# caddymgm:redirect-origin "):
			origin, err := strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:redirect-origin ")))
			if err != nil {
				return Site{}, fmt.Errorf("invalid redirect origin marker: %w", err)
			}
			site.RedirectOrigins = append(site.RedirectOrigins, origin)
		case line == "# caddymgm:hsts":
			site.HSTSEnabled = true
		case strings.HasPrefix(line, "header Strict-Transport-Security "):
			continue
		case strings.HasPrefix(line, "# caddymgm:security-header-profile "):
			site.SecurityHeaderProfile = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:security-header-profile "))
		case line == "# caddymgm:basic-auth":
			site.BasicAuthEnabled = true
		case strings.HasPrefix(line, "# caddymgm:tls-issuer "):
			site.ACMEIssuerID = strings.TrimSpace(strings.TrimPrefix(line, "# caddymgm:tls-issuer "))
		case line == "tls internal":
			site.TLSMode = "internal"
		case line == "tls {" || strings.HasPrefix(line, "tls {"):
			if site.TLSMode == "off" {
				site.TLSMode = "acme"
			}
			inTLS = true
		case strings.HasPrefix(line, "reverse_proxy ") && strings.HasSuffix(line, "{"):
			site.Mode = "proxy"
			site.Upstream = strings.TrimSpace(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "reverse_proxy "), "{")))
			inReverseProxy = true
			reverseProxyDepth = braceDelta(line)
			if reverseProxyDepth <= 0 {
				reverseProxyDepth = 1
			}
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

func renderManaged(sites []Site, issuers []ACMEIssuer, logDir string, webInterface WebInterface, accessProvider AccessOIDCProvider, caddyMode string, webPort string) string {
	var out strings.Builder
	out.WriteString(managedStart + "\n")
	out.WriteString(renderUnavailableSnippet())
	if block := renderWebInterface(webInterface, issuers, caddyMode, webPort); block != "" {
		out.WriteString(block)
	}
	if block := renderAccessGateway(accessProvider, issuers, effectiveWebInterfaceUpstream(webInterface, caddyMode)); block != "" {
		out.WriteString(block)
	}
	for _, site := range sites {
		out.WriteString("# caddymgm:site " + site.ID + "\n")
		out.WriteString(renderSite(site, issuers, logDir, effectiveWebInterfaceUpstream(webInterface, caddyMode)))
		out.WriteString("# caddymgm:end-site\n")
		if !site.Enabled {
			out.WriteString(renderUnavailableSite(site, issuers, logDir))
		}
	}
	out.WriteString(managedEnd + "\n")
	return out.String()
}

const unavailablePageHTML = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Website unavailable</title><style>:root{color-scheme:dark}*{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;padding:24px;background:#07080b;color:#f3f6fb;font:16px/1.5 Inter,system-ui,sans-serif}.card{width:min(560px,100%);padding:42px;border:1px solid #2b2d36;border-radius:18px;background:#191a20;box-shadow:0 24px 80px rgba(0,0,0,.4);text-align:center}.mark{display:grid;width:54px;height:54px;margin:0 auto 22px;place-items:center;border:1px solid rgba(91,184,255,.3);border-radius:50%;background:rgba(91,184,255,.1);color:#5bb8ff;font-size:26px}h1{margin:0 0 10px;font-size:28px}p{margin:0;color:#8d96a8}</style></head><body><main class="card"><div class="mark" aria-hidden="true">!</div><h1>Website unavailable</h1><p>This website is currently disabled. Please try again later.</p></main></body></html>`

func renderUnavailableSnippet() string {
	var out strings.Builder
	out.WriteString("(caddymgm_unavailable) {\n")
	out.WriteString("\theader {\n")
	out.WriteString("\t\tContent-Type \"text/html; charset=utf-8\"\n")
	out.WriteString("\t\tCache-Control \"no-store\"\n")
	out.WriteString("\t\tRetry-After \"3600\"\n")
	out.WriteString("\t\tX-Content-Type-Options \"nosniff\"\n")
	out.WriteString("\t\tX-Frame-Options \"DENY\"\n")
	out.WriteString("\t}\n")
	out.WriteString("\trespond " + strconv.Quote(unavailablePageHTML) + " 503\n")
	out.WriteString("}\n")
	return out.String()
}

func renderUnavailableSite(site Site, issuers []ACMEIssuer, logDir string) string {
	var out strings.Builder
	address := site.Address
	if site.TLSMode == "" || site.TLSMode == "off" {
		address = "http://" + strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://")
	}
	out.WriteString("# caddymgm:unavailable-site " + site.ID + "\n")
	out.WriteString(address + " {\n")
	out.WriteString("\timport caddymgm_unavailable\n")
	writeCompression(&out, "", site)
	if site.HSTSEnabled && site.TLSMode != "" && site.TLSMode != "off" {
		out.WriteString("\theader Strict-Transport-Security \"max-age=31536000\"\n")
	}
	tlsOptions := site.TLSMinVersion != "" || site.TLSMaxVersion != "" || len(site.TLSCipherSuites) > 0
	switch site.TLSMode {
	case "internal":
		if !tlsOptions {
			out.WriteString("\ttls internal\n")
		} else {
			out.WriteString("\ttls {\n")
			writeTLSOptions(&out, "", site)
			out.WriteString("\t\tissuer internal\n")
			out.WriteString("\t}\n")
		}
	case "acme":
		if issuer, ok := findACMEIssuer(issuers, site.ACMEIssuerID); ok {
			out.WriteString("\ttls {\n")
			writeTLSOptions(&out, "", site)
			out.WriteString("\t\tissuer acme {\n")
			out.WriteString("\t\t\tdir " + issuer.DirectoryURL + "\n")
			if issuer.Email != "" {
				out.WriteString("\t\t\temail " + issuer.Email + "\n")
			}
			if issuer.RootCAFile != "" {
				out.WriteString("\t\t\ttrusted_roots " + caddyfileQuote(issuer.RootCAFile) + "\n")
			}
			out.WriteString("\t\t}\n")
			out.WriteString("\t}\n")
		}
	}
	if site.LogsEnabled {
		out.WriteString("\tlog {\n")
		out.WriteString("\t\toutput file " + accessLogPath(logDir, site.ID) + " {\n")
		out.WriteString("\t\t\tmode 0644\n")
		out.WriteString("\t\t}\n")
		out.WriteString("\t\tformat json\n")
		out.WriteString("\t}\n")
	}
	out.WriteString("}\n")
	out.WriteString("# caddymgm:end-unavailable-site\n")
	return out.String()
}

func renderWebInterface(webInterface WebInterface, issuers []ACMEIssuer, caddyMode string, webPort string) string {
	host := strings.TrimSpace(webInterface.Host)
	upstream := effectiveWebInterfaceUpstream(webInterface, caddyMode)
	if upstream == "" {
		return ""
	}
	var out strings.Builder
	out.WriteString("# caddymgm:web-interface\n")
	if host == "" {
		out.WriteString("http://:" + webPort + " {\n")
	} else if webInterface.TLSEnabled {
		out.WriteString("https://" + host + ":" + webPort + " {\n")
	} else {
		out.WriteString("http://" + strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://") + ":" + webPort + " {\n")
	}
	out.WriteString("\treverse_proxy " + upstream + "\n")
	if webInterface.TLSEnabled {
		if issuer, ok := findACMEIssuer(issuers, webInterface.ACMEIssuerID); ok {
			out.WriteString("\t# caddymgm:tls-issuer " + issuer.ID + "\n")
			out.WriteString("\ttls {\n")
			out.WriteString("\t\tissuer acme {\n")
			out.WriteString("\t\t\tdir " + issuer.DirectoryURL + "\n")
			if issuer.Email != "" {
				out.WriteString("\t\t\temail " + issuer.Email + "\n")
			}
			if issuer.RootCAFile != "" {
				out.WriteString("\t\t\ttrusted_roots " + caddyfileQuote(issuer.RootCAFile) + "\n")
			}
			out.WriteString("\t\t}\n")
			out.WriteString("\t}\n")
		}
	}
	out.WriteString("}\n")
	out.WriteString("# caddymgm:end-web-interface\n")
	return out.String()
}

func renderAccessGateway(provider AccessOIDCProvider, issuers []ACMEIssuer, upstream string) string {
	if !provider.Enabled || upstream == "" {
		return ""
	}
	u, err := url.Parse(provider.GatewayURL)
	if err != nil || u.Host == "" {
		return ""
	}
	var out strings.Builder
	out.WriteString("# caddymgm:access-gateway\n")
	out.WriteString("https://" + u.Host + " {\n")
	out.WriteString("\trewrite / /.caddymgm/auth/portal\n")
	out.WriteString("\treverse_proxy /.caddymgm/auth/* " + upstream + "\n")
	if issuer, ok := findACMEIssuer(issuers, provider.ACMEIssuerID); ok {
		out.WriteString("\t# caddymgm:tls-issuer " + issuer.ID + "\n")
		out.WriteString("\ttls {\n")
		out.WriteString("\t\tissuer acme {\n")
		out.WriteString("\t\t\tdir " + issuer.DirectoryURL + "\n")
		if issuer.Email != "" {
			out.WriteString("\t\t\temail " + issuer.Email + "\n")
		}
		if issuer.RootCAFile != "" {
			out.WriteString("\t\t\ttrusted_roots " + caddyfileQuote(issuer.RootCAFile) + "\n")
		}
		out.WriteString("\t\t}\n")
		out.WriteString("\t}\n")
	}
	out.WriteString("}\n")
	out.WriteString("# caddymgm:end-access-gateway\n")
	return out.String()
}

func redirectRewriteRule(site Site) (string, string, bool) {
	upstream, err := url.Parse(strings.TrimSpace(site.Upstream))
	if err != nil || (upstream.Scheme != "http" && upstream.Scheme != "https") || upstream.Host == "" {
		return "", "", false
	}
	upstreamOrigin := upstream.Scheme + "://" + upstream.Host
	publicScheme := "http"
	if site.TLSMode != "" && site.TLSMode != "off" {
		publicScheme = "https"
	}
	publicOrigin := publicScheme + "://" + site.Address
	pattern := `(?i)^` + regexp.QuoteMeta(upstreamOrigin) + `([/?#].*)?` + "$"
	replacement := publicOrigin + "$" + "{1}"
	return pattern, replacement, true
}

func redirectRewriteRules(site Site) [][2]string {
	origins := make([]string, 0, len(site.RedirectOrigins)+1)
	if upstream, err := url.Parse(strings.TrimSpace(site.Upstream)); err == nil && (upstream.Scheme == "http" || upstream.Scheme == "https") && upstream.Host != "" {
		origins = append(origins, upstream.Scheme+"://"+upstream.Host)
	}
	origins = append(origins, site.RedirectOrigins...)
	publicScheme := "http"
	if site.TLSMode != "" && site.TLSMode != "off" {
		publicScheme = "https"
	}
	publicOrigin := publicScheme + "://" + site.Address
	rules := make([][2]string, 0, len(origins))
	seen := map[string]bool{}
	for _, origin := range origins {
		key := strings.ToLower(origin)
		if seen[key] {
			continue
		}
		seen[key] = true
		pattern := `(?i)^` + regexp.QuoteMeta(origin) + `([/?#].*)?` + "$"
		rules = append(rules, [2]string{pattern, publicOrigin + "$" + "{1}"})
	}
	return rules
}

func renderSite(site Site, issuers []ACMEIssuer, logDir, authGatewayUpstream string) string {
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
	if site.Comment != "" {
		out.WriteString(prefix + "\t# caddymgm:comment " + strconv.Quote(site.Comment) + "\n")
	}
	if site.BasicAuthEnabled {
		out.WriteString(prefix + "\t# caddymgm:basic-auth\n")
		out.WriteString(prefix + "\t# caddymgm:basic-auth-directive\n")
		out.WriteString(prefix + "\tbasic_auth {\n")
		out.WriteString(prefix + "\t\t" + site.BasicAuthUsername + " " + site.BasicAuthPasswordHash + "\n")
		out.WriteString(prefix + "\t}\n")
		out.WriteString(prefix + "\t# caddymgm:end-basic-auth-directive\n")
	}
	if site.AuthEnabled {
		out.WriteString(prefix + "\t# caddymgm:auth-provider " + site.AuthProviderID + "\n")
		out.WriteString(prefix + "\t# caddymgm:auth-directive\n")
		out.WriteString(prefix + "\t@caddymgmProtected not path /.caddymgm/auth/*\n")
		out.WriteString(prefix + "\tforward_auth @caddymgmProtected " + authGatewayUpstream + " {\n")
		out.WriteString(prefix + "\t\turi /.caddymgm/auth/check\n")
		out.WriteString(prefix + "\t\tcopy_headers X-Auth-User X-Auth-Email\n")
		out.WriteString(prefix + "\t}\n")
		out.WriteString(prefix + "\treverse_proxy /.caddymgm/auth/* " + authGatewayUpstream + "\n")
		out.WriteString(prefix + "\t# caddymgm:end-auth-directive\n")
	}
	if site.Mode == "static" {
		out.WriteString(prefix + "\troot * " + site.Root + "\n")
		out.WriteString(prefix + "\tfile_server\n")
	} else {
		if site.RewriteRedirects {
			out.WriteString(prefix + "\t# caddymgm:rewrite-redirects\n")
			for _, origin := range site.RedirectOrigins {
				out.WriteString(prefix + "\t# caddymgm:redirect-origin " + strconv.Quote(origin) + "\n")
			}
		} else {
			out.WriteString(prefix + "\t# caddymgm:no-rewrite-redirects\n")
		}
		redirectRules := redirectRewriteRules(site)
		canRewrite := len(redirectRules) > 0
		skipTLSVerify := site.SkipTLSVerify && !strings.HasPrefix(strings.ToLower(strings.TrimSpace(site.Upstream)), "http://")
		out.WriteString(prefix + "\treverse_proxy " + site.Upstream + " {\n")
		out.WriteString(prefix + "\t\theader_up Host {host}\n")
		if site.RewriteRedirects && canRewrite {
			for _, rule := range redirectRules {
				out.WriteString(prefix + "\t\theader_down Location " + caddyfileQuote(rule[0]) + " " + caddyfileQuote(rule[1]) + "\n")
			}
		}
		if skipTLSVerify {
			out.WriteString(prefix + "\t\t# caddymgm:skip-tls-verify\n")
			out.WriteString(prefix + "\t\ttransport http {\n")
			out.WriteString(prefix + "\t\t\ttls_insecure_skip_verify\n")
			out.WriteString(prefix + "\t\t}\n")
		}
		out.WriteString(prefix + "\t}\n")
	}
	if site.HSTSEnabled && site.TLSMode != "" && site.TLSMode != "off" {
		out.WriteString(prefix + "\t# caddymgm:hsts\n")
		out.WriteString(prefix + "\theader Strict-Transport-Security \"max-age=31536000\"\n")
	}
	if headers := securityHeaderProfileDirectives(site.SecurityHeaderProfile); len(headers) > 0 {
		out.WriteString(prefix + "\t# caddymgm:security-header-profile " + site.SecurityHeaderProfile + "\n")
		out.WriteString(prefix + "\t# caddymgm:security-header-directive\n")
		out.WriteString(prefix + "\theader {\n")
		for _, header := range headers {
			out.WriteString(prefix + "\t\t" + header + "\n")
		}
		out.WriteString(prefix + "\t}\n")
		out.WriteString(prefix + "\t# caddymgm:end-security-header-directive\n")
	}
	writeCompression(&out, prefix, site)
	tlsOptions := site.TLSMinVersion != "" || site.TLSMaxVersion != "" || len(site.TLSCipherSuites) > 0
	switch site.TLSMode {
	case "internal":
		if !tlsOptions {
			out.WriteString(prefix + "\ttls internal\n")
		} else {
			out.WriteString(prefix + "\ttls {\n")
			writeTLSOptions(&out, prefix, site)
			out.WriteString(prefix + "\t\tissuer internal\n")
			out.WriteString(prefix + "\t}\n")
		}
	case "acme":
		if issuer, ok := findACMEIssuer(issuers, site.ACMEIssuerID); ok {
			out.WriteString(prefix + "\t# caddymgm:tls-issuer " + issuer.ID + "\n")
			out.WriteString(prefix + "\ttls {\n")
			writeTLSOptions(&out, prefix, site)
			out.WriteString(prefix + "\t\tissuer acme {\n")
			out.WriteString(prefix + "\t\t\tdir " + issuer.DirectoryURL + "\n")
			if issuer.Email != "" {
				out.WriteString(prefix + "\t\t\temail " + issuer.Email + "\n")
			}
			if issuer.RootCAFile != "" {
				out.WriteString(prefix + "\t\t\ttrusted_roots " + caddyfileQuote(issuer.RootCAFile) + "\n")
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

func writeCompression(out *strings.Builder, prefix string, site Site) {
	switch site.CompressionProfile {
	case "gzip":
		out.WriteString(prefix + "\t# caddymgm:compression gzip\n")
		out.WriteString(prefix + "\tencode gzip\n")
	case "zstd-gzip":
		out.WriteString(prefix + "\t# caddymgm:compression zstd-gzip\n")
		out.WriteString(prefix + "\tencode zstd gzip\n")
	}
}

func writeTLSOptions(out *strings.Builder, prefix string, site Site) {
	if site.TLSMinVersion != "" || site.TLSMaxVersion != "" {
		minVersion := site.TLSMinVersion
		if minVersion == "" {
			minVersion = "tls1.2"
		}
		out.WriteString(prefix + "\t\tprotocols " + minVersion)
		if site.TLSMaxVersion != "" {
			out.WriteString(" " + site.TLSMaxVersion)
		}
		out.WriteString("\n")
	}
	if len(site.TLSCipherSuites) > 0 {
		out.WriteString(prefix + "\t\tciphers " + strings.Join(site.TLSCipherSuites, " ") + "\n")
	}
}

var supportedTLSCipherSuites = map[string]bool{
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":       true,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":         true,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       true,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         true,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": true,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   true,
}

var addressPattern = regexp.MustCompile(`^[A-Za-z0-9*._:-]+$`)

func normalizeSite(site *Site) error {
	site.Address = strings.TrimSpace(site.Address)
	site.Comment = strings.TrimSpace(site.Comment)
	site.Mode = strings.TrimSpace(site.Mode)
	site.Upstream = strings.TrimSpace(site.Upstream)
	site.Root = strings.TrimSpace(site.Root)
	site.ExtraDirectives = cleanExtraDirectives(site.ExtraDirectives)
	site.TLSMode = strings.TrimSpace(site.TLSMode)
	site.ACMEIssuerID = strings.TrimSpace(site.ACMEIssuerID)
	site.TLSMinVersion = strings.ToLower(strings.TrimSpace(site.TLSMinVersion))
	site.TLSMaxVersion = strings.ToLower(strings.TrimSpace(site.TLSMaxVersion))
	site.AuthProviderID = strings.TrimSpace(site.AuthProviderID)
	site.BasicAuthUsername = strings.TrimSpace(site.BasicAuthUsername)
	site.SecurityHeaderProfile = strings.ToLower(strings.TrimSpace(site.SecurityHeaderProfile))
	site.CompressionProfile = strings.ToLower(strings.TrimSpace(site.CompressionProfile))
	redirectOrigins := make([]string, 0, len(site.RedirectOrigins))
	seenRedirectOrigins := map[string]bool{}
	for _, rawOrigin := range site.RedirectOrigins {
		origin := strings.TrimSpace(rawOrigin)
		if origin == "" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
			return errors.New("redirect origins must be complete HTTP or HTTPS origins without paths")
		}
		normalized := strings.ToLower(parsed.Scheme) + "://" + parsed.Host
		key := strings.ToLower(normalized)
		if !seenRedirectOrigins[key] {
			seenRedirectOrigins[key] = true
			redirectOrigins = append(redirectOrigins, normalized)
		}
	}
	site.RedirectOrigins = redirectOrigins
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
		if site.SkipTLSVerify && strings.HasPrefix(strings.ToLower(site.Upstream), "http://") {
			return errors.New("skip upstream TLS verification requires an HTTPS upstream")
		}
	case "static":
		site.SkipTLSVerify = false
		site.RewriteRedirects = false
		site.RedirectOrigins = nil
		if site.Root == "" {
			return errors.New("root path is required")
		}
	default:
		return errors.New("mode must be proxy or static")
	}
	if !site.RewriteRedirects {
		site.RedirectOrigins = nil
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
	if site.TLSMode == "off" {
		site.HSTSEnabled = false
		site.TLSMinVersion = ""
		site.TLSMaxVersion = ""
		site.TLSCipherSuites = nil
	} else {
		for _, tlsVersion := range []string{site.TLSMinVersion, site.TLSMaxVersion} {
			if tlsVersion != "" && tlsVersion != "tls1.2" && tlsVersion != "tls1.3" {
				return errors.New("TLS protocol version must be tls1.2 or tls1.3")
			}
		}
		if site.TLSMinVersion == "tls1.3" && site.TLSMaxVersion == "tls1.2" {
			return errors.New("minimum TLS protocol version cannot exceed maximum version")
		}
		cipherSuites := make([]string, 0, len(site.TLSCipherSuites))
		seenCipherSuites := map[string]bool{}
		for _, cipherSuite := range site.TLSCipherSuites {
			cipherSuite = strings.ToUpper(strings.TrimSpace(cipherSuite))
			if !supportedTLSCipherSuites[cipherSuite] {
				return fmt.Errorf("unsupported TLS cipher suite %q", cipherSuite)
			}
			if !seenCipherSuites[cipherSuite] {
				seenCipherSuites[cipherSuite] = true
				cipherSuites = append(cipherSuites, cipherSuite)
			}
		}
		site.TLSCipherSuites = cipherSuites
	}
	switch site.SecurityHeaderProfile {
	case "", "standard", "strict":
	default:
		return errors.New("security header profile must be off, standard or strict")
	}
	switch site.CompressionProfile {
	case "", "gzip", "zstd-gzip":
	default:
		return errors.New("compression profile must be off, gzip or zstd-gzip")
	}
	if site.CompressionProfile != "" {
		for _, line := range strings.Split(site.ExtraDirectives, "\n") {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "encode ") {
				return errors.New("remove the manual encode directive before enabling managed compression")
			}
		}
	}
	if site.AuthEnabled {
		if site.TLSMode == "off" {
			return errors.New("website authentication requires TLS")
		}
		if site.AuthProviderID == "" {
			site.AuthProviderID = "oidc"
		}
	} else {
		site.AuthProviderID = ""
	}
	return nil
}

var basicAuthUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9._@-]{1,64}$`)

func prepareBasicAuth(site *Site, existing *Site) error {
	site.BasicAuthUsername = strings.TrimSpace(site.BasicAuthUsername)
	password := site.BasicAuthPassword
	site.BasicAuthPassword = ""
	if !site.BasicAuthEnabled {
		site.BasicAuthUsername = ""
		site.BasicAuthPasswordHash = ""
		return nil
	}
	if site.AuthEnabled {
		return errors.New("basic authentication and OIDC authentication cannot be enabled together")
	}
	if site.TLSMode == "" || site.TLSMode == "off" {
		return errors.New("basic authentication requires TLS")
	}
	if !basicAuthUsernamePattern.MatchString(site.BasicAuthUsername) {
		return errors.New("basic authentication username must use 1-64 letters, numbers, dots, underscores, @ signs or hyphens")
	}
	if password != "" {
		if len([]byte(password)) < 8 || len([]byte(password)) > 72 {
			return errors.New("basic authentication password must contain 8-72 bytes")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("could not secure basic authentication password")
		}
		site.BasicAuthPasswordHash = string(hash)
	} else if existing != nil && existing.BasicAuthEnabled {
		site.BasicAuthPasswordHash = existing.BasicAuthPasswordHash
	}
	if site.BasicAuthPasswordHash == "" {
		return errors.New("basic authentication password is required")
	}
	if _, err := bcrypt.Cost([]byte(site.BasicAuthPasswordHash)); err != nil {
		return errors.New("basic authentication password hash is invalid")
	}
	return nil
}

func securityHeaderProfileDirectives(profile string) []string {
	if profile == "standard" {
		return []string{
			`-Server`,
			`X-Content-Type-Options "nosniff"`,
			`Referrer-Policy "strict-origin-when-cross-origin"`,
			`X-Frame-Options "SAMEORIGIN"`,
		}
	}
	if profile == "strict" {
		return []string{
			`-Server`,
			`X-Content-Type-Options "nosniff"`,
			`Referrer-Policy "no-referrer"`,
			`X-Frame-Options "DENY"`,
			`Permissions-Policy "camera=(), geolocation=(), microphone=()"`,
			`Content-Security-Policy "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; object-src 'none'; img-src 'self' data: https:; font-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self' https: wss:"`,
		}
	}
	return nil
}

func duplicateSiteAddress(address string, sites []Site, excludedID string) bool {
	address = strings.ToLower(strings.TrimSpace(address))
	for _, site := range sites {
		if site.ID != excludedID && strings.ToLower(strings.TrimSpace(site.Address)) == address {
			return true
		}
	}
	return false
}

func deduplicateSitesByAddress(sites []Site) []Site {
	seen := make(map[string]struct{}, len(sites))
	unique := make([]Site, 0, len(sites))
	for _, site := range sites {
		address := strings.ToLower(strings.TrimSpace(site.Address))
		if _, exists := seen[address]; exists {
			continue
		}
		seen[address] = struct{}{}
		unique = append(unique, site)
	}
	return unique
}

func parseManagedComment(raw string) string {
	if raw == "" {
		return ""
	}
	if unquoted, err := strconv.Unquote(raw); err == nil {
		return strings.TrimSpace(unquoted)
	}
	return strings.TrimSpace(raw)
}

func (a *App) validateSiteLocked(site Site) error {
	if site.AuthEnabled {
		if validateAccessProvider(a.authProviders.OIDC) != nil || site.AuthProviderID != a.authProviders.OIDC.ID {
			return errors.New("selected website access provider is not enabled")
		}
		if effectiveWebInterfaceUpstream(a.settings.WebInterface, a.caddyMode) == "" {
			return errors.New("website authentication requires a reachable CaddyMGM web interface upstream")
		}
	}
	if site.Mode == "static" {
		if err := validateStaticRoot(site.Root, a.staticRootBase); err != nil {
			return err
		}
	}
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

func validateStaticRoot(root, base string) error {
	root = filepath.Clean(strings.TrimSpace(root))
	base = filepath.Clean(strings.TrimSpace(base))
	if root == "" {
		return errors.New("root path is required")
	}
	if base == "" || !filepath.IsAbs(base) {
		return errors.New("static root base must be an absolute path")
	}
	if strings.ContainsAny(root, "\r\n\t") {
		return errors.New("root path contains unsupported characters")
	}
	if !filepath.IsAbs(root) {
		return fmt.Errorf("root path must stay under %s", base)
	}
	rel, err := filepath.Rel(base, root)
	if err != nil {
		return fmt.Errorf("root path must stay under %s", base)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("root path must stay under %s", base)
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
		issuers[i].RootCAFile = strings.TrimSpace(issuers[i].RootCAFile)
		if issuers[i].ID == "letsencrypt" {
			issuers[i].Name = "Let's Encrypt"
			issuers[i].DirectoryURL = "https://acme-v02.api.letsencrypt.org/directory"
			issuers[i].RootCAFile = ""
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
		if err := validateRootCAFile(issuers[i].RootCAFile); err != nil {
			return nil, err
		}
	}
	return issuers, nil
}

func ensureBuiltInACMEIssuers(issuers []ACMEIssuer) []ACMEIssuer {
	for i := range issuers {
		if issuers[i].ID == "letsencrypt" {
			issuers[i].Name = "Let's Encrypt"
			issuers[i].DirectoryURL = "https://acme-v02.api.letsencrypt.org/directory"
			issuers[i].RootCAFile = ""
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

func validateRootCAFile(path string) error {
	if path == "" {
		return nil
	}
	if strings.Contains(path, "..") || strings.ContainsAny(path, "\n\r\t") {
		return errors.New("root CA file path is invalid")
	}
	if !strings.HasPrefix(path, "/ca-certificates/") || !isRootCAFile(path) {
		return errors.New("root CA file must be a .crt, .cer or .pem file under /ca-certificates")
	}
	return nil
}

func normalizeWebInterface(webInterface *WebInterface, caddyMode string) {
	webInterface.Host = strings.TrimSpace(webInterface.Host)
	webInterface.Upstream = strings.TrimSpace(webInterface.Upstream)
	webInterface.ACMEIssuerID = strings.TrimSpace(webInterface.ACMEIssuerID)
	if webInterface.Upstream == "" {
		webInterface.Upstream = defaultWebInterfaceUpstream(caddyMode)
	}
	if strings.HasPrefix(webInterface.Upstream, "http://") || strings.HasPrefix(webInterface.Upstream, "https://") {
		if parsed, err := url.Parse(webInterface.Upstream); err == nil && parsed.Host != "" {
			webInterface.Upstream = parsed.Host
		}
	}
	if !webInterface.TLSEnabled {
		webInterface.ACMEIssuerID = ""
	}
}

func validateWebInterface(webInterface WebInterface, issuers []ACMEIssuer) error {
	host := strings.TrimSpace(webInterface.Host)
	if host == "" && webInterface.TLSEnabled {
		return errors.New("web interface host is required when TLS is enabled")
	}
	parsedHost, _, err := splitHostPortLoose(host)
	if err != nil {
		return errors.New("web interface host is invalid")
	}
	if parsedHost != "" {
		host = parsedHost
	}
	host = strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://")
	if host != "" && !addressPattern.MatchString(host) {
		return errors.New("web interface host contains unsupported characters")
	}
	if upstream := strings.TrimSpace(webInterface.Upstream); upstream != "" {
		if _, err := normalizeProxyUpstream(upstream); err != nil {
			return fmt.Errorf("web interface upstream: %w", err)
		}
	}
	if webInterface.TLSEnabled {
		if webInterface.ACMEIssuerID == "" {
			return errors.New("web interface ACME authority is required when TLS is enabled")
		}
		if _, ok := findACMEIssuer(issuers, webInterface.ACMEIssuerID); !ok {
			return errors.New("web interface ACME authority was not found")
		}
	}
	return nil
}

func normalizeOIDCSettings(next, current OIDCSettings) OIDCSettings {
	next.IssuerURL = strings.TrimSpace(next.IssuerURL)
	next.ClientID = strings.TrimSpace(next.ClientID)
	next.ClientSecret = strings.TrimSpace(next.ClientSecret)
	next.RedirectURL = strings.TrimSpace(next.RedirectURL)
	next.Scopes = strings.TrimSpace(next.Scopes)
	if next.ClientSecret == "" {
		next.ClientSecret = current.ClientSecret
	}
	if next.Scopes == "" {
		next.Scopes = "openid profile email"
	}
	return next
}

func validateOIDCSettings(settings OIDCSettings) error {
	if !settings.Enabled {
		return nil
	}
	if err := validateOIDCURL(settings.IssuerURL, "issuer"); err != nil {
		return err
	}
	if strings.TrimSpace(settings.ClientID) == "" {
		return errors.New("oidc client id is required")
	}
	if strings.TrimSpace(settings.ClientSecret) == "" {
		return errors.New("oidc client secret is required")
	}
	if err := validateOIDCURL(settings.RedirectURL, "redirect"); err != nil {
		return err
	}
	return nil
}

func validateOIDCURL(raw, name string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fmt.Errorf("oidc %s url is required", name)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("oidc %s url is invalid", name)
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname()) {
		return nil
	}
	return fmt.Errorf("oidc %s url must use HTTPS unless it targets loopback", name)
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	address := net.ParseIP(strings.Trim(strings.TrimSpace(host), "[]"))
	return address != nil && address.IsLoopback()
}

func isRootCAFile(path string) bool {
	path = strings.ToLower(path)
	return strings.HasSuffix(path, ".crt") || strings.HasSuffix(path, ".cer") || strings.HasSuffix(path, ".pem")
}

func caddyfileQuote(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func normalizeCertificateBundle(content []byte) ([]byte, error) {
	var out bytes.Buffer
	remaining := content
	found := false
	for {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		remaining = rest
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.New("certificate file contains an invalid certificate")
		}
		_ = pem.Encode(&out, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
		found = true
	}
	if found {
		return out.Bytes(), nil
	}
	cert, err := x509.ParseCertificate(bytes.TrimSpace(content))
	if err != nil {
		return nil, errors.New("certificate file must contain a PEM or DER encoded certificate")
	}
	_ = pem.Encode(&out, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	return out.Bytes(), nil
}

func rootCAFilename(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		base = "root-ca"
	}
	var out strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(base) {
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
	safe := strings.Trim(out.String(), ".-")
	if safe == "" {
		safe = "root-ca"
	}
	return safe + ".crt"
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

func (a *App) readServiceLogsLocked() []LogEntry {
	limit := a.settings.LogRetention
	if limit <= 0 {
		limit = 100
	}
	lines, err := readLastLines(a.serviceLog, limit)
	if err != nil {
		return nil
	}
	entries := make([]LogEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entries = append(entries, serviceLogEntry(line))
	}
	return entries
}

func (a *App) serviceLogAvailableLocked() bool {
	if a.caddyMode != "docker" {
		return false
	}
	_, err := os.Stat(a.serviceLog)
	return err == nil
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

func serviceLogEntry(line string) LogEntry {
	entry := LogEntry{
		Time:    time.Now().Format(time.RFC3339),
		Action:  "service",
		Message: strings.TrimSpace(line),
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
	if msg, ok := payload["msg"].(string); ok && strings.TrimSpace(msg) != "" {
		entry.Message = msg
	}
	if logger, ok := payload["logger"].(string); ok && strings.TrimSpace(logger) != "" {
		entry.Action = logger
	}
	if level, ok := payload["level"].(string); ok && strings.TrimSpace(level) != "" {
		entry.Status = strings.ToUpper(level)
	}
	extra := make([]string, 0, 4)
	if identifier, ok := payload["identifier"].(string); ok && strings.TrimSpace(identifier) != "" {
		extra = append(extra, identifier)
	}
	if challengeType, ok := payload["challenge_type"].(string); ok && strings.TrimSpace(challengeType) != "" {
		extra = append(extra, challengeType)
	}
	if addr, ok := payload["addr"].(string); ok && strings.TrimSpace(addr) != "" {
		extra = append(extra, addr)
	}
	if ca, ok := payload["ca"].(string); ok && strings.TrimSpace(ca) != "" {
		extra = append(extra, ca)
	}
	if domains, ok := payload["domains"].([]any); ok {
		for _, value := range domains {
			if domain, ok := value.(string); ok && strings.TrimSpace(domain) != "" {
				extra = append(extra, domain)
			}
		}
	}
	if len(extra) > 0 {
		entry.Message = entry.Message + " - " + strings.Join(uniqueStrings(extra), " - ")
	}
	return entry
}

func accessLogPath(dir, siteID string) string {
	return filepath.Join(dir, siteID+".access.log")
}

func (a *App) populateCertificateMetadata(sites []Site) {
	for i := range sites {
		a.populateSiteCertificateMetadata(&sites[i])
	}
}

func (a *App) populateSiteCertificateMetadata(site *Site) {
	if site == nil {
		return
	}
	site.CertificateExpiresAt = ""
	expiresAt, err := a.certificateExpiresAt(site.Address)
	if err == nil && !expiresAt.IsZero() {
		site.CertificateExpiresAt = expiresAt.Format(time.RFC3339)
	}
}

func (a *App) certificateExpiresAt(domain string) (time.Time, error) {
	domain = strings.ToLower(siteIDFromAddress(domain))
	if domain == "" || a.caddyDataDir == "" {
		return time.Time{}, os.ErrNotExist
	}
	var newest time.Time
	err := filepath.WalkDir(a.certificateRoot(), func(path string, entry fs.DirEntry, err error) error {
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
	_ = filepath.WalkDir(a.certificateRoot(), func(path string, entry fs.DirEntry, err error) error {
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

func (a *App) certificateRoot() string {
	direct := filepath.Join(a.caddyDataDir, "certificates")
	if _, err := os.Stat(direct); err == nil {
		return direct
	}
	return filepath.Join(a.caddyDataDir, "caddy", "certificates")
}

func (a *App) triggerTLSHandshake(domain string) error {
	host := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(domain, "http://"), "https://"))
	if host == "" {
		return errors.New("missing host for tls handshake")
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName: host,
			},
		},
	}

	resp, err := client.Get("https://" + host)
	if err != nil {
		if isExpectedTLSHandshakeTriggerError(err) {
			return nil
		}
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func isExpectedTLSHandshakeTriggerError(err error) bool {
	if err == nil {
		return false
	}
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return true
	}
	var hostnameError x509.HostnameError
	if errors.As(err, &hostnameError) {
		return true
	}
	var invalidCertificate x509.CertificateInvalidError
	if errors.As(err, &invalidCertificate) {
		return true
	}
	return false
}

func noCacheFileServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
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

func validateListenAddress(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("listen address is required")
	}
	if strings.HasPrefix(value, ":") {
		if _, err := net.LookupPort("tcp", strings.TrimPrefix(value, ":")); err != nil {
			return errors.New("listen port is invalid")
		}
		return nil
	}
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		return errors.New("listen address must use host:port or :port")
	}
	if host == "" {
		return errors.New("listen host is invalid")
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return errors.New("listen port is invalid")
	}
	return nil
}

func splitHostPortLoose(value string) (string, string, error) {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(value, "http://"), "https://"))
	if value == "" {
		return "", "", nil
	}
	if strings.Count(value, ":") == 0 {
		return value, "", nil
	}
	if host, port, err := net.SplitHostPort(value); err == nil {
		return strings.Trim(host, "[]"), port, nil
	}
	host, port, ok := strings.Cut(value, ":")
	if !ok || host == "" || port == "" {
		return "", "", errors.New("invalid host")
	}
	if strings.Contains(port, ":") {
		return "", "", errors.New("invalid host")
	}
	return host, port, nil
}

func (a *App) syncManagedConfig() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, head, tail, err := a.load()
	if err != nil {
		return err
	}
	if err := a.save(head, sites, tail); err != nil {
		return err
	}
	return a.applyCaddyConfigLocked()
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

func defaultWebInterfaceUpstream(mode string) string {
	switch mode {
	case "docker":
		return "http://caddymgm:8080"
	case "native", "api":
		return "http://host.docker.internal:8080"
	default:
		return "http://caddymgm:8080"
	}
}

func effectiveWebInterfaceUpstream(webInterface WebInterface, caddyMode string) string {
	upstream := strings.TrimSpace(webInterface.Upstream)
	if upstream == "" {
		upstream = defaultWebInterfaceUpstream(caddyMode)
	}
	return upstream
}

func newID() string {
	return secureRandomHex(6)
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("hash password: %v", err))
	}
	return string(hash)
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.Lock()
		authEnabled := a.settings.AuthEnabled
		localEnabled := localAuthEnabledFromEnv()
		oidcEnabled := oidcAuthEnabledFromEnv() && a.settings.OIDC.Enabled
		authenticated := a.hasValidSessionLocked(r)
		a.mu.Unlock()

		if !authEnabled || authenticated || isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if !localEnabled && !oidcEnabled {
			writeError(w, http.StatusServiceUnavailable, errors.New("authentication is enabled but no login method is configured"))
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
	return passwordMatchesHash(password, a.settings.PasswordHash)
}

func (a *App) hasValidSessionLocked(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	session, ok := a.sessions[cookie.Value]
	if !ok {
		return false
	}
	if time.Now().After(session.ExpiresAt) {
		delete(a.sessions, cookie.Value)
		return false
	}
	return true
}

func isPublicPath(path string) bool {
	if strings.HasPrefix(path, "/.caddymgm/auth/") {
		return true
	}
	switch path {
	case "/login.html", "/login.css", "/login.js", "/styles.css", "/CaddyMGM.png", "/favicon.ico", "/world-map.svg", "/api/versions":
		return true
	case "/api/auth/login":
		return true
	case "/api/auth/config", "/api/auth/oidc/start", "/api/auth/oidc/callback", "/auth/oidc/callback":
		return true
	default:
		return false
	}
}

func newSessionToken() string {
	return secureRandomHex(32)
}

func secureRandomHex(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("generate cryptographic random value: %v", err))
	}
	return hex.EncodeToString(b)
}

func sessionCookie(value string, maxAge time.Duration, secure bool) *http.Cookie {
	// Secure is enabled for every non-loopback request; local HTTP is an intentional development exception.
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}

func oidcStateCookie(value string, maxAge time.Duration, secure bool) *http.Cookie {
	// Secure is enabled for every non-loopback request; local HTTP is an intentional development exception.
	return &http.Cookie{
		Name:     oidcStateCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
}

func csrfCookie(value string, maxAge time.Duration, secure bool) *http.Cookie {
	// Secure is enabled for every non-loopback request; local HTTP is an intentional development exception.
	return &http.Cookie{
		Name:     csrfCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   secure,
	}
}

func (a *App) createSessionLocked(w http.ResponseWriter, r *http.Request, username, provider string) {
	now := time.Now()
	a.cleanupSessionsLocked(now)
	if len(a.sessions) >= sessionLimit {
		a.evictEarliestSessionLocked()
	}
	token := newSessionToken()
	a.sessions[token] = Session{
		ExpiresAt: now.Add(sessionLifetime),
		Username:  username,
		Provider:  provider,
	}
	http.SetCookie(w, sessionCookie(token, sessionLifetime, a.isSecureRequest(r)))
}

func (a *App) cleanupSessionsLocked(now time.Time) {
	for token, session := range a.sessions {
		if !session.ExpiresAt.After(now) {
			delete(a.sessions, token)
		}
	}
}

func (a *App) evictEarliestSessionLocked() {
	earliestToken := ""
	var earliestExpiry time.Time
	for token, session := range a.sessions {
		if earliestToken == "" || session.ExpiresAt.Before(earliestExpiry) {
			earliestToken = token
			earliestExpiry = session.ExpiresAt
		}
	}
	if earliestToken != "" {
		delete(a.sessions, earliestToken)
	}
}

func (a *App) consumeOIDCStateLocked(r *http.Request, state string) error {
	cookie, err := r.Cookie(oidcStateCookieName)
	if err != nil || cookie.Value == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		return errors.New("state cookie mismatch")
	}
	expiresAt, ok := a.oidcStates[state]
	if !ok {
		return errors.New("state not found")
	}
	delete(a.oidcStates, state)
	if time.Now().After(expiresAt) {
		return errors.New("state expired")
	}
	return nil
}

func (a *App) cleanupOIDCStatesLocked(now time.Time) {
	for state, expiresAt := range a.oidcStates {
		if !expiresAt.After(now) {
			delete(a.oidcStates, state)
		}
	}
}

func (a *App) createOIDCStateLocked(now time.Time) (string, bool) {
	if a.oidcStates == nil {
		a.oidcStates = make(map[string]time.Time)
	}
	a.cleanupOIDCStatesLocked(now)
	if len(a.oidcStates) >= oidcStateLimit {
		return "", false
	}
	state := newSessionToken()
	a.oidcStates[state] = now.Add(oidcStateLifetime)
	return state, true
}

func (a *App) oidcRuntime(ctx context.Context) (*oidcRuntime, error) {
	a.mu.Lock()
	settings := a.settings
	cache := a.oidcCache
	a.mu.Unlock()

	if !oidcAuthEnabledFromEnv() || !settings.OIDC.Enabled {
		return nil, errors.New("oidc authentication is disabled")
	}
	if err := validateOIDCSettings(settings.OIDC); err != nil {
		return nil, err
	}
	cacheKey := strings.Join([]string{
		settings.OIDC.IssuerURL,
		settings.OIDC.ClientID,
		settings.OIDC.RedirectURL,
		settings.OIDC.Scopes,
	}, "|")
	if cached, ok := cache[cacheKey]; ok {
		return cached, nil
	}
	providerLoader := a.oidcProvider
	if providerLoader == nil {
		providerLoader = oidc.NewProvider
	}
	provider, err := providerLoader(ctx, settings.OIDC.IssuerURL)
	if err != nil {
		return nil, err
	}
	scopes := strings.Fields(settings.OIDC.Scopes)
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	runtime := &oidcRuntime{
		Provider: provider,
		Verifier: provider.Verifier(&oidc.Config{ClientID: settings.OIDC.ClientID}),
		Config: oauth2.Config{
			ClientID:     settings.OIDC.ClientID,
			ClientSecret: settings.OIDC.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  settings.OIDC.RedirectURL,
			Scopes:       scopes,
		},
	}

	a.mu.Lock()
	if cached, ok := a.oidcCache[cacheKey]; ok {
		a.mu.Unlock()
		return cached, nil
	}
	a.oidcCache[cacheKey] = runtime
	a.mu.Unlock()
	return runtime, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
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

func decodeJSONBody(w http.ResponseWriter, r *http.Request, destination any, limit int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("request body contains multiple JSON values")
		}
		return err
	}
	return nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func authEnabledFromEnv() bool {
	return envBool("CADDYMGM_AUTH_ENABLED", true)
}

func localAuthEnabledFromEnv() bool {
	return envBool("CADDYMGM_LOCALAUTH_ENABLED", true)
}

func oidcAuthEnabledFromEnv() bool {
	return envBool("CADDYMGM_OIDCAUTH_ENABLED", false)
}

func insecureHTTPAllowedFromEnv() bool {
	return envBool("CADDYMGM_ALLOW_INSECURE_HTTP", false)
}

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(env(key, ""))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "treu", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func passwordMatchesHash(password, stored string) bool {
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}
	sum := sha256.Sum256([]byte(password))
	legacy := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(legacy), []byte(stored)) == 1
}

func passwordNeedsUpgrade(stored string) bool {
	return !(strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$"))
}

func loginAttemptKeys(username, remoteAddr string) []string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.TrimSpace(remoteAddr)
	}
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))
	normalizedHost := strings.ToLower(strings.Trim(host, "[]"))
	return []string{
		hashLoginAttemptKey("username\x00" + normalizedUsername),
		hashLoginAttemptKey("username-source\x00" + normalizedUsername + "\x00" + normalizedHost),
	}
}

func hashLoginAttemptKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (l *loginLimiter) retryAfterAny(keys []string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.cleanupLocked(now)
	var longest time.Duration
	for _, key := range keys {
		attempt, ok := l.attempts[key]
		if !ok || !attempt.lockedUntil.After(now) {
			continue
		}
		attempt.lastSeen = now
		l.attempts[key] = attempt
		if remaining := attempt.lockedUntil.Sub(now); remaining > longest {
			longest = remaining
		}
	}
	return longest, longest > 0
}

func (l *loginLimiter) recordFailure(keys ...string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.cleanupLocked(now)
	var longest time.Duration
	for _, key := range keys {
		attempt := l.attempts[key]
		if attempt.windowStarted.IsZero() || now.Sub(attempt.windowStarted) >= l.failureWindow {
			attempt = loginAttempt{windowStarted: now}
		}
		attempt.failures++
		attempt.lastSeen = now
		if attempt.failures >= l.maxFailures {
			attempt.lockedUntil = now.Add(l.lockoutDuration)
		}
		if _, exists := l.attempts[key]; !exists && len(l.attempts) >= l.maxEntries {
			l.evictOldestLocked()
		}
		l.attempts[key] = attempt
		if remaining := attempt.lockedUntil.Sub(now); remaining > longest {
			longest = remaining
		}
	}
	return longest, longest > 0
}

func (l *loginLimiter) reset(keys ...string) {
	l.mu.Lock()
	for _, key := range keys {
		delete(l.attempts, key)
	}
	l.mu.Unlock()
}

func (l *loginLimiter) cleanupLocked(now time.Time) {
	if !l.nextCleanup.IsZero() && now.Before(l.nextCleanup) {
		return
	}
	for key, attempt := range l.attempts {
		windowExpired := !attempt.windowStarted.IsZero() && now.Sub(attempt.windowStarted) >= l.failureWindow
		lockExpired := attempt.lockedUntil.IsZero() || !attempt.lockedUntil.After(now)
		if windowExpired && lockExpired {
			delete(l.attempts, key)
		}
	}
	l.nextCleanup = now.Add(time.Minute)
}

func (l *loginLimiter) evictOldestLocked() {
	oldestKey := ""
	var oldestTime time.Time
	for key, attempt := range l.attempts {
		if oldestKey == "" || attempt.lastSeen.Before(oldestTime) {
			oldestKey = key
			oldestTime = attempt.lastSeen
		}
	}
	if oldestKey != "" {
		delete(l.attempts, oldestKey)
	}
}

func writeLoginRateLimit(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int64((retryAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.FormatInt(seconds, 10))
	writeError(w, http.StatusTooManyRequests, errors.New("too many login attempts; try again later"))
}

func newTrustedProxySet(value string) (*trustedProxySet, error) {
	set := &trustedProxySet{
		cache:  make(map[string]trustedProxyCacheEntry),
		lookup: net.DefaultResolver.LookupNetIP,
		now:    time.Now,
	}
	hostnamePattern := regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?$`)
	for _, entry := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\n' }) {
		entry = strings.TrimSpace(entry)
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			set.prefixes = append(set.prefixes, prefix.Masked())
			continue
		}
		if address, err := netip.ParseAddr(entry); err == nil {
			set.prefixes = append(set.prefixes, netip.PrefixFrom(address.Unmap(), address.Unmap().BitLen()))
			continue
		}
		if len(entry) > 253 || !hostnamePattern.MatchString(entry) {
			return nil, fmt.Errorf("invalid proxy address or hostname %q", entry)
		}
		set.hostnames = append(set.hostnames, strings.ToLower(entry))
	}
	return set, nil
}

func (s *trustedProxySet) contains(ctx context.Context, remoteAddr string) bool {
	if s == nil {
		return false
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.Trim(strings.TrimSpace(remoteAddr), "[]")
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	address = address.Unmap()
	for _, prefix := range s.prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	for _, hostname := range s.hostnames {
		if s.hostnameContains(ctx, hostname, address) {
			return true
		}
	}
	return false
}

func (s *trustedProxySet) hostnameContains(ctx context.Context, hostname string, address netip.Addr) bool {
	now := s.now()
	s.mu.Lock()
	entry, ok := s.cache[hostname]
	cacheValid := ok && entry.expiresAt.After(now)
	s.mu.Unlock()
	if !cacheValid {
		addresses, err := s.lookup(ctx, "ip", hostname)
		if err != nil {
			s.mu.Lock()
			delete(s.cache, hostname)
			s.mu.Unlock()
			return false
		}
		freshEntry := trustedProxyCacheEntry{addresses: addresses, expiresAt: now.Add(time.Minute)}
		s.mu.Lock()
		current, currentOK := s.cache[hostname]
		if !currentOK || !current.expiresAt.After(now) {
			s.cache[hostname] = freshEntry
			entry = freshEntry
		} else {
			entry = current
		}
		s.mu.Unlock()
	}
	for _, candidate := range entry.addresses {
		if candidate.Unmap() == address {
			return true
		}
	}
	return false
}

func (a *App) secureSessionTransportAllowed(r *http.Request) bool {
	return insecureHTTPAllowedFromEnv() || a.isSecureRequest(r) || isLoopbackRemoteAddr(r.RemoteAddr)
}

func (a *App) isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if a.trustedProxies == nil || !a.trustedProxies.contains(r.Context(), r.RemoteAddr) {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(forwardedProto(r))) {
	case "https", "wss":
		return true
	default:
		return false
	}
}

func forwardedProto(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); value != "" {
		if proto, _, ok := strings.Cut(value, ","); ok {
			return strings.TrimSpace(proto)
		}
		return value
	}
	for _, part := range strings.Split(r.Header.Get("Forwarded"), ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok || !strings.EqualFold(key, "proto") {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"`)
	}
	return ""
}

func (a *App) requestScheme(r *http.Request) string {
	if a.isSecureRequest(r) {
		return "https"
	}
	return "http"
}

func isLoopbackRemoteAddr(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		host = strings.Trim(strings.TrimSpace(remoteAddr), "[]")
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; img-src 'self' data:; object-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		if a.isSecureRequest(r) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) ensureCSRFCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if cookie, err := r.Cookie(csrfCookieName); err == nil && validCSRFToken(cookie.Value) {
			token = cookie.Value
		}
		if token == "" {
			token = newSessionToken()
			http.SetCookie(w, csrfCookie(token, 12*time.Hour, a.isSecureRequest(r)))
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) requireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !csrfProtectionRequired(r.Method) || !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if err := a.validateRequestOrigin(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		cookie, err := r.Cookie(csrfCookieName)
		if err != nil || !validCSRFToken(cookie.Value) {
			writeError(w, http.StatusForbidden, errors.New("missing csrf cookie"))
			return
		}
		token := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
		if !validCSRFToken(token) || subtle.ConstantTimeCompare([]byte(token), []byte(cookie.Value)) != 1 {
			writeError(w, http.StatusForbidden, errors.New("invalid csrf token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func csrfProtectionRequired(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func (a *App) validateRequestOrigin(r *http.Request) error {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	referer := strings.TrimSpace(r.Header.Get("Referer"))
	source := origin
	if source == "" {
		source = referer
	}
	if source == "" {
		return errors.New("missing request origin")
	}
	parsed, err := url.Parse(source)
	if err != nil || parsed.Host == "" {
		return errors.New("invalid request origin")
	}
	if !strings.EqualFold(parsed.Scheme, a.requestScheme(r)) {
		return errors.New("request origin scheme mismatch")
	}
	if !sameHost(parsed.Host, r.Host) {
		return errors.New("request origin host mismatch")
	}
	return nil
}

func sameHost(left, right string) bool {
	return strings.EqualFold(normalizeHost(left), normalizeHost(right))
}

func normalizeHost(value string) string {
	value = strings.TrimSpace(value)
	if host, port, err := net.SplitHostPort(value); err == nil {
		return strings.ToLower(net.JoinHostPort(strings.Trim(host, "[]"), port))
	}
	return strings.ToLower(strings.Trim(value, "[]"))
}

func validCSRFToken(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", sanitizeLogValue(r.Method), sanitizeLogValue(r.URL.Path), time.Since(start).Round(time.Millisecond))
	})
}

func sanitizeLogValue(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '_'
		}
		return r
	}, value)
}
