---
name: security-scan
description: Scan a repo with Trivy (deps, IaC, secrets, licenses, images) and gitleaks (git history), then triage findings. Use for a full-repo security audit beyond the stack's native checks.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/security-scan
  created: 2026-07-04
  updated: 2026-07-09
---

# Security Scanning

Scan a repository (and its images) for vulnerabilities, misconfigurations, secrets, and license issues. Layers **Trivy** (config: `dot_config/trivy/trivy.yaml` — `HIGH`/`CRITICAL`, `ignore-unfixed`, scanners: vuln/misconfig/secret/license) over **gitleaks** for deep git-history secret detection, complementing the language-native scanners (`govulncheck`, `pip-audit`).

## Workflow

1. **Repository scan** (dependencies, misconfig, secrets, licenses in one pass):
   ```bash
   trivy --config trivy.yaml fs .   # always pass --config: an exported TRIVY_CONFIG outranks ./trivy.yaml — see Gotchas
   ```
1. **Targeted config scan** (Dockerfiles, Kubernetes manifests, Terraform — optional if already done via `trivy fs`):
   ```bash
   trivy --config trivy.yaml config .
   ```
1. **Secrets in git history** (deeper than a working-tree scan — catches committed-then-deleted secrets):
   ```bash
   mise run check:leaks          # full history
   mise run check:leaks --staged # pre-commit scope
   ```
1. **Container image scan** (after a build — see [containerize](../containerize/SKILL.md)):
   ```bash
   trivy --config trivy.yaml image <registry>/<image>:<tag>
   ```
1. **Language-native depth** (already wired into `mise run check` as `check:vuln`):
   - Go: `go tool govulncheck ./...`
   - Python: `uv run pip-audit`
1. **Triage**: summarize findings by severity, separate fixable from `unfixed`, and propose the minimal upgrade/patch or a justified ignore.

## Mise Task Integration

Expose security scans in the project's canonical task vocabulary (`mise.toml`) per the [mise skill](../mise/SKILL.md). `check:leaks` and `check:scan` are the same in every repo; `check:vuln` is the language-native scanner where one exists (`govulncheck`/`pip-audit` — see [go-stack](../go-stack/SKILL.md)/[python-stack](../python-stack/SKILL.md)) and only falls back to `trivy fs` in polyglot or config-only repos with no native scanner:

```toml
[tasks."check:leaks"]
alias = "s"
description = "Audit codebase for leaked secrets (gitleaks)"
run = "gitleaks git --verbose" # history scan; `mise run check:leaks --staged` auto-appends --staged for pre-commit scope

[tasks."check:scan"]
description = "Scan configuration files (IaC, manifests, Dockerfiles) for misconfigurations"
run = "trivy --config trivy.yaml config ." # mise auto-appends extra path args; do NOT use `${@:-.}` (mise skill: it collapses to `.` and only appends)

[tasks."check:vuln"]
# Fallback for repos with no native scanner; a Go/Python repo uses govulncheck/pip-audit here instead.
description = "Scan project files and dependencies for vulnerabilities"
run = "trivy --config trivy.yaml fs ." # --config is mandatory: an exported TRIVY_CONFIG would otherwise win
```

## Gotchas

- **Signal over noise**: `ignore-unfixed: true` and `HIGH`/`CRITICAL` keep results actionable; widen severity only for a deliberate audit.
- **Trivy config resolution**: precedence is `--config` > `TRIVY_CONFIG` > `./trivy.yaml` > built-in defaults — trivy never auto-loads `~/.config/trivy/`. On this machine `dot_config/fish/conf.d/env.fish` exports `TRIVY_CONFIG=~/.config/trivy/trivy.yaml` globally, so **a bare `trivy fs .` silently ignores the project `trivy.yaml`** and scans with the global policy instead. Always pass `--config trivy.yaml` from a repo that ships one (this is why the `check:scan`/`check:vuln` tasks below do). Falling through to built-in defaults is the other failure mode: **all** severities, no `ignore-unfixed`, and the signal-over-noise settings gone. For reproducible local **and** CI scans, commit a project `trivy.yaml` (copy the global one — mirrors [dprint](../dprint/SKILL.md)).
- **Pre-commit gates need `--staged`**: `gitleaks git` scans committed history, so it can **not** block a secret in the commit being made — it isn't in history yet (returns clean). Reserve history mode for CI/on-demand and give the pre-commit hook its own staged scan so incoming secrets are actually gated. In [lefthook](../lefthook/SKILL.md) `pre-commit`, add (with a `priority` after the formatters):
  ```yaml
  check:leaks:
    priority: 20 # after the formatters (10), before the whole-tree `check` (30)
    run: mise run check:leaks --staged # → gitleaks git --verbose --staged
  ```
  Keep the history-mode `check:leaks` inside `mise run check` for CI's full-history pass.
- **Suppressions are auditable**: record accepted risks with a clear reason, never by lowering the global bar:
  - **Trivy**: Add CVEs or paths to `.trivyignore` (one per line, comments prefixed with `#`).
  - **Gitleaks**: Use inline `gitleaks:allow` comments next to false positives, or configure rules/ignores in `.gitleaks.toml` / `.gitleaksignore`.
- **CI parity**: CI runs `mise run check`, so the `check:leaks` scan runs there too — but a shallow CI checkout limits its git-history scope. For full-history secret + dependency scanning in CI, add a dedicated `security` job (`gitleaks git` + `trivy fs`, `fetch-depth: 0`) and gate on a non-zero exit; otherwise run `mise run check:leaks` and `trivy fs` on demand.
- **Secret rotation**: a leaked secret is compromised even after removal from history — rotate it, don't just delete the commit.

## Documentation

- [Trivy Documentation](https://trivy.dev) · [gitleaks](https://github.com/gitleaks/gitleaks)
- Companion skills:
  - [lefthook](../lefthook/SKILL.md) — the `check:leaks` pre-commit hook.
  - [github-actions](../github-actions/SKILL.md) — runs `mise run check` in CI (includes the `check:leaks` scan); add a `security` job for full-history/Trivy scanning.
  - [containerize](../containerize/SKILL.md) — image build + scan/sign.
  - [go-stack](../go-stack/SKILL.md) / [python-stack](../python-stack/SKILL.md) — native `govulncheck` / `pip-audit`.
