# Lurkarr

Lurkarr is a self-hosted media library companion for the \*Arr ecosystem. It automatically searches for missing and upgradeable media across your Sonarr, Radarr, Lidarr, Readarr, Whisparr, and Eros instances — with Prowlarr indexer monitoring, SABnzbd download management, and intelligent queue cleaning.

## Features

- **Multi-App Hunting** — Automatically searches for missing and upgradeable media across Sonarr, Radarr, Lidarr, Readarr, Whisparr (v2), and Eros (Whisparr v3)
- **Smart Search** — Configurable missing/upgrade counts, random or sequential selection, stateful tracking to avoid duplicate searches
- **API Rate Limiting** — Hourly caps per app to prevent indexer overload
- **Scheduling** — Time-based rules to enable/disable apps and adjust caps
- **Prowlarr Integration** — Monitor indexer health, status, and statistics
- **SABnzbd Monitoring** — View and manage the download queue, history, pause/resume
- **Queue Cleaning** — Automatically detect and remove stalled, slow, or metadata-stuck downloads
- **Auto Import** — Score-based deduplication and configurable import rules
- **Live Logs** — Real-time log streaming via WebSocket
- **2FA** — TOTP-based two-factor authentication
- **Modern UI** — SvelteKit + TailwindCSS dark-mode dashboard

## Getting Started

### Docker Compose (recommended)

```yaml
services:
  lurkarr:
    image: ghcr.io/lusoris/lurkarr:latest
    ports:
      - "9705:9705"
    environment:
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

A Helm chart is included under `deploy/helm/lurkarr/`.

```bash
helm install lurkarr deploy/helm/lurkarr \
  --set env.DATABASE_URL="postgres://..." \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=lurkarr.example.com
```

### Building from Source

```bash
# Backend
go build -o lurkarr ./cmd/lurkarr

# Frontend
cd frontend && npm ci && npm run build && cd ..

# Run (requires PostgreSQL)
DATABASE_URL="postgres://user:pass@localhost:5432/lurkarr?sslmode=disable" ./lurkarr
```

## Configuration

All configuration is done via environment variables.

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string (**required**) |
| `LISTEN_ADDR` | `:9705` | HTTP listen address |
| `SECRET_KEY` | *auto-generated* | Session cookie encryption key |
| `CSRF_KEY` | *auto-generated* | CSRF token key (32 bytes hex) |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |
| `PROXY_AUTH` | `false` | Trust reverse-proxy authentication headers |
| `PROXY_HEADER` | `Remote-User` | Header name for proxy auth |

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go, PostgreSQL 17, pgx/v5 |
| Frontend | Svelte 5, SvelteKit, TailwindCSS v4, TypeScript |
| Deployment | Docker (distroless), Helm, docker-compose |

## Contributing

Contributions are welcome. Please open an issue first to discuss larger changes.

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE).
