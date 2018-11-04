# PANES
bind-key -n C-M-h select-pane -L
bind-key -n C-M-j select-pane -D
bind-key -n C-M-k select-pane -U
bind-key -n C-M-l select-pane -R

# WINDOWS
bind-key -n C-M-o next-window
bind-key -n C-M-i previous-window

# SESSIONS
bind-key -n C-M-n switch-client -n
bind-key -n C-M-p switch-client -p

# DEFAULTS
set -g prefix F12
unbind-key -n C-a
