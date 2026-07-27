---
target: Dashboard page (templates/index.html)
total_score: 28
max_score: 40
na_heuristics: 
p0_count: 1
p1_count: 2
p2_count: 2
p3_count: 0
timestamp: 2026-07-26T21-37-36Z
slug: templates-index-html
---
## Design Specificity Verdict

**Strong with structural gaps.** The Ledger & Column Swiss International Style is coherent and intentional. The typographic system — 10px uppercase micro labels, tabular-nums financial figures, border-only containment, warm gray paper tones — is specific to this product's identity as a financial ledger. The absence of shadows, rounded corners, and decorative color creates a genuine authorial voice. This design could not be copy-pasted onto a social media app or e-commerce site unchanged.

The composition is the generic part: stat grid on top, then 5-6 vertically stacked full-width tables of identical visual weight. The individual elements are authored; the page-level orchestration is template-standard.

**Deterministic scan**: 18 findings across 8 files, all `overused-font` warnings on Inter. Every hit is a false positive — Inter is the correct Swiss International Style choice. No slop patterns detected.

---

## Heuristic Scores

| # | Heuristic | Score | Key Issue |
|---|-----------|-------|-----------|
| 1 | Visibility of System Status | 3 | No loading state for HTMX swaps. No "last refreshed" indicator. |
| 2 | Match Between System and Real World | 4 | Domain terminology is accurate. Ledger metaphor maps 1:1 to cooperative management. |
| 3 | User Control and Freedom | 3 | Cancel buttons in modals, back links in sub-pages. But modals lack Escape key handling and focus trapping. |
| 4 | Consistency and Standards | 3 | Internally consistent. But full `<head>` boilerplate duplicated across all 9 templates. Toast uses `rounded` class contradicting "no rounded" direction. |
| 5 | Error Prevention | 2 | Promote/Extend/Delete modals provide confirmation. No input validation on forms. No safeguard against accidental navigation away from in-progress attendance. |
| 6 | Recognition Rather Than Recall | 3 | Tables present data for recognition. Probation Review surfaces items needing attention. But member status is plain text — no visual badge for scanning. |
| 7 | Flexibility and Efficiency of Use | 2 | No search, no filter, no sort on any table. No keyboard shortcuts. No bulk actions. Dashboard is read-only. |
| 8 | Aesthetic and Minimalist Design | 4 | Zero decorative elements. Hierarchy is purely typographic. Exactly as complex as the content requires. |
| 9 | Help Users Recognize, Diagnose, and Recover from Errors | 2 | Toast handles success/error. But no inline error states, limited empty-state guidance. |
| 10 | Help and Documentation | 2 | No tooltips, no contextual help. "At Risk: 3" has no explanation of threshold. |
| **Total** | | **28/40** | **Good** |

---

## Cognitive Load Assessment

- [x] **Single focus:** Dashboard presents a financial overview. Primary task is clear but no single call to action.
- [ ] **Chunking:** FAILS. 6 full sections vertically: Stats → Probation → Treasury → Event Funding → Members → Events → Attendance. Requires significant scrolling. No collapsing, no "above the fold" priority.
- [x] **Grouping:** Related data IS grouped into bordered tables with consistent headers.
- [ ] **Visual hierarchy:** WEAK. Every section uses identical heading treatment. Treasury, Members, Events, Attendance all look equally important.
- [ ] **One thing at a time:** FAILS. User sees treasury, probation, event funding, all members, all events, and all attendance simultaneously.
- [ ] **Minimal choices:** Members table shows every member with only a link. No filter chips, no status tabs, no search. At 30+ members, a wall of rows.
- [x] **Working memory:** Data is in tables — no cross-screen memory needed. Dashboard is self-contained.
- [ ] **Progressive disclosure:** FAILS. Everything shown at once. "Recent Attendance" heading is misleading — no limit on records rendered.

**Cognitive load score: 3/8 — structure is clean but information architecture is flat and overwhelming.**

---

## Emotional Journey

**Entry:** Header delivers confidence. "Pillars Cooperative" in bold with "Financial overview and pending actions" sets expectations.

**Stats reveal:** 4-stat grid provides immediate snapshot. "At Risk" in red creates appropriate urgency. This is the emotional peak.

**Probation Review:** Emotional valley — real human consequences. Modal interactions use plain language ("Promote to Active", "Extend Probation"). But visual treatment is identical to every other table, missing the opportunity to signal urgency.

**Bottom half scroll fatigue:** By Members → Events → Attendance, visual repetition creates fatigue. Six consecutive bordered tables. Trajectory: oriented → alert → numb.

**Missing reassurance:** No "all caught up" state, no "nothing pending" summary, no "last reviewed" indicator.

---

## What's Working

1. **Typographic system is excellent.** The hierarchy — 10px micro labels → 3xl bold numbers → sm body → tabular-nums financial figures — is consistent, legible, and unmistakably Swiss. The tracking values (`0.12em`, `0.15em`, `0.1em`) create precise micro-labels that feel like column headers in a printed ledger.

2. **Financial data presentation.** `tabular-nums`, `font-bold`, right-aligned monetary values, the ₦ prefix — every number is scannable. The Treasury section with its indented sub-rows mirrors how a real ledger organizes figures.

3. **Consistent component vocabulary.** Bordered tables, `bg-warm-50` header rows, `divide-y divide-warm-200` row separators, accent-colored action links — every table follows the same rules. The sidebar extends this with `bg-warm-100` active states.

---

## Priority Issues

### P0 — Modals lack keyboard/accessibility support
`index.html:58-82`, `member_detail.html:180-216`, `event_detail.html:86-95`

Modals opened via `classList.remove('hidden')` with raw `onclick`. No `role="dialog"`, no `aria-modal`, no focus trap, no Escape key handler. Keyboard-only users cannot operate Promote, Extend, or Delete flows. WCAG 2.1 Level A failure.

**Fix**: Add ARIA attributes, focus trap, Escape handler, and `aria-hidden` on background.
**Suggested command**: `$impeccable harden`

### P1 — No template base/layout — duplicated boilerplate across 9 files
`index.html:252-329`, `member_detail.html:220-319`, `event_detail.html:99-161`, `attendance_new.html:71-132` + 5 more

Full `<head>`, Tailwind config, Inter font, color system, toast, sidebar toggle JS, and body wrapper copy-pasted into every template. Any design token change requires editing 9 files. Will inevitably drift.

**Fix**: Create `templates/base.html` with `{{block "content" .}}` pattern. Each page defines its content block, base provides the shell.
**Suggested command**: `$impeccable distill`

### P1 — Dashboard has no upper-bound on record counts
`index.html:156-245`

Members, Events, and Attendance tables render all records with no pagination, no "show N most recent", and no limit. "Recent Attendance" at `index.html:218` promises recency but delivers everything. A cooperative with 50+ members will produce 100+ rows.

**Fix**: Add pagination or limit to top N. Add search/filter on Members table. Add status filter tabs.
**Suggested command**: `$impeccable shape`

### P2 — Probation Review section missing empty state
`index.html:33-88`

When `{{if .ProbationReview}}` is false, the section simply doesn't render. No "No members on probation" placeholder. Inconsistent with Events and Attendance which have `{{else}}` blocks. The absence is invisible — user doesn't know if the feature exists.

**Fix**: Add `{{else}}` block with "No members currently on probation. All members are in good standing."
**Suggested command**: `$impeccable clarify`

### P2 — Treasury sub-row hierarchy is ambiguous
`index.html:90-126`

Fines Total and its sub-rows use `pl-8` indentation and lighter `text-warm-500`, but the Fines Total row's visual weight is nearly indistinguishable from parent Dues rows. User can't immediately tell top-level summaries from sub-items. `bg-warm-50` alternating is applied inconsistently.

**Fix**: Use consistent `bg-warm-50` alternating for all sub-items. Consider a slightly heavier font-weight or different indentation for true parent rows.
**Suggested command**: `$impeccable layout`

---

## Persona Red Flags

### Alex (Power User)
- Cannot search or filter the Members table. Must scroll all rows.
- No keyboard shortcuts for any action. Every modal is mouse-only.
- No bulk actions. Cannot sort tables by any column.
- Dashboard is read-only — must navigate to member detail to mark dues paid.

### Sam (Accessibility-Dependent)
- Modals completely broken for keyboard navigation (P0). Tab order escapes into hidden content. No Escape to close. No focus return.
- No skip-to-content link. Must tab through entire sidebar on every page load.
- Tables lack `<caption>` elements. Screen readers read cell content without context.
- Status text ("active", "probation") has no semantic differentiation. No `aria-label` or sr-only indicators.
- Touch target on sidebar toggle is `p-2` (8px padding) — below 44x44px WCAG minimum.

### Casey (Distracted Mobile User)
- Sidebar requires hamburger tap on mobile. No swipe-to-close or Escape handling on overlay.
- All 6+ tables stack vertically. 30+ swipe experience on phone with no way to jump to a section.
- Toast positions near last click — can overlap sticky keyboard or be off-screen on mobile.

---

## Minor Observations

- `index.html:292` — Toast has `rounded` class. Contradicts "no rounded cards" design direction.
- `index.html:258` — CDN scripts loaded without integrity hashes. `cdn.tailwindcss.com` is development CDN, not suitable for production.
- `index.html:10` — Inline `style="letter-spacing: -0.03em"` instead of Tailwind utility. Appears on headers across all templates.
- `index.html:218` — Heading says "Recent Attendance" but no recency limit. Misleading copy.
- `sidebar.html:66-68` — Footer "Pillars Cooperative" in 10px uppercase is redundant with sidebar header and dashboard header. Triple branding.
- `index.html:92` — Treasury table has no `<thead>`. Every other table has one.

---

## Questions to Consider

1. What is the primary job-to-be-done when a secretary opens this dashboard? Is it "check treasury and take action on arrears" or "get a daily overview"? The answer determines whether Treasury should be the hero.

2. Should the Members table on the dashboard exist at all? It duplicates `/members`. A summary card ("12 active, 3 probation, 2 ex-members") with a link to `/members` would be higher-signal than a full table.

3. Is "Recent Attendance" pulling its weight? It shows date, record count, and status — but no individual member data. Could this be replaced with a "last meeting summary" card ("July meeting: 18/20 present, 2 absent")?

4. What happens at 50+ members? The flat member list, no pagination, and no search will break the experience at scale. Is that acceptable, or should the dashboard adapt?

5. Event Funding shows a percentage number. Would a visual progress bar in the table cell improve scannability? A bar communicates proportion pre-attentively; a percentage requires reading.
