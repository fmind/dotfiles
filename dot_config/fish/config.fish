# Add common user paths (including Mise global shims to fix Zellij/scripts)
fish_add_path .venv/bin ~/.local/bin ~/.local/share/mise/shims ~/.tfenv/bin ~/.pub-cache/bin ~/flutter/bin /opt/homebrew/bin /opt/homebrew/sbin /usr/local/bin /usr/local/sbin

# CONFIGS
set -g fish_greeting ''

# PRIVATE ENV
if test -e ~/.private
    source ~/.private
end

# KEYBINDINGS (Hybrid vi mode)
function fish_hybrid_key_bindings
    for mode in default insert visual
        fish_default_key_bindings -M $mode
    end
    fish_vi_key_bindings --no-erase
end
set -g fish_key_bindings fish_hybrid_key_bindings

if status is-interactive
    if command -v mise >/dev/null
        mise activate fish | source
    end
    if command -v fzf >/dev/null
        fzf --fish | source
    end
    if command -v zoxide >/dev/null
        zoxide init fish | source
    end
    if command -v starship >/dev/null
        starship init fish | source
    end
end
