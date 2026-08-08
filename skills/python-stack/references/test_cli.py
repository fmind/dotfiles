from __future__ import annotations

import sys

import pytest
from typer.testing import CliRunner

from <package> import app, main

runner = CliRunner()


def test_greet() -> None:
    result = runner.invoke(app, ["--name", "Ada"])

    assert result.exit_code == 0
    assert result.stdout == "Hello, Ada!\n"


def test_console_entrypoint(monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.setattr(sys, "argv", ["<slug>", "--name", "Ada"])

    with pytest.raises(SystemExit) as exit_info:
        main()

    assert exit_info.value.code == 0
    assert capsys.readouterr().out == "Hello, Ada!\n"
