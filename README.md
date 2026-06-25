# caddymgm

Management web interface for Caddy websites.

The container serves a small UI and API. The Caddy configuration stays outside
the container and is mounted into `/config`.

Docker Compose can run only CaddyMGM, or CaddyMGM plus a separate Caddy service.
This allows the Caddy reverse proxy to be updated independently from the
CaddyMGM management interface.

## Run

```bash
docker compose up -d --build
```

Open:

```text
http://localhost:8080
```

The management interface is exposed by Docker Compose on port `8080`:

```yaml
ports:
  - "8080:8080"
```

Caddy is exposed by Docker Compose on the standard HTTP and HTTPS ports:

```yaml
ports:
  - "80:80"
  - "443:443"
  - "443:443/udp"
```

The host ports can be changed in `.env`:

```text
CADDY_HTTP_PORT=80
CADDY_HTTPS_PORT=443
```

By default CaddyMGM only writes the Caddyfile and does not reload Caddy:

```text
CADDYMGM_CADDY_MODE=file
```

For a native or remote Caddy server, enable the Caddy Admin API integration:

```text
CADDYMGM_CADDY_MODE=native
CADDYMGM_CADDY_API_URL=http://192.0.2.10:2019
```

For the optional Docker Caddy service from this Compose file:

```text
COMPOSE_PROFILES=docker-caddy
CADDYMGM_CADDY_MODE=docker
CADDYMGM_CADDY_API_URL=http://caddy:2019
```

The application protects the web interface and API with a login page and
HttpOnly session cookie.
The initial credentials come from Docker Compose environment variables:

```text
CADDYMGM_ADMIN_USER=admin
CADDYMGM_ADMIN_PASSWORD=changeme
CADDYMGM_AUTH_ENABLED=true
```

Change the password after the first start in `Settings`.
Authentication defaults to enabled. It can be disabled only through the
`CADDYMGM_AUTH_ENABLED=false` environment variable in `.env`.

## Config Access

The Caddy configuration lives outside the container on the host:

```text
./config/Caddyfile
```

Docker Compose mounts this directory into the container:

```yaml
volumes:
  - ./config:/config
```

Caddy receives the same Caddyfile as read-only configuration:

```yaml
volumes:
  - ./config:/etc/caddy:ro
```

Inside the container the application reads and writes:

```text
/config/Caddyfile
```

The path can be changed with:

```text
CADDY_CONFIG_PATH=/config/Caddyfile
```

CaddyMGM stores its own settings next to the Caddyfile:

```text
./config/caddymgm-settings.json
```

Inside the container this file is read and written at:

```text
/config/caddymgm-settings.json
```

The container is started with the host user and group from `.env`:

```text
PUID=1000
PGID=1000
```

This keeps `./config/Caddyfile` writable by the container while still leaving the
file visible and editable on the host. If the host user has different ids, set
them with:

```bash
id -u
id -g
```

## Configuration

CaddyMGM manages proxy host entries in the configured Caddyfile.

Access logs are enabled by default for newly created sites by writing Caddy's
`log` directive into the site block. They can be disabled manually per site in
the proxy host editor.

TLS is disabled by default for newly created sites. CaddyMGM writes these sites
as `http://example.com` blocks so Caddy does not request Let's Encrypt
certificates automatically.

TLS can be enabled per site:

- `Internes Zertifikat` writes `tls internal`.
- `ACME Zertifizierungsstelle` writes a Caddy ACME issuer block for the selected
  certificate authority.

Custom ACME certificate authorities can be managed in `Certificates`.

## Caddy Integration

CaddyMGM supports three Caddy integration modes:

| Mode | Behavior |
| --- | --- |
| `file` | Write `/config/Caddyfile` only, without reloading Caddy |
| `native` | Write `/config/Caddyfile` and load it into a native or remote Caddy via Admin API |
| `docker` | Write `/config/Caddyfile` and load it into the Compose Caddy service via Admin API |

Caddy's Admin API exposes `POST /load`, which replaces the active configuration
without downtime and rolls back if loading fails. CaddyMGM sends the generated
Caddyfile to this endpoint with `Content-Type: text/caddyfile`.

For native Caddy, the proxy server does not need Docker. It only needs Caddy's
Admin API reachable from the CaddyMGM container:

```text
CADDYMGM_CADDY_MODE=native
CADDYMGM_CADDY_API_URL=http://proxy.example.internal:2019
```

Protect the Caddy Admin API. It should only be reachable from trusted hosts or
over a private management network.

The optional Docker Caddy service is enabled with the `docker-caddy` Compose
profile:

```yaml
services:
  caddymgm:
    image: ghcr.io/thetaran/caddymgm:latest

  caddy:
    image: ${CADDY_IMAGE:-caddy:2-alpine}
    profiles:
      - docker-caddy
```

CaddyMGM reads and writes the external Caddyfile at `/config/Caddyfile`.
Caddy reads the same file at `/etc/caddy/Caddyfile`.

Caddy stores certificates and runtime data in Docker named volumes:

```text
caddy_data
caddy_config
```

To start CaddyMGM with the optional Docker Caddy service:

```bash
COMPOSE_PROFILES=docker-caddy docker compose up -d
```

To update only the Docker Caddy service:

```bash
docker compose pull caddy
docker compose up -d --no-deps caddy
```

To update only CaddyMGM:

```bash
docker compose pull caddymgm
docker compose up -d --no-deps caddymgm
```

The Caddy image can be changed independently in `.env`:

```text
CADDY_IMAGE=caddy:2-alpine
```

## Views

- `Dashboard` lists all proxy hosts.
- `Proxy Hosts` shows and edits the configuration of individual websites.
- `Certificates` is reserved for TLS certificate management.
- `Logs` shows CaddyMGM host events. Real Caddy access log ingestion is not
  connected yet.
- `Settings` shows and updates CaddyMGM settings, including authentication.

## Docker

GitHub Actions runs tests for pull requests. Docker images are built and pushed
to GitHub Container Registry only when a version tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

```text
ghcr.io/thetaran/caddymgm:latest
```

To use the image without building locally:

```bash
docker compose pull
docker compose up -d
```

Environment variables:

| Name | Default | Description |
| --- | --- | --- |
| `PUID` | `1000` | User id used by Docker Compose for host file ownership |
| `PGID` | `1000` | Group id used by Docker Compose for host file ownership |
| `CADDYMGM_ADMIN_USER` | `admin` | Initial admin user when no settings file exists |
| `CADDYMGM_ADMIN_PASSWORD` | `changeme` | Initial admin password when no settings file exists |
| `CADDYMGM_AUTH_ENABLED` | `true` | Enable or disable login/session authentication |
| `CADDYMGM_LISTEN` | `:8080` | HTTP listen address inside the container |
| `CADDY_CONFIG_PATH` | `/config/Caddyfile` | Path to the mounted Caddyfile |
| `CADDYMGM_SETTINGS_PATH` | `/config/caddymgm-settings.json` | Path to the mounted CaddyMGM settings file |
| `CADDYMGM_CADDY_MODE` | `file` | Caddy integration mode: `file`, `native`, `docker`, or `api` |
| `CADDYMGM_CADDY_API_URL` | empty | Caddy Admin API base URL for `native`, `docker`, or `api` mode |
| `COMPOSE_PROFILES` | empty | Set to `docker-caddy` to start the optional Compose Caddy service |
| `CADDY_IMAGE` | `caddy:2-alpine` | Caddy Docker image used by the separate Caddy service |
| `CADDY_HTTP_PORT` | `80` | Host HTTP port forwarded to the Caddy container |
| `CADDY_HTTPS_PORT` | `443` | Host HTTPS port forwarded to the Caddy container |

## Current scope

- List managed sites
- Add, edit, enable, disable and delete sites
- Generate reverse proxy and static file Caddy blocks
- Enable Caddy access logs by default for new sites, with a per-site toggle
- Keep TLS disabled by default so Caddy does not request public certificates
  unless enabled per site
- Manage custom ACME certificate authorities under Certificates
- Load generated Caddyfiles through the Caddy Admin API for native or Docker
  Caddy deployments
- Login page with session-cookie authentication for the management interface
- Editable CaddyMGM settings
- In-memory CaddyMGM host event logs
