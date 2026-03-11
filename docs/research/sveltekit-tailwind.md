# SvelteKit 5 + TailwindCSS v4 Reference

## SvelteKit 5

> Docs: https://kit.svelte.dev/docs
> Svelte 5 Docs: https://svelte.dev/docs/svelte

### Runes (Svelte 5)

```svelte
<script lang="ts">
  // Reactive state
  let count = $state(0);

  // Derived values (replaces $: reactive declarations)
  let doubled = $derived(count * 2);

  // Complex derived
  let stats = $derived.by(() => {
    return { total: count, doubled: count * 2 };
  });

  // Effects (replaces $: side effects)
  $effect(() => {
    console.log('count changed:', count);
  });

  // Props
  let { data, onsubmit } = $props<{ data: PageData; onsubmit: () => void }>();

  // Bindable props
  let { value = $bindable('') } = $props();
</script>
```

### Routing (Lurkarr uses static adapter)

```
src/routes/
├── +layout.svelte       # Root layout (nav, auth check)
├── +layout.ts            # Universal load (runs during prerender)
├── +page.svelte          # Home page
├── apps/
│   └── +page.svelte      # /apps
├── settings/
│   └── +page.svelte      # /settings
└── login/
    └── +page.svelte      # /login
```

### Load Functions

```ts
// +page.ts (universal — runs during prerender)
export const load: PageLoad = async ({ fetch }) => {
  const res = await fetch('/api/settings');
  return { settings: await res.json() };
};
```

### Static Adapter (Lurkarr config)

```js
// svelte.config.js
import adapter from '@sveltejs/adapter-static';
export default {
  kit: {
    adapter: adapter({
      pages: 'build',
      assets: 'build',
      fallback: undefined, // no SPA fallback — prerendered
    }),
  },
};
```

### API Calls Pattern

```ts
// lib/api.ts
const BASE = '/api';
export async function getSettings(): Promise<Settings> {
  const res = await fetch(`${BASE}/settings`);
  if (!res.ok) throw new Error('Failed');
  return res.json();
}
```

### Stores Pattern

```ts
// lib/stores/settings.ts
import { writable } from 'svelte/store';
// Or with Svelte 5 runes:
export const settings = $state<Settings | null>(null);
```

## TailwindCSS v4

> Docs: https://tailwindcss.com/docs
> Upgrade guide: https://tailwindcss.com/docs/upgrade-guide

### Key Changes from v3 → v4

- CSS-first configuration (no tailwind.config.js)
- Use `@import "tailwindcss"` instead of `@tailwind`
- `@theme` directive for custom values
- Container queries built-in
- Color-mix() for opacity

### Setup in SvelteKit

```css
/* app.css */
@import "tailwindcss";

@theme {
  --color-primary: #3b82f6;
  --color-danger: #ef4444;
  --radius-default: 0.5rem;
}
```

### Utility Classes Commonly Used

```html
<!-- Layout -->
<div class="flex items-center justify-between gap-4">
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">

<!-- Card pattern -->
<div class="rounded-lg border bg-card p-6 shadow-sm">

<!-- Form input -->
<input class="w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary">

<!-- Button -->
<button class="inline-flex items-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary/90">

<!-- Dark mode -->
<div class="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
```

## Frontend Architecture Notes

### Current Pages (10)

| Route | Purpose | Status |
|-------|---------|--------|
| / | Dashboard | ✅ |
| /apps | Arr instance management | ✅ |
| /downloads | Active downloads | ✅ |
| /history | Hunt history | ✅ |
| /login | Authentication | ✅ |
| /logs | Log viewer | ⚠️ Remove (use Grafana) |
| /queue | Queue management | ✅ |
| /scheduling | Schedule config | ✅ |
| /settings | General settings | ✅ |
| /user | User profile / 2FA | ✅ |

### Missing Pages

| Route | Purpose |
|-------|---------|
| /notifications | Notification provider config |
| /seerr | Overseerr/Jellyseerr config |
| /download-clients | DL client management |
| /monitoring | Grafana embed/links |
| /auto-import | Auto-import config |

### Component Structure

```
src/lib/components/
├── Nav.svelte           # Navigation sidebar
├── AppCard.svelte       # Arr instance card
├── SettingsForm.svelte  # Reusable settings form
└── ...
```
