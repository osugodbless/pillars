# AGENTS.md — Pillars Cooperative

## Commands
- **Run server**: `go run ./cmd/server` (or `./server` — prebuilt static binary)
- **Tests**: `go test ./...`
- **Build**: `go build -o server ./cmd/server`
- Must run from repo root (templates loaded from relative `templates/`)

## Architecture
- **Entrypoint**: `cmd/server/main.go` — `http.ServeMux` on `:8080`
- **All business logic in one package**: `internal/app/` (store, handlers, render)
- **Store is dual-path**: in-memory Go slices + optional SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- **DB**: auto-created at `./data/pillars.db` (WAL mode, `busy_timeout=5000`, `foreign_keys=ON`)

## Tech
- Go 1.26.3, module `pillars`
- **No framework** — standard `net/http` + `html/template`
- **HTMX**: handlers detect `HX-Request` header to return partials vs full page
- **Tailwind CSS** in templates
- **PDF**: `gofpdf` for attendance & contribution reports

## Domain quirks
- Event creation auto-generates pending `Contribution` records for every active member
- Contribution & fine deductions work by consuming existing paid dues (resets them to `pending`/`partially_paid`, carries forward remainder as `owed`)

## Default settings (in code, no config file)
AbsenceFine=1000, LateFine=500, DuesAmount=2000, ProbationPeriod=90 days

## Tests
- `store_test.go` — 14 tests, mix of in-memory (`NewStore()`) and SQLite-backed (`NewStoreWithPath()`)
- SQLite tests use `t.TempDir()` — no cleanup needed
- No integration test prerequisites

## Repo surface
```
cmd/server/main.go         — HTTP server & route setup
internal/app/store.go      — domain models, persistence, business logic
internal/app/handlers.go   — HTTP handlers (HTMX-aware)
internal/app/store_test.go — tests
templates/*.html           — Go templates with Tailwind
data/pillars.db            — SQLite database (committed)
server                     — prebuilt static binary (build artifact)
```

## gstack
- Use the `/browse` skill from gstack for all web browsing
- Never use `mcp__claude-in-chrome__*` tools
- Available skills: /office-hours, /plan-ceo-review, /plan-eng-review, /plan-design-review, /design-consultation, /design-shotgun, /design-html, /review, /ship, /land-and-deploy, /canary, /benchmark, /browse, /connect-chrome, /qa, /qa-only, /design-review, /setup-browser-cookies, /setup-deploy, /setup-gbrain, /retro, /investigate, /document-release, /document-generate, /codex, /cso, /autoplan, /plan-devex-review, /devex-review, /careful, /freeze, /guard, /unfreeze, /gstack-upgrade, /learn
