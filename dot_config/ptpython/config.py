"""Configuration for ptpython."""

from __future__ import unicode_literals


def configure(repl):
    repl.vi_mode = True
    repl.prompt_style = "ipython"
    repl.show_status_bar = True
    repl.show_line_numbers = True
    repl.highlight_matching_parenthesis = True
    repl.use_code_colorscheme("tango")
