# Core system setup
set -gx EDITOR nvim
set -gx LANG en_US.UTF-8
set -gx MANPAGER "nvim +Man!"
set -gx PAGER bat
set -gx VISUAL nvim

# Ripgrep config path
set -gx RIPGREP_CONFIG_PATH "$HOME/.ripgreprc"

# FZF default to fd, hide git
set -gx FZF_ALT_C_COMMAND "fd --type d --hidden --exclude .git"
set -gx FZF_ALT_C_OPTS "--preview 'eza --tree --color=always {} | head -200'"
set -gx FZF_CTRL_T_COMMAND "$FZF_DEFAULT_COMMAND"
set -gx FZF_CTRL_T_OPTS "--preview 'bat -n --color=always {}'"
set -gx FZF_DEFAULT_COMMAND "fd --type f --hidden --exclude .git"
