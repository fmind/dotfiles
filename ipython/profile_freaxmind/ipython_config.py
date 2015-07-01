# Configuration file for ipython.

c = get_config()

c.InteractiveShellApp.exec_lines = [
    'import datetime',
    'import time',
    'import math',
    'import sys',
    'import os',
    'import re',
]

c.TerminalInteractiveShell.editor = 'vim'
c.TerminalInteractiveShell.colors = 'Linux'
c.TerminalIPythonApp.display_banner = False
c.TerminalInteractiveShell.confirm_exit = False
c.PromptManager.in_template = '{color.DarkGray}\T>{color.Green}Fx{color.Cyan}~In[{color.Yellow}\N{color.Cyan}]{color.normal}: '
