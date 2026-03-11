# Arr Stack API Reference

> All *arr apps share a common Servarr API base. Version: v3

## Common Endpoints (Sonarr, Radarr, Lidarr, Readarr)

### Authentication
- Header: `X-Api-Key: <apikey>`
- All endpoints require auth

### Queue
- `GET /api/v3/queue?page=1&pageSize=50&includeUnknownSeriesItems=true`
- `DELETE /api/v3/queue/{id}?removeFromClient=true&blocklist=true`
- `DELETE /api/v3/queue/bulk` (body: `{ids: [1,2,3]}`)

### Wanted/Missing
- Sonarr: `GET /api/v3/wanted/missing?page=1&pageSize=50&sortKey=airDateUtc&sortDirection=descending&monitored=true`
- Radarr: `GET /api/v3/wanted/missing?page=1&pageSize=50&sortKey=digitalRelease&sortDirection=descending&monitored=true`

### Search (trigger)
- Sonarr: `POST /api/v3/command` body: `{"name": "EpisodeSearch", "episodeIds": [1,2,3]}`
- Radarr: `POST /api/v3/command` body: `{"name": "MoviesSearch", "movieIds": [1,2,3]}`

### Cutoff Unmet (upgrades)
- `GET /api/v3/wanted/cutoff?page=1&pageSize=50&monitored=true`

### Manual Import
- `GET /api/v3/manualimport?downloadId=<id>&filterExistingFiles=true`

### System Health
- `GET /api/v3/health` — returns array of health checks
- `GET /api/v3/system/status` — version, branch, build time

### History
- `GET /api/v3/history?pageSize=50&page=1&sortKey=date&sortDirection=descending`

## Sonarr-Specific

- `GET /api/v3/series` — all series
- `GET /api/v3/episode?seriesId={id}` — episodes for series
- Queue fields: `trackedDownloadStatus`, `trackedDownloadState`, `statusMessages`

## Radarr-Specific

- `GET /api/v3/movie` — all movies
- `GET /api/v3/movie/{id}` — single movie
- Radarr uses `digitalRelease` and `physicalRelease` dates

## Whisparr v2 (Sonarr-based)

> Source: https://github.com/Whisparr/Whisparr (branch: `v2-develop`)
> Port: 6969, API version: v3

Whisparr v2 is built on Sonarr. Studios are "series", scenes and movies are "episodes".

### Core Resources

- `GET /api/v3/series` — all studios (series)
- `GET /api/v3/series/{id}` — single studio
- `GET /api/v3/episode?seriesId={id}` — scenes/movies for a studio
- `GET /api/v3/episode/{id}` — single scene/movie

### Wanted/Missing

```
GET /api/v3/wanted/missing?sortKey=releaseDate&sortDirection=descending&pageSize=1000&monitored=true
```

Returns `EpisodeResourcePagingResource`:
```json
{
  "page": 1,
  "pageSize": 1000,
  "totalRecords": 42,
  "records": [
    {
      "id": 1,
      "seriesId": 10,
      "title": "Scene Title",
      "seasonNumber": 1,
      "monitored": true,
      "hasFile": false,
      "releaseDate": "2024-01-15"
    }
  ]
}
```

### Cutoff Unmet (Upgrades)

```
GET /api/v3/wanted/cutoff?sortKey=releaseDate&sortDirection=descending&pageSize=1000&monitored=true
```

Same paginated response format as wanted/missing.

### Search Commands

```json
POST /api/v3/command
{"name": "EpisodeSearch", "episodeIds": [1, 2, 3]}
{"name": "SeasonSearch", "seriesId": 10, "seasonNumber": 1}
{"name": "SeriesSearch", "seriesId": 10}
```

### Special Fields

- `Actor` resource has `tpdbId` (ThePornDB ID) and `Gender` enum
- `EpisodeResource` has `releaseDate` (not `airDateUtc` like Sonarr)
- Queue endpoint same as Sonarr: `GET /api/v3/queue?page=1&pageSize=50`

## Whisparr v3 / Eros (Radarr-based)

> Source: https://github.com/Whisparr/Whisparr (branch: `eros`)
> Port: 6969, API version: v3

Eros is built on Radarr. Scenes and movies are individual `MovieResource` items with an `itemType` field.

### Core Resources

- `GET /api/v3/movie` — all items (scenes + movies)
- `GET /api/v3/movie/{id}` — single item
- `GET /api/v3/performer` — all performers
- `GET /api/v3/performer/{id}` — single performer
- `GET /api/v3/studio` — all studios
- `GET /api/v3/studio/{id}` — single studio

### MovieResource

```json
{
  "id": 1,
  "title": "Title",
  "itemType": "scene",
  "monitored": true,
  "hasFile": false,
  "studio": {"id": 5, "title": "Studio Name"},
  "credits": [{"personName": "...", "character": "..."}],
  "stashId": "abc-123"
}
```

`itemType` values: `"movie"` or `"scene"`

### PerformerResource

```json
{
  "id": 1,
  "fullName": "Name",
  "gender": "female",
  "ethnicity": "...",
  "careerStart": 2020,
  "careerEnd": null,
  "hasMovies": true,
  "hasScenes": true,
  "sceneCount": 42,
  "foreignId": "..."
}
```

### StudioResource

```json
{
  "id": 1,
  "title": "Studio Name",
  "hasMovies": true,
  "hasScenes": true,
  "totalSceneCount": 100,
  "foreignId": "..."
}
```

### Wanted/Missing — NOT AVAILABLE

Eros has **no** `/api/v3/wanted/missing` or `/api/v3/wanted/cutoff` endpoints.
To find missing items, fetch all from `/api/v3/movie` and filter client-side:
- Missing: `monitored == true && hasFile == false`
- No cutoff/upgrade equivalent exists

### Search Commands

```json
POST /api/v3/command
{"name": "MoviesSearch", "movieIds": [1, 2, 3]}
```

### Lookup

- `GET /api/v3/lookup/scene?term=query` — search for scenes
- `GET /api/v3/lookup/movie?term=query` — search for movies

### Naming Config

```json
{
  "renameMovies": true,
  "renameScenes": true,
  "standardMovieFormat": "{Movie Title} ({Release Year})",
  "standardSceneFormat": "{Scene Title} ({Release Year})",
  "sceneFolderFormat": "...",
  "sceneImportFolderFormat": "..."
}
```

### Import Exclusion Types

`ImportExclusionType` enum: `scene`, `movie`, `studio`, `performer`

### Add Movie Methods

`AddMovieMethod` enum: `manual`, `list`, `performer`, `studio`

## Prowlarr-Specific

- `GET /api/v1/indexer` — list indexers
- `GET /api/v1/indexerstats` — indexer statistics
- `GET /api/v1/search?query=test&indexerIds=1,2` — live search
- API version: v1 (not v3)

## SABnzbd API

> Protocol: HTTP REST with `apikey` query param
> Docs: https://sabnzbd.org/wiki/advanced/api

### Key Endpoints

```
GET /api?mode=queue&apikey=KEY&output=json          # Queue status
GET /api?mode=history&apikey=KEY&output=json         # Download history
GET /api?mode=status&apikey=KEY&output=json          # Server status
GET /api?mode=pause&apikey=KEY                       # Pause all
GET /api?mode=resume&apikey=KEY                      # Resume all
GET /api?mode=queue&name=delete&value=ID&apikey=KEY  # Delete from queue
```

### Queue Item Fields
- `status`: Downloading, Queued, Grabbing, Paused
- `timeleft`: "0:02:30" or "1.23:59:59" (D.HH:MM:SS)
- `mb`, `mbleft`: total/remaining size
- `priority`: -100 (force), -1 (low), 0 (normal), 1 (high), 2 (force)

## qBittorrent Web API

> Docs: https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)

### Authentication
```
POST /api/v2/auth/login   body: username=admin&password=pass
```
Returns cookie `SID=xxx` for subsequent requests.

### Key Endpoints

```
GET  /api/v2/torrents/info?filter=all
GET  /api/v2/torrents/properties?hash=xxx
POST /api/v2/torrents/pause       body: hashes=xxx|yyy
POST /api/v2/torrents/resume      body: hashes=xxx|yyy
POST /api/v2/torrents/delete      body: hashes=xxx&deleteFiles=true
GET  /api/v2/transfer/info        # Global transfer info (speeds)
```

### Torrent Fields
- `state`: uploading, downloading, pausedUP, pausedDL, stalledUP, stalledDL, error
- `ratio`: current seed ratio
- `seeding_time`: seconds seeding
- `dlspeed`, `upspeed`: bytes/sec

## Transmission RPC

> Docs: https://github.com/transmission/transmission/blob/main/docs/rpc-spec.md

### Protocol
- POST to `/transmission/rpc`
- Header: `X-Transmission-Session-Id: <session-id>` (409 to get new one)
- Body: JSON RPC

```json
{
  "method": "torrent-get",
  "arguments": {
    "fields": ["id", "name", "status", "rateDownload", "percentDone", "seedRatioLimit"]
  }
}
```

### Methods
- `torrent-get` — query torrents
- `torrent-start` / `torrent-stop` — pause/resume
- `torrent-remove` — delete (with `delete-local-data: true` to delete files)
- `session-get` — global settings

## Deluge Web UI API

> JSON-RPC over HTTP

```json
POST /json
{
  "method": "core.get_torrents_status",
  "params": [{"id": ["hash1"]}, ["name", "progress", "ratio"]],
  "id": 1
}
```

### Auth
```json
{"method": "auth.login", "params": ["password"], "id": 0}
```

## NZBGet XML-RPC

> Docs: https://nzbget.com/info/api/

### Protocol
- POST to `/xmlrpc` or `/jsonrpc`
- Basic auth: `username:password`

### Key Methods

```
listgroups         # Active downloads
history            # Completed downloads
pausedownload(id)  # Pause
resumedownload(id) # Resume
editqueue("GroupDelete", "", [id])  # Delete
status             # Server status (speed, remaining)
```

## Seerr API

> Source: https://github.com/seerr-team/seerr
> Port: 5055, API version: v1

### Authentication
- Header: `X-Api-Key: <apikey>`
- Or cookie-based session auth via `POST /api/v1/auth/local`

### Key Endpoints

```
GET  /api/v1/request?take=20&skip=0&sort=added  # List requests
GET  /api/v1/request/count                       # Request counts
POST /api/v1/request/{id}/approve                # Approve request
POST /api/v1/request/{id}/decline                # Decline request
GET  /api/v1/media?take=20&skip=0                # Media library
GET  /api/v1/user                                # Users
GET  /api/v1/status                              # Server status
GET  /api/v1/settings/main                       # Main settings
```

### MediaStatus

| Code | Constant |
|------|----------|
| 1 | Unknown |
| 2 | Pending |
| 3 | Processing |
| 4 | Partially Available |
| 5 | Available |
| 6 | Deleted |

### MediaRequestStatus

| Code | Constant |
|------|----------|
| 1 | Pending |
| 2 | Approved |
| 3 | Declined |
