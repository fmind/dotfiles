# Security Policy

## Scope

This repository holds personal dotfiles. It is provided **as-is**, without warranty (see [LICENSE](LICENSE)). Only the tip of `main` is supported — there are no maintained release branches, and older tags receive no fixes.

In scope:

- Secrets or credentials committed to this repository.
- The `dot` CLI ([`dot/`](dot/)) — command injection, path traversal, unsafe file permissions.
- Bootstrap and hook scripts (`install.sh`, `run_once_after_*`, `modify_*`) — anything that executes during `chezmoi apply`.
- Agent skills under [`skills/`](skills/) that instruct an agent to run commands.

Out of scope: vulnerabilities in the upstream tools this repo installs and configures (report those to their own maintainers), and the deliberate trade-offs documented below.

## Known trade-offs

These are intentional and not vulnerabilities in this repository:

- Agent CLIs are configured to skip permission prompts (`bypassPermissions`, `approval_policy = "never"`, `sandbox_mode = "danger-full-access"`). This is a single-user workstation setup; do not adopt it on shared or production hosts.
- `run_once_after_install-antigravity-cli.sh.tmpl` pipes a vendor install script from `antigravity.google` into `bash`.
- `install.sh` pipes the official mise installer from `mise.run` into `bash`.

## Reporting a vulnerability

Report privately via [GitHub Security Advisories](https://github.com/fmind/dotfiles/security/advisories/new), or by email to <mederic.hurier@fmind.dev>. Please do not open a public issue for a suspected vulnerability.

Include the affected file or command, the impact, and reproduction steps. Expect an initial response within 7 days.

## Automated scanning

Every commit is gated by `mise run check`, which runs `gitleaks` over git history and `trivy` across dependencies, IaC, licenses, and secrets. CI re-runs the same tasks on Linux and macOS, and Dependabot opens grouped weekly updates for GitHub Actions and Go modules.
