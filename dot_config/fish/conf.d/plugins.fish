# PLUGINS
if command -q mise
    mise activate fish | source
    if not status is-interactive
        mise hook-env -s fish | source
    end
end

if status is-interactive
    if command -q carapace
        carapace _carapace | source
    end
    if command -q fzf
        fzf --fish | source
    end
    if command -q atuin
        atuin init fish | source
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
