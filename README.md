# caddymgm

Management web interface for Caddy websites.

The container serves a small UI and API with Go's built-in `net/http` server.
No nginx or Apache is bundled in the CaddyMGM container. The Caddy configuration
stays outside the container and is mounted into `/config`.

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

Caddy access logs are stored outside the containers:

```text
./logs
```

Docker Compose mounts this directory into CaddyMGM and the optional Docker Caddy
service:

```yaml
volumes:
  - ./logs:/logs
```

Caddy certificate storage and runtime data are stored outside the containers:

```text
./caddy-data
./caddy-config
./site
```

Docker Compose uses the official Caddy image mount layout:

```yaml
volumes:
  - ./config:/etc/caddy:ro
  - ./site:/srv
  - ./caddy-data:/data
  - ./caddy-config:/config
```

Caddy stores certificates under `./caddy-data/caddy/certificates`. CaddyMGM
also mounts `./caddy-data` at `/caddy-data` to show certificate expiration dates
and clean up managed certificate files when a web host is deleted.

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

CaddyMGM manages web host entries in the configured Caddyfile.

Access logs are enabled by default for newly created sites by writing Caddy's
`log` directive into the site block. CaddyMGM writes one JSON access-log file per
site under `/logs/<domain>.access.log`. They can be disabled manually per site in
the web host editor.

TLS is disabled by default for newly created sites. CaddyMGM writes these sites
as `http://example.com` blocks so Caddy does not request Let's Encrypt
certificates automatically.

TLS can be enabled per site:

- Enable `TLS enabled` in the web host editor.
- Select an ACME authority. CaddyMGM writes a Caddy ACME issuer block for the
  selected certificate authority. Caddy performs the ACME order itself; `acme.sh`
  is not used.

Custom ACME certificate authorities can be managed in `Certificates`.
`Let's Encrypt` is available as a built-in ACME authority.

The `Certificates` view contains:

- `ACME Authorities` for built-in and custom ACME issuers.
- `Issued Certificates` for managed hosts with TLS enabled.

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

Caddy stores certificates and runtime data in host directories:

```text
./caddy-data
./caddy-config
./site
```

For the optional Docker Caddy service, Compose runs Caddy with the same
`PUID`/`PGID` as CaddyMGM so generated certificate metadata can be read by the
management interface.
The `caddy-init` service prepares ownership for `./config`, `./logs` and the
Caddy host directories before the optional Caddy container starts.

### Custom Root CAs for ACME

If a custom ACME authority uses a certificate signed by an internal Root CA,
place the Root CA certificate in:

```text
./ca-certificates
```

Root CA certificates can also be uploaded from the `Certificates` view when
editing an ACME authority. Uploaded PEM or DER certificates are stored under
`./ca-certificates` as PEM `.crt` files and the `Root CA file` field is filled
automatically.

For manual placement, use PEM encoded files with a `.crt`, `.cer` or `.pem`
extension, for example:

```text
./ca-certificates/internal-root-ca.cer
```

Then set the matching ACME authority's `Root CA file` field to the container
path:

```text
/ca-certificates/internal-root-ca.cer
```

CaddyMGM writes this into the Caddyfile as an ACME issuer `trusted_roots`
setting. Reload the affected web host configuration after adding or changing
Root CA files:

```bash
docker compose up -d --force-recreate caddy
```

For native or remote Caddy deployments, install the same Root CA on the server
running Caddy.

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

- `Dashboard` lists all web hosts.
- `Web Hosts` shows and edits the configuration of individual websites.
- `Certificates` manages ACME authorities and lists managed TLS certificates.
- `Logs` shows website access logs from Caddy.
- `Settings` shows and updates CaddyMGM settings, including authentication.
  Settings are grouped into `Authentication`, `Web Interface` and `Logs`.
  `Web Interface` can publish the CaddyMGM UI through Caddy with optional ACME
  TLS, using the same ACME authorities as web hosts.

## Docker

GitHub Actions runs tests for pull requests. Docker images are built only for
two-part master release tags like `v0.4` or `0.4`:

```bash
git tag v0.4
git push origin v0.4
```

The image receives these tags from that Git tag:

```text
latest
0.4
```

Three-part Git tags like `v0.4.1` are minor development tags and do not trigger
the Docker workflow.

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
| `CADDYMGM_ACCESS_LOG_DIR` | `/logs` | Directory where CaddyMGM reads website access logs |
| `CADDYMGM_CADDY_DATA_DIR` | `/caddy-data` | Caddy data directory used for certificate metadata and cleanup |
| `CADDYMGM_CA_CERT_DIR` | `/ca-certificates` | Directory where uploaded Root CA certificates are stored |
| `CADDYMGM_WEB_INTERFACE_UPSTREAM` | `http://caddymgm:8080` | Default upstream used when publishing the CaddyMGM web interface through Caddy |
| `CADDY_ACCESS_LOG_DIR` | `/logs` | Directory written into generated Caddy log directives |
| `COMPOSE_PROFILES` | empty | Set to `docker-caddy` to start the optional Compose Caddy service |
| `CADDY_IMAGE` | `caddy:2-alpine` | Caddy Docker image used by the separate Caddy service |
| `CADDY_HTTP_PORT` | `80` | Host HTTP port forwarded to the Caddy container |
| `CADDY_HTTPS_PORT` | `443` | Host HTTPS port forwarded to the Caddy container |

## Current scope

- List managed sites
- Add, edit, enable, disable and delete sites
- Generate reverse proxy and static file Caddy blocks
- Enable Caddy access logs by default for new sites, with a per-site toggle
- Store website access logs outside Docker under `./logs`
- Store Caddy certificate/runtime data outside Docker under `./caddy-data`
- Keep TLS disabled by default so Caddy does not request public certificates
  unless enabled per site
- Manage custom ACME certificate authorities under Certificates
- Trust custom ACME Root CAs by placing `.crt`, `.cer` or `.pem` files under `./ca-certificates`
- Load generated Caddyfiles through the Caddy Admin API for native or Docker
  Caddy deployments
- Login page with session-cookie authentication for the management interface
- Editable CaddyMGM settings
- Website access log viewer backed by Caddy JSON log files
