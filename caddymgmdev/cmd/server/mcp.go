package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sort"
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

// handleUpdateMCPToken changes only the metadata and permissions of an
// existing token. Its ID and secret hash intentionally remain untouched, so
// connected agents can continue using the same bearer token immediately.
func (a *App) handleUpdateMCPToken(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/mcp/tokens/"))
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, errors.New("MCP token ID is required"))
		return
	}
	var payload struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := decodeJSONBody(w, r, &payload, 64<<10); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid MCP token request"))
		return
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" || len(name) > 80 {
		writeError(w, http.StatusBadRequest, errors.New("MCP token name must be between 1 and 80 characters"))
		return
	}
	scopes, err := normalizeMCPScopes(payload.Scopes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for index := range a.settings.MCPTokens {
		token := &a.settings.MCPTokens[index]
		if token.ID != id {
			continue
		}
		token.Name = name
		token.Scopes = scopes
		if err := a.saveSettingsLocked(); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		updated := *token
		updated.TokenHash = ""
		writeJSON(w, http.StatusOK, updated)
		return
	}
	writeError(w, http.StatusNotFound, errors.New("MCP token not found"))
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
		return map[string]any{"tools": mcpToolsForToken(token)}, nil
	case "tools/call":
		name, _ := request.Params["name"].(string)
		arguments, err := mcpArguments(request.Params)
		if err != nil {
			return nil, err
		}
		return a.callMCPTool(r, token, name, arguments)
	default:
		return nil, errors.New("unsupported MCP method")
	}
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	Scope       string         `json:"-"`
}

func mcpToolsForToken(token MCPToken) []mcpTool {
	object := func(properties map[string]any) map[string]any {
		return map[string]any{"type": "object", "properties": properties}
	}
	stringField := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}
	tools := []mcpTool{
		{"list_web_hosts", "List managed Caddy web hosts and their enabled state.", object(map[string]any{}), "hosts:read"},
		{"get_web_host", "Read one managed web host by its ID.", object(map[string]any{"id": stringField("Web host ID.")}), "hosts:read"},
		{"create_web_host", "Create a web host. The change is staged and must be applied separately in CaddyMGM.", object(map[string]any{"site": map[string]any{"type": "object"}}), "hosts:write"},
		{"update_web_host", "Replace a web host configuration. The change is staged and must be applied separately in CaddyMGM.", object(map[string]any{"id": stringField("Web host ID."), "site": map[string]any{"type": "object"}}), "hosts:write"},
		{"set_web_host_enabled", "Enable or disable a web host. The change is staged and must be applied separately in CaddyMGM.", object(map[string]any{"id": stringField("Web host ID."), "enabled": map[string]any{"type": "boolean"}}), "hosts:write"},
		{"delete_web_host", "Delete a web host. The change is staged and must be applied separately in CaddyMGM.", object(map[string]any{"id": stringField("Web host ID.")}), "hosts:write"},
		{"get_web_protection", "Read the global web protection policy and manual IP lists.", object(map[string]any{}), "protection:read"},
		{"update_global_web_protection", "Update the global country protection policy. IP lists remain unchanged. The change is staged.", object(map[string]any{"webProtection": map[string]any{"type": "object"}}), "protection:write"},
		{"update_manual_ip_lists", "Replace global manual allow/block IP and DNS lists. The change is staged.", object(map[string]any{"manualIpLists": map[string]any{"type": "array"}}), "protection:write"},
		{"list_certificates", "List configured ACME authorities and managed certificate expiry information.", object(map[string]any{}), "certificates:read"},
		{"replace_acme_authorities", "Replace configured ACME authorities. The change is staged and must be applied separately in CaddyMGM.", object(map[string]any{"authorities": map[string]any{"type": "array"}}), "certificates:write"},
		{"get_access_logs", "Read recent access logs for a web host. Sensitive request headers are not included.", object(map[string]any{"siteId": stringField("Web host ID."), "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 200}}), "logs:read"},
		{"get_settings", "Read non-secret CaddyMGM settings.", object(map[string]any{}), "settings:read"},
		{"update_log_settings", "Update log retention. This does not reload Caddy.", object(map[string]any{"logRetention": map[string]any{"type": "integer", "minimum": 25}}), "settings:write"},
		{"update_client_ip_settings", "Update visitor IP source and forwarding settings. The change is staged.", object(map[string]any{"clientIp": map[string]any{"type": "object"}}), "settings:write"},
		{"get_component_versions", "Read current CaddyMGM component versions.", object(map[string]any{}), "reports:read"},
		{"get_pending_changes", "Read all changes that still need applying to Caddy.", object(map[string]any{}), "reports:read"},
		{"get_security_report", "Create a security report from retained access logs, including request and protection-block trends.", object(map[string]any{"period": map[string]any{"type": "string", "enum": []string{"1h", "6h", "1d", "7d", "30d"}, "description": "Report period; defaults to 1d."}, "siteId": stringField("Optional web host ID."), "eventLimit": map[string]any{"type": "integer", "minimum": 0, "maximum": 100, "description": "Maximum managed protection events; defaults to 20."}}), "reports:read"},
		{"get_geo_access_report", "Create a geographic access report with the most active source IPs and locations from retained logs.", object(map[string]any{"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 200, "description": "Maximum locations and source IPs; defaults to 25."}}), "reports:read"},
		{"get_mcp_usage_report", "Read MCP request totals, success/failure counts, response bytes, and recent tool usage.", object(map[string]any{}), "reports:read"},
	}
	result := make([]mcpTool, 0, len(tools))
	for _, tool := range tools {
		if mcpScopeAllowed(token, tool.Scope) {
			result = append(result, tool)
		}
	}
	return result
}

func mcpArguments(params map[string]any) (map[string]any, error) {
	if params == nil {
		return map[string]any{}, nil
	}
	if value, ok := params["arguments"]; ok {
		arguments, ok := value.(map[string]any)
		if !ok {
			return nil, errors.New("tool arguments must be an object")
		}
		return arguments, nil
	}
	return map[string]any{}, nil
}

func mcpRequire(token MCPToken, scope string) error {
	if !mcpScopeAllowed(token, scope) {
		return fmt.Errorf("MCP token lacks %s", scope)
	}
	return nil
}

func mcpDecodeArguments(arguments map[string]any, target any) error {
	content, err := json.Marshal(arguments)
	if err != nil {
		return errors.New("invalid tool arguments")
	}
	if err := json.Unmarshal(content, target); err != nil {
		return errors.New("invalid tool arguments")
	}
	return nil
}

func (a *App) callMCPTool(r *http.Request, token MCPToken, name string, arguments map[string]any) (any, error) {
	switch name {
	case "list_web_hosts":
		if err := mcpRequire(token, "hosts:read"); err != nil {
			return nil, err
		}
		sites, err := a.readSites()
		if err != nil {
			return nil, err
		}
		a.populateCertificateMetadata(sites)
		return mcpTextResult(sites), nil
	case "get_web_host":
		if err := mcpRequire(token, "hosts:read"); err != nil {
			return nil, err
		}
		id, err := mcpRequiredID(arguments)
		if err != nil {
			return nil, err
		}
		site, err := a.mcpSiteByID(id)
		if err != nil {
			return nil, err
		}
		a.populateSiteCertificateMetadata(&site)
		return mcpTextResult(site), nil
	case "create_web_host":
		if err := mcpRequire(token, "hosts:write"); err != nil {
			return nil, err
		}
		return a.mcpCreateSite(arguments)
	case "update_web_host":
		if err := mcpRequire(token, "hosts:write"); err != nil {
			return nil, err
		}
		return a.mcpUpdateSite(arguments)
	case "set_web_host_enabled":
		if err := mcpRequire(token, "hosts:write"); err != nil {
			return nil, err
		}
		return a.mcpSetSiteEnabled(arguments)
	case "delete_web_host":
		if err := mcpRequire(token, "hosts:write"); err != nil {
			return nil, err
		}
		return a.mcpDeleteSite(arguments)
	case "get_web_protection":
		if err := mcpRequire(token, "protection:read"); err != nil {
			return nil, err
		}
		a.mu.Lock()
		settings := a.publicSettingsLocked()
		a.mu.Unlock()
		return mcpTextResult(map[string]any{"webProtection": settings.WebProtection, "manualIpLists": settings.ManualIPLists, "externalBlocklists": settings.ExternalBlocklists, "externalBlockedIpCount": settings.ExternalBlockedIPCount}), nil
	case "update_global_web_protection":
		if err := mcpRequire(token, "protection:write"); err != nil {
			return nil, err
		}
		var payload struct {
			WebProtection WebProtection `json:"webProtection"`
		}
		if err := mcpDecodeArguments(arguments, &payload); err != nil {
			return nil, err
		}
		if err := normalizeWebProtection(&payload.WebProtection); err != nil {
			return nil, err
		}
		return a.mcpUpdateSettings(func(next *Settings) error {
			payload.WebProtection.BlockedIPs = manualIPListEntries(next.ManualIPLists, "block")
			payload.WebProtection.AllowedIPs = manualIPListEntries(next.ManualIPLists, "allow")
			next.WebProtection = payload.WebProtection
			return nil
		})
	case "update_manual_ip_lists":
		if err := mcpRequire(token, "protection:write"); err != nil {
			return nil, err
		}
		var payload struct {
			Lists ManualIPLists `json:"manualIpLists"`
		}
		if err := mcpDecodeArguments(arguments, &payload); err != nil {
			return nil, err
		}
		return a.mcpUpdateSettings(func(next *Settings) error {
			lists, err := normalizeManualIPLists(payload.Lists)
			if err != nil {
				return err
			}
			lists, _, err = resolveManualAllowlistDNS(r.Context(), lists, next.ManualIPLists)
			if err != nil {
				return err
			}
			next.ManualIPLists = lists
			next.WebProtection.BlockedIPs = manualIPListEntries(lists, "block")
			next.WebProtection.AllowedIPs = manualIPListEntries(lists, "allow")
			return nil
		})
	case "list_certificates":
		if err := mcpRequire(token, "certificates:read"); err != nil {
			return nil, err
		}
		sites, err := a.readSites()
		if err != nil {
			return nil, err
		}
		a.populateCertificateMetadata(sites)
		a.mu.Lock()
		issuers := append([]ACMEIssuer(nil), a.publicSettingsLocked().ACMEIssuers...)
		a.mu.Unlock()
		return mcpTextResult(map[string]any{"authorities": issuers, "certificates": sites}), nil
	case "replace_acme_authorities":
		if err := mcpRequire(token, "certificates:write"); err != nil {
			return nil, err
		}
		var payload struct {
			Authorities []ACMEIssuer `json:"authorities"`
		}
		if err := mcpDecodeArguments(arguments, &payload); err != nil {
			return nil, err
		}
		issuers, err := normalizeACMEIssuers(payload.Authorities)
		if err != nil {
			return nil, err
		}
		return a.mcpUpdateSettings(func(next *Settings) error {
			next.ACMEIssuers = issuers
			return validateWebInterface(next.WebInterface, issuers)
		})
	case "get_access_logs":
		if err := mcpRequire(token, "logs:read"); err != nil {
			return nil, err
		}
		return a.mcpAccessLogs(arguments)
	case "get_settings":
		if err := mcpRequire(token, "settings:read"); err != nil {
			return nil, err
		}
		a.mu.Lock()
		settings := a.publicSettingsLocked()
		a.mu.Unlock()
		return mcpTextResult(mcpSafeSettings(settings)), nil
	case "update_log_settings":
		if err := mcpRequire(token, "settings:write"); err != nil {
			return nil, err
		}
		var payload struct {
			LogRetention int `json:"logRetention"`
		}
		if err := mcpDecodeArguments(arguments, &payload); err != nil {
			return nil, err
		}
		if payload.LogRetention < 25 {
			return nil, errors.New("log retention must be at least 25")
		}
		return a.mcpUpdateSettingsOnly(func(next *Settings) error { next.LogRetention = payload.LogRetention; return nil })
	case "update_client_ip_settings":
		if err := mcpRequire(token, "settings:write"); err != nil {
			return nil, err
		}
		var payload struct {
			ClientIP ClientIPSettings `json:"clientIp"`
		}
		if err := mcpDecodeArguments(arguments, &payload); err != nil {
			return nil, err
		}
		if err := normalizeClientIPSettings(&payload.ClientIP); err != nil {
			return nil, err
		}
		return a.mcpUpdateSettings(func(next *Settings) error { next.ClientIP = payload.ClientIP; return nil })
	case "get_component_versions":
		if err := mcpRequire(token, "reports:read"); err != nil {
			return nil, err
		}
		return mcpTextResult(map[string]string{"caddymgm": version, "caddy": a.currentCaddyVersion(), "go": runtime.Version()}), nil
	case "get_pending_changes":
		if err := mcpRequire(token, "reports:read"); err != nil {
			return nil, err
		}
		a.mu.Lock()
		changes, err := a.pendingConfigSummaryLocked()
		a.mu.Unlock()
		if err != nil {
			return nil, err
		}
		return mcpTextResult(map[string]any{"pending": len(changes) > 0, "changes": changes}), nil
	case "get_security_report":
		if err := mcpRequire(token, "reports:read"); err != nil {
			return nil, err
		}
		return a.mcpSecurityReport(arguments)
	case "get_geo_access_report":
		if err := mcpRequire(token, "reports:read"); err != nil {
			return nil, err
		}
		return a.mcpGeoAccessReport(arguments)
	case "get_mcp_usage_report":
		if err := mcpRequire(token, "reports:read"); err != nil {
			return nil, err
		}
		return mcpTextResult(a.mcpUsageReport()), nil
	default:
		return nil, errors.New("unknown MCP tool")
	}
}

func (a *App) mcpSecurityReport(arguments map[string]any) (any, error) {
	period, _ := arguments["period"].(string)
	siteID, _ := arguments["siteId"].(string)
	eventLimit, err := mcpReportLimit(arguments, "eventLimit", 20, 100, true)
	if err != nil {
		return nil, err
	}
	overview, err := a.securityOverview(period, strings.TrimSpace(siteID), true)
	if err != nil {
		return nil, err
	}
	if eventLimit < len(overview.Events) {
		overview.Events = overview.Events[:eventLimit]
	}
	return mcpTextResult(map[string]any{
		"report":           "security",
		"period":           normalizedSecurityPeriod(period),
		"siteId":           strings.TrimSpace(siteID),
		"generatedAt":      time.Now().UTC().Format(time.RFC3339),
		"securityOverview": overview,
	}), nil
}

func (a *App) mcpGeoAccessReport(arguments map[string]any) (any, error) {
	limit, err := mcpReportLimit(arguments, "limit", 25, 200, false)
	if err != nil {
		return nil, err
	}
	report, err := a.geoMapReport()
	if err != nil {
		return nil, err
	}
	if len(report.TopIPs) > limit {
		report.TopIPs = report.TopIPs[:limit]
	}
	if len(report.Locations) > limit {
		report.Locations = report.Locations[:limit]
	}
	for index := range report.Locations {
		if len(report.Locations[index].IPs) > 10 {
			report.Locations[index].IPs = report.Locations[index].IPs[:10]
		}
	}
	return mcpTextResult(map[string]any{"report": "geo_access", "generatedAt": time.Now().UTC().Format(time.RFC3339), "geoAccess": report}), nil
}

func mcpReportLimit(arguments map[string]any, name string, fallback, maximum int, allowZero bool) (int, error) {
	raw, found := arguments[name]
	if !found {
		return fallback, nil
	}
	value, ok := raw.(float64)
	if !ok || value != float64(int(value)) {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	limit := int(value)
	minimum := 1
	if allowZero {
		minimum = 0
	}
	if limit < minimum || limit > maximum {
		return 0, fmt.Errorf("%s must be between %d and %d", name, minimum, maximum)
	}
	return limit, nil
}

func normalizedSecurityPeriod(period string) string {
	if strings.TrimSpace(period) == "" {
		return "1d"
	}
	return period
}

func (a *App) mcpUsageReport() map[string]any {
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
	recent := append([]MCPRequestRecord(nil), a.settings.MCPRequests...)
	return map[string]any{"requests": requests, "successful": successful, "failed": requests - successful, "responseBytes": bytes, "byTool": byTool, "recent": recent}
}

func mcpRequiredID(arguments map[string]any) (string, error) {
	id, _ := arguments["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		return "", errors.New("id is required")
	}
	return id, nil
}

func (a *App) mcpSiteByID(id string) (Site, error) {
	sites, err := a.readSites()
	if err != nil {
		return Site{}, err
	}
	for _, site := range sites {
		if site.ID == id {
			return site, nil
		}
	}
	return Site{}, errors.New("web host not found")
}

func (a *App) mcpCreateSite(arguments map[string]any) (any, error) {
	var payload struct {
		Site sitePayload `json:"site"`
	}
	if err := mcpDecodeArguments(arguments, &payload); err != nil {
		return nil, err
	}
	site := payload.Site.site(newID(), true)
	if err := normalizeSite(&site); err != nil {
		return nil, err
	}
	if err := prepareBasicAuth(&site, nil); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, head, tail, err := a.load()
	if err != nil {
		return nil, err
	}
	if err := a.validateSiteLocked(site); err != nil {
		return nil, err
	}
	if duplicateSiteAddress(site.Address, sites, "") {
		return nil, errors.New("a web host with this address already exists")
	}
	site.ID = uniqueSiteID(site.Address, sites, "")
	sites = append(sites, site)
	if err := a.stageCaddyConfigLocked(head, sites, tail); err != nil {
		return nil, err
	}
	a.addLogLocked(site, "created via MCP", "Web host created; pending apply")
	a.populateSiteCertificateMetadata(&site)
	return mcpTextResult(map[string]any{"site": site, "pending": true}), nil
}

func (a *App) mcpUpdateSite(arguments map[string]any) (any, error) {
	id, err := mcpRequiredID(arguments)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Site sitePayload `json:"site"`
	}
	if err := mcpDecodeArguments(arguments, &payload); err != nil {
		return nil, err
	}
	updated := payload.Site.site(id, false)
	if err := normalizeSite(&updated); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, head, tail, err := a.load()
	if err != nil {
		return nil, err
	}
	var existing *Site
	for i := range sites {
		if sites[i].ID == id {
			existing = &sites[i]
			break
		}
	}
	if existing == nil {
		return nil, errors.New("web host not found")
	}
	if err := prepareBasicAuth(&updated, existing); err != nil {
		return nil, err
	}
	if err := a.validateSiteLocked(updated); err != nil {
		return nil, err
	}
	if duplicateSiteAddress(updated.Address, sites, id) {
		return nil, errors.New("a web host with this address already exists")
	}
	for i := range sites {
		if sites[i].ID == id {
			updated.ID = uniqueSiteID(updated.Address, sites, id)
			sites[i] = updated
			break
		}
	}
	if err := a.stageCaddyConfigLocked(head, sites, tail); err != nil {
		return nil, err
	}
	a.addLogLocked(updated, "updated via MCP", "Web host updated; pending apply")
	a.populateSiteCertificateMetadata(&updated)
	return mcpTextResult(map[string]any{"site": updated, "pending": true}), nil
}

func (a *App) mcpSetSiteEnabled(arguments map[string]any) (any, error) {
	id, err := mcpRequiredID(arguments)
	if err != nil {
		return nil, err
	}
	enabled, ok := arguments["enabled"].(bool)
	if !ok {
		return nil, errors.New("enabled must be a boolean")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, head, tail, err := a.load()
	if err != nil {
		return nil, err
	}
	var changed *Site
	for i := range sites {
		if sites[i].ID == id {
			sites[i].Enabled = enabled
			changed = &sites[i]
			break
		}
	}
	if changed == nil {
		return nil, errors.New("web host not found")
	}
	if err := a.stageCaddyConfigLocked(head, sites, tail); err != nil {
		return nil, err
	}
	a.addLogLocked(*changed, "enabled via MCP", "Web host state changed; pending apply")
	return mcpTextResult(map[string]any{"site": *changed, "pending": true}), nil
}

func (a *App) mcpDeleteSite(arguments map[string]any) (any, error) {
	id, err := mcpRequiredID(arguments)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, head, tail, err := a.load()
	if err != nil {
		return nil, err
	}
	next := make([]Site, 0, len(sites)-1)
	var removed *Site
	for i := range sites {
		if sites[i].ID == id {
			copy := sites[i]
			removed = &copy
			continue
		}
		next = append(next, sites[i])
	}
	if removed == nil {
		return nil, errors.New("web host not found")
	}
	if err := a.stageCaddyConfigLocked(head, next, tail); err != nil {
		return nil, err
	}
	a.addLogLocked(*removed, "deleted via MCP", "Web host deleted; pending apply")
	return mcpTextResult(map[string]any{"deleted": id, "pending": true}), nil
}

// mcpUpdateSettings changes settings that affect the generated Caddyfile. It
// deliberately stages, rather than loads, the new configuration.
func (a *App) mcpUpdateSettings(mutate func(*Settings) error) (any, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	next := a.settings
	if err := mutate(&next); err != nil {
		return nil, err
	}
	sites, head, tail, err := a.load()
	if err != nil {
		return nil, err
	}
	previous := a.settings
	previousFile, err := json.MarshalIndent(previous, "", "  ")
	if err != nil {
		return nil, err
	}
	a.settings = next
	if err := a.saveSettingsLocked(); err != nil {
		a.settings = previous
		return nil, err
	}
	if err := a.stageCaddyConfigLocked(head, sites, tail); err != nil {
		a.settings = previous
		_ = writeFileAtomically(a.settingsPath, append(previousFile, '\n'), 0o600)
		return nil, fmt.Errorf("unable to stage settings change: %w", err)
	}
	return mcpTextResult(map[string]any{"pending": true, "settings": mcpSafeSettings(a.publicSettingsLocked())}), nil
}

// mcpUpdateSettingsOnly is for settings that do not change Caddy's runtime
// configuration and must therefore not create an unnecessary pending reload.
func (a *App) mcpUpdateSettingsOnly(mutate func(*Settings) error) (any, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	next := a.settings
	if err := mutate(&next); err != nil {
		return nil, err
	}
	previous := a.settings
	a.settings = next
	if err := a.saveSettingsLocked(); err != nil {
		a.settings = previous
		return nil, err
	}
	a.trimLogsLocked()
	return mcpTextResult(map[string]any{"pending": a.hasPendingConfigLocked(), "settings": mcpSafeSettings(a.publicSettingsLocked())}), nil
}

func (a *App) mcpAccessLogs(arguments map[string]any) (any, error) {
	siteID, err := mcpRequiredID(map[string]any{"id": arguments["siteId"]})
	if err != nil {
		return nil, errors.New("siteId is required")
	}
	limit := 50
	if raw, ok := arguments["limit"].(float64); ok {
		limit = int(raw)
	}
	if limit < 1 || limit > 200 {
		return nil, errors.New("limit must be between 1 and 200")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	sites, _, _, err := a.load()
	if err != nil {
		return nil, err
	}
	found := false
	for _, site := range sites {
		if site.ID == siteID {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("web host not found")
	}
	entries := a.readAccessLogsLocked(siteID, sites)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Time > entries[j].Time })
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return mcpTextResult(map[string]any{"logs": entries}), nil
}

func mcpSafeSettings(settings Settings) map[string]any {
	return map[string]any{
		"appName": settings.AppName, "logRetention": settings.LogRetention,
		"hideLocalDashboardIps": settings.HideLocalDashboardIPs, "clientIp": settings.ClientIP,
		"caddyMode": settings.CaddyMode, "webInterface": settings.WebInterface,
		"pendingChanges": settings.PendingChanges,
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
