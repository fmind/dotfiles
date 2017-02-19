# Configuration file for ipython.

c = get_config()

c.InteractiveShellApp.exec_lines = [
    'import math',
    'import re',
    'import os'
]

c.TerminalInteractiveShell.colors = 'Linux'
c.TerminalIPythonApp.display_banner = False
c.TerminalInteractiveShell.confirm_exit = False
