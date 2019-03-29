# CONFIGS {{{
set -g fish_greeting ''
set -g fish_prompt_pwd_dir_length 0
# }}}
# EXPORTS {{{
set -gx EDITOR vim
set -gx LANG en_US.UTF-8
set -gx PATH $HOME/bin $HOME/.local/bin /usr/local/bin /usr/local/sbin /usr/bin /usr/sbin /bin /sbin
# }}}
# PROMPTS {{{
function fish_prompt
    set -l last_status $status

    set_color brred
    printf "%s" (whoami)
    set_color normal
    printf " at "
    set_color brgreen
    printf "%s" (hostname)
    set_color normal
    printf " in "
    set_color brblue
    printf "%s" (prompt_pwd)
    set_color brmagenta
    printf "%s" (__fish_git_prompt)
    set_color normal
    printf "\n"

    if test $last_status -ne 0
        printf "! "
    else if test (id -u) -eq 0
        printf "# "
    else
        printf "\$ "
    end
end

function fish_right_prompt
    set_color cyan

    if set -q VIRTUAL_ENV
        printf "[%s]" (basename $VIRTUAL_ENV)
    end

    set_color normal
end
# }}}
# SESSIONS {{{
status --is-login; and status --is-interactive; and exec byobu-launcher
# }}}
