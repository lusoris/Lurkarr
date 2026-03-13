# Feature Matrix — Lurkarr

**Date:** 2026-03-13 | **Branch:** develop

## Backend ↔ Frontend Wiring Status

### ✅ Fully Wired (Backend + Frontend + API)

| Feature | Backend | API | Frontend | Tests |
|---------|---------|-----|----------|-------|
| User login (local) | ✅ bcrypt | POST /auth/login | ✅ /login | ✅ |
| User setup (initial admin) | ✅ | POST /auth/setup | ✅ /login | ✅ |
| OIDC login | ✅ PKCE | GET/POST /auth/oidc/* | ✅ /login | ✅ |
| TOTP 2FA enable/disable/verify | ✅ | POST /auth/2fa/* | ✅ /user | ✅ |
| Recovery codes | ✅ | POST /auth/2fa/recovery-codes | ✅ /user + /login | ⚠️ New |
| Session management | ✅ | GET/DEL /sessions | ✅ /user | ⚠️ New |
| Admin user CRUD | ✅ | /admin/users/* | ✅ /admin/users | ⚠️ New |
| Arr instance CRUD | ✅ | /instances/* | ✅ /apps | ✅ |
| Arr instance health | ✅ | /instances/{id}/health | ✅ /apps cards | ✅ |
| Arr instance test | ✅ | POST /instances/test | ✅ /apps modal | ✅ |
| Download client CRUD | ✅ | /download-clients/* | ✅ /apps | ✅ |
| Download client health | ✅ | /download-clients/{id}/health | ✅ /apps + /downloads | ✅ |
| Prowlarr settings | ✅ | /prowlarr/settings | ✅ /apps modal | ✅ |
| Seerr settings | ✅ | /seerr/settings | ✅ /apps modal | ✅ |
| Lurk settings per app | ✅ | /settings/{app} | ✅ /lurk | ✅ |
| General settings | ✅ | /settings/general | ✅ /settings | ✅ |
| Schedules CRUD | ✅ | /schedules/* | ✅ /scheduling | ✅ |
| Schedule execution history | ✅ | /schedules/history | ✅ /scheduling modal | ✅ |
| Lurk history (search) | ✅ FTS | /history | ✅ /history | ✅ |
| Dashboard stats | ✅ | /stats | ✅ / (dashboard) | ✅ |
| Hourly API caps | ✅ | /stats/hourly-caps | ✅ / (dashboard) | ✅ |
| Stats reset | ✅ | POST /stats/reset | ✅ / (dashboard) | ✅ |
| Queue cleaner settings | ✅ | /queue/settings/{app} | ✅ /queue tab 1 | ✅ |
| Scoring profile | ✅ | /queue/scoring/{app} | ✅ /queue tab 2 | ✅ |
| Blocklist log | ✅ | /queue/blocklist/{app} | ✅ /queue tab 3 | ✅ |
| Import log | ✅ | /queue/imports/{app} | ✅ /queue tab 4 | ✅ |
| Notification providers CRUD | ✅ | /notifications/providers/* | ✅ /notifications | ✅ |
| Notification test | ✅ | POST /notifications/providers/{id}/test | ✅ /notifications | ✅ |
| Health/readiness probes | ✅ | /healthz, /readyz | ✅ /monitoring | ✅ |
| OpenAPI docs | ✅ | /api/docs, /api/spec | ✅ /monitoring links | ✅ |
| Prometheus metrics | ✅ | /metrics | ✅ /monitoring links | ✅ |

### ⚠️ Backend Exists, Frontend Missing

| Feature | Backend Endpoint | Status |
|---------|-----------------|--------|
| Blocklist sources CRUD | GET/POST/PUT/DEL /blocklist/sources/* | No UI for managing sources |
| Blocklist rules CRUD | GET/POST/DEL /blocklist/rules/* | No UI for managing rules |
| Seerr request list | GET /seerr/requests | No display in frontend |
| Seerr request count | GET /seerr/count | No display in frontend |
| Prowlarr indexer list | GET /prowlarr/indexers | No display (config only) |
| Prowlarr indexer stats | GET /prowlarr/stats | No display |
| SABnzbd queue | GET /sabnzbd/queue | No display (only generic DL client view) |
| SABnzbd history | GET /sabnzbd/history | No display |
| SABnzbd stats | GET /sabnzbd/stats | No display |
| SABnzbd pause/resume | POST /sabnzbd/pause/resume | No UI buttons |
| State reset per instance | POST /state/{app}/instances/{id}/reset | Only global reset |
| History delete by app | DELETE /history/{app} | No delete button in UI |
| Download client settings (legacy) | GET/PUT /queue/download-client/{app} | Uses new multi-instance instead |

### ❌ Not Implemented (Backend or Frontend)

| Feature | Status |
|---------|--------|
| WebAuthn/Passkeys | Model only, no handlers/endpoints/UI |
| Proxy auth UI config | Backend supports it via env vars only |
| OIDC group refresh on request | Groups cached until session ends |
| Real-time WebSocket updates | Not implemented |
| Dark/Light mode toggle | Hardcoded dark mode |

## Background Services Status

| Service | Running | Config | Monitoring |
|---------|---------|--------|------------|
| Lurking Engine | ✅ On boot | Per-app settings | Prometheus metrics |
| Queue Cleaner | ✅ On boot | Per-app settings | Prometheus metrics |
| Auto-Importer | ✅ On boot | 5min interval | Prometheus metrics |
| Seerr Sync | ✅ On boot | Configurable interval | Logs only |
| Health Poller | ✅ On boot | 5min interval | Internal (not exposed) |
| Maintenance | ✅ On boot | Hourly | Logs only |
| Scheduler | ✅ On boot | Cron-based | Execution history |
