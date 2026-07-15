# Anweisungen für Security Tests
- Act as a Principal Security Engineer.
- Systematically scan the entire codebase systematically, file by file. 
- List every vulnerability found including its file path, line number, severity (Critical/High/Medium/Low), and technical explanation.
- Directly fix each vulnerability in the code using safe coding practices (e.g., input sanitization, parameterized queries, strict boundary enforcement). 
- After fixing, summarize the changes and flag any complex architectural decisions or secrets.
- Before finalizing, critique your own work: list 3 ways your mitigation could be bypassed, address them, and produce the final secure diff.

# CaddyMGM Security Findings

This document records security findings from the local CaddyMGM security review, their remediation status, and the checks used to verify fixes.

## Review scope

- Go backend and embedded web interface
- Authentication, sessions, OIDC, and CSRF protection
- Caddy and Docker Compose integration
- Dependency vulnerabilities
- Runtime exposure and security headers

The review used static code inspection, `go test`, `go vet`, `govulncheck`, Docker image builds, Compose validation, and non-destructive HTTP checks against the local deployment.

## Findings

### 1. Vulnerable OIDC JOSE dependency

- Severity: High
- Status: Fixed in `v0.8.2`
- Finding: `github.com/go-jose/go-jose/v4@v4.0.2` was affected by `GO-2026-4945`. `govulncheck` identified a reachable path through the OIDC callback that could cause a denial of service.
- Fix: Updated `go-jose` to `v4.1.4`. Because the fixed dependency requires Go 1.24 or newer, the module and Docker build stage were updated to Go 1.24.
- Verification: `go test ./...`, `go vet ./...`, Docker image build, and `govulncheck` completed successfully. No reachable Go vulnerabilities remained.

### 2. Caddy Admin API exposed across the Docker network

- Severity: Medium
- Status: Fixed in `v0.8.2`
- Finding: The unauthenticated Caddy Admin API listened on `0.0.0.0:2019` and was reachable by every container on the shared Compose network.
- Fix: Added a dedicated internal Docker management network. Caddy binds the Admin API to the internal `caddy-admin` network alias, and CaddyMGM uses `http://caddy-admin:2019` explicitly.
- Verification: `docker compose config` passed, and an isolated container test confirmed that CaddyMGM-style access to the Admin API works through the internal alias.

### 3. Missing HTTP server timeouts

- Severity: Medium
- Status: Fixed in `v0.8.2`
- Finding: The Go HTTP server did not define connection or request timeouts, increasing exposure to slow-client resource exhaustion.
- Fix: Added the following limits:
  - `ReadHeaderTimeout`: 10 seconds
  - `ReadTimeout`: 30 seconds
  - `WriteTimeout`: 30 seconds
  - `IdleTimeout`: 60 seconds
  - `MaxHeaderBytes`: 1 MiB
- Verification: Go build, `go test`, `go vet`, and the Docker image build completed successfully.

### 4. Missing local-login rate limiting

- Severity: Medium
- Status: Fixed for `v0.8.3`
- Finding: When local authentication is enabled, login attempts are not rate-limited or temporarily blocked.
- Fix: Added an in-memory limiter that aggregates failures both by normalized username and by username/source-IP combination. Five failures within 15 minutes cause a 15-minute lockout. Rate-limited responses use HTTP `429` and include `Retry-After`; successful authentication resets the related counters.
- Resource controls: Attempt keys are SHA-256 hashes, expired entries are removed automatically, and storage is capped at 10,000 entries.
- Verification: Unit tests cover the threshold, expiry, successful-login reset, failure-window reset, concurrent requests, bounded storage, normalized keys, and source-IP rotation. Normal tests, the race detector, `go vet`, Docker build, and `govulncheck` passed.
- Current deployment note: Local authentication was disabled during the review, reducing current exposure but not removing the application-level issue.

### 5. Unrestricted trust in forwarded proxy headers

- Severity: Low to Medium
- Status: Fixed for `v0.8.4`
- Finding: `X-Forwarded-Proto` and `Forwarded` are accepted regardless of the request source. If CaddyMGM is ever exposed directly or placed behind a proxy that preserves attacker-supplied headers, transport and Secure-cookie decisions could be influenced.
- Fix: Added `CADDYMGM_TRUSTED_PROXIES`, accepting explicit proxy IPs, CIDRs, and hostnames. Forwarded transport headers are ignored unless the direct peer matches that allowlist. The Compose deployment trusts only the internal `caddy-admin` alias.
- Additional hardening: The local-HTTP login exception now checks the actual TCP peer for a loopback address instead of trusting a client-controlled `Host` header.
- Verification: Tests cover trusted and untrusted sources, exact IPs, CIDRs, hostname resolution and caching, direct TLS, Secure-cookie behavior, invalid configuration, and spoofed localhost hostnames.

### 6. Fixed default administrator password on first start

- Severity: High
- Status: Fixed for `v0.8.5`
- Finding: If the settings file and `CADDYMGM_ADMIN_PASSWORD` are both absent, the application creates the administrator hash from the fixed password `changeme`.
- Location: `caddymgmdev/cmd/server/main.go:944` and `caddymgmdev/cmd/server/main.go:998`.
- Risk: A fresh installation with authentication enabled can be accessed with publicly predictable credentials.
- Fix: Removed the fixed fallback. First startup now aborts with a clear error unless `CADDYMGM_ADMIN_PASSWORD` is provided; existing settings without a password hash are rejected as well. Settings updates cannot clear the administrator password hash.
- Verification: Added tests for missing, configured, and incomplete administrator credentials. Normal tests, the race detector, and `go vet` passed.

### 7. Unbounded JSON request bodies

- Severity: Medium
- Status: Fixed for `v0.9.1`
- Finding: Several handlers decode `r.Body` directly with `json.NewDecoder` without an explicit request-body limit. This includes the public login endpoint.
- Locations: `caddymgmdev/cmd/server/main.go:341`, `:387`, `:502`, and `:719`.
- Risk: Oversized JSON requests can consume unnecessary memory and CPU and may enable resource-exhaustion attacks.
- Fix: Added a shared bounded JSON decoder using `http.MaxBytesReader`. Login payloads are limited to 64 KiB; administrative JSON payloads are limited to 1 MiB. Trailing additional JSON documents are rejected.
- Verification: Added tests for oversized payloads, multiple JSON values, and valid trailing whitespace. Normal tests, the race detector, and `go vet` passed.

### 8. Unbounded OIDC state storage

- Severity: Medium
- Status: Fixed for `v0.8.6`
- Finding: Every request to the public OIDC start endpoint adds an entry to `a.oidcStates`; expired entries are not periodically removed.
- Locations: `caddymgmdev/cmd/server/main.go:811` and `:830`.
- Risk: Repeated OIDC start requests can grow the in-memory state map without a fixed bound.
- Fix: Expired states are removed before new state creation. Outstanding states are capped at 1,000; when capacity is exhausted, the endpoint returns HTTP `429` with `Retry-After: 60` and does not allocate another state.
- Verification: Added tests for expiry cleanup, capacity enforcement, capacity recovery, and the ten-minute state lifetime.

### 9. DNS lookup while holding the trusted-proxy cache mutex

- Severity: Low
- Status: Open
- Finding: Hostname-based trusted-proxy checks perform DNS resolution while holding the cache mutex.
- Location: `caddymgmdev/cmd/server/main.go:2763`.
- Risk: Slow or unavailable DNS can briefly block concurrent trusted-proxy checks and request processing.
- Recommended fix: Resolve outside the mutex and atomically publish the result afterward. Prefer fixed IPs or CIDRs for production trusted-proxy configuration.

### 10. Local `.env` permissions

- Classification: Server-level hardening, not an application finding
- Status: Excluded from application remediation
- Observation: The local `.env` file uses mode `0664` and may contain secrets. It remains excluded from Git.
- Decision: File ownership and permissions are managed as a server concern and were intentionally not changed by the application fixes.
- Server recommendation: Review permissions, owner/group membership, backup access, and host-level secret-management policy separately.

## Positive controls observed

- Anonymous API access is rejected when authentication is enabled.
- HTTPS deployment provides HSTS, CSP, frame protection, MIME sniffing protection, and a restrictive referrer policy.
- Session cookies use `HttpOnly`, `Secure` over HTTPS, and `SameSite=Lax`.
- State-changing API requests use origin validation and a CSRF token.
- OIDC state values are cryptographically generated, expire, and are single-use.
- Password hashes use bcrypt.
- Certificate uploads are size-limited, parsed as X.509 certificates, and written with sanitized filenames.
- Runtime secrets, settings, Caddy configuration, logs, and certificate data are excluded from Git.

## Release record

- Security fixes 1-3 were released in commit `bbbf327` with tag `v0.8.2`.
- Finding 4 was fixed for release `v0.8.3`.
- Finding 5 was fixed for release `v0.8.4`.
- Finding 6 was fixed for release `v0.8.5`.
- Finding 7 was fixed for release `v0.9.1`.
- Finding 8 was fixed for release `v0.8.6`.
- Finding 9 remains open for a future security release.
- Finding 10 belongs to server administration and must not be changed as part of an application update without separate approval.
