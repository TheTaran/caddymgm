package main

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

func TestHandleUploadRootCARejectsOversizedMultipartBody(t *testing.T) {
	tempDir := t.TempDir()
	app := &App{caCertDir: tempDir}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("certificate", "root-ca.pem")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	oversized := strings.Repeat("A", rootCAUploadLimit+1)
	if _, err := io.WriteString(part, oversized); err != nil {
		t.Fatalf("write multipart body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/certificates/root-ca", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()

	app.handleUploadRootCA(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	matches, err := filepath.Glob(filepath.Join(tempDir, "*"))
	if err != nil {
		t.Fatalf("glob uploaded files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("unexpected files written for oversized upload: %v", matches)
	}
}

func TestOIDCRuntimeLookupDoesNotHoldAppMutex(t *testing.T) {
	lookupStarted := make(chan struct{})
	releaseLookup := make(chan struct{})
	app := &App{
		settings: Settings{
			OIDC: OIDCSettings{
				Enabled:      true,
				IssuerURL:    "https://issuer.example.com",
				ClientID:     "client-id",
				ClientSecret: "secret",
				RedirectURL:  "https://app.example.com/auth/oidc/callback",
				Scopes:       "openid profile email",
			},
		},
		oidcCache: make(map[string]*oidcRuntime),
		oidcProvider: func(ctx context.Context, issuer string) (*oidc.Provider, error) {
			close(lookupStarted)
			<-releaseLookup
			return nil, context.DeadlineExceeded
		},
	}
	t.Setenv("CADDYMGM_OIDCAUTH_ENABLED", "true")

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = app.oidcRuntime(context.Background())
	}()

	select {
	case <-lookupStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("OIDC provider lookup did not start")
	}

	locked := make(chan struct{})
	go func() {
		app.mu.Lock()
		app.mu.Unlock()
		close(locked)
	}()

	select {
	case <-locked:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("app mutex remained locked during OIDC provider lookup")
	}

	close(releaseLookup)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("OIDC runtime lookup did not finish")
	}
}
