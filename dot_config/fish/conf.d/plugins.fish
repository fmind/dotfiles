# PLUGINS
if status is-interactive
    if type -q atuin
        atuin init fish | source
    end
    if type -q carapace
        set -gx CARAPACE_BRIDGES 'zsh,fish,bash,inshellisense'
        carapace _carapace | source
    end
    if type -q fzf
        fzf --fish | source
    end
    if type -q mise
        mise activate fish | source
    end
    if type -q starship
        starship init fish | source
    end
    if type -q zellij; and not set -q ZELLIJ; and not set -q TMUX; and test "$TERM_PROGRAM" != "vscode"
        if zellij attach --create main
            exit
        end
    end
    if type -q zoxide
        zoxide init fish | source
    end
end
