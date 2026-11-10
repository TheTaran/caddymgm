package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const mcpJSONBodyLimit = 256 << 10

var mcpAllowedScopes = []string{"hosts:read", "hosts:write", "protection:read", "protection:write", "certificates:read", "certificates:write", "logs:read", "settings:read", "settings:write", "reports:read"}

type mcpRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

func (a *App) handleMCP(w http.ResponseWriter, r *http.Request) {
	token, ok := a.authorizeMCP(r)
	if !ok {
		w.Header().Set("WWW-Authenticate", `Bearer realm="CaddyMGM MCP"`)
		writeError(w, http.StatusUnauthorized, errors.New("MCP bearer token required"))
		return
	}
	var request mcpRequest
	if err := decodeJSONBody(w, r, &request, mcpJSONBodyLimit); err != nil || request.JSONRPC != "2.0" || request.Method == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"jsonrpc": "2.0", "error": map[string]any{"code": -32600, "message": "invalid JSON-RPC request"}, "id": nil})
		return
	}
	if request.ID == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	result, err := a.mcpResult(r, request, token)
	a.recordMCPRequest(token.ID, request.Params, err == nil, result)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"jsonrpc": "2.0", "id": request.ID, "error": map[string]any{"code": -32602, "message": err.Error()}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
}

func (a *App) authorizeMCP(r *http.Request) (MCPToken, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(value, "Bearer ") {
		return MCPToken{}, false
	}
	value = strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
	hash := sha256.Sum256([]byte(value))
	for index := range a.settings.MCPTokens {
		token := &a.settings.MCPTokens[index]
		if len(token.TokenHash) == len(fmt.Sprintf("%x", hash)) && subtle.ConstantTimeCompare([]byte(token.TokenHash), []byte(fmt.Sprintf("%x", hash))) == 1 {
			token.LastUsed = time.Now().UTC().Format(time.RFC3339)
			token.Requests++
			_ = a.saveSettingsLocked()
			return *token, true
		}
	}
	return MCPToken{}, false
}

func (a *App) handleCreateMCPToken(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := decodeJSONBody(w, r, &payload, 64<<10); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid MCP token request"))
		return
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		name = "MCP token"
	}
	if len(name) > 80 {
		writeError(w, http.StatusBadRequest, errors.New("MCP token name is too long"))
		return
	}
	scopes, err := normalizeMCPScopes(payload.Scopes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	token := secureRandomHex(32)
	hash := sha256.Sum256([]byte(token))
	a.settings.MCPTokens = append(a.settings.MCPTokens, MCPToken{ID: newID(), Name: name, TokenHash: fmt.Sprintf("%x", hash), Scopes: scopes, CreatedAt: time.Now().UTC().Format(time.RFC3339)})
	if err := a.saveSettingsLocked(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"token": token, "endpoint": "/mcp"})
}

func (a *App) handleDeleteMCPToken(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := strings.TrimPrefix(r.URL.Path, "/api/mcp/tokens/")
	if id == "" || r.URL.Path == "/api/mcp/token" {
		a.settings.MCPTokens = nil
	} else {
		remaining := a.settings.MCPTokens[:0]
		for _, token := range a.settings.MCPTokens {
			if token.ID != id {
				remaining = append(remaining, token)
			}
		}
		a.settings.MCPTokens = remaining
	}
	if err := a.saveSettingsLocked(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleMCPReport(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	requests, successful, bytes := len(a.settings.MCPRequests), 0, 0
	byTool := map[string]int{}
	for _, request := range a.settings.MCPRequests {
		if request.Success {
			successful++
		}
		bytes += request.Bytes
		byTool[request.Tool]++
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": requests, "successful": successful, "failed": requests - successful, "responseBytes": bytes, "byTool": byTool, "recent": a.settings.MCPRequests})
}

func normalizeMCPScopes(scopes []string) ([]string, error) {
	allowed := map[string]bool{}
	for _, scope := range mcpAllowedScopes {
		allowed[scope] = true
	}
	result := make([]string, 0, len(scopes))
	seen := map[string]bool{}
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if !allowed[scope] {
			return nil, errors.New("invalid MCP scope")
		}
		if !seen[scope] {
			seen[scope] = true
			result = append(result, scope)
		}
	}
	if len(result) == 0 {
		return nil, errors.New("select at least one MCP scope")
	}
	return result, nil
}

func (a *App) mcpResult(r *http.Request, request mcpRequest, token MCPToken) (any, error) {
	switch request.Method {
	case "initialize":
		return map[string]any{"protocolVersion": "2025-06-18", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "caddymgm", "version": version}}, nil
	case "tools/list":
		return map[string]any{"tools": []map[string]any{
			{"name": "list_web_hosts", "description": "List managed Caddy web hosts and their enabled state.", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}}},
			{"name": "get_component_versions", "description": "Read current and available CaddyMGM component versions.", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}}},
		}}, nil
	case "tools/call":
		name, _ := request.Params["name"].(string)
		switch name {
		case "list_web_hosts":
			if !mcpScopeAllowed(token, "hosts:read") {
				return nil, errors.New("MCP token lacks hosts:read")
			}
			sites, err := a.readSites()
			if err != nil {
				return nil, err
			}
			return mcpTextResult(sites), nil
		case "get_component_versions":
			if !mcpScopeAllowed(token, "reports:read") {
				return nil, errors.New("MCP token lacks reports:read")
			}
			current := map[string]string{"caddymgm": version, "caddy": a.currentCaddyVersion(), "go": runtime.Version()}
			return mcpTextResult(current), nil
		default:
			return nil, errors.New("unknown MCP tool")
		}
	default:
		return nil, errors.New("unsupported MCP method")
	}
}

func mcpScopeAllowed(token MCPToken, scope string) bool {
	for _, current := range token.Scopes {
		if current == scope {
			return true
		}
	}
	return false
}

func (a *App) recordMCPRequest(tokenID string, params map[string]any, success bool, result any) {
	tool, _ := params["name"].(string)
	content, _ := json.Marshal(result)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.settings.MCPRequests = append(a.settings.MCPRequests, MCPRequestRecord{Time: time.Now().UTC().Format(time.RFC3339), TokenID: tokenID, Tool: tool, Success: success, Bytes: len(content)})
	if len(a.settings.MCPRequests) > 10000 {
		a.settings.MCPRequests = a.settings.MCPRequests[len(a.settings.MCPRequests)-10000:]
	}
	_ = a.saveSettingsLocked()
}

func mcpTextResult(value any) map[string]any {
	content, _ := json.Marshal(value)
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(content)}}}
}
