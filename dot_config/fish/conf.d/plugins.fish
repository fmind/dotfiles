# PLUGINS
if command -q mise
    if status is-interactive
        mise activate fish | source
    else
        mise activate fish --shims | source
    end
end

if status is-interactive
    set -l fish_cache_dir "$HOME/.cache/fish"
    if set -q XDG_CACHE_HOME
        set fish_cache_dir "$XDG_CACHE_HOME/fish"
    end
    if command -q carapace
        set -l carapace_init "$fish_cache_dir/carapace-init.fish"
        if test -r "$carapace_init"
            source "$carapace_init"
        else
            carapace _carapace | source
        end
    end
    if command -q fzf
        fzf --fish | source
    end
    if command -q atuin
        # Atuin's UUID helper is expensive; use the OS UUID source before loading
        # the cached integration, while preserving its one-session-per-shell rule.
        if not set -q ATUIN_SESSION; or test "$ATUIN_SHLVL" != "$SHLVL"
            if test -r /proc/sys/kernel/random/uuid
                set -gx ATUIN_SESSION (string trim </proc/sys/kernel/random/uuid)
            else
                set -gx ATUIN_SESSION (command -q uuidgen; and uuidgen | string lower; or atuin uuid)
            end
            set -gx ATUIN_SHLVL $SHLVL
        end
        set -l atuin_init "$fish_cache_dir/atuin-init.fish"
        if test -r "$atuin_init"
            source "$atuin_init"
        else
            atuin init fish | source
        end
    end
    if command -q starship
        starship init fish | source
    end
    if command -q zoxide
        zoxide init fish | source
    end
    # Auto-start Zellij cleanly after entire shell initialization is complete
    function auto_zellij --on-event fish_prompt
        functions -e auto_zellij
        if command -q zellij; and not set -q ZELLIJ; and not set -q TMUX; and test "$TERM_PROGRAM" != vscode; and not set -q NVIM; and not set -q SSH_CONNECTION; and not set -q SSH_CLIENT; and not set -q SSH_TTY
            exec zellij attach --create main
        end
    end
end
