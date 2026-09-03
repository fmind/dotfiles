---
name: ast-grep
description: Search and rewrite code structurally with ast-grep patterns, meta-variables, YAML rules, and JSON output. Use for AST-aware code search, lint rules, or safe bulk refactors.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/ast-grep
  created: "2026-09-03"
  updated: "2026-09-03"
---

# ast-grep

Structural code search and rewrite: a pattern is real code with meta-variables, matched against the syntax tree, so it ignores formatting and never matches inside strings or comments. Use it where `rg` gives false positives and where a refactor must touch every call site exactly once; plain text search stays with `rg`.

## Commands

```bash
ast-grep run -p 'print($A)' -l python                                          # search; run is the default subcommand
ast-grep run -p 'console.log($$$ARGS)' -r 'logger.info($$$ARGS)' -l ts         # dry run: prints the diff, changes nothing
ast-grep run -p 'console.log($$$ARGS)' -r 'logger.info($$$ARGS)' -l ts --update-all   # apply after reviewing the dry run (-i to confirm per hunk)
ast-grep run -p 'func _() { os.Getenv($K) }' --selector call_expression -l go  # Go calls need a function context (see Gotchas)
ast-grep run -p 'print($A)' -l python --json=compact                           # structured output; --json=stream gives one object per line
ast-grep scan                                                                  # every rule in sgconfig.yml
ast-grep scan -r rules/no-print.yml --format github                            # one rule file; GitHub annotations in CI
```

## Workflow

1. **Write the pattern as code**: `$NAME` matches one node, `$$$NAME` a sequence (arguments, statements), `$_` a node without binding; always pass `-l <lang>` so the pattern parses in the right grammar, and use `--debug-query=ast` when a pattern that should match does not.
1. **Search first**: run without `-r`, read the matches with `-C 2` for context, and tune `--globs` or `--no-ignore` when files are skipped.
1. **Rewrite in two steps**: add `-r` to see the diff, then `--update-all` (or `-i` for an interactive session); captured meta-variables are reused in the replacement.
1. **Promote to a rule**: for a lint or a repeated refactor, `ast-grep new project` scaffolds `sgconfig.yml` and `rules/`; a rule file has `id`, `language`, `rule` (`pattern`, `kind`, `inside`, `has`, `not`), optional `fix`, `severity`, and `message`; `ast-grep test` runs its `valid` and `invalid` cases.
1. **Wire into the gate**: run `ast-grep scan` inside `check:lint` (see [mise](../mise/SKILL.md)) so hooks and CI apply the same rules.

## Gotchas

- **Go call patterns**: a bare `pkg.Func($A)` parses as a type conversion and matches nothing; wrap it as `func _() { pkg.Func($A) }` with `--selector call_expression` (in a rule: `pattern: {context: ..., selector: call_expression}`).
- **Meta-variables are uppercase**: `$a` is plain text; `$A`, `$ARGS`, `$_` are meta-variables.
- **Pattern must be a complete node**: `foo(` does not parse; match `foo($$$)` and narrow with `--selector`.
- **Rewrite scope**: `-r` replaces the whole matched node, not a substring inside it.
- **Language ids**: `go`, `python`, `typescript` (`ts`), `tsx`, `rust`; under `scan` the file extension picks the grammar.

## Official Skills

Upstream: `ast-grep/agent-skill`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add ast-grep/agent-skill --list
skills add ast-grep/agent-skill --skill <name> -y
```

## Documentation

- [ast-grep guide](https://ast-grep.github.io/guide/introduction.html) · [Pattern syntax](https://ast-grep.github.io/guide/pattern-syntax.html) · [Rule reference](https://ast-grep.github.io/reference/rule.html) · [Languages](https://ast-grep.github.io/reference/languages.html)
- Companion skills: [reduce-complexity](../reduce-complexity/SKILL.md) (bulk simplifications), [go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md), [typescript-stack](../typescript-stack/SKILL.md) (language linters).
