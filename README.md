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

The Caddy configuration stays outside the container.
Caddy can run separately and can be updated independently from CaddyMGM.

## Quick Start

1. Copy the example environment file:

```bash
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

## Web UI Notes

- New web hosts have logs enabled by default.
- New web hosts have TLS disabled by default.
- TLS is only requested when you enable it for a host and select an ACME authority.
- `Certificates` shows configured ACME authorities and issued certificates.
- `Logs` shows realtime Caddy service logs and website access logs.

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
| `CADDYMGM_ADMIN_PASSWORD` | `changeme` | Initial local admin password |
| `CADDYMGM_AUTH_ENABLED` | `true` | Enables the login portal |
| `CADDYMGM_LOCALAUTH_ENABLED` | `true` | Enables local username/password login |
| `CADDYMGM_OIDCAUTH_ENABLED` | `false` | Enables OIDC login |
| `CADDYMGM_CADDY_MODE` | `file` | Caddy integration mode |
| `CADDYMGM_CADDY_API_URL` | `http://caddy:2019` in example | Caddy Admin API URL for reloads |
| `CADDYMGM_ACCESS_LOG_DIR` | `/logs` | Directory from which CaddyMGM reads website logs |
| `CADDYMGM_CADDY_DATA_DIR` | `/caddy-data` | Directory used to inspect certificate data |
| `CADDYMGM_CA_CERT_DIR` | `/ca-certificates` | Directory for uploaded or mounted Root CAs |
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
CADDYMGM_ADMIN_PASSWORD=changeme
CADDYMGM_AUTH_ENABLED=true
CADDYMGM_LOCALAUTH_ENABLED=true
CADDYMGM_OIDCAUTH_ENABLED=false
CADDYMGM_CADDY_MODE=file
CADDYMGM_CADDY_API_URL=http://caddy:2019
CADDYMGM_WEB_PORT=8080
COMPOSE_PROFILES=docker-caddy
```

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
