---
name: python-script
description: Write standalone single-file Python scripts using PEP 723 inline metadata and uv. Use when creating a quick CLI script that needs dependencies without a full project.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/python-script
  created: 2026-07-09
  updated: 2026-08-02
---

# PEP 723 Standalone Python Scripts

Write single-file Python CLI scripts with inline dependency metadata (PEP 723) executed via `uv run` — no virtualenv, no `pyproject.toml`, no project scaffolding required.

## When to Use

- Agent scratch scripts (store in `.agents/tmp/`).
- Any situation where a full Python project is overkill.
- Quick automation, data processing, or one-off CLI tools.

## Template

Base every script on [script.py](references/script.py). Key elements:

1. **Shebang + PEP 723 block** (must be at the top of the file):
   ```python
   #!/usr/bin/env -S uv run --quiet --script
   # /// script
   # requires-python = ">=3.14"
   # dependencies = [
   #     "rich>=15.0.0",
   #     "typer>=0.27.0",
   # ]
   # ///
   ```
1. **CLI framework**: Use `Typer` with `Rich` for argument parsing and formatted output.
1. **Dual consoles**: `Console()` for stdout results, `Console(stderr=True)` for logs/errors.
1. **Typed arguments**: Use `Annotated[..., typer.Argument/Option(...)]` with help text.
1. **Error handling**: Catch exceptions at the CLI boundary with `err.print_exception(show_locals=False)` and exit via `raise typer.Exit(code=1) from None`; never render locals because they can contain secrets.

## Execution

```bash
# Direct execution (after chmod +x)
./script.py input.txt

# Or via uv explicitly
uv run script.py input.txt --verbose
```

`uv` resolves and caches the declared dependencies automatically — no manual install step.

For a durable script, commit its adjacent lockfile so future runs reuse the same resolution:

```bash
uv lock --script script.py
uv run --locked --script script.py input.txt
```

## Guidelines

- **Pin `requires-python`** to the minimum version you need (e.g., `>=3.14`).
- **Declare dependency lower bounds** (e.g., `rich>=15.0.0`) for compatibility. Lower bounds alone are not reproducible; use `uv lock --script` for a durable script.
- **Keep it single-file** — if the script grows beyond ~200 lines or needs multiple modules, switch to a full project via the [python-stack](../python-stack/SKILL.md) skill.
- **No bare `except`** — outside the CLI boundary handler above (which prints the full traceback before exiting non-zero), let unexpected errors propagate.
