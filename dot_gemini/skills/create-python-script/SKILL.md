name: create-python-script
description: Generate clean, standalone Python CLI scripts with uv, Typer, and Loguru

---

# Create Python Script

This skill guides you in creating elegant, standalone, production-grade Python CLI scripts using **uv**, **Typer**, and **Loguru**.

## Script Template (Python 3.14+)

```python
#!/usr/bin/env -S uv run --quiet
# /// script
# requires-python = ">=3.14"
# dependencies = [
#     "httpx>=0.28.1",
#     "loguru>=0.7.3",
#     "rich>=14.3.3",
#     "tenacity>=9.1.4",
#     "typer>=0.24.1",
# ]
# [tool.uv]
# exclude-newer = "2026-04-10T00:00:00Z"
# ///

import sys
from pathlib import Path
from typing import Annotated, Optional

import typer
from loguru import logger
from rich.console import Console

app = typer.Typer(rich_markup_mode="rich")
console = Console()

@app.command()
def main(
    input_file: Annotated[Path, typer.Argument(help="Path to process", exists=True, dir_okay=False)],
    output_dir: Annotated[Optional[Path], typer.Option("--output", "-o", help="Output directory")] = None,
    verbose: Annotated[bool, typer.Option("--verbose", "-v", help="Show debug logs")] = False,
):
    """ A concise description of what this script does goes here."""
    # Simple loguru configuration
    logger.remove()
    logger.add(sys.stderr, level="DEBUG" if verbose else "INFO")

    logger.debug(f"Starting process for {input_file}")
    try:
        # Implementation logic
        console.print(f"[green]✓[/green] Successfully processed {input_file}")
    except Exception:
        logger.exception("An unexpected fatal error occurred")
        raise typer.Exit(code=1)

if __name__ == "__main__":
    app()
```

## Core Principles

1.  **Minimalist & Functional**: Rely on **Typer** for all CLI arguments, options, and validation.
2.  **Reproducibility**: Always use `[tool.uv]` with `exclude-newer` to lock dependencies in time.
3.  **Modern Python 3.14+**: Leverage modern idioms like `t-strings` and `Annotated` for clean, self-documenting code.
4.  **CLI UX**: Use **Rich** for user-facing feedback and **Loguru** for technical logs.
5.  **Robustness**: Use **httpx** for I/O and **tenacity** for retries. Handle all fatal errors with `logger.exception` to capture stack traces.
6.  **Self-Contained**: The `uv` shebang and PEP 723 metadata block are MANDATORY.

## AI Agent Instructions

- **Zero Setup**: Always provide scripts that can be run directly with `uv run script.py`.
- **Verification**: After creating the script, ALWAYS test it using `uv run script.py [args]` to ensure dependencies resolve and basic functionality works.
- **Dependency Management**: Use `uv add --script <filename> "<pkg>"` to add new dependencies to the metadata block.
- **Fatal Errors**: Always use `logger.exception` in `try...except` blocks for fatal errors to provide a full stack trace for debugging.
- **Type Safety**: Use `Annotated` for all CLI arguments and options.
