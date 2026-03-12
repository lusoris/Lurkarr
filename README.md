# Lurkarr

Lurkarr is a self-hosted media library companion for the \*Arr ecosystem. It automatically searches for missing and upgradeable media across your Sonarr, Radarr, Lidarr, Readarr, Whisparr, and Eros instances — with Prowlarr indexer monitoring, download client management, intelligent queue cleaning, and a modern dark-mode dashboard.

## Features

### Core

- **Multi-App Lurking** — Automatically searches for missing and upgradeable media across Sonarr, Radarr, Lidarr, Readarr, Whisparr (v2), and Eros (Whisparr v3)
- **Multi-Instance** — Manage multiple instances per app type, each with independent settings
- **Smart Search** — Configurable missing/upgrade counts, random or sequential selection, stateful tracking to avoid duplicate searches
- **API Rate Limiting** — Hourly caps per app/instance to prevent indexer overload
- **Scheduling** — Time-based rules to enable/disable apps and adjust caps (days + hour/minute granularity)

### Queue Management

- **Queue Cleaning** — Automatically detect and remove stalled, slow, failed-import, or metadata-stuck downloads
- **Strike System** — Configurable strike thresholds with time windows before removal
- **Score Deduplication** — Detect duplicate queue items and keep the highest-scored release
- **Auto Import** — Detect stuck importPending items and trigger manual imports
- **Seeding Rules** — Enforce ratio and time limits on torrent clients (and/or mode, skip private trackers)
- **Orphan Detection** — Find downloads with no matching queue item (with grace period and exclusions)
- **Hardlink Protection** — Skip deletion if files have hardlinks
- **Cross-Seed Awareness** — Detect cross-seeded torrents sharing content
- **Blocklist System** — Community blocklist sources (URL sync with ETag), pattern matching (release group, title contains/regex, indexer)
- **Cross-Arr Blocklist Sync** — Propagate blocklist removals across instances

### Integrations

- **Prowlarr** — Monitor indexer health, statistics, and sync
- **SABnzbd** — View queue/history, pause/resume, server statistics
- **Seerr** — Media request monitoring with configurable auto-approval
- **Download Clients** — qBittorrent, Transmission, Deluge, SABnzbd, NZBGet

### Notifications

8 notification providers with per-event subscription:
Discord, Telegram, Pushover, Gotify, ntfy, Apprise, Email, Webhook

### Monitoring

- **Prometheus** — `/metrics` endpoint with lurk, queue cleaner, download client, scheduler, HTTP, and auto-import metrics
- **Grafana** — Pre-built dashboards (overview, system/runtime, Loki logs) in `deploy/`
- **API Documentation** — Interactive Scalar API reference at `/api/docs`, OpenAPI spec at `/api/spec`

### Authentication

- **Local Auth** — Username/password with bcrypt hashing
- **TOTP 2FA** — QR code setup, verify-on-login
- **OIDC/SSO** — Authorization code + PKCE, auto-create users, group/role mapping
- **Proxy Auth** — Trusted proxy headers with IP allowlist (e.g. Authelia, Authentik)
- **Reverse Proxy** — `BASE_PATH` sub-path support, `X-Forwarded-For/Proto` handling

### UI

- **Modern Dashboard** — SvelteKit 5 + TailwindCSS v4 dark-mode interface
- **Mobile Responsive** — Hamburger drawer sidebar, responsive grids, scrollable tables
- **Live Logs** — Real-time log streaming via WebSocket

## Getting Started

### Docker Compose (recommended)

```yaml
services:
  lurkarr:
    image: ghcr.io/lusoris/lurkarr:latest
    ports:
      - "9705:9705"
    environment:
      LISTEN_ADDR: ":9705"
      DATABASE_URL: postgres://lurkarr:lurkarr@db:5432/lurkarr?sslmode=disable
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

  db:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: lurkarr
      POSTGRES_PASSWORD: lurkarr
      POSTGRES_DB: lurkarr
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U lurkarr"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
```

```bash
docker compose up -d
```

Open `http://localhost:9705` and create your admin account on first launch.

### Helm

```bash
helm install lurkarr oci://ghcr.io/lusoris/lurkarr-helm/lurkarr \
  --set env.DATABASE_URL="postgres://..." \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=lurkarr.example.com
```

Or from the source chart:

```bash
helm install lurkarr deploy/helm/lurkarr \
  --set env.DATABASE_URL="postgres://..." 
```

### Building from Source

```bash
# Frontend
cd frontend && npm ci && npm run build && cd ..

# Backend
go build -o lurkarr ./cmd/lurkarr

# Run (requires PostgreSQL)
DATABASE_URL="postgres://user:pass@localhost:5432/lurkarr?sslmode=disable" ./lurkarr
```

## Configuration

All configuration is done via environment variables.

### Required

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |

### Server

| Variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8484` | HTTP listen address (Docker image uses `:9705`) |
| `CSRF_KEY` | *auto-generated* | CSRF token key (32 bytes hex). Set for session persistence across restarts. |
| `ALLOWED_ORIGINS` | — | CORS allowed origins (comma-separated) |
| `BASE_PATH` | — | URL sub-path prefix for reverse proxy hosting (e.g. `/lurkarr`) |
| `SECURE_COOKIE` | `false` | Set `true` when behind HTTPS |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `DB_MAX_CONNS` | *auto (CPUs×4)* | PostgreSQL connection pool size |

### Proxy Authentication

| Variable | Default | Description |
|---|---|---|
| `PROXY_AUTH` | `false` | Trust proxy auth headers |
| `PROXY_HEADER` | `Remote-User` | Header(s) containing username (comma-separated) |
| `TRUSTED_PROXIES` | *RFC1918 ranges* | Trusted proxy CIDRs (comma-separated) |

### OIDC / SSO

| Variable | Default | Description |
|---|---|---|
| `OIDC_ENABLED` | `false` | Enable OIDC authentication |
| `OIDC_ISSUER_URL` | — | OIDC provider issuer URL |
| `OIDC_CLIENT_ID` | — | OAuth2 client ID |
| `OIDC_CLIENT_SECRET` | — | OAuth2 client secret |
| `OIDC_REDIRECT_URL` | — | Callback URL (e.g. `https://lurkarr.example.com/api/auth/oidc/callback`) |
| `OIDC_SCOPES` | `openid,profile,email` | OIDC scopes (comma-separated) |
| `OIDC_AUTO_CREATE_USER` | `true` | Auto-create local user on first OIDC login |
| `OIDC_ADMIN_GROUP` | — | OIDC group claim that maps to admin role |

## Reverse Proxy Examples

### Traefik

```yaml
labels:
  - "traefik.http.routers.lurkarr.rule=Host(`lurkarr.example.com`)"
  - "traefik.http.services.lurkarr.loadbalancer.server.port=9705"
```

### Caddy

```
lurkarr.example.com {
    reverse_proxy lurkarr:9705
}
```

### nginx

```nginx
location / {
    proxy_pass http://lurkarr:9705;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

For sub-path hosting (e.g. `/lurkarr`), set `BASE_PATH=/lurkarr` and adjust your reverse proxy accordingly.

## API Documentation

Interactive API documentation is available at `/api/docs` (powered by [Scalar](https://github.com/scalar/scalar)). The raw OpenAPI 3.1 spec is served at `/api/spec`.

## Monitoring

### Prometheus

Lurkarr exposes a `/metrics` endpoint with metrics for lurk operations, queue cleaning, download clients, scheduling, HTTP requests, and auto-imports.

### Grafana Dashboards

Pre-built dashboards are included in `deploy/`:

- **Lurkarr Overview** (`deploy/grafana/dashboards/lurkarr.json`) — Search rates, missing/upgrade trends, error rates, durations
- **Lurkarr System** (`deploy/grafana/dashboards/lurkarr-system.json`) — Go runtime: goroutines, heap, GC, CPU, threads
- **Lurkarr Logs** (`deploy/grafana/dashboards/lurkarr-logs.json`) — Loki log exploration, volume by level, error aggregation

A monitoring stack (Prometheus + Loki + Grafana) is included in `deploy/docker-compose.monitoring.yml`.

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26, PostgreSQL 17, pgx/v5, goose migrations |
| Frontend | Svelte 5 (runes), SvelteKit, TailwindCSS v4, TypeScript |
| Auth | bcrypt, gorilla/csrf, pquerna/otp, go-oidc/v3 |
| Scheduling | gocron/v2 |
| DI | Uber FX |
| Monitoring | Prometheus client_golang, Grafana, Loki |
| Deployment | Docker (distroless), Helm, docker-compose |

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE).
