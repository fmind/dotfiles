"""Configuration for ptpython."""

from __future__ import unicode_literals


def configure(repl) -> None:
    repl.vi_mode = True
    repl.prompt_style = "ipython"
