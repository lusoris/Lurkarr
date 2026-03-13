# Frontend Analysis — Lurkarr

**Date:** 2026-03-13 | **Branch:** develop

## Tech Stack

- **SvelteKit 2.50.2** with **Svelte 5.51** (runes: `$state`, `$derived`, `$effect`)
- **Tailwind CSS 4.2.1** with custom theme (lurk-* blue, surface-* grayscale)
- **TypeScript 5.9.3** (strict mode)
- **Vite 7.3.1** with static adapter (prerendered SPA)
- **Fonts**: Inter (sans), JetBrains Mono (mono)
- **Icons**: Heroicons outline 24x24 (inline SVG paths)
- **Dark mode only** (hardcoded in app.html)

## Component Library (`$lib/components/ui/`)

| Component | Props | Used In |
|-----------|-------|---------|
| **Button** | variant (primary/secondary/danger/ghost), size (sm/md/lg), disabled, loading | Every page (20+) |
| **Input** | value (bindable), type, placeholder, label, hint, error, disabled | Settings, modals (15+) |
| **Card** | onclick (optional), class | Every page (20+) |
| **Badge** | variant (default/success/warning/error/info) | 8 pages |
| **Toggle** | checked (bindable), label, hint, disabled, onchange | 4 pages |
| **Modal** | open (bindable), title, onclose | 5 pages |
| **Sidebar** | N/A (app-specific) | Root layout |
| **Toast** | N/A (global store) | Root layout |

## Stores

| Store | Type | State | Methods |
|-------|------|-------|---------|
| `auth.svelte.ts` | Svelte 5 runes | user, loading | check(), login(), logout() |
| `toast.svelte.ts` | Svelte 5 runes | toasts[] | success(), error(), info(), warning(), remove() |

## Pages & Status

| # | Route | Title | Lines | API Wired | Status |
|---|-------|-------|-------|-----------|--------|
| 1 | `/` | Dashboard | 161 | ✅ stats, hourly-caps, instances | ✅ Complete |
| 2 | `/login` | Login | ~200 | ✅ setup, login, oidc/info | ✅ Complete |
| 3 | `/apps` | Connections | ~600 | ✅ instances, download-clients, prowlarr, seerr | ✅ Complete |
| 4 | `/lurk` | Lurk Settings | 110 | ✅ settings/{app} | ✅ Complete |
| 5 | `/scheduling` | Schedules | ~180 | ✅ schedules, schedules/history | ✅ Complete |
| 6 | `/history` | History | ~100 | ✅ history (search, debounced) | ✅ Complete |
| 7 | `/downloads` | Downloads | ~80 | ✅ download-clients, health | ✅ Complete |
| 8 | `/queue` | Queue Management | 472 | ✅ queue/settings, scoring, blocklist, imports | ✅ Complete |
| 9 | `/notifications` | Notifications | 306 | ✅ notifications/providers CRUD, test | ✅ Complete |
| 10 | `/monitoring` | Monitoring | ~80 | ✅ healthz, readyz | ✅ Complete |
| 11 | `/settings` | Settings | 62 | ✅ settings/general | ✅ Complete |
| 12 | `/user` | Profile | ~250 | ✅ user, sessions, 2FA | ✅ Complete |
| 13 | `/admin/users` | User Management | ~200 | ✅ admin/users CRUD | ✅ Complete |

**All 13 pages: ✅ Complete and API-wired**

## Layout Quality Assessment

| Page | Layout Type | Quality | Notes |
|------|-------------|---------|-------|
| Dashboard | Card grid | ⭐⭐⭐⭐⭐ | Responsive, clean |
| Login | Centered form | ⭐⭐⭐⭐⭐ | Adaptive (setup/login/2FA/OIDC) |
| Connections | Sections + modals | ⭐⭐⭐⭐ | Well organized but see issue #1 |
| Lurk Settings | App tabs + flat form | ⭐⭐⭐ | **Needs sections** — just a flat list |
| Scheduling | List + modal | ⭐⭐⭐⭐ | Good |
| History | Search + table | ⭐⭐⭐⭐ | Simple, effective |
| Downloads | Client cards | ⭐⭐⭐⭐ | Good |
| Queue Management | Tabs + grouped sections | ⭐⭐⭐⭐⭐ | **Exemplary** — best layout in app |
| Notifications | List + modal | ⭐⭐⭐⭐ | Modal config needs grouping for email |
| Monitoring | Info cards | ⭐⭐⭐⭐ | Simple, effective |
| Settings | Single section | ⭐⭐⭐⭐⭐ | Appropriately minimal |
| Profile | Sections | ⭐⭐⭐⭐ | Good |
| Admin Users | Table + modals | ⭐⭐⭐⭐ | Good |

## User-Reported Issues

### 1. Connections Page Not Dynamic (User: "just one add button and an empty list")
**Current behavior**: Page shows ALL app types (Sonarr, Radarr, Lidarr, Readarr, Whisparr) with per-type "Add" buttons, and an empty "No instances configured" under each. Download clients has a single "Add" button.

**User wants**: 
- Single "Add" button with dropdown to pick app type/dl client type
- Only show categories after instances are added (no empty sections)
- Cards appear grouped under their category only after first instance

### 2. Settings Pages Need Better Design
**Priority redesign targets**:
- **Lurk Settings** (`/lurk`) — flat list of inputs, no visual grouping. Needs sections like queue has.
- **Notifications modal** — complex provider forms (email has 6 fields) need field grouping.

## Missing Frontend Features (vs Backend)

1. **WebAuthn/Passkeys** — Backend has model, no frontend
2. **Health status in dashboard** — Backend polls but UI doesn't show cached results
3. **Blocklist sources CRUD** — Backend has full CRUD, frontend only shows blocklist log table (no source management UI)
4. **Seerr requests list** — Backend has endpoint, frontend doesn't show individual requests
5. **SABnzbd/Prowlarr dashboards** — Backend has dedicated handlers, frontend only has config modals
6. **State reset per instance** — Backend supports it, no dedicated UI (just global reset on dashboard)

## No Unused npm Dependencies
All packages in package.json are actively used.

## Performance Notes
- Static prerendered build (no SSR)
- Client-side data fetching (SPA pattern)
- Debounced search (300ms in history)
- Lazy tab loading (queue page)
- Parallel API calls (`Promise.all` for instances)
