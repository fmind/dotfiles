function dirinfos() {
    NDIR=`find . -maxdepth 1 -type d | wc -l`
    NFILE=`find . -maxdepth 1 -type f | wc -l`
    PERM=`stat -c %a .`

    echo "δ${NDIR} φ${NFILE} π${PERM}"
}

function gitinfos() {
    ISGIT=`git branch >/dev/null 2>/dev/null && echo '±'`
    echo "${ISGIT}" 
}

function viminfos() {
    VIMODE="%{$fg_bold[yellow]%}NORMAL%{$reset_color%}"
    echo "${${KEYMAP/vicmd/$VIMODE}/(main|viins)/}"
}

function suinfos () {
    #if [[ $EUID -eq 0 ]]; then
    #    echo "∮"
    #else
    #    echo "∫"
    #fi
}

PROMPT='%{%B%}%{$fg[red]%}∀%m%{$fg[green]%}∃%n%{$fg[blue]%}∧%1d%{$fg[white]%}$(suinfos)%{$reset_color%}%{%b%} '
RPROMPT='%{%b%}%{$fg[white]%} $(viminfos) $(gitinfos) $(dirinfos)%{$reset_color%}'
