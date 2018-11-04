# set -g prefix F12
# unbind-key -n C-a

# is_vim="ps -o state= -o comm= -t '#{pane_tty}' \
    # | grep -iqE '^[^TXZ ]+ +(\\S+\\/)?g?(view|n?vim?x?)(diff)?$'"

# PANES
# bind -n C-h run "(tmux display-message -p '#{pane_current_command}' | grep -iqE '(^|\/)n?vim(diff)?$|emacs.*$' && tmux send-keys C-h) || tmux select-pane -L"
# bind -n C-j run "(tmux display-message -p '#{pane_current_command}' | grep -iqE '(^|\/)n?vim(diff)?$|emacs.*$' && tmux send-keys C-j) || tmux select-pane -D"
# bind -n C-k run "(tmux display-message -p '#{pane_current_command}' | grep -iqE '(^|\/)n?vim(diff)?$|emacs.*$' && tmux send-keys C-k) || tmux select-pane -U"
# bind -n C-l run "(tmux display-message -p '#{pane_current_command}' | grep -iqE '(^|\/)n?vim(diff)?$|emacs.*$' && tmux send-keys C-l) || tmux select-pane -R"

# WINDOWS
# bind-key -n M-l next-window
# bind-key -n M-h previous-window

# SESSIONS
# bind-key -n M-k switch-client -n
# bind-key -n M-j switch-client -p
