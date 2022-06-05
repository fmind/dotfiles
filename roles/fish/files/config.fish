# ALIASES
source $HOME/.aliases
# CONFIGS
set -g fish_greeting ''
# KEYBIND
function fish_hybrid_key_bindings
    for mode in default insert visual
        fish_default_key_bindings -M $mode
    end
    fish_vi_key_bindings --no-erase
end
set -g fish_key_bindings fish_hybrid_key_bindings
# EXPORTS
set -gx EDITOR vim
set -gx LANG en_US.UTF-8
set -gx PYTHONBREAKPOINT ipdb.set_trace
set -gx PATH .venv/bin $HOME/.pyenv/bin $HOME/.local/bin /snap/bin /usr/local/bin /usr/local/sbin /usr/bin /usr/sbin /bin /sbin
# EXTENDS
direnv hook fish | source
starship init fish | source
status is-login; and pyenv init --path | source
status is-interactive; and pyenv init - | source
status --is-login; and status --is-interactive; and exec byobu-launcher
