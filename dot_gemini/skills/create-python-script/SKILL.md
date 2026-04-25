---
name: create-python-script
description: Generate a standalone Python CLI script with uv, Rich, and Typer
---

# Create Python Script

This skill guides you in creating elegant, standalone, production-grade Python CLI scripts using **uv**, **Rich**, and **Typer**.

## Script Template

```python
#!/usr/bin/env -S uv run --quiet --script
# /// script
# requires-python = ">=3.14"
# dependencies = [
#     "httpx>=0.28.1",
#     "rich>=15.0.0",
#     "tenacity>=9.1.4",
#     "typer>=0.24.2",
# ]
# ///

from pathlib import Path
from typing import Annotated, Optional

import typer
from rich.console import Console

app = typer.Typer(add_completion=False, rich_markup_mode="rich")
err = Console(stderr=True)

@app.command()
def main(
    input_file: Annotated[Path, typer.Argument(help="Path to process", exists=True, dir_okay=False)],
    output_dir: Annotated[Optional[Path], typer.Option("--output", "-o", help="Output directory")] = None,
    verbose: Annotated[bool, typer.Option("--verbose", "-v", help="Show debug logs")] = False,
) -> None:
    """A concise description of what this script does goes here."""
    try:
        err.print(f"[green]✓[/green] Successfully processed {input_file}")
    except Exception:
        err.print_exception(show_locals=True)
        raise typer.Exit(code=1)

if __name__ == "__main__":
    app()
```

## Core Principles

- **Minimalist & Functional**: Rely on **Typer** for all CLI arguments, options, and validation.
- **Modern Python 3.14+**: Leverage modern idioms like `t-strings` and `Annotated` for clean, self-documenting code.
- **CLI UX**: Use **Rich** (`err = Console(stderr=True)`) for outputs, feedback, and error reporting.
- **Robustness**: Use **httpx** for I/O and **tenacity** for retries.
- **Self-Contained**: The `uv` shebang and PEP 723 metadata block are MANDATORY.

## AI Agent Instructions

- **Zero Setup**: Always provide scripts that can be run directly with `uv run script.py`.
- **Verification**: Run `uv run script.py [args]` to ensure dependencies resolve and basic functionality works.
- **Dependency Management**: Use `uv add --script <filename> "<pkg>"` to add new dependencies to the metadata block.
- **Fatal Errors**: Always use `err.print_exception(show_locals=True)` in `try...except` blocks for fatal errors to provide a full stack trace with local variables for debugging.
- **Type Safety**: Use `Annotated` for all CLI arguments and options.
