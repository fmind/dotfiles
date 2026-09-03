---
name: python-script
description: Write standalone single-file Python scripts with PEP 723 inline metadata run by uv. Use for a quick CLI script that needs dependencies without a full project.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/python-script
  created: "2026-07-09"
  updated: "2026-09-03"
---

# PEP 723 Standalone Python Scripts

Single-file Python CLI scripts with inline dependency metadata (PEP 723) run by `uv run` — no virtualenv, no `pyproject.toml`; a script that outgrows one file moves to [python-stack](../python-stack/SKILL.md).

## Workflow

1. **Start from the template**: copy [script.py](references/script.py); its shebang (`#!/usr/bin/env -S uv run --quiet --script`) and `# /// script` block declare `requires-python` and dependency lower bounds.
1. **Parse arguments with Typer**: `Annotated[..., typer.Argument/Option(...)]` with help text; Rich `Console()` for stdout results and `Console(stderr=True)` for logs and errors.
1. **Handle errors at the boundary**: catch in the command, `err.print_exception(show_locals=False)` (locals can hold secrets), then `raise typer.Exit(code=1) from None`; elsewhere let errors propagate.
1. **Run**: `chmod +x script.py && ./script.py input.txt`, or `uv run script.py input.txt`; uv resolves and caches the dependencies on first run.
1. **Lock a durable script**: `uv lock --script script.py`, then `uv run --locked --script script.py`; lower bounds alone are not reproducible.

## Gotchas

- **Agent scratch scripts** live in `.agents/tmp/`.
- **One file**: past ~200 lines or a second module, switch to a full project.

## Documentation

- [PEP 723](https://peps.python.org/pep-0723/) · [uv scripts](https://docs.astral.sh/uv/guides/scripts/) · [Typer](https://typer.tiangolo.com/)
- Companion skills: [python-stack](../python-stack/SKILL.md) (full projects), [cli-contracts](../cli-contracts/SKILL.md) (flags, streams, exit codes).
