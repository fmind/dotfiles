# Core setup
set -gx EDITOR nvim
set -gx VISUAL nvim
set -gx LANG en_US.UTF-8
set -gx PAGER "bat --plain"
set -gx MANPAGER "nvim +Man!"

# Ripgrep config
set -gx RIPGREP_CONFIG_PATH ~/.ripgreprc

# GitHub Copilot setup
set -gx COPILOT_ALLOW_ALL true
