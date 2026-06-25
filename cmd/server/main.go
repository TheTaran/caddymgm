package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
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
	Enabled         bool   `json:"enabled"`
}

type App struct {
	mu         sync.Mutex
	configPath string
}

func main() {
	configPath := env("CADDY_CONFIG_PATH", "/config/Caddyfile")
	addr := env("CADDYMGM_LISTEN", ":8080")

	app := &App{configPath: configPath}
	if err := app.ensureConfig(); err != nil {
		log.Fatalf("prepare config: %v", err)
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

	log.Printf("caddymgm listening on %s, config=%s", addr, configPath)
	log.Fatal(http.ListenAndServe(addr, logRequest(mux)))
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
	var site Site
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	site.ID = newID()
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
	writeJSON(w, http.StatusCreated, site)
}

func (a *App) handleUpdateSite(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	if id == "" {
		writeError(w, http.StatusBadRequest, errors.New("missing site id"))
		return
	}

	var updated Site
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	updated.ID = id
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
	for _, site := range sites {
		if site.ID == id {
			found = true
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

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
