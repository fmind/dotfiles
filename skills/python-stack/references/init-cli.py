"""<description>."""

from __future__ import annotations

from typing import Annotated

import typer

__version__ = "0.1.0"
app = typer.Typer(no_args_is_help=True, pretty_exceptions_show_locals=False)


@app.command()
def greet(name: Annotated[str, typer.Option(help="Name to greet")] = "world") -> None:
    """Print a greeting."""
    typer.echo(f"Hello, {name}!")


def main() -> None:
    """Entry point invoked by the `<slug>` console script or `uv run`."""
    app()
