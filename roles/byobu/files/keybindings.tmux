# PANES
bind-key -n C-M-h select-pane -L
bind-key -n C-M-j select-pane -D
bind-key -n C-M-k select-pane -U
bind-key -n C-M-l select-pane -R

# WINDOWS
bind-key -n M-L next-window
bind-key -n M-H previous-window

# SESSIONS
bind-key -n M-J switch-client -n
bind-key -n M-K switch-client -p

# DEFAULTS
set -g prefix F12
unbind-key -n C-a
