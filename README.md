# Lurkarr 🐸

**Automated media library hunter for the *Arr ecosystem.**

Lurkarr searches for missing and upgradeable media across your Sonarr, Radarr, Lidarr, Readarr, Whisparr (v2), and Eros (Whisparr v3) instances — with Prowlarr indexer integration and SABnzbd download monitoring.

## Features

- **Multi-App Support** — Sonarr, Radarr, Lidarr, Readarr, Whisparr, Eros
- **Prowlarr Integration** — Monitor indexer status and statistics
- **SABnzbd Monitoring** — View download queue, history, pause/resume
- **Smart Hunting** — Configurable missing/upgrade counts, random/sequential selection
- **Hourly API Caps** — Prevent overloading your indexers
- **Scheduling** — Enable/disable apps and set API caps on a schedule
- **Stateful Tracking** — Remembers what's been searched to avoid duplicates
- **Live Logs** — WebSocket-powered real-time log viewer
- **2FA Support** — TOTP-based two-factor authentication
- **Modern UI** — SvelteKit + TailwindCSS dark theme

## Quick Start

### Docker Compose

```bash
curl -O https://raw.githubusercontent.com/lusoris/lurkarr/main/docker-compose.yml
docker compose up -d
```

Open `http://localhost:9705` and create your admin account.

### From Source

```bash
# Backend
go build -o lurkarr ./cmd/lurkarr

# Frontend
cd frontend && npm ci && npm run build && cd ..

# Run (requires PostgreSQL)
DATABASE_URL="postgres://user:pass@localhost:5432/lurkarr?sslmode=disable" ./lurkarr
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *required* | PostgreSQL connection string |
| `LISTEN_ADDR` | `:9705` | HTTP listen address |
| `SECRET_KEY` | *auto-generated* | Session cookie encryption key |
| `CSRF_KEY` | *auto-generated* | CSRF token key (32 bytes hex) |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |
| `PROXY_AUTH` | `false` | Trust proxy authentication headers |
| `PROXY_HEADER` | `Remote-User` | Proxy auth header name |

## Tech Stack

- **Backend:** Go 1.26, PostgreSQL 17, pgx/v5
- **Frontend:** Svelte 5, SvelteKit, TailwindCSS v4, TypeScript
- **Deployment:** Docker (distroless), Helm chart, docker-compose

## API

All endpoints under `/api/` require authentication (session cookie).

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/setup` | First-run setup |
| GET | `/api/stats` | Hunt statistics |
| GET | `/api/instances/{app}` | List app instances |
| GET | `/api/history` | Hunt history |
| GET | `/api/logs` | Query logs |
| WS | `/ws/logs` | Live log stream |
| GET | `/api/prowlarr/indexers` | Prowlarr indexers |
| GET | `/api/sabnzbd/queue` | SABnzbd download queue |

## License

See [LICENSE](LICENSE).
