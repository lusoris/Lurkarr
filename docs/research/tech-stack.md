# Technology Stack Reference

## Backend

| Technology | Version | Purpose | Docs |
|-----------|---------|---------|------|
| Go | 1.25+ | Language | https://go.dev/doc/ |
| pgx/v5 | v5.8.0 | PostgreSQL driver | https://pkg.go.dev/github.com/jackc/pgx/v5 |
| goose/v3 | v3.27.0 | DB migrations | https://pressly.github.io/goose/ |
| gocron/v2 | v2.19.1 | Task scheduling | https://pkg.go.dev/github.com/go-co-op/gocron/v2 |
| otter/v2 | v2.3.0 | L1 cache (W-TinyLFU) | https://pkg.go.dev/github.com/maypok86/otter/v2 |
| gorilla/csrf | v1.7.3 | CSRF protection | https://pkg.go.dev/github.com/gorilla/csrf |
| pquerna/otp | v1.5.0 | TOTP 2FA | https://pkg.go.dev/github.com/pquerna/otp |
| coder/websocket | v1.8.14 | WebSocket | https://pkg.go.dev/github.com/coder/websocket |
| x/time/rate | latest | Rate limiting | https://pkg.go.dev/golang.org/x/time/rate |
| x/crypto | v0.48.0 | bcrypt, scrypt | https://pkg.go.dev/golang.org/x/crypto |
| prometheus/client_golang | v1.23.2 | Metrics | https://pkg.go.dev/github.com/prometheus/client_golang |
| uber.org/mock | v0.6.0 | Mock generation | https://pkg.go.dev/go.uber.org/mock |
| go-qrcode/v2 | v2.2.5 | QR code for TOTP | https://pkg.go.dev/github.com/yeqown/go-qrcode/v2 |

## Planned Backend

| Technology | Version | Purpose | Docs |
|-----------|---------|---------|------|
| uber.org/fx | v1.24.0 | Dependency injection | https://uber-go.github.io/fx/ |
| ogen-go/ogen | v1.20+ | OpenAPI codegen | https://ogen.dev/ |
| @scalar/api-reference | latest | API docs UI | https://github.com/scalar/scalar |

## Frontend

| Technology | Version | Purpose | Docs |
|-----------|---------|---------|------|
| SvelteKit | 2.x | Framework | https://kit.svelte.dev/docs |
| Svelte | 5.x | UI framework | https://svelte.dev/docs |
| TailwindCSS | 4.x | CSS utility | https://tailwindcss.com/docs |
| TypeScript | 5.x | Type safety | https://www.typescriptlang.org/docs/ |
| Vite | 7.x | Build tool | https://vite.dev/guide/ |

## Infrastructure

| Technology | Purpose | Docs |
|-----------|---------|------|
| PostgreSQL 17 | Primary database | https://www.postgresql.org/docs/17/ |
| Docker | Containerization | https://docs.docker.com/ |
| Helm | K8s package manager | https://helm.sh/docs/ |
| Prometheus | Metrics collection | https://prometheus.io/docs/ |
| Grafana | Dashboard/visualization | https://grafana.com/docs/grafana/latest/ |
| Loki | Log aggregation | https://grafana.com/docs/loki/latest/ |
| Coder | Dev environments | https://coder.com/docs |
| GitHub Actions | CI/CD | https://docs.github.com/en/actions |
| GHCR | Container registry | https://docs.github.com/en/packages |

## External Services Integrated

| Service | API Version | Purpose |
|---------|-------------|---------|
| Sonarr | v3 | TV series management |
| Radarr | v3 | Movie management |
| Lidarr | v3 | Music management |
| Readarr | v3 | Book management |
| Whisparr | v3 | Adult content management |
| Prowlarr | v1 | Indexer management |
| SABnzbd | custom REST | Usenet downloader |
| qBittorrent | v2 | Torrent client |
| Transmission | RPC | Torrent client |
| Deluge | JSON-RPC | Torrent client |
| NZBGet | XML-RPC | Usenet downloader |
| Seerr | v1 | Request management |
