package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	accessCookieName      = "caddymgm_access"
	accessStateCookieName = "caddymgm_access_state"
)

type AuthProvidersConfig struct {
	OIDC AccessOIDCProvider `json:"oidc"`
}

type AccessOIDCProvider struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	IssuerURL    string `json:"issuerUrl"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret,omitempty"`
	Scopes       string `json:"scopes"`
	GatewayURL   string `json:"gatewayUrl"`
	ACMEIssuerID string `json:"acmeIssuerId,omitempty"`
}

type AccessSession struct {
	ExpiresAt time.Time
	SiteID    string
	Username  string
	Email     string
	LoginID   string
}

type AccessState struct {
	ExpiresAt   time.Time
	SiteID      string
	ReturnPath  string
	RedirectURL string
}

type AccessTicket struct {
	ExpiresAt  time.Time
	SiteID     string
	ReturnPath string
	Username   string
	Email      string
	LoginID    string
}

func (a *App) ensureAuthProviders() error {
	content, err := os.ReadFile(a.authProvidersPath)
	if errors.Is(err, os.ErrNotExist) {
		a.authProviders = AuthProvidersConfig{OIDC: AccessOIDCProvider{ID: "oidc", Name: "OIDC", Scopes: "openid profile email"}}
		return nil
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, &a.authProviders); err != nil {
		return err
	}
	a.normalizeAuthProviderLocked()
	return nil
}

func (a *App) normalizeAuthProviderLocked() {
	p := &a.authProviders.OIDC
	p.ID = "oidc"
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		p.Name = "OIDC"
	}
	p.IssuerURL = strings.TrimRight(strings.TrimSpace(p.IssuerURL), "/")
	p.ClientID = strings.TrimSpace(p.ClientID)
	p.GatewayURL = strings.TrimRight(strings.TrimSpace(p.GatewayURL), "/")
	p.ACMEIssuerID = strings.TrimSpace(p.ACMEIssuerID)
	p.Scopes = strings.Join(strings.Fields(p.Scopes), " ")
	if p.Scopes == "" {
		p.Scopes = "openid profile email"
	}
}

func validateAccessProvider(p AccessOIDCProvider) error {
	if !p.Enabled {
		return nil
	}
	if p.IssuerURL == "" || p.ClientID == "" || p.ClientSecret == "" || p.GatewayURL == "" || p.ACMEIssuerID == "" {
		return errors.New("enabled access provider requires issuer URL, client ID, client secret, SSO base URL and ACME authority")
	}
	issuer, err := url.Parse(p.IssuerURL)
	if err != nil || issuer.Scheme != "https" || issuer.Host == "" || issuer.User != nil || issuer.RawQuery != "" || issuer.Fragment != "" {
		return errors.New("access provider issuer URL must be an HTTPS origin")
	}
	gateway, err := url.Parse(p.GatewayURL)
	if err != nil || gateway.Scheme != "https" || gateway.Host == "" || gateway.User != nil || gateway.RawQuery != "" || gateway.Fragment != "" || (gateway.Path != "" && gateway.Path != "/") {
		return errors.New("SSO base URL must be an HTTPS origin without a path")
	}
	return nil
}

func (a *App) saveAuthProvidersLocked() error {
	if !a.authProviders.OIDC.Enabled {
		err := os.Remove(a.authProvidersPath)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.authProvidersPath), 0o750); err != nil {
		return err
	}
	content, err := json.MarshalIndent(a.authProviders, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomically(a.authProvidersPath, append(content, '\n'), 0o600)
}

func (a *App) publicAuthProvidersLocked() AuthProvidersConfig {
	result := a.authProviders
	result.OIDC.ClientSecret = ""
	return result
}

func (a *App) handleGetAuthProviders(w http.ResponseWriter, _ *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	writeJSON(w, http.StatusOK, a.publicAuthProvidersLocked())
}

func (a *App) handleUpdateAuthProviders(w http.ResponseWriter, r *http.Request) {
	var next AuthProvidersConfig
	if err := decodeJSONBody(w, r, &next, adminJSONBodyLimit); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.TrimSpace(next.OIDC.ClientSecret) == "" {
		next.OIDC.ClientSecret = a.authProviders.OIDC.ClientSecret
	}
	previous := a.authProviders
	a.authProviders = next
	a.normalizeAuthProviderLocked()
	if err := validateAccessProvider(a.authProviders.OIDC); err != nil {
		a.authProviders = previous
		writeError(w, http.StatusBadRequest, err)
		return
	}
	sites, head, tail, err := a.load()
	if err != nil {
		a.authProviders = previous
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !a.authProviders.OIDC.Enabled {
		for _, site := range sites {
			if site.Enabled && site.AuthEnabled {
				a.authProviders = previous
				writeError(w, http.StatusConflict, errors.New("access provider is still used by a protected website"))
				return
			}
		}
	} else {
		if a.authProviders.OIDC.ACMEIssuerID != "" {
			if _, ok := findACMEIssuer(a.settings.ACMEIssuers, a.authProviders.OIDC.ACMEIssuerID); !ok {
				a.authProviders = previous
				writeError(w, http.StatusBadRequest, errors.New("selected SSO ACME authority does not exist"))
				return
			}
		}
		gatewayHost := accessProviderGatewayHost(a.authProviders.OIDC)
		for _, site := range sites {
			siteHost, _, _ := splitHostPortLoose(site.Address)
			if strings.EqualFold(siteHost, gatewayHost) {
				a.authProviders = previous
				writeError(w, http.StatusConflict, errors.New("SSO base URL must not use a managed website hostname"))
				return
			}
		}
	}
	if err := a.saveAuthProvidersLocked(); err != nil {
		a.authProviders = previous
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := a.saveAndApplyCaddyConfigLocked(head, sites, tail); err != nil {
		authErr := err
		a.authProviders = previous
		if restoreErr := a.saveAuthProvidersLocked(); restoreErr != nil {
			authErr = fmt.Errorf("%w; restoring auth provider failed: %v", authErr, restoreErr)
		}
		writeError(w, http.StatusBadGateway, authErr)
		return
	}
	a.oidcCache = make(map[string]*oidcRuntime)
	writeJSON(w, http.StatusOK, a.publicAuthProvidersLocked())
}

func accessProviderGatewayHost(p AccessOIDCProvider) string {
	u, err := url.Parse(p.GatewayURL)
	if err != nil {
		return ""
	}
	host, _, err := splitHostPortLoose(u.Host)
	if err != nil {
		return ""
	}
	return strings.ToLower(host)
}

func accessRequestHost(r *http.Request) string {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	if parsed, _, err := splitHostPortLoose(host); err == nil && parsed != "" {
		return strings.ToLower(parsed)
	}
	return strings.ToLower(strings.TrimSpace(host))
}

func (a *App) accessSite(r *http.Request) (Site, error) {
	host := accessRequestHost(r)
	sites, err := a.readSites()
	if err != nil {
		return Site{}, err
	}
	for _, site := range sites {
		siteHost, _, _ := splitHostPortLoose(site.Address)
		if site.Enabled && site.AuthEnabled && strings.EqualFold(siteHost, host) {
			return site, nil
		}
	}
	return Site{}, errors.New("protected website not found")
}

func accessCookie(name, value string, age time.Duration) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: "/", MaxAge: int(age.Seconds()), HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}
}

func (a *App) handleAccessCheck(w http.ResponseWriter, r *http.Request) {
	if a.trustedProxies == nil || !a.trustedProxies.contains(r.Context(), r.RemoteAddr) {
		writeError(w, http.StatusForbidden, errors.New("trusted proxy required"))
		return
	}
	site, err := a.accessSite(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	a.mu.Lock()
	var session AccessSession
	if cookie, cookieErr := r.Cookie(accessCookieName); cookieErr == nil {
		session = a.accessSessions[cookie.Value]
		if !session.ExpiresAt.After(time.Now()) {
			delete(a.accessSessions, cookie.Value)
			session = AccessSession{}
		}
	}
	a.mu.Unlock()
	if session.SiteID != site.ID {
		returnPath := safeAccessReturnPath(r.Header.Get("X-Forwarded-Uri"))
		a.mu.Lock()
		gatewayURL := a.authProviders.OIDC.GatewayURL
		providerEnabled := a.authProviders.OIDC.Enabled
		a.mu.Unlock()
		if !providerEnabled || gatewayURL == "" {
			writeError(w, http.StatusServiceUnavailable, errors.New("OIDC access provider unavailable"))
			return
		}
		target := gatewayURL + "/.caddymgm/auth/start?site=" + url.QueryEscape(site.ID) + "&return=" + url.QueryEscape(returnPath)
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
	w.Header().Set("X-Auth-User", session.Username)
	w.Header().Set("X-Auth-Email", session.Email)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) accessRuntime(ctx context.Context, redirectURL string) (*oidcRuntime, error) {
	a.mu.Lock()
	p := a.authProviders.OIDC
	if err := validateAccessProvider(p); err != nil {
		a.mu.Unlock()
		return nil, err
	}
	key := "access|" + p.IssuerURL + "|" + p.ClientID + "|" + p.Scopes
	base, ok := a.oidcCache[key]
	a.mu.Unlock()
	if !ok {
		provider, err := a.oidcProvider(ctx, p.IssuerURL)
		if err != nil {
			return nil, err
		}
		base = &oidcRuntime{Provider: provider, Verifier: provider.Verifier(&oidc.Config{ClientID: p.ClientID}), Config: oauth2.Config{ClientID: p.ClientID, ClientSecret: p.ClientSecret, Endpoint: provider.Endpoint(), Scopes: strings.Fields(p.Scopes)}}
		a.mu.Lock()
		if cached := a.oidcCache[key]; cached != nil {
			base = cached
		} else {
			a.oidcCache[key] = base
		}
		a.mu.Unlock()
	}
	result := *base
	result.Config.RedirectURL = redirectURL
	return &result, nil
}

func safeAccessReturnPath(value string) string {
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "/"
	}
	return value
}

func (a *App) accessProxyAllowed(r *http.Request) bool {
	return a.trustedProxies != nil && a.trustedProxies.contains(r.Context(), r.RemoteAddr)
}

func (a *App) accessGatewayRequest(r *http.Request) bool {
	if !a.accessProxyAllowed(r) {
		return false
	}
	a.mu.Lock()
	provider := a.authProviders.OIDC
	a.mu.Unlock()
	return provider.Enabled && accessProviderGatewayHost(provider) != "" && strings.EqualFold(accessRequestHost(r), accessProviderGatewayHost(provider))
}

func (a *App) handleAccessStart(w http.ResponseWriter, r *http.Request) {
	if !a.accessGatewayRequest(r) {
		writeError(w, http.StatusNotFound, errors.New("SSO gateway not found"))
		return
	}
	siteID := strings.TrimSpace(r.URL.Query().Get("site"))
	var site Site
	var err error
	if siteID != "" {
		site, err = a.siteByID(siteID)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
	}
	returnPath := safeAccessReturnPath(r.URL.Query().Get("return"))
	if session, ok := a.centralAccessSession(r); ok {
		if siteID == "" {
			a.mu.Lock()
			gatewayURL := a.authProviders.OIDC.GatewayURL
			a.mu.Unlock()
			http.Redirect(w, r, gatewayURL, http.StatusFound)
		} else if err := a.redirectWithAccessTicket(w, r, site, returnPath, session.Username, session.Email, session.LoginID); err != nil {
			writeError(w, http.StatusTooManyRequests, err)
		}
		return
	}
	a.mu.Lock()
	redirectURL := a.authProviders.OIDC.GatewayURL + "/.caddymgm/auth/callback"
	a.mu.Unlock()
	runtime, err := a.accessRuntime(r.Context(), redirectURL)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("OIDC provider unavailable"))
		return
	}
	state := newSessionToken()
	a.mu.Lock()
	now := time.Now()
	for key, item := range a.accessStates {
		if !item.ExpiresAt.After(now) {
			delete(a.accessStates, key)
		}
	}
	if len(a.accessStates) >= accessStateLimit {
		a.mu.Unlock()
		writeError(w, http.StatusTooManyRequests, errors.New("too many pending logins"))
		return
	}
	a.accessStates[state] = AccessState{ExpiresAt: now.Add(oidcStateLifetime), SiteID: siteID, ReturnPath: returnPath, RedirectURL: redirectURL}
	a.mu.Unlock()
	auditSite := site.Address
	if auditSite == "" {
		auditSite = "CaddyMGM SSO"
	}
	a.recordOIDCAudit(r, "login_started", "pending", "", "", auditSite, "Website SSO login")
	http.SetCookie(w, accessCookie(accessStateCookieName, state, oidcStateLifetime))
	http.Redirect(w, r, runtime.Config.AuthCodeURL(state), http.StatusFound)
}

func (a *App) handleAccessCallback(w http.ResponseWriter, r *http.Request) {
	if !a.accessGatewayRequest(r) {
		writeError(w, http.StatusNotFound, errors.New("SSO gateway not found"))
		return
	}
	auditSuccess := false
	defer func() {
		if !auditSuccess {
			a.recordOIDCAudit(r, "login_failed", "failed", "", "", "CaddyMGM SSO", "Website SSO login")
		}
	}()
	state, code := strings.TrimSpace(r.URL.Query().Get("state")), strings.TrimSpace(r.URL.Query().Get("code"))
	cookie, err := r.Cookie(accessStateCookieName)
	if err != nil || state == "" || code == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		writeError(w, http.StatusBadRequest, errors.New("invalid OIDC response"))
		return
	}
	a.mu.Lock()
	pending, ok := a.accessStates[state]
	delete(a.accessStates, state)
	a.mu.Unlock()
	http.SetCookie(w, accessCookie(accessStateCookieName, "", -time.Hour))
	if !ok || !pending.ExpiresAt.After(time.Now()) {
		writeError(w, http.StatusBadRequest, errors.New("OIDC state expired"))
		return
	}
	runtime, err := a.accessRuntime(r.Context(), pending.RedirectURL)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("OIDC provider unavailable"))
		return
	}
	token, err := runtime.Config.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, errors.New("OIDC exchange failed"))
		return
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("missing ID token"))
		return
	}
	idToken, err := runtime.Verifier.Verify(r.Context(), raw)
	if err != nil {
		writeError(w, http.StatusUnauthorized, errors.New("invalid ID token"))
		return
	}
	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		writeError(w, http.StatusUnauthorized, errors.New("invalid claims"))
		return
	}
	textClaim := func(name string) string { value, _ := claims[name].(string); return strings.TrimSpace(value) }
	username := firstNonEmpty(textClaim("preferred_username"), textClaim("email"), textClaim("name"), textClaim("sub"))
	if username == "" {
		writeError(w, http.StatusUnauthorized, errors.New("OIDC identity is missing"))
		return
	}
	var site Site
	if pending.SiteID != "" {
		site, err = a.siteByID(pending.SiteID)
		if err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
	}
	loginID := newSessionToken()
	centralToken := a.storeAccessSession("", username, textClaim("email"), loginID)
	http.SetCookie(w, accessCookie(accessCookieName, centralToken, sessionLifetime))
	auditSuccess = true
	auditSite := site.Address
	if auditSite == "" {
		auditSite = "CaddyMGM SSO"
	}
	a.recordOIDCAudit(r, "login_success", "success", username, textClaim("email"), auditSite, "Website SSO login")
	if pending.SiteID == "" {
		a.mu.Lock()
		gatewayURL := a.authProviders.OIDC.GatewayURL
		a.mu.Unlock()
		http.Redirect(w, r, gatewayURL, http.StatusFound)
		return
	}
	if err := a.redirectWithAccessTicket(w, r, site, pending.ReturnPath, username, textClaim("email"), loginID); err != nil {
		writeError(w, http.StatusTooManyRequests, err)
	}
}

func (a *App) centralAccessSession(r *http.Request) (AccessSession, bool) {
	cookie, err := r.Cookie(accessCookieName)
	if err != nil {
		return AccessSession{}, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	session, ok := a.accessSessions[cookie.Value]
	if !ok || session.SiteID != "" || !session.ExpiresAt.After(time.Now()) {
		delete(a.accessSessions, cookie.Value)
		return AccessSession{}, false
	}
	return session, true
}

func (a *App) storeAccessSession(siteID, username, email, loginID string) string {
	token := newSessionToken()
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for key, item := range a.accessSessions {
		if !item.ExpiresAt.After(now) {
			delete(a.accessSessions, key)
		}
	}
	if len(a.accessSessions) >= accessSessionLimit {
		for key := range a.accessSessions {
			delete(a.accessSessions, key)
			break
		}
	}
	a.accessSessions[token] = AccessSession{ExpiresAt: now.Add(sessionLifetime), SiteID: siteID, Username: username, Email: email, LoginID: loginID}
	return token
}

func (a *App) redirectWithAccessTicket(w http.ResponseWriter, r *http.Request, site Site, returnPath, username, email, loginID string) error {
	ticket := newSessionToken()
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for key, item := range a.accessTickets {
		if !item.ExpiresAt.After(now) {
			delete(a.accessTickets, key)
		}
	}
	if len(a.accessTickets) >= accessTicketLimit {
		return errors.New("too many pending access tickets")
	}
	a.accessTickets[ticket] = AccessTicket{ExpiresAt: now.Add(accessTicketLifetime), SiteID: site.ID, ReturnPath: safeAccessReturnPath(returnPath), Username: username, Email: email, LoginID: loginID}
	target := "https://" + site.Address + "/.caddymgm/auth/complete?ticket=" + url.QueryEscape(ticket)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(w, r, target, http.StatusFound)
	return nil
}

func (a *App) handleAccessComplete(w http.ResponseWriter, r *http.Request) {
	if !a.accessProxyAllowed(r) {
		writeError(w, http.StatusForbidden, errors.New("trusted proxy required"))
		return
	}
	site, err := a.accessSite(r)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("ticket"))
	a.mu.Lock()
	ticket, ok := a.accessTickets[token]
	if ok && ticket.SiteID == site.ID && ticket.ExpiresAt.After(time.Now()) {
		delete(a.accessTickets, token)
	} else {
		ok = false
	}
	a.mu.Unlock()
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("invalid or expired access ticket"))
		return
	}
	sessionToken := a.storeAccessSession(site.ID, ticket.Username, ticket.Email, ticket.LoginID)
	http.SetCookie(w, accessCookie(accessCookieName, sessionToken, sessionLifetime))
	a.recordOIDCAudit(r, "access_granted", "success", ticket.Username, ticket.Email, site.Address, "Protected web host session created")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(w, r, ticket.ReturnPath, http.StatusFound)
}

func (a *App) siteByID(id string) (Site, error) {
	sites, err := a.readSites()
	if err != nil {
		return Site{}, err
	}
	for _, site := range sites {
		if site.ID == id && site.Enabled && site.AuthEnabled {
			return site, nil
		}
	}
	return Site{}, errors.New("protected website no longer exists")
}
