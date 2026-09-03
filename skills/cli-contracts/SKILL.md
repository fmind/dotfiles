---
name: cli-contracts
description: Design a CLI's command, flag, stream, exit-code, and JSON contract before coding it. Use when adding or changing a CLI surface.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/cli-contracts
  created: "2026-08-30"
  updated: "2026-09-03"
---

# CLI Contracts

Define a command-line interface as a stable human and machine contract before implementing it. [go-stack](../go-stack/SKILL.md) and [python-stack](../python-stack/SKILL.md) own the language code; [systematic-debugging](../systematic-debugging/SKILL.md) owns unknown failures.

## Workflow

1. **Name the jobs**: one purpose per command, a predictable noun-verb hierarchy, and aliases only when they cannot trigger a different class of action.
1. **Specify inputs**: positional arguments, flags, environment variables, config precedence, defaults, validation, mutual exclusions, and behavior without a TTY.
1. **Separate streams**: requested data to stdout, diagnostics and progress to stderr; `--json` stays machine-readable, schema-stable, and free of decoration.
1. **Define outcomes**: map success, usage errors, authentication failures, remote failures, partial results, and cancellation to documented exit codes; never turn an error into an empty success.
1. **Make help executable**: root and subcommand help, examples, deprecations, and shell completions generate from the same command definition.
1. **Protect sensitive context**: errors show safe command and argument context, never collected values, tokens, provider response bodies, or unredacted stderr.
1. **Handle process behavior**: respect signals, cancellation, timeouts, piping, broken pipes, and cleanup; a destructive command needs a confirmation or an explicit non-interactive opt-in such as `--yes`.
1. **Preserve compatibility deliberately**: command names, flags, JSON fields, exit codes, and script-consumed stdout are public API; prefer deprecation and migration guidance over silent changes.
1. **Test the contract**: help, parsing, validation, streams, exit codes, JSON schema, cancellation, and representative success and failure paths, then `mise run check` and `mise run test`.
1. **Exercise the installed binary**: from a clean shell, piped and non-TTY; direct function tests alone do not prove the contract.

## Documentation

- [Command Line Interface Guidelines](https://clig.dev)
- Companion skills: [go-stack](../go-stack/SKILL.md) (Go CLIs), [python-stack](../python-stack/SKILL.md) (Python CLIs), [systematic-debugging](../systematic-debugging/SKILL.md) (failures of unknown cause).
