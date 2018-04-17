# Configuration file for ipython.

c = get_config()

c.TerminalInteractiveShell.colors = 'Linux'
c.TerminalInteractiveShell.editing_mode = 'vi'
c.TerminalInteractiveShell.confirm_exit = False
c.InteractiveShellApp.extensions = ['autoreload']
c.InteractiveShellApp.exec_lines = ['%autoreload 2']
