# Bazarr Integration Research

## Overview
Bazarr is an automated subtitle manager for Sonarr and Radarr. It monitors wanted subtitles,
searches subtitle providers, and downloads/manages them. Unlike the *arr apps, it doesn't
participate in the search/grab/download pipeline — it's a companion service.

## Integration Type
**Service integration** (like Prowlarr/Seerr), not a full `AppType`. Reasons:
- Single instance (not multi-instance like Radarr/Sonarr)
- No lurking/search engine participation (subtitles, not media)
- No queue cleaner participation (no download queue)
- No cross-instance dedup participation

Follows the `ProwlarrHandler`/`SeerrHandler` pattern: dedicated handler struct,
single-row settings table, dedicated client package.

## Docker Image
`ghcr.io/hotio/bazarr:release` (hotio, per convention)

## Authentication
API key via `X-API-Key` header (same as *arr apps). Stored in container at
`/config/config/config.ini` under `[auth]` section:
```ini
[auth]
apikey = <hex-string>
```

Extraction:
```bash
docker exec lurkarr-bazarr-1 grep -i apikey /config/config/config.ini | head -1 | cut -d= -f2 | tr -d ' '
```

## API Endpoints

### System
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/system/status` | Status info (version, startup time, bazarr/python/sonarr paths) |
| GET | `/api/system/health` | Health check items |
| GET | `/api/system/languages` | Available subtitle languages |
| GET | `/api/system/languages/profiles` | Configured language profiles |

### Subtitles — Episodes
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/episodes/wanted` | Episodes missing subtitles |
| GET | `/api/episodes` | List all tracked episodes (with subtitle status) |
| PATCH | `/api/episodes/subtitles` | Trigger subtitle search for episodes |

### Subtitles — Movies
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/movies/wanted` | Movies missing subtitles |
| GET | `/api/movies` | List all tracked movies |
| PATCH | `/api/movies/subtitles` | Trigger subtitle search for movies |

### Providers
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/providers` | Configured subtitle providers |
| GET | `/api/providers/episodes/{id}` | Search providers for episode subtitles |
| GET | `/api/providers/movies/{id}` | Search providers for movie subtitles |

### History
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/history/episodes` | Episode subtitle download history |
| GET | `/api/history/movies` | Movie subtitle download history |

## Lurkarr Dashboard Integration
Display on the dashboard / monitoring:
- **Wanted counts** — Episodes + movies missing subtitles
- **Health status** — Provider connectivity
- **History** — Recent subtitle downloads

## Database Schema
Single-row settings following ProwlarrSettings pattern:
```sql
CREATE TABLE bazarr_settings (
    id         SERIAL PRIMARY KEY,
    url        TEXT NOT NULL DEFAULT '',
    api_key    TEXT NOT NULL DEFAULT '',
    enabled    BOOLEAN NOT NULL DEFAULT false,
    timeout    INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO bazarr_settings (id) VALUES (1);
```

## Implementation Plan
1. `internal/bazarrclient/client.go` — HTTP client with `TestConnection`, `GetStatus`, `GetWantedEpisodes`, `GetWantedMovies`, `GetHealth`, `GetHistory`
2. `internal/database/models.go` — `BazarrSettings` struct
3. `internal/database/migrations/XXX_bazarr_settings.sql` — migration
4. `internal/api/bazarr.go` — handlers: GetSettings, UpdateSettings, TestConnection, GetWanted, GetHealth
5. `internal/api/bazarr_settings.go` — settings CRUD (following prowlarr_settings.go pattern)
6. `internal/server/server.go` — route registration
7. Frontend: Bazarr section on Apps/Connections page (service block, like Prowlarr)
