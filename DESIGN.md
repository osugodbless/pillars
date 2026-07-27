# Design System — Pillars Cooperative

## Brand Identity

**Name:** Pillars Cooperative  
**Platform:** Web (Go + HTMX + Tailwind)  
**Mode:** Operate — committee dashboard for daily cooperative management  
**Locale:** Nigeria (English, Naira ₦)

---

## Color System

### Primitive Palette (OKLCH-derived)

| Token | Hex | OKLCH | Role |
|-------|-----|-------|------|
| `near-black` | `#1a1a1a` | 0.15 0 0 | Primary text, headings |
| `warm-900` | `#2d2a26` | 0.18 0.015 45 | Elevated text |
| `warm-700` | `#5c5955` | 0.38 0.02 45 | Secondary text, labels |
| `warm-500` | `#787571` | 0.52 0.015 45 | Muted text, placeholders |
| `warm-300` | `#d4d1cc` | 0.78 0.01 45 | Borders, dividers |
| `warm-200` | `#e8e6e1` | 0.86 0.008 45 | Light borders, hover backgrounds |
| `warm-100` | `#eeedea` | 0.92 0.006 45 | Active nav, selected rows |
| `warm-50` | `#f7f6f3` | 0.96 0.004 45 | Section backgrounds |
| `white` | `#ffffff` | 1 0 0 | Card backgrounds |

### Brand Accent

| Token | Hex | OKLCH | Role |
|-------|-----|-------|------|
| `accent` | `#003d8f` | 0.38 0.18 260 | Primary actions, links, focus rings |
| `accent-hover` | `#002f73` | 0.32 0.16 260 | Hover state |
| `accent-light` | `#eef3fb` | 0.94 0.03 260 | Accent backgrounds, badges |

### Semantic Colors

| Token | Hex | OKLCH | Role |
|-------|-----|-------|------|
| `success` | `#007a33` | 0.42 0.12 145 | Paid, complete, positive balance |
| `success-light` | `#e8f5ec` | 0.93 0.025 145 | Success backgrounds |
| `warning` | `#b85c00` | 0.55 0.11 70 | Pending, partially paid, attention needed |
| `warning-light` | `#fef3e2` | 0.95 0.025 70 | Warning backgrounds |
| `info` | `#0066cc` | 0.45 0.15 250 | Information, neutral actions |
| `info-light` | `#eef4fb` | 0.94 0.025 250 | Info backgrounds |
| `danger` | `#c41230` | 0.45 0.18 15 | Errors, destructive, unpaid, absent |
| `danger-light` | `#fcebee` | 0.93 0.025 15 | Danger backgrounds |

### Status Colors (Member/Financial)

| Status | Text Token | Background Token | Border Token |
|--------|------------|------------------|--------------|
| `active` / `bonafide` | `success` | `success-light` | `success` |
| `probation` | `warning` | `warning-light` | `warning` |
| `ex-member` | `warm-500` | `warm-100` | `warm-300` |
| `paid` / `settled` | `success` | `success-light` | `success` |
| `pending` / `not_paid` | `danger` | `danger-light` | `danger` |
| `partially_paid` | `warning` | `warning-light` | `warning` |
| `owed` | `danger` | `danger-light` | `danger` |
| `outstanding` | `danger` | `danger-light` | `danger` |
| `present` | `success` | `success-light` | `success` |
| `absent` | `danger` | `danger-light` | `danger` |
| `not_recorded` | `warm-500` | `warm-100` | `warm-300` |

### Data Visualization (Categorical)

| Token | Hex | OKLCH | Use |
|-------|-----|-------|-----|
| `chart-1` | `#003d8f` | 0.38 0.18 260 | Primary series (accent) |
| `chart-2` | `#007a33` | 0.42 0.12 145 | Secondary (success) |
| `chart-3` | `#b85c00` | 0.55 0.11 70 | Tertiary (warning) |
| `chart-4` | `#c41230` | 0.45 0.18 15 | Quaternary (danger) |
| `chart-5` | `#6b4c9a` | 0.42 0.12 285 | Quinary |

---

## Typography

**Font Family:** Inter (400, 500, 600, 700, 800, 900)  
**Base Size:** 14px (0.875rem)  
**Scale:** 12 / 14 / 16 / 18 / 24 / 30 / 36 / 48px  
**Tracking:** -0.03em display, -0.02em headings, 0 body, +0.05em caps labels  
**Tabular Numerals:** Required for all monetary/date columns (`.tabular-nums`)

---

## Spacing & Layout

**Base Unit:** 4px (0.25rem)  
**Container Max:** 80rem (1280px) — `max-w-5xl`  
**Sidebar Width:** 15rem (240px) — `w-60`  
**Card Radius:** 0.5rem (8px) — `rounded-lg`  
**Border Width:** 1px default, 2px focus  
**Shadow:** Subtle elevation only (`shadow-sm` for cards)

---

## Component Tokens

### Buttons
| Variant | Background | Text | Border | Hover | Focus Ring |
|---------|------------|------|--------|-------|------------|
| Primary | `accent` | `white` | none | `accent-hover` | `accent` |
| Secondary | `white` | `near-black` | `warm-300` | `warm-50` | `accent` |
| Danger | `danger` | `white` | none | `danger/90` | `danger` |
| Ghost | transparent | `warm-700` | none | `warm-100` | `accent` |

### Form Inputs
- Border: `warm-300` default, `accent` focus
- Background: `white`
- Text: `near-black`
- Placeholder: `warm-500`
- Error border: `danger`
- Error text: `danger`

### Tables
- Header: `warm-50` bg, `warm-300` border, `warm-500` text (uppercase, 10px)
- Row hover: `warm-50`
- Divider: `warm-200`
- Numeric: `tabular-nums`, right-aligned

### Badges / Pills
- Padding: `px-2 py-0.5`
- Radius: `rounded-full`
- Text: `text-xs font-medium`

---

## Motion

**Duration:** 150ms default, 200ms complex  
**Easing:** `ease-out` (cubic-bezier(0.25, 0.46, 0.45, 0.94))  
**Reduced Motion:** Respect `prefers-reduced-motion` — disable all transitions/animations

---

## Accessibility

- All text meets WCAG AA (4.5:1 body, 3:1 large)
- Focus indicators: 2px solid `accent`, offset 2px
- Color never sole carrier of meaning — paired with text/icon/shape
- Semantic HTML + ARIA where needed (modals, toasts, live regions)

---

## Implementation Notes

Colors defined in `templates/base.html` Tailwind config. All templates consume semantic tokens (e.g., `text-success`, `bg-warning-light`, `border-danger`) not primitive values. When adding new UI, extend the token system rather than hardcoding hex.