package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const socketMode = 0o666

var releaseVersionPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+(?:\.[0-9]+)?$`)

type updater struct {
	mu         sync.Mutex
	projectDir string
	compose    string
	running    bool
}

func main() {
	socketPath := env("CADDYMGM_UPDATER_SOCKET", "/run/caddymgm-updater/updater.sock")
	projectDir := env("CADDYMGM_PROJECT_DIR", "/project")
	composeFile := env("CADDYMGM_COMPOSE_FILE", filepath.Join(projectDir, "compose.yml"))

	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		log.Fatalf("create socket directory: %v", err)
	}
	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("remove old socket: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("listen on updater socket: %v", err)
	}
	if err := os.Chmod(socketPath, socketMode); err != nil {
		listener.Close()
		log.Fatalf("set updater socket permissions: %v", err)
	}
	defer os.Remove(socketPath)

	app := &updater{projectDir: projectDir, compose: composeFile}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/", app.handleUpdate)
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	log.Printf("caddymgm updater listening on %s, project=%s", socketPath, projectDir)
	log.Fatal(server.Serve(listener))
}

func (u *updater) handleUpdate(w http.ResponseWriter, r *http.Request) {
	service := strings.TrimPrefix(r.URL.Path, "/update/")
	if service != "caddy" && service != "caddymgm" {
		http.Error(w, "unsupported service", http.StatusBadRequest)
		return
	}

	version := strings.TrimSpace(r.URL.Query().Get("version"))
	if service == "caddymgm" && !releaseVersionPattern.MatchString(version) {
		http.Error(w, "valid release version is required", http.StatusBadRequest)
		return
	}

	u.mu.Lock()
	if u.running {
		u.mu.Unlock()
		http.Error(w, "an update is already running", http.StatusConflict)
		return
	}
	u.running = true
	u.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "started", "service": service})

	go func() {
		defer func() {
			u.mu.Lock()
			u.running = false
			u.mu.Unlock()
		}()
		if err := u.runUpdate(service, version); err != nil {
			log.Printf("update %s failed: %v", service, err)
			return
		}
		log.Printf("update %s completed", service)
	}()
}

func (u *updater) runUpdate(service, version string) error {
	if service == "caddymgm" {
		tag := strings.TrimPrefix(version, "v")
		sourceImage := "ghcr.io/thetaran/caddymgm:" + tag
		if err := u.runDocker("pull", sourceImage); err != nil {
			return err
		}
		if err := u.runDocker("image", "tag", sourceImage, "ghcr.io/thetaran/caddymgm:latest"); err != nil {
			return err
		}
	} else if err := u.runCompose("pull", service); err != nil {
		return err
	}
	return u.runCompose("up", "-d", "--no-deps", "--no-build", "--force-recreate", service)
}

func (u *updater) runDocker(arguments ...string) error {
	command := exec.Command("docker", arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func (u *updater) runCompose(arguments ...string) error {
	args := append([]string{"compose", "--project-directory", u.projectDir, "-f", u.compose}, arguments...)
	command := exec.Command("docker", args...)
	command.Dir = u.projectDir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
