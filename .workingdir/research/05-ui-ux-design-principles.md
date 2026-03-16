# UI/UX Design Principles — Deep Research Reference

## 1. Dark Mode Best Practices

### 1.1 Research Findings (NNGroup, Academic Studies)
- **Light mode is faster to read** for most users (positive polarity advantage)
- **Dark mode is preferred** by many users; ~80% of developers use dark themes
- **Dark mode reduces eye strain** in low-light environments
- **Always offer both modes** — let user choose, default to system preference
- **Key recommendation**: Support dark mode but don't make it the only option

### 1.2 Dark Mode Implementation Rules
- **Never use pure black (#000000)** — use dark gray (e.g., `#0a0a0a` to `#1a1a1a`)
- **Reduce white intensity** — use off-white (`#e0e0e0` to `#f0f0f0`) instead of `#ffffff`
- **Maintain contrast ratios**: WCAG AA minimum 4.5:1 for body text, 3:1 for large text
- **Reduce elevation shadows** — in dark mode, use lighter surface colors for elevation instead of shadows
- **Desaturate colors** slightly in dark mode — vivid colors on dark backgrounds cause visual vibration
- **Test with both modes** — components must look correct in both themes

### 1.3 Color System (oklch-based — TailwindCSS v4)
- **Use oklch()** for perceptually uniform color manipulation
- **Define semantic tokens**: `--color-surface`, `--color-surface-elevated`, `--color-text-primary`, `--color-text-secondary`, `--color-accent`
- **Map tokens to light/dark values** via CSS custom properties and `@media (prefers-color-scheme: dark)`
- **Status colors**: Success (green), Warning (amber), Danger (red), Info (blue) — ensure accessibility in both modes

---

## 2. Dashboard Design Patterns

### 2.1 Information Architecture
- **Progressive disclosure**: Show summary → allow drill-down to details
- **Critical info first**: Status indicators, error counts, active operations at top
- **Group by function**: Download clients together, arr apps together, notifications separate
- **Consistent layout**: Cards/panels with consistent spacing, alignment, and hierarchy

### 2.2 Data Display
- **Tables**: Use for structured data (queue, history, blocklist); sortable columns, pagination
- **Cards**: Use for entity summaries (apps, download clients); show name, status, key metric
- **Charts**: Sparingly — only where trends matter (download speeds, queue depth over time)
- **Status Indicators**: Traffic light (green/amber/red) dots, badges, or icons
- **Empty States**: Always show meaningful empty states ("No items in queue" with suggested action)

### 2.3 Real-time Updates
- **Polling intervals**: Match urgency — queue updates every 10-30s, history every 60s, health every 30s
- **Visual feedback**: Subtle pulse/animation when data refreshes
- **Last updated timestamp**: Show when data was last fetched
- **Loading indicators**: Skeleton screens for first load, subtle spinners for refreshes

---

## 3. Accessibility (WCAG 2.1)

### 3.1 Contrast Requirements
- **AA Standard** (minimum target):
  - Normal text: 4.5:1 contrast ratio
  - Large text (18px+ or 14px+ bold): 3:1
  - UI components and graphics: 3:1
- **AAA Standard** (ideal):
  - Normal text: 7:1
  - Large text: 4.5:1
- **Tool**: Use oklch manipulation in TailwindCSS v4 to generate accessible palette

### 3.2 Keyboard Navigation
- **All interactive elements** must be keyboard-focusable
- **Visible focus indicators**: Never `outline: none` without replacement
- **Tab order**: Logical flow matching visual layout
- **Escape to close**: All modals, popovers, dropdowns
- **bits-ui handles this** for its primitives — leverage it

### 3.3 Screen Reader Support
- **Semantic HTML**: Use `<nav>`, `<main>`, `<aside>`, `<section>`
- **ARIA labels**: For icon-only buttons, status indicators
- **Live regions**: `aria-live="polite"` for toast notifications, status updates
- **bits-ui handles ARIA** — don't override unless necessary

### 3.4 Motion & Animation
- **Respect `prefers-reduced-motion`**: Reduce/disable animations
- **Subtle transitions**: 150-300ms for most interactions
- **No autoplay animations** that can't be paused

---

## 4. Layout & Spacing

### 4.1 Responsive Design
- **Mobile-first** approach — design for small screens, enhance for large
- **Breakpoints** (Tailwind defaults):
  - `sm`: 640px (tablet portrait)
  - `md`: 768px (tablet landscape)
  - `lg`: 1024px (desktop)
  - `xl`: 1280px (wide desktop)
  - `2xl`: 1536px (ultra-wide)
- **Sidebar pattern**: Collapsible sidebar on mobile, persistent on desktop
- **Table responsive**: Horizontal scroll on mobile, full display on desktop

### 4.2 Spacing System
- **Use consistent spacing scale**: Tailwind's 4px base (0.25rem increments)
- **Component spacing**: `p-4` (16px) for cards, `gap-4` (16px) between elements
- **Section spacing**: `py-6` to `py-8` between page sections
- **Comfortable click targets**: Minimum 44×44px (WCAG, Apple HIG)

### 4.3 Typography
- **Font sizes**: Use Tailwind's scale (`text-sm` = 14px, `text-base` = 16px, `text-lg` = 18px)
- **Body text**: 16px minimum for readability
- **Line height**: 1.5 for body text, 1.25 for headings
- **Font weight**: 400 (normal) for body, 600 (semibold) for labels/headings
- **Monospace**: For IDs, hashes, file paths, technical values

---

## 5. Interaction Design

### 5.1 Forms
- **Inline validation**: Validate on blur, not on every keystroke
- **Error messages**: Below the field, in danger color, with clear description
- **Success feedback**: Toast notification on successful submission
- **Disabled state**: Clear visual distinction, tooltip explaining why
- **Auto-save** where appropriate (settings pages)

### 5.2 Destructive Actions
- **Confirmation dialogs**: For delete, remove, disconnect operations
- **Red/danger styling**: Destructive buttons use danger color variant
- **Undo when possible**: Prefer soft-delete with undo option over permanent delete
- **Type to confirm**: For critical operations (e.g., "Type DELETE to confirm")

### 5.3 Loading States
- **Skeleton screens**: For initial page loads (mimics content layout)
- **Inline spinners**: For button actions (disable button + show spinner)
- **Progress bars**: For operations with known duration
- **Optimistic updates**: Update UI immediately, revert on failure

### 5.4 Error States
- **User-friendly language**: "Unable to connect to Sonarr" not "ECONNREFUSED 192.168.1.5:8989"
- **Actionable messages**: Include what the user can do ("Check that Sonarr is running")
- **Retry option**: Every error state should have a retry/refresh action
- **Don't break the page**: Partial failures should only affect the failing component

---

## 6. Material Design 3 Color System (Reference)

### 6.1 Baseline Tokens (Adapted for Lurkarr)
- **Primary**: Accent color for key actions, active states, important highlights
- **Secondary**: Supporting elements, less prominent components
- **Tertiary**: Complementary accents, balance visuals
- **Surface**: Background colors with elevation tiers
- **Error**: Error states and destructive actions
- **On-[color]**: Text/icon color on respective background
- **[Color]-Container**: Subdued version for backgrounds (e.g., error-container for error summary backgrounds)

### 6.2 Dynamic Color
- **Derive full palette** from 1-2 key colors using oklch manipulation
- **Consistent across modes**: Light and dark variants of same palette
- **Use CSS custom properties**: `var(--primary)`, `var(--surface-elevated)`

---

## 7. Navigation Patterns

### 7.1 Sidebar Navigation (Current Pattern)
- **Primary nav**: Sidebar with icon + label for each section
- **Grouping**: Apps, Activity (Queue/History/Downloads), Settings
- **Active state**: Highlight current route
- **Badge counts**: Errors, pending items on nav items
- **Collapse**: Icon-only mode for more content space

### 7.2 Page Structure
```
┌─────────────────────────────────┐
│  Header (breadcrumb, actions)   │
├─────────────────────────────────┤
│  Page Title + Description       │
├─────────────────────────────────┤
│  Content Area                   │
│  ┌─────────┐ ┌─────────┐      │
│  │  Card 1  │ │  Card 2  │      │
│  └─────────┘ └─────────┘      │
│  ┌──────────────────────┐      │
│  │     Table / List      │      │
│  └──────────────────────┘      │
└─────────────────────────────────┘
```

---

## 8. Performance Perception

- **< 100ms**: Feels instant — aim for all UI interactions
- **100-300ms**: Slight delay — acceptable for data fetches
- **300ms-1s**: Noticeable — show loading indicator after 300ms
- **> 1s**: Show progress — skeleton screens or progress bar
- **> 5s**: Keep user informed — show percentage or step progress

---

## 9. Notification/Toast System

### 9.1 Toast Types (svelte-sonner)
- **Success**: Green accent, auto-dismiss after 3-5s
- **Error**: Red accent, persist until user dismisses (errors need attention)
- **Warning**: Amber accent, auto-dismiss after 5-7s
- **Info**: Blue accent, auto-dismiss after 3-5s
- **Loading/Promise**: Show loading → success/error transition

### 9.2 Placement
- **Bottom-right**: Least intrusive for primary content
- **Toast stacking**: Max 3-5 visible, queue rest
- **Don't duplicate**: Deduplicate identical messages within short window
