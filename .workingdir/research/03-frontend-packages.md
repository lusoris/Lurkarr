# Frontend Packages — Deep Research Reference

## Core Framework

### SvelteKit `2.50.2` + Svelte `5.51.0`
- **Routing**: File-based routing in `src/routes/`
- **Rendering**: SPA mode (via `@sveltejs/adapter-static`)
- **Svelte 5 Runes**: Modern reactivity system
  - `$state()` — reactive state declaration
  - `$derived()` — computed values (replaces `$:`)
  - `$effect()` — side effects (replaces `afterUpdate`)
  - `$props()` — component props declaration
  - `$bindable()` — two-way binding props
  - `$inspect()` — development debugging
- **Best Practices**:
  - Use runes exclusively (no legacy `let` reactivity)
  - Prefer `$derived` over `$effect` for derived state
  - Use `{#snippet}` blocks for reusable template fragments (replaces slots)
  - Use `{@render}` to render snippets
  - Avoid `$effect` for data fetching — use `load` functions
  - Event handlers: use `onclick` not `on:click` (Svelte 5 syntax)
  - Component composition: `{@render children()}` not `<slot />`

### Vite `7.3.1`
- **Purpose**: Build tool and dev server
- **Config**: `vite.config.ts` with SvelteKit plugin
- **Features**: HMR, pre-bundling, tree-shaking, code splitting
- **Best Practices**:
  - Use `import.meta.env` for environment variables
  - Configure `build.target` for browser compatibility

### TypeScript `5.8.3`
- **Config**: `tsconfig.json` extends SvelteKit defaults
- **Best Practices**:
  - Strict mode enabled
  - Use `satisfies` for type checking without widening
  - Prefer type inference where unambiguous
  - Use `unknown` over `any`

---

## UI Component Libraries

### bits-ui `2.16.3`
- **Purpose**: Headless UI components (Svelte port of Radix primitives)
- **Components**: Dialog, Popover, Select, Tabs, Tooltip, Accordion, Calendar, etc.
- **Key Features**:
  - Fully accessible (WAI-ARIA compliant)
  - Unstyled / headless — styled via TailwindCSS
  - Svelte 5 compatible with runes
- **Best Practices**:
  - Use composed parts (Root, Trigger, Content) pattern
  - Let bits-ui handle accessibility (focus management, keyboard nav)
  - Style with Tailwind utility classes

### shadcn-svelte `1.1.1`
- **Purpose**: Pre-built component recipes using bits-ui + Tailwind
- **Pattern**: Copy-paste components into `$lib/components/ui/`
- **Components**: Button, Card, Input, Label, Table, Badge, Alert, etc.
- **Best Practices**:
  - Customize via CSS variables (not direct overrides)
  - Use `cn()` utility (tailwind-merge + clsx) for conditional classes
  - Components are owned — modify freely
  - Use `components.json` for CLI configuration

### lucide-svelte `0.577.0`
- **Purpose**: Icon library (Lucide icons as Svelte components)
- **Usage**: `import { Icon } from 'lucide-svelte'`
- **Best Practices**: Tree-shakes automatically; import individual icons

---

## Styling

### TailwindCSS `4.2.1` (v4)
- **Major Changes in v4**:
  - CSS-first configuration (no `tailwind.config.js`)
  - `@import "tailwindcss"` in main CSS file
  - `@theme` directive for custom tokens
  - Automatic content detection (no `content` array needed)
  - New color system: `oklch()` based
  - `@variant` for custom variants
  - Native container queries support
- **Best Practices**:
  - Configure tokens in `app.css` via `@theme { }` block
  - Use `@source` for explicit content paths if needed
  - Use new `text-*` utilities with `text-wrap-balance` / `text-wrap-pretty`
  - Use `size-*` for width+height simultaneously
  - Dark mode via `dark:` variant

### tailwind-merge `3.3.0`
- **Purpose**: Intelligently merge Tailwind classes (resolves conflicts)
- **Usage**: `twMerge('px-2 py-1', 'p-3')` → `'p-3'`

### tailwind-variants `0.3.1`
- **Purpose**: Variant-based component styling (cva alternative)
- **Usage**: Define `tv()` variants for component props → className mapping

### tw-animate-css `1.3.4`
- **Purpose**: Tailwind CSS animation utilities
- **Usage**: Pre-built animation classes for transitions

### mode-watcher `0.5.2`
- **Purpose**: Dark/light mode detection and persistence
- **Features**: System preference detection, manual toggle, localStorage persistence
- **Best Practices**: Use `mode` store for current theme, respect system preference

---

## Toast / Notifications

### svelte-sonner `2.0.1`
- **Purpose**: Toast notification library (Sonner port for Svelte)
- **Usage**: `toast.success('Done!')`, `toast.error('Failed')`
- **Features**: Stacking, swipe to dismiss, custom components, promise toasts
- **Best Practices**: Place `<Toaster />` in root layout once

---

## Testing

### vitest `4.1.0`
- **Purpose**: Unit testing framework (Vite-native)
- **Features**: ESM-first, HMR-aware, JSDom/Happy-DOM, snapshot testing
- **Best Practices**:
  - Co-locate test files: `Component.test.ts` next to `Component.svelte`
  - Use `vi.mock()` for module mocking
  - Use `vi.fn()` for function spies
  - Prefer `toEqual` for objects, `toBe` for primitives 

### @testing-library/svelte `5.2.7`
- **Purpose**: DOM-based component testing
- **Best Practices**:
  - Query by role/text (accessible queries first)
  - `render(Component, { props })` for mounting
  - Use `screen.getByRole`, `screen.getByText`
  - `await fireEvent.click(element)` for interactions
  - `cleanup()` after each test (auto with vitest)

### @testing-library/jest-dom `6.6.3`
- **Purpose**: Extended DOM matchers
- **Matchers**: `toBeVisible`, `toHaveTextContent`, `toBeDisabled`, `toHaveAttribute`

---

## Build & Tooling

### @sveltejs/adapter-static `3.0.8`
- **Purpose**: SvelteKit adapter for static/SPA output
- **Config**: `prerender` all routes, `fallback: 'index.html'` for SPA routing
- **Output**: `build/` directory with HTML + assets (embedded into Go binary)

### @sveltejs/kit `2.21.5`
- **SPA Mode**: All routes pre-rendered as static HTML
- **Embedding**: `frontend/embed.go` uses `//go:embed build/*` for Go embedding

---

## Package Version Matrix

| Package | Version | Category |
|---------|---------|----------|
| svelte | 5.51.0 | Framework |
| @sveltejs/kit | 2.21.5 | Framework |
| vite | 7.3.1 | Build |
| typescript | 5.8.3 | Language |
| tailwindcss | 4.2.1 | Styling |
| bits-ui | 2.16.3 | Components |
| shadcn-svelte | 1.1.1 | Components |
| lucide-svelte | 0.577.0 | Icons |
| svelte-sonner | 2.0.1 | Toasts |
| vitest | 4.1.0 | Testing |
| @testing-library/svelte | 5.2.7 | Testing |
| mode-watcher | 0.5.2 | Theme |
| tailwind-merge | 3.3.0 | Utility |
| tailwind-variants | 0.3.1 | Utility |
