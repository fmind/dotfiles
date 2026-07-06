"""Configuration for ptpython."""


def configure(repl) -> None:
    """Configure ptpython REPL."""
    # --- Vi Mode ---
    repl.vi_mode = True
    # --- Behavior ---
    repl.confirm_exit = False
    # --- UI Settings ---
    repl.enable_mouse_support = True
    repl.highlight_matching_parenthesis = True
    repl.prompt_style = "ipython"
    repl.show_docstring = True
    repl.show_line_numbers = True
    repl.show_signature = True
    repl.use_code_colorscheme("material")
    # --- Completion & Suggestion ---
    repl.enable_auto_suggest = True
    repl.enable_history_search = True
    repl.enable_fuzzy_completion = True
