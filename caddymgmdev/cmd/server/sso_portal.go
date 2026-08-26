package main

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type accessPortalSite struct {
	Name string
	URL  string
}

type accessPortalData struct {
	Authenticated bool
	Username      string
	Email         string
	Sites         []accessPortalSite
}

var accessPortalTemplate = template.Must(template.New("access-portal").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CaddyMGM SSO</title>
  <style>
    :root{color-scheme:dark;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#07090d;color:#f5f7fb}
    *{box-sizing:border-box}body{min-height:100vh;margin:0;display:grid;place-items:center;padding:24px;background:radial-gradient(circle at 50% 15%,rgba(35,137,255,.16),transparent 34%),#07090d}
    main{width:min(460px,100%);border:1px solid #2b303b;border-radius:18px;background:#17191f;box-shadow:0 28px 90px rgba(0,0,0,.48);overflow:hidden}
    header{display:flex;align-items:center;gap:14px;padding:16px 18px;border-bottom:1px solid #2b303b;background:#12151b}
    .logo{display:grid;place-items:center;width:38px;height:38px;border:1px solid #2589ff;border-radius:13px;background:linear-gradient(145deg,#153a68,#091421);color:#48b7ff;font-size:22px;font-weight:800}
    h1,h2,p{margin:0}h1{font-size:20px}.eyebrow{display:block;margin-bottom:3px;color:#8c96a8;font-size:11px;letter-spacing:.12em;text-transform:uppercase}
    .content{display:grid;gap:16px;padding:22px}
    .lead{color:#c1c8d4;font-size:15px;line-height:1.55}.status{display:flex;align-items:center;gap:9px;color:#65d69b;font-weight:650}.dot{width:9px;height:9px;border-radius:50%;background:#39c985;box-shadow:0 0 16px rgba(57,201,133,.8)}
    .identity{display:grid;gap:5px;padding:16px;border:1px solid #303542;border-radius:12px;background:#111319}.identity span{color:#929cac;font-size:13px}
    .button{display:inline-flex;align-items:center;justify-content:center;width:100%;min-height:44px;border:1px solid #51b7ff;border-radius:10px;background:linear-gradient(180deg,#62c3ff,#2799ed);color:#07101a;font-weight:750;text-decoration:none}
    .logout{display:flex;justify-content:flex-end}.logout button{min-height:38px;border:1px solid #7e3548;border-radius:9px;padding:0 16px;background:#32131d;color:#ff8ca8;font:inherit;font-weight:700;cursor:pointer}.logout button:hover{background:#451925}
    .apps{display:grid;gap:10px}.apps h2{font-size:14px}.app{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:13px 15px;border:1px solid #2b303b;border-radius:10px;background:#111319;color:#eaf1fb;text-decoration:none}.app:hover{border-color:#2589ff;background:#151c27}.arrow{color:#48b7ff}
    footer{padding:13px 18px;border-top:1px solid #2b303b;background:#12151b;color:#8c96a8;font-size:12px;line-height:1.4;text-align:center}
  </style>
</head>
<body>
<main>
  <header><div class="logo">C</div><div><span class="eyebrow">Central authentication</span><h1>CaddyMGM SSO</h1></div></header>
  <section class="content">
  {{if .Authenticated}}
    <div class="status"><span class="dot"></span>SSO session active</div>
    <div class="identity"><strong>{{.Username}}</strong>{{if .Email}}<span>{{.Email}}</span>{{end}}</div>
    <form class="logout" method="post" action="/.caddymgm/auth/logout"><button type="submit">Sign out</button></form>
    <div class="apps"><h2>Protected web hosts</h2>{{range .Sites}}<a class="app" href="{{.URL}}"><span>{{.Name}}</span><span class="arrow">Open →</span></a>{{else}}<p class="lead">No protected web hosts are currently enabled.</p>{{end}}</div>
  {{else}}
    <p class="lead">Authenticate with OIDC SSO.</p>
    <a class="button" href="/.caddymgm/auth/start">Login</a>
  {{end}}
  </section>
  <footer>Secure access provided by CaddyMGM</footer>
</main>
</body>
</html>`))

func (a *App) handleAccessLogout(w http.ResponseWriter, r *http.Request) {
	if !a.accessGatewayRequest(r) {
		http.NotFound(w, r)
		return
	}
	a.mu.Lock()
	gatewayURL := a.authProviders.OIDC.GatewayURL
	a.mu.Unlock()
	if origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/"); origin == "" || !strings.EqualFold(origin, gatewayURL) {
		writeError(w, http.StatusForbidden, errors.New("invalid logout origin"))
		return
	}
	if cookie, err := r.Cookie(accessCookieName); err == nil {
		a.mu.Lock()
		session, ok := a.accessSessions[cookie.Value]
		if ok && session.SiteID == "" && session.LoginID != "" {
			for token, item := range a.accessSessions {
				if item.LoginID == session.LoginID {
					delete(a.accessSessions, token)
				}
			}
			for token, item := range a.accessTickets {
				if item.LoginID == session.LoginID {
					delete(a.accessTickets, token)
				}
			}
		}
		a.mu.Unlock()
	}
	http.SetCookie(w, accessCookie(accessCookieName, "", -time.Hour))
	http.SetCookie(w, accessCookie(accessStateCookieName, "", -time.Hour))
	w.Header().Set("Cache-Control", "no-store")
	a.recordOIDCAudit(r, "logout", "success", "", "", "CaddyMGM SSO", "Central website session ended")
	http.Redirect(w, r, gatewayURL, http.StatusSeeOther)
}

func (a *App) handleAccessPortal(w http.ResponseWriter, r *http.Request) {
	if !a.accessGatewayRequest(r) {
		http.NotFound(w, r)
		return
	}
	data := accessPortalData{}
	if session, ok := a.centralAccessSession(r); ok {
		data.Authenticated = true
		data.Username = session.Username
		data.Email = session.Email
		sites, err := a.readSites()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		for _, site := range sites {
			if site.Enabled && site.AuthEnabled {
				data.Sites = append(data.Sites, accessPortalSite{Name: site.Address, URL: "https://" + site.Address + "/"})
			}
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := accessPortalTemplate.Execute(w, data); err != nil {
		return
	}
}
