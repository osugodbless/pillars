# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

**Committee (admin/secretary):** The primary operator. Manages the full lifecycle of the cooperative — members, attendance, dues, fines, events, contributions, probation reviews, and financial reports. Works from a single dashboard. Likely one person or a very small committee handling day-to-day operations.

**Members:** See their own standing — attendance record, dues paid/owed, fines, and event contributions. Access via individual detail pages.

## Product Purpose

Pillars Cooperative is the operational backbone of a real cooperative organization. It replaces spreadsheet or paper-based tracking with a single source of truth for membership, financial obligations, attendance, and collective events. Success means the committee can manage the cooperative without losing track of who owes what, who attended what, and which members are in good standing.

## Positioning

Unlike generic accounting tools or spreadsheet templates, Pillars is purpose-built for the cooperative domain: it understands the relationship between dues, fines, and contributions (fines and contributions are deducted from paid dues, carrying forward remainders as owed), the probation-to-active member lifecycle, and the committee reporting workflows that cooperatives actually need.

## Operating Context

- Weekly or biweekly cooperative meetings where attendance is taken
- Dues collected on a recurring schedule (amount set in code: ₦2,000)
- Fines levied for unapproved absence (₦1,000) or lateness (₦500)
- Events created with funding goals; contributions auto-generated for every active member
- Probation period (90 days) before new members are promoted to active/bonafide status
- Committee needs PDF reports for attendance and contributions at meetings
- Arrears and committee financial reports for governance

## Capabilities and Constraints

- **Member management:** Add, view, delete members; track status (probation/active/bonafide); probation review workflow (promote or extend)
- **Attendance:** Record attendance per meeting date with status (present/absent/absent without permission); export to PDF
- **Dues:** Track dues per member with paid/pending/partially_paid/owed status; mark as paid
- **Fines:** Levy fines with reasons; deduct from paid dues; mark as paid
- **Events & Contributions:** Create events with funding goals; auto-generate pending contributions for active members; track collection progress; settle events
- **Financial Reports:** Committee report, arrears report with aging buckets (0-30, 31-60, 61-90, 90+ days)
- **Member Dashboard:** Individual view showing attendance summary, dues, fines, contributions
- **Event Dashboard:** Per-event view showing contributions collected vs goal
- **PDF Export:** Attendance records and contribution records exportable as PDF
- **Health endpoint:** `/health` for uptime monitoring
- **Currency:** Nigerian Naira (₦), formatted with comma separators
- **Persistence:** SQLite (WAL mode) at `./data/pillars.db`, with in-memory fallback
- **No authentication** — single-user admin access assumed (committee member uses the app directly)
- **No multi-tenancy** — single cooperative instance

## Brand Commitments

- Name: Pillars Cooperative
- Currency: Nigerian Naira (₦)
- Language: English
- No logo, color palette, or typography constraints established yet

## Evidence on Hand

- Working Go + HTMX + Tailwind application with full CRUD flows
- 4 HTML templates: index (dashboard), member_detail, event_detail, attendance_detail
- 14 passing tests covering attendance fine creation, member balance, and SQLite persistence
- `financial_summaries_suggestions.md` — wishlist for additional PDF statement types (member statements, consolidated reports)
- All monetary values already use Naira formatting

## Product Principles

1. **One source of truth** — every number in the cooperative (who attended, who paid, who owes) lives in one place, not scattered across spreadsheets
2. **Domain-first, not generic** — the app understands cooperative-specific rules (fine deductions consume paid dues, probation lifecycle, contribution auto-generation) rather than forcing users to model them manually
3. **Committee-first workflow** — the primary interface serves the admin/secretary who runs everything; member views are secondary
4. **Simple infrastructure** — single binary, single SQLite file, no external dependencies the host machine can't run
5. **Financial accuracy over speed** — balance calculations and deduction logic must be correct; the cooperative's money is at stake
