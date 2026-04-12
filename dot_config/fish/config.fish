# PATHS
fish_add_path -g .venv/bin ~/.local/bin ~/.local/share/mise/shims /opt/homebrew/bin /opt/homebrew/sbin /usr/local/bin /usr/local/sbin /usr/bin /usr/sbin /bin /sbin

# CONFIGS
set -g fish_greeting ''

# PRIVATES
if test -e ~/.private.fish
    source ~/.private.fish
end

# KEYBINDINGS
function fish_hybrid_key_bindings
    for mode in default insert visual
        fish_default_key_bindings -M $mode
    end
    fish_vi_key_bindings --no-erase
end
set -g fish_key_bindings fish_hybrid_key_bindings

if status is-interactive
    if command -v atuin >/dev/null
        atuin init fish | source
    end
    if command -v carapace >/dev/null
        set -gx CARAPACE_BRIDGES 'zsh,fish,bash,inshellisense'
        carapace _carapace | source
    end
    if command -v fzf >/dev/null
        fzf --fish | source
    end
    if command -v mise >/dev/null
        mise activate fish | source
    end
    if command -v starship >/dev/null
        starship init fish | source
    end
    if command -v zellij >/dev/null; and not set -q ZELLIJ; and not set -q TMUX
        zellij attach --create main
        if test $status -eq 0
            exit
        end
    end
    if command -v zoxide >/dev/null
        zoxide init fish | source
    end
end
