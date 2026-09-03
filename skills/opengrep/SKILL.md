---
name: opengrep
description: Run semantic SAST with opengrep and a pinned opengrep-rules checkout as the opt-in check:sast task, with inline suppressions and JSON or SARIF reports. Use for any opengrep scan or rule pin.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/opengrep
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Opengrep

Semantic code-pattern scanning (SAST) with opengrep, the LGPL fork of the Semgrep engine; it is opt-in as `check:sast`, and [secure](../secure/SKILL.md) decides when a repository adopts it.

## Commands

```bash
opengrep scan --error --quiet --config .opengrep/rules/go --config .opengrep/rules/python .  # check:sast: exit 1 on findings, findings only
opengrep scan --error --quiet --config .opengrep/rules/go --json --output opengrep.json .    # machine-readable report
opengrep scan --error --quiet --config .opengrep/rules/go --sarif --output opengrep.sarif .  # SARIF for code scanning
opengrep scan --experimental --config git+https://github.com/opengrep/opengrep-rules#<ref> . # clone-on-run pin, no checkout task
PAGER=cat opengrep scan --help                                                               # the help is a man page
```

## Mise Tasks

The rules are pinned by commit into an ignored directory so the check stays offline and reproducible; add `.opengrep/` to `.gitignore`:

```toml
[tasks."install:sast"]
description = "Fetch the pinned opengrep rules"
run = "test -d .opengrep/rules || (git init -q .opengrep/rules && git -C .opengrep/rules fetch -q --depth 1 https://github.com/opengrep/opengrep-rules <commit> && git -C .opengrep/rules checkout -q FETCH_HEAD)"

[tasks."check:sast"]
description = "Scan source for insecure code patterns (opengrep)"
depends = ["install:sast"]
run = "opengrep scan --error --quiet --config .opengrep/rules/go --config .opengrep/rules/python ." # one --config per language directory
```

The rules repository ships one directory per language, including `go/`, `python/`, `typescript/`, `terraform/`, `dockerfile/`, and `yaml/`.

## Workflow

1. **Adopt**: pin `opengrep` in `mise.toml` `[tools]`, pin an opengrep-rules commit in `install:sast`, and add `check:sast` to the `depends` of `check` per [mise](../mise/SKILL.md).
1. **Triage**: fix each finding; suppress a true false positive inline with `// nosemgrep: <rule id>` plus a reason, where the id is the `id` field of the rule file. Exclude generated paths in `.semgrepignore`.
1. **Report**: pass `--json` or `--sarif` when a report is needed; the gate itself stays text-only.

## Gotchas

- **`--config auto` phones home**: it logs in to the Semgrep registry with the project URL, and registry names (`p/...`) fetch from semgrep.dev; never use them in hooks or CI.
- **`git+` needs `--experimental`**: it clones on every run and pins a branch or tag only; opengrep-rules ships no tags and is archived (frozen since 2025-11), so the commit-pinned checkout task is the reproducible form.
- **A bare `nosemgrep` silences every rule on the line**: always name the rule id.
- **Help is a man page**: `opengrep scan --help` renders through a pager and blocks without a TTY; run it with `PAGER=cat`.

## Documentation

- [opengrep](https://github.com/opengrep/opengrep) · [Wiki](https://github.com/opengrep/opengrep/wiki) · [opengrep-rules](https://github.com/opengrep/opengrep-rules) · [Rule syntax](https://semgrep.dev/docs/writing-rules/rule-syntax)
- Companion skills: [secure](../secure/SKILL.md) (when to adopt), [mise](../mise/SKILL.md) (the `check:sast` name), [trivy](../trivy/SKILL.md) (dependency and IaC scans, not code patterns).
