# Arr Stack API Reference

> All *arr apps share a common Servarr API base. Version: v3

## Common Endpoints (Sonarr, Radarr, Lidarr, Readarr, Whisparr)

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

## Overseerr/Jellyseerr API

> Docs: https://api-docs.overseerr.dev/

### Authentication
- Header: `X-Api-Key: <apikey>`

### Key Endpoints

```
GET  /api/v1/request?take=20&skip=0&sort=added  # List requests
GET  /api/v1/request/count                       # Request counts
POST /api/v1/request/{id}/approve                # Approve request
POST /api/v1/request/{id}/decline                # Decline request
GET  /api/v1/media?take=20&skip=0                # Media library
GET  /api/v1/user                                # Users
GET  /api/v1/status                              # Server status
```

### Request Status Codes
| Code | Meaning |
|------|---------|
| 1 | Pending |
| 2 | Approved |
| 3 | Declined |
| 4 | Available |
| 5 | Completed (processing done) |
