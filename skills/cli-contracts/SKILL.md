---
name: cli-contracts
description: "Engineer CLI contracts across languages: commands, flags, help, streams, exit codes, JSON, completions, cancellation, compatibility, and safe errors."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/cli-contracts
  created: 2026-08-30
  updated: 2026-08-30
---

# Engineer CLI Contracts

Define a command-line interface as a stable human and machine contract before implementing it. Pair with [go-stack](../go-stack/SKILL.md) or [python-stack](../python-stack/SKILL.md) for language-specific code and [systematic-debugging](../systematic-debugging/SKILL.md) for unknown failures.

## Contract

1. **Name the jobs:** Give each command one purpose, a predictable noun/verb hierarchy, and aliases only when they cannot trigger a different class of action.
1. **Specify inputs:** Define positional arguments, flags, environment variables, config precedence, defaults, validation, mutual exclusions, and behavior without a TTY.
1. **Separate streams:** Send requested data to stdout and diagnostics or progress to stderr. Keep `--json` machine-readable, schema-stable, and free of decoration.
1. **Define outcomes:** Map success, usage errors, authentication failures, remote failures, partial results, and cancellation to documented exit codes. Never convert an error into an empty success.
1. **Make help executable:** Ensure root and subcommand help, examples, deprecations, and shell completions are generated from the same command definition.
1. **Protect sensitive context:** Render safe command and argument context for actionable errors, but never include collected values, tokens, provider response bodies, or unredacted stderr.
1. **Handle process behavior:** Respect signals, cancellation, timeouts, piping, broken pipes, and cleanup. Destructive commands need a confirmation mechanism or explicit non-interactive opt-in, while agent execution still requires separate target-scoped user authority; a `--yes` flag is never authority.
1. **Preserve compatibility deliberately:** Treat command names, flags, JSON fields, exit codes, and stdout text consumed by scripts as public API. Prefer deprecation and migration guidance over silent changes.

## Verification

1. Write contract tests for help, parsing, validation, streams, exit codes, JSON schema, cancellation, and representative success and failure paths.
1. Exercise the installed binary from a clean shell, including piped and non-TTY use; do not rely only on direct function tests.
1. Run the focused tests, completion generation, documentation checks, and the repository's full gate.

## Handoff

Summarize the user-visible contract, compatibility impact, exact binary exercised, error-path coverage, and any real provider boundary that remains untested.
