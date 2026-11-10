package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPTokenHashPersistsInSettings(t *testing.T) {
	settings := Settings{MCPTokens: []MCPToken{{ID: "test", TokenHash: "persisted-hash"}}}
	encoded, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}
	var restored Settings
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}
	if len(restored.MCPTokens) != 1 || restored.MCPTokens[0].TokenHash != "persisted-hash" {
		t.Fatalf("MCP token hash was not persisted: %#v", restored.MCPTokens)
	}
}

func TestMCPRequiresBearerToken(t *testing.T) {
	hash := sha256.Sum256([]byte("test-token"))
	app := &App{settings: Settings{MCPTokens: []MCPToken{{ID: "test", TokenHash: fmt.Sprintf("%x", hash), Scopes: []string{"hosts:read", "reports:read"}}}}}
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	res := httptest.NewRecorder()
	app.handleMCP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestMCPInitializeAndToolsList(t *testing.T) {
	hash := sha256.Sum256([]byte("test-token"))
	app := &App{settings: Settings{MCPTokens: []MCPToken{{ID: "test", TokenHash: fmt.Sprintf("%x", hash), Scopes: []string{"hosts:read", "reports:read"}}}}}
	for _, body := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer test-token")
		res := httptest.NewRecorder()
		app.handleMCP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", res.Code, http.StatusOK, res.Body.String())
		}
	}
}
