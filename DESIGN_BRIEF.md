# Design Brief: Pillars Cooperative — Full UI Redesign

## 1. Job and Audience

**Who:** The committee admin/secretary — one person running the cooperative's operations daily.
**Context:** Sitting at a laptop or phone, managing members, recording attendance after monthly meetings, chasing dues, issuing fines, tracking event contributions. Often working quickly between meetings or during committee sessions.
**Need:** A calm, scannable tool where every section is one click away, nothing is buried, and the financial picture is always visible without scrolling through 10 competing sections.
**Visitor mode:** Operate.

## 2. Outcome and Proof

**Primary task:** At a glance, know the cooperative's financial health and what needs attention today (who's in probation review, who owes money, which events need funding). Drill into any section instantly via sidebar.
**Success:** The admin opens the app, immediately sees what matters, and can act on it within 2-3 clicks. No hunting through a mega-dashboard. No "where was that form again?"
**Real evidence:** Naira currency throughout, Nigerian cooperative context, 90-day probation, ₦2,000 dues, ₦1,000 absence fine, ₦500 late fine.

## 3. Selected Direction

### THESIS
The category default is a friendly dashboard with cards and soft colors. Ledger & Column refuses that: it is a typographic grid system where hierarchy comes from weight, size, and spacing alone — never from color fills, rounded cards, or decorative elements. The interface IS the ledger. Every pixel earns its place through information density and structural clarity.

### OWN-WORLD
**Palette:** Restrained — near-black (#1a1a1a) on white (#ffffff) with one accent color. The accent is a deep, authoritative blue (#003d8f) — not electric, not playful. Supporting neutrals: warm gray 50 (#f7f6f3) for surfaces, warm gray 200 (#e8e6e1) for hairlines, warm gray 500 (#8a8680) for secondary text. No green. No teal. No decorative color.

**Typography:** The primary weapon. Inter (or a comparable workhorse) at heavy weights for hierarchy — 700-900 for headings, 400 for body, 500 for labels. Tight letter-spacing on large text (-0.02em to -0.04em). Uppercase small-caps for section labels. Numbers in tabular-lingular figures for financial data alignment. The type system does the work that color and cards do in other systems.

**Components:** No rounded cards. No shadows. No gradient fills. Borders are 1px hairlines in warm gray 200. Buttons are flat rectangles with 2px borders or solid fills. Inputs are underlined or bordered boxes with sharp corners (border-radius: 0-2px). Tables are the primary data container — dense, striped alternating rows, no card wrappers. Status indicators use text weight and the accent color, never colored pill badges.

**Spacing:** Tight and deliberate. 4px base unit. Content density higher than typical SaaS — this is a tool for someone who uses it daily, not a landing page. Generous whitespace above headings, tight below.

### STORY
The admin opens Pillars and sees an official record — not a friendly dashboard, but a financial instrument. The treasury balance reads like a bank statement. The member list reads like a registry. Every action (add member, record attendance, issue fine) feels like filling out an official form. The weight of the interface communicates: these numbers are real, these records matter, this is the authoritative source.

### FIRST VIEWPORT
Left sidebar (fixed, 240px): Pillars wordmark at top in 700 weight, section links below in 500 weight, active state uses accent blue left border + bold text. No icons — text-only navigation.

Content area: 
- Top row: 4 stat blocks in a tight grid. Each is a number (tabular, 900 weight, 32px) above a label (500 weight, 12px uppercase tracking). No cards, no borders — just type on white, separated by vertical hairlines.
- Below stats: Two-column layout. Left column (60%): Treasury summary as a dense table — rows for dues paid, dues owed, fines paid, fines owed, with right-aligned tabular numbers. Right column (40%): Probation review list, each member as a row with name, probation end date, and action links (Promote / Extend) styled as underlined text links, not buttons.
- Below that: Event funding as a table — event name, goal, collected, progress as a simple text fraction (₦12,000 / ₦50,000), no progress bars.

Primary action: "New" links in the sidebar for each section (New Member, Record Attendance, New Event, Issue Fine) — styled as underlined text with accent blue, not buttons.

### FORM
Swiss International Style / Neue Grafik. Position: #1 on grounded list (top-ranked by resonance for this subject). Catalog challengers: none (degraded roll). Seed key: f529b80a.

### CROSS-SURFACE REACH
This direction works for every page: member detail becomes a structured record card, event detail becomes a funding ledger, attendance becomes a dense table. The grammar is universal — typographic hierarchy, grid, hairlines, tabular data. Future surfaces (PDF reports, member statements) inherit the same system naturally.

### HONEST RISK
The density may feel austere to users accustomed to friendly SaaS dashboards. Mitigation: the tightness is purposeful — it communicates professionalism and efficiency, not coldness. The warm gray palette (not pure gray) softens the Swiss precision just enough. The risk is real but aligned with the product's truth: this is a financial record system, not a social app.

## 4. Scope and Boundaries

**Fidelity:** Production-ready screen — all 4 current pages rebuilt, plus new dedicated pages for creation workflows.
**Breadth:** Full UI rebuild — layout, navigation, all templates, the Tailwind config/design tokens.
**What remains untouched:** Go backend, store logic, handlers (only template names/paths change if needed), tests, SQLite schema.
**Anti-goals:** No new backend features. No new database tables. No authentication system. No SPA framework. No new dependencies.

## 5. States and Ranges

**Members:** 0-50 typical, up to ~100 max.
**Attendance:** Monthly meetings, 12 records per member per year. Records per meeting: 0-100+ members.
**Dues:** Recurring ₦2,000 records per member. Status: paid, pending, partially_paid, owed.
**Fines:** Any number outstanding. ₦500-₦1,000 amounts.
**Events:** Any number per year, each with 0-100% contribution progress.
**Probation review:** Any number of members at any time.

## 6. Interaction and Layout

**Sidebar (persistent, left):** Fixed 240px on desktop, collapsible drawer on mobile. Text-only nav, no icons. Active state: accent blue left border + bold text.

**Dashboard:** Overview only — stats, treasury table, probation review, event funding table, recent activity. NO forms.

**Dedicated creation pages:** /members/new, /attendance/new, /events/new, /fines/new. Each is a focused form page.

**Detail pages:** Same sidebar layout. Structured record display with dense tables for sub-data.

**Responsiveness:** Desktop: sidebar + content. Tablet: collapsible sidebar. Mobile: hamburger menu.

**Feedback:** HTMX partial swaps, subtle opacity transitions, toast confirmations.

## 7. Constraints

- **Platform:** Web (Go + HTMX + Tailwind CDN)
- **Font:** Inter (workhorse, already in use)
- **Color:** Restrained — near-black, white, warm grays, one accent blue (#003d8f)
- **No auth:** Single-user
- **PDF exports:** Keep existing routes, update button placement
- **Reports:** Dedicated sidebar section

## Direction Contract

```
<!-- impeccable:direction 1 -->
<!-- THESIS: Typographic grid system where hierarchy comes from weight, size, and spacing alone. The interface IS the ledger. -->
<!-- OWN-WORLD: Near-black on white, one accent blue (#003d8f), warm gray neutrals. Inter at heavy weights. No rounded cards, no shadows, no decorative color. Hairlines and tabular data. -->
<!-- STORY: The admin opens an official financial record, not a friendly dashboard. Every number matters, every record is authoritative. -->
<!-- FIRST VIEWPORT: Fixed sidebar (240px) with text-only nav. Content: 4 stat blocks (type only, no cards), treasury table, probation review list, event funding table. -->
<!-- FORM: Swiss International Style / Neue Grafik. Seed f529b80a. -->
```
