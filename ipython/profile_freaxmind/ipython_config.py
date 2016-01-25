# Configuration file for ipython.

c = get_config()

c.InteractiveShellApp.exec_lines = [
    'from vocabulary import Vocabulary as vb',
    'import math',
    'import re',
    'import os'
]

c.TerminalInteractiveShell.editor = 'vim'
c.TerminalInteractiveShell.colors = 'Linux'
c.TerminalIPythonApp.display_banner = False
c.TerminalInteractiveShell.confirm_exit = False
c.PromptManager.in_template = '{color.Green}Fx~{color.Cyan}In[{color.Yellow}\\N{color.Cyan}]{color.normal}: '
c.PromptManager.out_template = '  Out[{color.Yellow}\\N{color.Cyan}]{color.normal}: '
