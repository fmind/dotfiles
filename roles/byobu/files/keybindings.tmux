# CONFIGS
set -g prefix F12
unbind-key -n C-a
# BINDINGS
bind-key -n M-H select-pane -L
bind-key -n M-J select-pane -D
bind-key -n M-K select-pane -U
bind-key -n M-L select-pane -R
bind-key -n C-M-l next-window
bind-key -n C-M-h previous-window
bind-key -n C-M-j switch-client -n
bind-key -n C-M-k switch-client -p
