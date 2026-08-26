<p align="center">
  <img src="./caddymgmdev/cmd/server/web/CaddyMGM.png" alt="CaddyMGM" width="220" />
</p>

# caddymgm

<p align="center">Management web interface for Caddy web hosts.</p>

## Overview

CaddyMGM gives you a web UI to:

- view all managed web hosts
- create and edit reverse proxies or static file hosts
- manage ACME authorities
- inspect website logs
- manage CaddyMGM authentication and web interface settings
- protect individual web hosts through one central OIDC provider

The Caddy configuration stays outside the container.
Caddy can run separately and can be updated independently from CaddyMGM.

## Planned Upcoming Features

_Last updated: 2026-08-26_

- Integration of LDAP for authentication of websites or OIDC
- GEO IP Map on Dashboard for Accessed Websites
- Security features to protect managed websites, including:
  - security headers
  - HTTP Strict Transport Security (HSTS)
  - access policies for everyone, allowed IPs, or blocked IPs
  - IP address and CIDR lists
  - basic authentication with username and password
  - maximum request body size limits
  - allowed HTTP methods
  - blocked paths

## Screenshots

| View | Preview |
| --- | --- |
| **Dashboard**<br>See all managed web hosts and their current state at a glance. | <img src="./caddymgmdev/cmd/server/web/dashboard.png" alt="CaddyMGM Dashboard interface" width="760" /> |
| **Web Hosts**<br>Create and edit reverse proxies and static websites from the rule editor. | <img src="./caddymgmdev/cmd/server/web/web-hosts.png" alt="CaddyMGM Web Hosts interface" width="760" /> |
| **Certificates**<br>Manage built-in and custom ACME authorities and inspect issued certificates. | <img src="./caddymgmdev/cmd/server/web/certificates.png" alt="CaddyMGM Certificates interface" width="760" /> |
| **Logs**<br>Follow the Caddy service log and per-host access logs in realtime. | <img src="./caddymgmdev/cmd/server/web/logs.png" alt="CaddyMGM Logs interface" width="760" /> |
| **Settings**<br>Configure authentication, the management interface, and log retention. | <img src="./caddymgmdev/cmd/server/web/settings.png" alt="CaddyMGM Settings interface" width="760" /> |

## Quick Start

1. Create the local Compose and environment files from their templates:

```bash
cp compose-template.yml compose.yml
cp .env.example .env
```

2. Start CaddyMGM:

```bash
docker compose up -d --build
```

3. Open:

```text
http://localhost:8080
```

4. Log in with the credentials from `.env`.

For remote access, publish CaddyMGM only behind HTTPS.
Local `http://localhost:8080` access remains supported for development on the same machine.

## Optional Caddy Service

If you also want Docker Compose to run Caddy, enable the profile:

```bash
COMPOSE_PROFILES=docker-caddy docker compose up -d --build
```

This publishes:

- `80` for HTTP
- `443` for HTTPS
- `8080` for the management entrypoint through Caddy

## Runtime Folders

The repository root is the runtime structure:

| Path | Purpose |
| --- | --- |
| `./caddy-config` | External Caddy configuration |
| `./caddy-data` | Caddy data, certificates, state, Root CAs, and local static site files |
| `./caddy-logs` | Website access logs and Caddy service log |
| `./caddymgm` | CaddyMGM settings file |
| `./caddymgmdev` | Go source code, frontend, and Docker build context |

## Important Files

| File | Description |
| --- | --- |
| `./caddy-config/Caddyfile` | Main Caddyfile managed by CaddyMGM |
| `./caddymgm/caddymgm-settings.json` | Persistent CaddyMGM settings |
| `./caddymgm/auth-providers.json` | Automatically managed central OIDC provider configuration; exists only while Central Website SSO is enabled |
| `./caddymgm/oidc-audit.log` | Automatically managed OIDC authentication audit log; no Compose variable is required |
| `./caddy-logs/caddy-service.log` | Caddy runtime/service log |
| `./caddy-logs/<domain>.access.log` | Per-host JSON access log |
| `./caddy-data/site` | Document roots for local static sites served by Caddy under `/srv` |
| `./caddy-data/ca-certificates` | Custom Root CA certificates mounted for Caddy and CaddyMGM |

## Caddy Integration Modes

| Mode | Behavior |
| --- | --- |
| `file` | Only writes the Caddyfile |
| `native` | Writes the Caddyfile and reloads a native or remote Caddy via Admin API |
| `docker` | Writes the Caddyfile and reloads the Compose Caddy service via Admin API |
| `api` | Same behavior as API-driven mode for a reachable Caddy Admin API |

## Central Website SSO

CaddyMGM can protect multiple web hosts through one central OIDC client. Configure an HTTPS SSO Base URL such as `https://sso.example.com` under `Settings -> Authentication -> SSO / OIDC`, then register exactly this callback at the identity provider:

```text
https://sso.example.com/.caddymgm/auth/callback
```

Enable `SSO / OIDC` on each web host that should require authentication. The first login creates a central SSO session. Access to another protected host uses a short-lived, single-use, host-bound ticket to establish a separate secure cookie for that host. The identity provider controls which users may use the OIDC client.

## Web UI Notes

- New web hosts have logs enabled by default.
- New web hosts have TLS disabled by default.
- TLS is only requested when you enable it for a host and select an ACME authority.
- Static websites must use a document root under `/srv` inside the Caddy container.
- `Certificates` shows configured ACME authorities and issued certificates.
- `Logs` uses horizontal tabs for realtime Caddy service logs, website access logs, and OIDC authentication events.

## Authentication

CaddyMGM supports:

- local username/password login
- OIDC login
- both at the same time

The login page changes automatically based on your `.env` settings.

## Custom ACME Root CAs

If your ACME server uses a private Root CA:

1. place the certificate in `./caddy-data/ca-certificates`
2. or upload it in `Certificates`
3. reference it in the ACME authority

Supported file types:

- `.crt`
- `.cer`
- `.pem`

Inside the container these files are available under:

```text
/ca-certificates
```

## Compose and Environment Variables

The following variables are used by Docker Compose and CaddyMGM.

| Variable | Default | Description |
| --- | --- | --- |
| `PUID` | `1000` | User id used for container file ownership |
| `PGID` | `1000` | Group id used for container file ownership |
| `CADDYMGM_ADMIN_USER` | `admin` | Initial local admin username |
| `CADDYMGM_ADMIN_PASSWORD` | required for first startup | Initial local admin password; no default password is generated |
| `CADDYMGM_AUTH_ENABLED` | `true` | Enables the login portal |
| `CADDYMGM_LOCALAUTH_ENABLED` | `true` | Enables local username/password login |
| `CADDYMGM_ALLOW_INSECURE_HTTP` | `false` | Allows remote login over unencrypted HTTP for initial setup; disable after HTTPS is configured |
| `CADDYMGM_OIDCAUTH_ENABLED` | `false` | Enables OIDC login |
| `CADDYMGM_CADDY_MODE` | `file` | Caddy integration mode |
| `CADDYMGM_CADDY_API_URL` | `http://caddy:2019` in example | Caddy Admin API URL for reloads |
| `CADDYMGM_TRUSTED_PROXIES` | empty | Comma-separated proxy IPs, CIDRs, or hostnames allowed to supply forwarded HTTPS headers |
| `CADDYMGM_ACCESS_LOG_DIR` | `/logs` | Directory from which CaddyMGM reads website logs |
| `CADDYMGM_CADDY_DATA_DIR` | `/caddy-data` | Directory used to inspect certificate data |
| `CADDYMGM_CA_CERT_DIR` | `/ca-certificates` | Directory for uploaded or mounted Root CAs |
| `CADDYMGM_STATIC_ROOT_BASE` | `/srv` | Allowed base path for static website document roots |
| `CADDYMGM_WEB_LISTEN` | `:8080` | Internal listen address of the Go web server |
| `CADDYMGM_WEB_PORT` | `8080` | Public management port used through Caddy |
| `CADDYMGM_SETTINGS_PATH` | `/caddymgm/caddymgm-settings.json` | Path to the CaddyMGM settings file |
| `CADDY_ACCESS_LOG_DIR` | `/logs` | Directory written into generated Caddy log directives |
| `COMPOSE_PROFILES` | `docker-caddy` in example | Start optional Compose services such as Caddy |
| `CADDY_IMAGE` | `caddy:2-alpine` | Caddy Docker image |
| `CADDY_HTTP_PORT` | `80` | Host port mapped to Caddy HTTP |
| `CADDY_HTTPS_PORT` | `443` | Host port mapped to Caddy HTTPS and HTTP/3 UDP |

## Example `.env`

```text
PUID=1000
PGID=1000
CADDYMGM_ADMIN_USER=admin
CADDYMGM_ADMIN_PASSWORD=replace-with-a-strong-password
CADDYMGM_AUTH_ENABLED=true
CADDYMGM_LOCALAUTH_ENABLED=true
CADDYMGM_ALLOW_INSECURE_HTTP=true
CADDYMGM_OIDCAUTH_ENABLED=false
CADDYMGM_CADDY_MODE=file
CADDYMGM_CADDY_API_URL=http://caddy:2019
CADDYMGM_TRUSTED_PROXIES=caddy,caddy-admin
CADDYMGM_STATIC_ROOT_BASE=/srv
CADDYMGM_WEB_PORT=8080
COMPOSE_PROFILES=docker-caddy
```

`CADDYMGM_ALLOW_INSECURE_HTTP=true` is intended only for initial setup on a trusted network.
Set it to `false` and restart CaddyMGM as soon as HTTPS is available. Credentials and session
cookies are otherwise transmitted without transport encryption.

## Update

Update only CaddyMGM:

```bash
docker compose pull caddymgm
docker compose up -d --no-deps caddymgm
```

Update only Caddy:

```bash
docker compose pull caddy
docker compose up -d --no-deps caddy
```

## Source Layout

Development files live under:

```text
./caddymgmdev
```

Main source paths:

| Path | Description |
| --- | --- |
| `./caddymgmdev/cmd/server` | Go backend |
| `./caddymgmdev/cmd/server/web` | Frontend assets |
| `./caddymgmdev/Dockerfile` | CaddyMGM image build |

## Current Scope

- web host management
- reverse proxy and static file configuration
- per-host access logs
- ACME authority management
- certificate visibility and forced renew action
- local auth and OIDC auth
- Caddy Admin API integration
