# Core system setup
set -gx EDITOR nvim
set -gx LANG en_US.UTF-8
set -gx MANPAGER "nvim +Man!"
set -gx PAGER bat
set -gx VISUAL nvim

# Ripgrep config path
set -gx RIPGREP_CONFIG_PATH "$HOME/.ripgreprc"

# Catppuccin Mocha colors for tools
set -gx GLAMOUR_STYLE catppuccin-mocha
set -gx JQ_COLORS "1;30:0;39:0;39:0;39:0;32:1;39:1;39"
