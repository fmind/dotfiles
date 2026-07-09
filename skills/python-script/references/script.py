#!/usr/bin/env -S uv run --quiet --script
# /// script
# requires-python = ">=3.14"
# dependencies = [
#     "rich>=15.0.0",
#     "typer>=0.24.2",
# ]
# ///

from pathlib import Path
from typing import Annotated

import typer
from rich.console import Console

app = typer.Typer(add_completion=False, rich_markup_mode="rich")
err = Console(stderr=True)  # stderr: logs and errors
out = Console()  # stdout: results


@app.command()
def main(
    input_file: Annotated[Path, typer.Argument(help="Path to process", exists=True, dir_okay=False)],
    output_dir: Annotated[Path | None, typer.Option("--output", "-o", help="Output directory")] = None,
    verbose: Annotated[bool, typer.Option("--verbose", "-v", help="Show debug logs")] = False,
) -> None:
    """A concise description of what this script does goes here."""
    try:
        target = output_dir or input_file.parent
        if verbose:
            err.print(f"[dim]Processing {input_file} -> {target}[/dim]")
        # ... do the real work here ...
        out.print(f"[green]✓[/green] Successfully processed {input_file}")
    except Exception:
        err.print_exception(show_locals=True)
        raise typer.Exit(code=1) from None


if __name__ == "__main__":
    app()
