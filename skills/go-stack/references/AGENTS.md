# AGENTS.md (Project)

Context and rules for AI agents working in this repository. Humans should start with `README.md`.

## Project overview

- **Name**: <slug>
- **Description**: <1-2 sentences on what this project does.>
- **Language**: Go 1.26+ (`go.mod`), module `<import_path>`.

## Setup & core commands

All work goes through `mise` (see `mise.toml`); git hooks and CI call the same tasks.

- Install: `mise run install` — tidy modules, install git hooks, vendor assets.
- Format: `mise run format` — `goimports` + `gofumpt`, `templ fmt`, `dprint`.
- Check: `mise run check` — `golangci-lint`, `govulncheck`, `dprint check`, `gitleaks`.
- Test: `mise run test` — `gotestsum` with race detector and coverage.
- Build: `mise run build` — compile the binary into `bin/`.
- Watch: `mise run watch` — live reload for local development (web only).

## Definition of done

A change is complete only when, locally, `mise run format` is clean, `mise run check` reports no findings, `mise run test` is green, and new or changed behavior has a test. Fix root causes — never weaken an assertion, add a skip, loosen a type, or suppress a lint error to force a green result.

## Conventions & idioms

- **Errors as values**: wrap with `fmt.Errorf("doing x: %w", err)`; inspect with `errors.Is`/`errors.As`; no panic-based control flow.
- **Context first**: pass `ctx context.Context` as the first argument to any I/O or long-running call, and propagate it down the stack.
- **Config from env**: parse environment variables into a typed struct (`caarlos0/env`); validate and fail fast on startup.
- **Logging**: standard `log/slog` — `TextHandler` in development, `JSONHandler` in production.
- **No hardcoded values**: lift magic numbers, strings, and durations into typed `const` or config.
- **Resource cleanup**: pair acquisition with a deferred `Close()` immediately after the error check.
- **Commits**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`); no attribution in commit messages.

## Repository layout

- `cmd/<slug>/` — entry point (`main.go`): CLI, web daemon, or ADK agent launcher.
- `config/` — typed, env-parsed configuration (`caarlos0/env`), validated fail-fast at startup.
- `<slug>.go` — core library and business logic; `<slug>_test.go` — its tests.
- `mise.toml` — task runner and pinned toolchain; `lefthook.yml` — git hooks.
- `.golangci.yml` — linter and formatter config; `dprint.json` — config/markup formatter.
- `.env` / `.env.example` — environment configuration (never commit secrets).
- (web) `server.go`, `middleware.go`, `telemetry.go`, `templates/`, `static/`, `.air.toml`.
