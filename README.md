# caddymgm

Management web interface for Caddy websites.

The container serves a small UI and API. The Caddy configuration stays outside
the container and is mounted into `/config`.

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

The app only owns the block between these markers:

```caddyfile
# caddymgm:start
# caddymgm:end
```

Manual Caddy directives outside this block are preserved.

Example managed reverse proxy entry:

```caddyfile
# caddymgm:start
# caddymgm:site 4f197c4ea9bd
example.local {
	reverse_proxy http://app:3000
	log
	encode zstd gzip
}
# caddymgm:end-site
# caddymgm:end
```

Example managed static file entry:

```caddyfile
# caddymgm:site 7a4d9c1a8302
files.example.local {
	root * /srv/www/files
	file_server
	log
}
# caddymgm:end-site
```

Disabled sites are kept in the Caddyfile as commented blocks, so they can be
enabled again from the UI.

Access logs are enabled by default for newly created sites by writing Caddy's
`log` directive into the site block. They can be disabled manually per site in
the proxy host editor.

## Caddy Integration

This project currently writes the Caddyfile only. Reloading Caddy is not
automated yet.

Recommended deployment model:

- Run Caddy in its own container or on the host.
- Mount the same external Caddyfile into Caddy.
- After changes, reload Caddy with your chosen mechanism.

Possible reload mechanisms:

- `caddy reload --config /path/to/Caddyfile`
- Caddy Admin API
- a Docker-side helper that sends a reload command to the Caddy container

The management container does not need access to TLS certificates, Caddy data,
or the Docker socket for the current scope.

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

## Current scope

- List managed sites
- Add, edit, enable, disable and delete sites
- Generate reverse proxy and static file Caddy blocks
- Enable Caddy access logs by default for new sites, with a per-site toggle
- Preserve manual config outside the managed block
- Login page with session-cookie authentication for the management interface
- Editable CaddyMGM settings
- In-memory CaddyMGM host event logs

Reloading Caddy is intentionally not automated yet. In a production setup this
should be wired to the chosen Caddy deployment model, for example a shared
volume plus `caddy reload`, or the Caddy admin API.
