---
name: go-stack
description: Canonical Go development stack — tooling, scaffolding, web (GOTH), CLI/TUI, and ADK agents. Use for any Go (Golang) project, library, or application.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/go-stack
  created: 2026-06-23
  updated: 2026-07-06
---

# Go Stack Standard (Go 1.26+)

Canonical guidelines for Go development: project scaffolding, libraries, CLI/TUI tools, web apps (GOTH), and ADK agents.

## 1. Core & Quality Stack

- **Go Version**: Target latest stable (Go 1.26+). Use modern idioms: `new(expr)` to allocate an initialized value, self-referential generic constraints (F-bounded), and the rewritten `go fix` modernizers (e.g. `omitempty` → `omitzero`).
- **Dependency & Tooling**: Manage heavy CLI tools (`golangci-lint`, `lefthook`, `gotestsum`) via `mise.toml` ([mise.toml](references/mise.toml)) to prevent `go.mod` dependency bloat and compile errors. Use native `tool` directive in `go.mod` (Go 1.24+) only for code generators and small utilities (`templ`, `goimports`, `gofumpt`, `govulncheck`). Run via `go tool <name>`.
- **Project Tools**:
  - **Via `mise.toml`**: `golangci-lint` (v2+), `lefthook`, `gotestsum`.
  - **Via `go.mod`**: `templ`, `sqlc`, `gomodifytags`, `impl`, `goimports`, `gofumpt`, `govulncheck`.
- **Task Runner & Hooks**: Use `mise.toml` ([mise.toml](references/mise.toml)) for commands (`install`, `format`, `check`, `test`, `build`, `watch`) per the [mise skill](../mise/SKILL.md). Use `lefthook` ([lefthook.yml](references/lefthook.yml)) for pre-commit (format → `check:leaks --staged` → check) and pre-push (test) per the [lefthook skill](../lefthook/SKILL.md). Order the hook commands with explicit `priority` (formatters low, `check` high) so formatters restage before `check` reads the files — lefthook sorts alphabetically otherwise.
- **Linting & Formatting**: Clean markdown/JSON/YAML with `dprint` using the configuration maintained by the [dprint skill](../dprint/SKILL.md). Format Go files with `goimports`/`gofumpt`. Enforce zero-warning rule with `golangci-lint` ([golangci.yml](references/golangci.yml)).
- **Testing**: Standard library `testing` (starters stay dependency-free); reach for `stretchr/testify` when you want richer assertions. Run with `gotestsum`.
  - Use `testing/synctest` (Go 1.25+) for deterministic concurrent testing (virtualized clocks).
  - Use `testing/cryptotest.SetGlobalRandom` (Go 1.26+) to make crypto key generation deterministic in tests.
- **Logging**: Use standard `log/slog`. Local: `slog.TextHandler` (or `charmbracelet/log`) with dynamic `slog.LevelVar`. Production: `slog.JSONHandler` with OpenTelemetry span/trace context propagation.
- **Diagnostics & Observability**: Continuous trace buffering via `runtime/trace.FlightRecorder` (Go 1.25+) to dump on error/panic. Standardize on OpenTelemetry (`otel`) for traces — wired for **web and agents only**, not CLIs. The web starter's `SetupOTel` ([telemetry.go](references/telemetry.go)) installs an OTLP/HTTP exporter and wraps the router in `otelhttp` for per-request spans; `OtelHandler` then stamps the active `trace_id`/`span_id` onto every `slog` record. Tracing activates only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set, so local runs stay quiet. ADK agents inherit tracing from the launcher (same OTLP env var, or `--otel_to_cloud` for GCP Cloud Trace).
- **Container Awareness**: Go 1.25+ is cgroups/container-aware—`GOMAXPROCS` aligns automatically (do not use `automaxprocs`).
- **Omit Zero Struct Tags**: Use native `json:",omitzero"` (Go 1.24+) on structs/types (e.g., `time.Time`) instead of `omitempty`.

## 2. Project Scaffolding Workflow

1. **Information**: Define project `Slug`, `Import Path` (e.g. `github.com/username/slug`), and `Holder/Year`.
1. **Go Pinning**: Pin target version in `go.mod`.
1. **Bootstrap**: Run `go mod init <import_path>` in the project root.
1. **Config Initialization**: Copy and customize configurations:
   - `mise.toml` ([mise.toml](references/mise.toml)) — this reference is the **web** variant. For **CLI/agent** projects, delete the web-only tasks (`build:css`, `build:templ`, `format:templ`, `install:vendor`, `watch`/`watch:css`/`watch:air`) and the `tailwindcss` tool, then remove the now-dangling `depends` references to them: `build:go` (drop the whole line), `build:image` (drop the whole line), `format` (remove `format:templ`), `install:go` (drop the whole line), `install` (remove `install:vendor`), and `test`/`test:watch` (drop the whole line).
   - `.golangci.yml` ([golangci.yml](references/golangci.yml))
   - `lefthook.yml` ([lefthook.yml](references/lefthook.yml))
   - `dprint.json` (setup as instructed in the [dprint skill](../dprint/SKILL.md))
   - `.air.toml` ([air.toml](references/air.toml)) (for web)
   - `.env`/`.env.example` ([env.example](references/env.example)), `LICENSE` ([LICENSE](references/LICENSE)), and `.gitignore` ([gitignore](references/gitignore))
   - `AGENTS.md` ([AGENTS.md](references/AGENTS.md)) — Go-specific agent context (trim the web-only layout line for CLI/agent projects)
   - Run `mise trust` so the mise-shimmed `go` and `mise run` tasks are allowed to execute against the new project config, then add tool dependencies: `go get -tool golang.org/x/tools/cmd/goimports mvdan.cc/gofumpt golang.org/x/vuln/cmd/govulncheck` (**web** projects also need the templ generator: `go get -tool github.com/a-h/templ/cmd/templ`).
1. **Scaffold Sources**:
   - Write entry point `cmd/<slug>/main.go` using reference template [main.go](references/main.go) (web), [cli.go](references/cli.go) (CLI), or [agent.go](references/agent.go) (ADK agent — also run `go get google.golang.org/adk/v2`).
   - Write core library files `<slug>.go` using [lib.go](references/lib.go) and `<slug>_test.go` using [lib_test.go](references/lib_test.go).
   - Write the typed config package `config/config.go` using [config.go](references/config.go) (`go mod tidy` pulls in `caarlos0/env`). CLI/agent projects may drop the web-only `Port` field; the [agent.go](references/agent.go) starter adds Vertex `GOOGLE_CLOUD_PROJECT`/`GOOGLE_CLOUD_LOCATION` (auth via ADC — `gcloud auth application-default login`), so put those in the agent's `.env.example` instead of `PORT`.
   - For web:
     - Copy web server components: `server.go` ([server.go](references/server.go)), `server_test.go` ([server_test.go](references/server_test.go)), `middleware.go` ([middleware.go](references/middleware.go)), and `telemetry.go` ([telemetry.go](references/telemetry.go)).
     - Copy templates ([layout.templ](references/layout.templ), [home.templ](references/home.templ)), stylesheet ([styles.css](references/styles.css)), and run `go run scripts/vendor.go` (using [vendor.go](references/vendor.go)) to self-host assets.
1. **Git & Validation**:
   - Run `git init --initial-branch=main`.
   - Run verification sequence:
     ```bash
     mise run install
     mise run format
     mise run check
     mise run test
     ```
   - Initialize workspace agent context via the [agent-project skill](../agent-project/SKILL.md), but **keep the Go-specific [AGENTS.md](references/AGENTS.md) copied in step 4** — skip agent-project's generic AGENTS.md template so it does not overwrite the stack file.
   - Create `README.md` (humans) and keep it in sync with `AGENTS.md` via the [readme-agents skill](../readme-agents/SKILL.md); the layouts (§8) list it but no reference template ships one.
   - Commit changes: `git add . && git commit -m "chore: initial commit"`.

## 3. Database & Persistence Stack

- **ORM Rejection**: Write raw SQL and generate type-safe structures using `sqlc`.
- **Postgres Driver**: Use `jackc/pgx/v5` and `jackc/pgx/v5/pgxpool` with explicit pool bounds and connection timeouts.
- **Schema Management**: Use `Atlas` for declarative schema linting and migrations.
- **Versioned Migrations**: Use `goose` (`github.com/pressly/goose/v3`) for versioned SQL- or Go-based database migrations.
- **High-Availability**: Deploy PostgreSQL using the **CloudNativePG** operator on Kubernetes.

## 4. Web Stack & Serving Standard

- **Router**: Use `http.ServeMux` for native path-value routing (Go 1.22+). For middleware-heavy applications, use `go-chi/chi/v5`.
- **Type-Safe REST**: Use **Huma** (`github.com/danielgtaylor/huma/v2`) to build APIs with automatic OpenAPI 3.1 & JSON Schema validation.
- **UI Components (GOTH)**: Co-locate CSS/Alpine.js with `Templ`. Pass server-side structures to client-side components using JSON serialization.
- **Tailwind CSS v4**: CSS-First configuration compiled via the standalone `tailwindcss` binary managed by `mise` (no Node.js/JavaScript dependencies).
- **Static Assets (Self-Hosted)**:
  - Serve all assets (HTMX, Alpine, CSS) locally from `/static/` (embedded via `go:embed`).
  - Cache-bust via in-memory content hash (`?v=hash`) and set long-term cache headers.
- **Production HTTP**: Configure explicit `http.Server` timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`).
- **Tracing**: Call `SetupOTel(ctx, serviceName)` in `main` and defer its shutdown; `NewAppHandler` wraps the router in `otelhttp` so each request is a span and logs carry its `trace_id` (see §1 Diagnostics). Requires `go.opentelemetry.io/otel/{sdk,exporters/otlp/otlptrace/otlptracehttp}` and `contrib/instrumentation/net/http/otelhttp` — pulled in by `go mod tidy`.

## 5. CLI & TUI Stack

- **CLI Framework**: Use **urfave/cli/v3** for declarative, callback-driven multi-command CLIs — used by the [cli.go](references/cli.go) starter.
- **TUI**: Use **Bubble Tea** (`charm.land/bubbletea/v2`) and the Charm layout tools (`charm.land/lipgloss/v2`, `charm.land/bubbles/v2`) for rich interactive terminals/forms. Note: the stable v2 modules import from `charm.land`, not `github.com/charmbracelet`.
- **Completions**: Generate dynamic shell completions from CLI framework native integrations.

## 6. AI Agent Stack (ADK v2)

- **Framework**: Use the **Agent Development Kit for Go** (`google.golang.org/adk/v2`, Go 1.25+) — the code-first, model-agnostic toolkit optimized for Gemini. Scaffold with the [agent.go](references/agent.go) starter (`go get google.golang.org/adk/v2`).
- **Agents**: Build with `llmagent.New(llmagent.Config{...})` — set `Name`, `Model`, `Instruction`, and `Tools`; delegate to `SubAgents` for multi-agent trees.
- **Models & Auth**: `gemini.NewModel(ctx, "gemini-3.5-flash", &genai.ClientConfig{...})`. The [agent.go](references/agent.go) starter defaults to **Vertex AI with Application Default Credentials** — `Backend: genai.BackendVertexAI` plus `GOOGLE_CLOUD_PROJECT`/`GOOGLE_CLOUD_LOCATION` (parsed fail-fast via `caarlos0/env`), credentials from `gcloud auth application-default login` locally or the attached service account in prod (no key stored). For AI Studio instead, drop the Vertex fields and set `APIKey` from `GOOGLE_API_KEY`. Logging goes through the shared `config` package (env-aware `slog` handler), not `log.Fatalf`.
- **Tools**: Wrap typed Go functions with `functiontool.New` (generic `[TArgs, TResults]`; `jsonschema` struct tags document fields). Add built-ins from `tool/geminitool` (Google Search) or connect external servers via `tool/mcptoolset` (MCP).
- **Entry Point**: `full.NewLauncher()` serves the agent through a CLI with two top-level modes — `console` (interactive) and `web` (HTTP server); the `web` mode hosts the `webui` (dev UI), `api` (REST), and `a2a` sublaunchers — no custom wiring. Invoking with an **unrecognized** mode prints the usage and exits non-zero; invoking with **no** arguments defaults to interactive `console` mode. Note: ADK owns this CLI; reserve `urfave/cli/v3` for non-agent tools.
- **Streaming**: Agent runs return `iter.Seq2[*session.Event, error]`; consume with `for event, err := range …` — never collect events into a slice.
- **Observability**: The launcher wires OpenTelemetry itself — pin `service.name` via `launcher.Config.TelemetryOptions` (`telemetry.WithResource`), then export by setting `OTEL_EXPORTER_OTLP_ENDPOINT` (OTLP) or passing `--otel_to_cloud` (GCP Cloud Trace, reusing the Vertex ADC). ADK emits spans for model and tool calls; no manual `TracerProvider` needed (unlike the web variant).

## 7. Configuration Standard

- **Environment-First**: Use environment variables for all configuration (Twelve-Factor App).
- **Strong Typing**: Parse configs into a typed `Config` struct in a dedicated `config` package — see [config.go](references/config.go) for the starter's `Config`, `Load()`, and env-aware `NewHandler` (drives the log level/format and, for web, the listen port and HSTS). Model the environment as a typed enum so `"development"`/`"production"` never appear as literals at call sites.
- **Parsing**: Use `caarlos0/env/v11` for parsing env vars (`env.ParseAs[Config]()`; `env:"NAME,required"` / `envDefault` tags).
- **Fail-Fast**: Validate configuration immediately on startup via `config.Load()`; log through `slog` and `os.Exit(1)` if validation fails.

## 8. Project Layouts

### CLI + Library Project Layout

```text
<slug>/
├── cmd/
│   └── <slug>/
│       └── main.go         // CLI entry point
├── config/
│   └── config.go           // Typed environment configuration
├── .env
├── .env.example
├── .gitignore
├── .golangci.yml
├── dprint.json
├── lefthook.yml
├── mise.toml
├── AGENTS.md
├── LICENSE
├── <slug>.go               // Library entry point
├── <slug>_test.go          // Unit tests
└── README.md
```

### Web + Library Project Layout

```text
<slug>/
├── cmd/
│   └── <slug>/
│       └── main.go         // Daemon entry point
├── config/
│   └── config.go           // Typed environment configuration
├── .env
├── .env.example
├── .gitignore
├── .golangci.yml
├── .air.toml
├── dprint.json
├── lefthook.yml
├── mise.toml
├── AGENTS.md
├── LICENSE
├── <slug>.go               // Core business logic / client
├── <slug>_test.go          // Core business logic tests
├── server.go               // HTTP handler and asset serving definitions
├── server_test.go          // HTTP routing and integration tests
├── middleware.go           // Standard HTTP middlewares
├── telemetry.go            // OpenTelemetry setup (SetupOTel) + slog trace correlation
├── README.md
├── scripts/
│   └── vendor.go
├── static/
│   ├── css/
│   │   ├── styles.css
│   │   └── dist.css
│   └── vendor/
│       ├── htmx.min.js
│       └── alpine.min.js
└── templates/
    ├── home.templ
    └── layout.templ
```

## 9. Go Coding Standards & Idioms

- **No Hardcoded Values**: Extract magic values (numbers, strings, durations, error messages) into meaningful, typed constants (`const`) or structured configurations (`config.Config`).
  - Use file- or package-level `const` blocks for internal static values.
  - Use structured, environment-parsed config structs for operational values.
- **Conscious Concurrency**: Use goroutines (`go func()`) intentionally to leverage concurrency.
  - Always manage goroutine lifecycles using `context.Context` for cancellation and timeouts.
  - Avoid goroutine leaks by ensuring every spawned goroutine has a guaranteed exit condition.
  - Use safe synchronization primitives: standard channels, `sync.Mutex`, `sync.WaitGroup`, or `golang.org/x/sync/errgroup` (for groups of goroutines returning errors).
  - Test concurrent logic deterministically using `testing/synctest`.
- **Explicit Error Wrapping**: Avoid flat, string-only errors. Wrap lower-level errors with context using `fmt.Errorf("doing action: %w", err)`.
  - Inspect wrapped errors using `errors.Is` or `errors.As`.
  - Avoid panic-based control flows; handle errors as values.
- **Context Propagation**: Always pass `ctx context.Context` as the first argument in any function that performs network, database, or filesystem I/O, or long-running computations.
  - Propagate context down the call stack to respect caller-initiated cancellation and timeouts.
- **Immediate Resource Cleanup**: Pair resource acquisition (file handles, network connections, HTTP response bodies, database transactions) with a deferred close statement (`defer resource.Close()`) immediately after checking for errors to prevent leaks.
- **Allocation Optimization**: Pre-allocate slices and maps using `make([]T, 0, capacity)` or `make(map[K]V, capacity)` when the target size or capacity is known in advance to avoid redundant heap allocations and resize overhead.
- **Consistent Receiver Types**: Maintain consistency in struct receiver types. Use pointer receivers (`*T`) if the struct modifies state, contains synchronization fields (like `sync.Mutex`), or is large. Use value receivers (`T`) for small, immutable data transfer objects. Never mix receiver types on the same struct.
- **Zero-Value Usability**: Design structs so their zero-state is safe and ready to use immediately without explicit constructors where possible (e.g., standard libraries like `sync.Mutex` or `bytes.Buffer`).

## Gotchas & Guidelines

- **Go Tool Directive**: Tool management in Go 1.24+ requires direct `go tool` invocations. For `golangci-lint` (v2+), use module `github.com/golangci/golangci-lint/v2/cmd/golangci-lint` with a `version: "2"` config schema.
- **Port Conflicts**: Default address is `:8080`.
- **Tailwind v4 CLI**: Compiled via the standalone `tailwindcss` CLI executable, installed automatically via `mise` (using the `github` backend: `"github:tailwindlabs/tailwindcss"`).
- **Embedded Asset Updates**: Since assets are embedded via `go:embed`, running `go run` does not hot-reload static assets. Use `air` or rebuild assets to see updates.
- **Committed Generated Code**: The reference `.gitignore` does not exclude `*_templ.go`, so commit the generated Templ code — `check`/CI compile `server.go` (which imports `templates`) without first running `build:templ` (only `test`/`build` regenerate it). `mise run test`'s `build:templ` keeps it fresh and CI's `git diff --exit-code` catches staleness.
- **Self-Hosted Assets**: Never reference CDNs; serve all assets locally.
- **No JS/TS Toolchain**: The GOTH stack is deliberately Node-free — Tailwind is the standalone `tailwindcss` binary, HTMX/Alpine are vendored. Never introduce `npm`/`npx`/`node`.
- **Container Builds (`ko`)**: `mise run build:image` needs `ko` from the [containerize skill](../containerize/SKILL.md) (`go get -tool github.com/google/ko`); it is not installed by default.
- **ADK Agents**: Require Go 1.25+ and a `GOOGLE_API_KEY` (AI Studio) or Vertex AI ADC. `full.NewLauncher()` is cobra-based and separate from `urfave/cli/v3`.

## Documentation

- [Go Documentation](https://go.dev/doc/)
- [Templ Documentation](https://templ.guide)
- [HTMX Documentation](https://htmx.org)
- [Alpine.js Documentation](https://alpinejs.dev)
- [Tailwind CSS v4](https://tailwindcss.com)
- [ADK for Go](https://google.github.io/adk-docs/get-started/go/) — Agent Development Kit (agents).
- Companion skills:
  - [github-actions](../github-actions/SKILL.md) — CI that runs these same `mise run` gates.
  - [security-scan](../security-scan/SKILL.md) — audit dependencies, secrets, and licenses.
  - [containerize](../containerize/SKILL.md) — package the binary into a minimal, signed image.
  - [readme-agents](../readme-agents/SKILL.md) — keep `README.md` and `AGENTS.md` in sync.
