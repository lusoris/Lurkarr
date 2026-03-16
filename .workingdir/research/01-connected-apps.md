# Connected Apps — Deep Research Reference

## 1. Arr Stack (Servarr Ecosystem)

All *arr apps share a common Servarr API base (v3 for most, v1 for Prowlarr).

### 1.1 Sonarr (TV Series Management)
- **API Version**: v3 (v4 beta available with breaking changes)
- **Port**: 8989
- **Key Endpoints Used by Lurkarr**:
  - `GET /api/v3/wanted/missing` — paginated missing episodes
  - `GET /api/v3/wanted/cutoff` — upgradeable episodes (cutoff unmet)
  - `POST /api/v3/command` — trigger searches (`EpisodeSearch`, `SeriesSearch`)
  - `GET /api/v3/queue` — download queue with status tracking
  - `DELETE /api/v3/queue/{id}` — remove from queue (with blocklist option)
  - `DELETE /api/v3/queue/bulk` — bulk queue removal
  - `GET /api/v3/health` — system health checks
  - `GET /api/v3/system/status` — version/build info
  - `GET /api/v3/history` — download history with pagination
  - `GET /api/v3/manualimport` — manual import for stuck items
  - `GET /api/v3/series` — all series metadata
  - `GET /api/v3/episode?seriesId={id}` — episodes for series
- **Auth**: `X-Api-Key` header
- **Best Practices**:
  - Respect rate limits — avoid hammering missing/cutoff endpoints
  - Use `pageSize` wisely (50-250 max per request)
  - Include `monitored=true` filter to avoid searching unmonitored
  - Check `trackedDownloadStatus` and `trackedDownloadState` for queue health
  - Use `includeUnknownSeriesItems=true` for complete queue view
  - v4 note: Sonarr v4 is now stable; API differences exist in episode search commands

### 1.2 Radarr (Movie Management)
- **API Version**: v3
- **Port**: 7878
- **Key Endpoints Used by Lurkarr**:
  - `GET /api/v3/wanted/missing` — missing movies (sortKey: `digitalRelease`)
  - `GET /api/v3/wanted/cutoff` — upgradeable movies
  - `POST /api/v3/command` — trigger `MoviesSearch`
  - `GET /api/v3/queue` — queue management
  - `DELETE /api/v3/queue/{id}` / `DELETE /api/v3/queue/bulk`
  - `GET /api/v3/health`, `GET /api/v3/system/status`
  - `GET /api/v3/movie` — all movies
  - `GET /api/v3/manualimport`
- **Notes**: Uses `digitalRelease` and `physicalRelease` dates for sorting

### 1.3 Lidarr (Music Management)
- **API Version**: v3 (some endpoints use v1 internally)
- **Port**: 8686
- **Key Differences**:
  - Missing endpoint: `GET /api/v1/wanted/missing`
  - Cutoff endpoint: `GET /api/v1/wanted/cutoff`
  - Search command: `AlbumSearch` with `albumIds`
  - Has `artist` and `album` resources instead of series/movie

### 1.4 Readarr (Book Management — RETIRED)
- **Status**: **Officially retired** by the Servarr team
- **Reason**: Metadata became unusable; no active development
- **Port**: 8787
- **API**: v1 (not v3 like other apps)
- **Impact on Lurkarr**: Should gracefully handle connections to Readarr; consider displaying retirement notice in UI
- **Alternative**: Third-party metadata mirrors exist (e.g., `rreading-glasses`)

### 1.5 Whisparr v2 (Sonarr-based)
- **Port**: 6969
- **API Version**: v3 (Sonarr-compatible)
- **Key Differences**:
  - Studios = "series", Scenes/Movies = "episodes"
  - Uses `releaseDate` instead of `airDateUtc`
  - Search: same `EpisodeSearch` command as Sonarr
- **All Sonarr endpoints work identically**

### 1.6 Eros (Whisparr v3)
- **Port**: 6969 (configurable)
- **API Version**: v3 (Radarr-based fork)
- **Key Differences from Whisparr v2**:
  - Movie-based model (like Radarr) instead of series-based
  - Uses `MoviesSearch` command
  - Different missing/cutoff endpoints

### 1.7 Prowlarr (Indexer Management)
- **API Version**: v1
- **Port**: 9696
- **Key Endpoints**:
  - `GET /api/v1/indexer` — list all indexers
  - `GET /api/v1/indexerstats` — indexer statistics
  - `GET /api/v1/indexerstatus` — indexer health status
  - `GET /api/v1/health` — system health
  - `GET /api/v1/system/status` — version info
  - `GET /api/v1/history` — search/grab history
- **Integration Pattern**: Read-only monitoring (Lurkarr doesn't manage indexers)

### 1.8 Bazarr (Subtitle Management)
- **API**: REST JSON
- **Port**: 6767
- **Key Endpoints**:
  - `GET /api/system/status` — version/health
  - `GET /api/episodes/wanted` — missing subtitles for episodes
  - `GET /api/movies/wanted` — missing subtitles for movies
  - `GET /api/system/health` — system health
- **Auth**: `X-API-KEY` header

### 1.9 Kapowarr (Comic Management)  
- **Custom REST API**
- **Port**: 5656
- **Similar to Readarr but for comics/manga**

### 1.10 Shoko (Anime Management)
- **Custom REST API**
- **Port**: 8111
- **Specialized for anime with AniDB/MyAnimeList integration**

---

## 2. Download Clients

### 2.1 qBittorrent
- **API Version**: WebUI API v2 (v2.8.3+ for latest features)
- **Port**: 8080 (default)
- **Auth**: Cookie-based (SID) via `POST /api/v2/auth/login`
- **Key Endpoints for Lurkarr**:
  - `GET /api/v2/torrents/info` — list torrents with filtering (state, category, tag, hash)
  - `GET /api/v2/torrents/properties` — detailed torrent properties
  - `GET /api/v2/torrents/files` — file list with priorities
  - `GET /api/v2/transfer/info` — global transfer stats
  - `POST /api/v2/torrents/delete` — remove torrents (with `deleteFiles` option)
  - `POST /api/v2/torrents/pause` / `POST /api/v2/torrents/resume`
  - `POST /api/v2/torrents/setShareLimits` — ratio/seeding time limits
  - `GET /api/v2/sync/maindata` — efficient delta sync
  - `GET /api/v2/app/preferences` — configuration
- **Torrent States**: `error`, `missingFiles`, `uploading`, `pausedUP`, `queuedUP`, `stalledUP`, `checkingUP`, `forcedUP`, `allocating`, `downloading`, `metaDL`, `pausedDL`, `queuedDL`, `stalledDL`, `checkingDL`, `forcedDL`, `checkingResumeData`, `moving`, `unknown`
- **Best Practices**:
  - Use `Referer` header matching the WebUI domain
  - Handle session expiry (re-login on 403)
  - Use `isPrivate` field to detect private tracker torrents
  - `sync/maindata` with `rid` for efficient polling
  - Check `seeding_time` for seeding rule enforcement

### 2.2 Transmission
- **Protocol**: JSON-RPC over HTTP
- **Port**: 9091
- **Auth**: Basic HTTP auth + CSRF token (`X-Transmission-Session-Id`)
- **Key Methods**:
  - `torrent-get` — list torrents with fields selection
  - `torrent-remove` — remove torrents
  - `torrent-start`, `torrent-stop`
  - `session-get` — server configuration
  - `session-stats` — transfer statistics
- **Best Practices**:
  - Must handle 409 response to get CSRF token, then retry
  - Select only needed fields for performance

### 2.3 Deluge
- **Protocol**: JSON-RPC
- **Port**: 8112
- **Auth**: `auth.login` method with password
- **Key Methods**:
  - `core.get_torrents_status` — list with filters
  - `core.remove_torrent` — remove with data option
  - `core.pause_torrent`, `core.resume_torrent`
  - `core.get_session_status` — transfer stats
- **Note**: Returns request IDs that must be tracked

### 2.4 SABnzbd (Usenet)
- **Protocol**: REST API with `mode` parameter
- **Port**: 8080 (default)
- **Auth**: API key in query string (`apikey=`)
- **Key Endpoints**:
  - `mode=queue` — full queue with slots, speed, disk space
  - `mode=history` — completed/failed downloads with stats
  - `mode=server_stats` — per-server download statistics
  - `mode=status` — full status including server connections
  - `mode=pause` / `mode=resume` — global pause/resume
  - `mode=queue&name=delete&value=NZO_ID` — delete queue item
  - `mode=queue&name=priority&value=NZO_ID&value2=PRIORITY`
  - `mode=history&name=delete&value=NZO_ID`
  - `mode=retry&value=NZO_ID` — retry failed item
- **Output**: JSON (default) or XML via `output=json`
- **Priority Values**: -100 (Default), -3 (Duplicate), -2 (Paused), -1 (Low), 0 (Normal), 1 (High), 2 (Force)
- **Best Practices**:
  - Always include `output=json` 
  - Use `start`/`limit` for pagination
  - Use `last_history_update` for efficient polling
  - Server stats provide per-server article success rates

### 2.5 NZBGet (Usenet)
- **Protocol**: JSON-RPC / XML-RPC
- **Port**: 6789
- **Auth**: Basic HTTP auth
- **Key Methods**:
  - `listgroups` — queue items
  - `history` — completed items
  - `status` — server status
  - `editqueue` — modify queue items

### 2.6 rTorrent
- **Protocol**: XML-RPC (SCGI)
- **Auth**: Varies (usually behind nginx/ruTorrent)
- **Uses**: `github.com/autobrr/go-rtorrent` Go client library

### 2.7 µTorrent
- **Protocol**: WebUI API (legacy)
- **Port**: Configurable
- **Auth**: Basic HTTP auth + CSRF token

---

## 3. Seerr (Media Request Management)

### Overseerr / Jellyseerr
- **API Version**: v1
- **Base URL**: `/api/v1`
- **Port**: 5055
- **Auth**: `X-Api-Key` header or cookie-based session
- **Key Endpoints for Lurkarr**:
  - `GET /api/v1/request` — all media requests (paginated)
  - `GET /api/v1/request/count` — request counts
  - `POST /api/v1/request/{requestId}/{status}` — approve/decline requests
  - `DELETE /api/v1/request/{requestId}` — delete request
  - `GET /api/v1/status` — server status (public, no auth)
  - `GET /api/v1/settings/main` — main settings
  - `GET /api/v1/media` — media items
- **Request Statuses**: pending, approved, declined, available
- **Lurkarr Integration**:
  - Sync pending requests
  - Auto-approve based on configurable rules
  - Cleanup old/completed requests
  - Duplicate detection across request sources

---

## 4. Notification Providers

### 4.1 Discord
- **Protocol**: Webhook POST
- **Format**: JSON with `embeds` array
- **Rate Limit**: 30 requests/60 seconds per webhook
- **Best Practices**: Rich embeds with color-coded severity, link to Lurkarr UI

### 4.2 Telegram
- **Protocol**: Bot API HTTPS
- **Endpoint**: `https://api.telegram.org/bot{token}/sendMessage`
- **Format**: Markdown or HTML parse mode
- **Best Practices**: Use `parse_mode=HTML` for rich formatting

### 4.3 Pushover
- **Protocol**: HTTPS POST
- **Endpoint**: `https://api.pushover.net/1/messages.json`
- **Fields**: `token`, `user`, `message`, `title`, `priority`, `sound`

### 4.4 Gotify
- **Protocol**: REST API
- **Endpoint**: `POST /message`
- **Auth**: Application token
- **Supports**: Markdown extras for rich content

### 4.5 ntfy
- **Protocol**: REST / SSE / WebSocket
- **Endpoint**: `POST /{topic}`
- **Features**: Actions, priority levels, tags, click URLs

### 4.6 Apprise
- **Protocol**: Apprise API gateway
- **Endpoint**: `POST /notify`
- **Supports**: 90+ notification services through unified API

### 4.7 Email (SMTP)
- **Standard**: SMTP with TLS/STARTTLS
- **Best Practices**: Template-based HTML emails, plain-text fallback

### 4.8 Webhook
- **Protocol**: HTTP POST
- **Format**: JSON payload with event type, details, timestamp
- **Best Practices**: HMAC signature verification, retry with backoff

---

## 5. Monitoring Stack

### 5.1 Prometheus
- **Endpoint**: `/metrics` (OpenMetrics format)
- **Metrics Types**: Counter, Gauge, Histogram, Summary
- **Naming Convention**: `lurkarr_<subsystem>_<metric>_<unit>`
- **Best Practices**: Use labels sparingly, avoid high cardinality

### 5.2 Grafana
- **Dashboards**: Pre-built for overview, system/runtime, Loki logs
- **Data Sources**: Prometheus, Loki
- **Best Practices**: Alert rules, variable templates, annotation markers

### 5.3 Loki
- **Protocol**: HTTP push API
- **Integration**: Via Promtail or Docker logging driver
- **Label Strategy**: Match Prometheus labels for correlation
