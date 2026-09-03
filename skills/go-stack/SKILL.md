---
name: go-stack
description: Build Go projects, libraries, CLIs, TUIs, GOTH web apps, or ADK agents with the standard package layout and pinned tooling. Use for any Go work.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/go-stack
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Go Stack Standard

Canonical Go development: scaffolding, libraries, CLI/TUI tools, GOTH web apps, and the Go API side of ADK agents. The agent workflow belongs to [google-adk](../google-adk/SKILL.md); content sites belong to [hugo](../hugo/SKILL.md).

## 1. Core & Quality Stack

- **Go version**: target latest stable; adopt current idioms mechanically with `go fix ./...` and `gofumpt` rather than hand-porting.
- **Tooling split**: standalone CLIs (`golangci-lint`, `gotestsum`, `air`, `esbuild`) come from `mise.toml`; `tool` directives in `go.mod` hold `templ`, `goimports`, `gofumpt`, and `govulncheck`, run as `go tool <name>`.
- **Tasks and hooks**: [mise.toml](references/mise.toml) (web) or [mise-cli.toml](references/mise-cli.toml) (CLI/agent) per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) or [lefthook-cli.yml](references/lefthook-cli.yml) per [lefthook](../lefthook/SKILL.md).
- **Linting and formatting**: `goimports` + `gofumpt` format Go; `golangci-lint` ([golangci.yml](references/golangci.yml)) enforces zero warnings; `dprint` keeps config and markup per [dprint](../dprint/SKILL.md).
- **Testing**: standard `testing` (starters stay dependency-free), `stretchr/testify` when richer assertions pay off, run through `gotestsum`.
- **Logging**: `log/slog` — `TextHandler` locally with a dynamic `slog.LevelVar`, `JSONHandler` in production with trace context.
- **Observability**: OpenTelemetry for web and agents only; `SetupOTel` ([telemetry.go](references/telemetry.go)) exports OTLP and stamps trace ids on logs when `OTEL_EXPORTER_OTLP_ENDPOINT` is set — see [observability](../observability/SKILL.md).
- **Security checks**: `govulncheck` is `check:vuln` and `gitleaks` is `check:leaks`; SAST is opt-in per [opengrep](../opengrep/SKILL.md).
- **Runtime defaults**: Go is container-aware (no `automaxprocs`); use `json:",omitzero"` instead of `omitempty`.

## 2. Project Scaffolding Workflow

1. **Information**: define `Slug`, `Import Path` (e.g. `github.com/username/slug`), and `Holder/Year`.
1. **Bootstrap**: `go mod init <import_path>` records the active toolchain version in `go.mod`.
1. **Tasks and hooks by project type**, saved as `mise.toml` and `lefthook.yml`:
   - Web: [mise.toml](references/mise.toml) + [lefthook.yml](references/lefthook.yml) (templ, Tailwind, vendor, and watch tasks).
   - CLI/agent: [mise-cli.toml](references/mise-cli.toml) + [lefthook-cli.yml](references/lefthook-cli.yml) (same vocabulary, no web tasks).
1. **Config files**:
   - `.golangci.yml` from [golangci.yml](references/golangci.yml) — replace `<import_path>` there and in the `format:go` task so `format` and `check` agree.
   - `dprint.json` per [dprint](../dprint/SKILL.md); `.air.toml` from [air.toml](references/air.toml) (web only).
   - `.env.example` from [env.example](references/env.example) (uncomment what the project type uses), `.gitignore` from [gitignore](references/gitignore).
   - `AGENTS.md` from [AGENTS.md](references/AGENTS.md) (drop the `(web)` lines for CLI/agent); `LICENSE` per [project-license](../project-license/SKILL.md).
1. **Toolchain**: `mise trust && mise install`, then `go get -tool golang.org/x/tools/cmd/goimports mvdan.cc/gofumpt golang.org/x/vuln/cmd/govulncheck` (web adds `github.com/a-h/templ/cmd/templ`).
1. **Sources**:
   - `cmd/<slug>/main.go` from [main.go](references/main.go) (web), [cli.go](references/cli.go) (CLI), or [agent.go](references/agent.go) (agent, plus `go get google.golang.org/adk/v2`).
   - `<slug>.go` from [lib.go](references/lib.go) with [lib_test.go](references/lib_test.go); `config/config.go` from [config.go](references/config.go) (CLI/agent may drop `Port`).
   - Web: [server.go](references/server.go), [server_test.go](references/server_test.go), [middleware.go](references/middleware.go), [telemetry.go](references/telemetry.go).
   - Web templates and assets: [layout.templ](references/layout.templ), [home.templ](references/home.templ), [styles.css](references/styles.css), [app.js](references/app.js), [user-card.js](references/user-card.js).
   - Web vendoring: `scripts/vendor.go` from [vendor.go](references/vendor.go), run once by `install:vendor`.
1. **Validate**: `git init --initial-branch=main`, then `mise run install`, `mise run format`, `mise run check`, `mise run test`; `check:leaks` prints `no commits yet` until the first commit.
1. **Finish**: keep this stack's `AGENTS.md` when running [agent-project](../agent-project/SKILL.md), write `README.md` per [readme-agents](../readme-agents/SKILL.md), then `git add . && git commit -m "chore: initial commit"`.

## 3. Database & Persistence

- **SQL first**: raw SQL with `sqlc`-generated types over `jackc/pgx/v5` (`pgxpool` with explicit bounds and timeouts); schema linting and migrations per [atlas](../atlas/SKILL.md), or `goose` for versioned SQL migrations.

## 4. Web Stack (GOTH)

- **Router**: `http.ServeMux` path-value routing; `go-chi/chi/v5` when middleware stacks grow.
- **Type-safe REST**: Huma (`github.com/danielgtaylor/huma/v2`) for OpenAPI 3.1 and JSON Schema validation.
- **UI components**: Templ co-locates markup, styling, and state; an Alpine component moves to `assets/js/components/<name>.js` once it grows methods or is reused ([home.templ](references/home.templ) shows both).
- **Tailwind CSS v4**: CSS-first config compiled by the standalone `tailwindcss` binary from mise; no Node toolchain.
- **JavaScript**: `esbuild` bundles `assets/js/app.js` into `static/js/dist.js`; skip the bundler while the client side is a couple of inline snippets, and record that decision in `AGENTS.md`.
- **`assets/` vs `static/`**: authored sources live in `assets/`; only build output and vendored libraries live in `static/`, which `server.go` embeds whole — a source left in `static/` ships in the binary.
- **Self-hosted assets**: HTMX, Alpine, CSS, and JS are served from embedded `/static/` with content-hash cache busting; one binary, no CDN, no runtime fetch.
- **Production HTTP**: explicit `http.Server` timeouts; `SetupOTel` in `main` and `NewAppHandler` wrapping the router in `otelhttp` (see §1).

## 5. CLI & TUI

- **Framework**: `urfave/cli/v3` ([cli.go](references/cli.go)); flags, streams, exit codes, and completions follow [cli-contracts](../cli-contracts/SKILL.md).
- **Dual CLI/library**: domain logic and types in the root package, command wiring in `cmd/<slug>/main.go`.
- **TUI**: Bubble Tea and the Charm layout tools import from `charm.land/{bubbletea,lipgloss,bubbles}/v2`, not `github.com/charmbracelet`.

## 6. ADK Agents (Go API)

The agent workflow, `agents-cli`, model-pin rationale, and deployment live in [google-adk](../google-adk/SKILL.md); this section keeps the Go API notes behind the [agent.go](references/agent.go) starter.

- **Module**: `google.golang.org/adk/v2` (requires Go 1.26+); read its API from `~/go/pkg/mod`, not memory.
- **Agents and tools**: `llmagent.New(llmagent.Config{...})` (`SubAgents` for trees); `functiontool.New` wraps typed functions with `jsonschema` tags; `tool/geminitool` and `tool/mcptoolset` add built-ins and MCP.
- **Model and auth**: `gemini.NewModel` on the Gemini Enterprise Agent Platform (formerly Vertex AI) with ADC (`GOOGLE_CLOUD_PROJECT`/`GOOGLE_CLOUD_LOCATION` parsed by `caarlos0/env`); the pinned model name lives in the starter.
- **Entry point**: `full.NewLauncher()` owns the CLI (`console` and `web` modes; `web` hosts `webui`, `api`, `a2a`, and Cloud triggers) and parses its own flags, so keep `urfave/cli/v3` for non-agent tools.
- **Streaming and tracing**: consume runs with `for event, err := range …` (`iter.Seq2`); the launcher wires OpenTelemetry itself (`OTEL_EXPORTER_OTLP_ENDPOINT` or `--otel_to_cloud`).

## 7. Configuration

- **Environment-first**: `caarlos0/env/v11` parses env vars into a typed `Config` in the `config` package ([config.go](references/config.go)); `config.Load()` validates on startup and exits 1 through `slog` on failure.
- **Typed environments**: model `development`/`production` as an enum so the literals never appear at call sites.

## 8. Project Layouts

The CLI + library and web + library trees live in [layouts.md](references/layouts.md); every file there maps to a reference in §2.

## 9. Go Idioms

- **Deterministic concurrency tests**: `testing/synctest` virtualizes clocks for goroutine-heavy code.
- **Receiver consistency**: pointer receivers for state, sync fields, or large structs; value receivers for small immutable values; never mix on one type.
- **Zero-value usability**: design structs so the zero value works without a constructor (`sync.Mutex`, `bytes.Buffer`).
- **Pre-allocate**: `make([]T, 0, n)` / `make(map[K]V, n)` when the size is known (`prealloc` lints it).

## Gotchas

- **Tools are not auto-installed**: `run_auto_install = false` means `mise install` must run once after `mise trust`.
- **Embedded assets do not hot-reload**: use `mise run watch` (air + Tailwind + esbuild watchers); `.air.toml` excludes `assets/` on purpose so only the watchers' writes into `static/` trigger a rebuild.
- **Commit generated code**: `*_templ.go`, `static/css/dist.css`, and `static/js/dist.js` are committed because `check` compiles `server.go` without running generators; CI's clean-tree check catches staleness.
- **Vendored libraries**: `scripts/vendor.go` pins HTMX and Alpine by URL and sha256 and `install:vendor` skips when present; bump a version by editing URL and hash together, never through npm.
- **Alpine load order**: [layout.templ](references/layout.templ) loads `dist.js` before `alpine.min.js` because `Alpine.data()` registrations must exist at `alpine:init`; reversing the tags fails silently.
- **`ko` is per project**: `build:image` needs `go get -tool github.com/google/ko` per [containerize](../containerize/SKILL.md).
- **Transitive vulnerabilities**: `govulncheck` flags modules you never imported (e.g. `grpc` via OTel/ADK); fix with `go get -u <module> && go mod tidy`, not a `require` pin — fresh ADK agents hit this on the first `check`.

## Documentation

- [Go](https://go.dev/doc/) · [Templ](https://templ.guide) · [HTMX](https://htmx.org) · [Alpine.js](https://alpinejs.dev) · [Tailwind CSS](https://tailwindcss.com) · [esbuild](https://esbuild.github.io/) · [ADK for Go](https://google.github.io/adk-docs/get-started/go/)
- Companion skills: [google-adk](../google-adk/SKILL.md) (agent workflow), [cli-contracts](../cli-contracts/SKILL.md), [containerize](../containerize/SKILL.md), [github-actions](../github-actions/SKILL.md), [secure](../secure/SKILL.md), [readme-agents](../readme-agents/SKILL.md).
