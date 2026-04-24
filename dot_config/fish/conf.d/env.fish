set -gx BAT_THEME "Catppuccin Mocha"
set -gx CARAPACE_BRIDGES 'zsh,fish,bash,inshellisense'
set -gx COLORFGBG "15;0"
set -gx COPILOT_ALLOW_ALL true
set -gx COREPACK_ENABLE_AUTO_PIN 0
set -gx DELTA_PAGER "less -RFX"
set -gx EDITOR nvim
set -gx LANG en_US.UTF-8
set -gx LC_ALL en_US.UTF-8
set -gx MANPAGER "nvim +Man!"
set -gx PAGER "bat --plain"
set -gx RIPGREP_CONFIG_PATH ~/.ripgreprc
set -gx VISUAL $EDITOR

if test -e ~/.private.fish
    source ~/.private.fish
end
