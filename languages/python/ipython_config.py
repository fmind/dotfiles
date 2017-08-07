# Configuration file for ipython.

c = get_config()

c.InteractiveShellApp.exec_lines = [
    'from collections import *',
    'from itertools import *',
    'from functools import *',
    'import math',
    'import re',
    'import os',
]

c.TerminalInteractiveShell.colors = 'Linux'
c.TerminalInteractiveShell.editing_mode = 'vi'
c.TerminalInteractiveShell.confirm_exit = False
