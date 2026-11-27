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

func TestMCPToolsListHonorsTokenScopes(t *testing.T) {
	readOnly := MCPToken{Scopes: []string{"hosts:read", "reports:read"}}
	tools := mcpToolsForToken(readOnly)
	seen := map[string]bool{}
	for _, tool := range tools {
		seen[tool.Name] = true
	}
	if !seen["list_web_hosts"] || !seen["get_component_versions"] || !seen["get_security_report"] || !seen["get_geo_access_report"] || !seen["get_mcp_usage_report"] {
		t.Fatalf("expected permitted tools, got %#v", seen)
	}
	for _, name := range []string{"create_web_host", "get_settings", "get_access_logs", "list_certificates"} {
		if seen[name] {
			t.Fatalf("unexpected tool %q for a read-only token", name)
		}
	}
}

func TestMCPReportToolsReturnDashboardData(t *testing.T) {
	app := &App{demoData: true}
	token := MCPToken{Scopes: []string{"reports:read"}}
	request := httptest.NewRequest(http.MethodPost, "/mcp", nil)

	for name, arguments := range map[string]map[string]any{
		"get_security_report":   {"period": "1h", "eventLimit": float64(3)},
		"get_geo_access_report": {"limit": float64(2)},
		"get_mcp_usage_report":  {},
	} {
		result, err := app.callMCPTool(request, token, name, arguments)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		content, ok := result.(map[string]any)["content"].([]map[string]string)
		if !ok || len(content) != 1 || !json.Valid([]byte(content[0]["text"])) {
			t.Fatalf("%s returned invalid MCP content: %#v", name, result)
		}
	}
}

func TestMCPWriteToolRequiresWriteScope(t *testing.T) {
	app := &App{}
	_, err := app.callMCPTool(httptest.NewRequest(http.MethodPost, "/mcp", nil), MCPToken{Scopes: []string{"hosts:read"}}, "delete_web_host", map[string]any{"id": "example"})
	if err == nil || !strings.Contains(err.Error(), "hosts:write") {
		t.Fatalf("error = %v, want missing hosts:write", err)
	}
}
