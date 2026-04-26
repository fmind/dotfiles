# Locale + editors
set -gx EDITOR nvim
set -gx VISUAL $EDITOR
set -gx LANG en_US.UTF-8
set -gx LC_ALL en_US.UTF-8

# Pagers
set -gx PAGER "bat --plain"
set -gx MANPAGER "nvim +Man!"
set -gx DELTA_PAGER "less -RFX"
set -gx BAT_PAGER "less -RF"
set -gx LESS -FRSXMK
set -gx LESSHISTFILE -

# Terminal
set -gx CARAPACE_BRIDGES 'zsh,fish,bash,inshellisense'
set -gx COLORFGBG "15;0"
set -gx COPILOT_ALLOW_ALL true
set -gx COREPACK_ENABLE_AUTO_PIN 0
set -gx RIPGREP_CONFIG_PATH ~/.config/ripgrep/config
