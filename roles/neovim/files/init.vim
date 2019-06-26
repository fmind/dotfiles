" PROG
set shell=/bin/bash
set clipboard=unnamedplus
if executable('ag')
  set grepprg=ag\ --nogroup\ --nocolor
endif
" MENU
set wildmode=list:longest,full
set completeopt=menuone,longest
" SPELL
set thesaurus+=/usr/share/dict/theses
set dictionary+=/usr/share/dict/words
" BUFFER
set hidden
set confirm
set autoread
set autowrite
" SEARCH
set path+=**
set hlsearch
set incsearch
set smartcase
set ignorecase
" INDENT
set expandtab
set shiftround
set shiftwidth=4
" WINDOW
set number
set linebreak
set shortmess=I
set scrolloff=10
set relativenumber
" FOLDER
set foldmethod=syntax
set foldlevelstart=99
" PLUGIN
let g:loaded_netrw = 1
let g:loaded_matchparen=1
let g:loaded_netrwPlugin = 1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'benmills/vimux'
Plug 'itchyny/lightline.vim'
Plug 'JuliaEditorSupport/julia-vim'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
Plug 'justinmk/vim-sneak'
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
Plug 'szw/vim-g'
Plug 'tomasr/molokai'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-rsi'
Plug 'tpope/vim-surround'
Plug 'Valloric/YouCompleteMe', {'do': './install.py'}
let g:ycm_collect_identifiers_from_tags_files = 0
let g:ycm_auto_trigger = 0
Plug 'w0rp/ale'
let g:ale_set_quickfix = 1
let g:ale_sign_column_always = 1
let b:ale_fixers = {'python': ['black', 'isort']}
let b:ale_linters = {'python': ['mypy', 'pylint']}
call plug#end()
" COLOR
try
    colorscheme molokai
catch
    colorscheme zellner
endtry
" KEYMAP
noremap j gj
noremap k gk
noremap <cr> :
noremap gl :nohl<cr>
" LEADER
let mapleader=" "
let maplocalleader=";"
noremap <leader>a :Ag<cr>
noremap <leader>b :Buffers<cr>
noremap <leader>c :Colors<cr>
noremap <leader>d :YcmCompleter GetDoc<cr>
noremap <leader>e :YcmCompleter GoToDeclaration<cr>
noremap <leader>f :Files<cr>
noremap <leader>g :GFiles<cr>
noremap <leader>h :bprevious<cr>
noremap <leader>i :Lines<cr>
noremap <leader>j :cnext<cr>
noremap <leader>k :cprevious<cr>
noremap <leader>l :bnext<cr>
noremap <leader>m :Marks<cr>
noremap <leader>n :BCommits<cr>
noremap <leader>o :YcmCompleter GoToDefinition<cr>
noremap <leader>p :Commands<cr>
noremap <leader>q :bdelete<cr>
noremap <leader>r :History<cr>
noremap <leader>s :Tags<cr>
noremap <leader>t :YcmCompleter GoTo<cr>
noremap <leader>u :YcmCompleter GoToReferences<cr>
noremap <leader>v :BTags<cr>
noremap <leader>w :Windows<cr>
noremap <leader>x :History:<cr>
noremap <leader>y :YcmCompleter GetType<cr>
noremap <leader>z :Filetypes<cr>
noremap <leader>` :Locate 
noremap <leader>- :Maps<cr>
noremap <leader>= :ALEFix<cr>
noremap <leader>[ :ALEPreviousWrap<cr>
noremap <leader>] :ALENextWrap<cr>
noremap <leader>' :Helptags<cr>
noremap <leader>, :edit $MYVIMRC<cr>
noremap <leader>. :Google 
noremap <leader>/ :History/<cr>
noremap <leader>\ :BLines<cr>
noremap <leader><cr> :make<cr>
noremap <leader><tab> :b#<cr>
noremap <leader><space> :make 
noremap <localleader>; :make %<cr>
noremap <localleader>a :make add<cr>
noremap <localleader>b :call VimuxSlime(join(getline(1, '$'), "\n"))<cr>
noremap <localleader>c :make clean<cr>
noremap <localleader>d :call VimuxSlime("pydoc ".input("Symbol: "))<cr>
noremap <localleader>e "vy :call VimuxSlime(@v)<cr>
noremap <localleader>f :make format<cr>
noremap <localleader>g :make hook<cr>
noremap <localleader>h :call VimuxSlime("%paste")<cr>
noremap <localleader>i :VimuxInterruptRunner<cr>
noremap <localleader>j "vY :call VimuxSlime(@v)<cr>
noremap <localleader>k :call VimuxSlime("python ".bufname("%"))<cr>
noremap <localleader>l :call VimuxSlime("pylint ".bufname("%"))<cr>
noremap <localleader>m :make all<cr>
noremap <localleader>n :VimuxInspectRunner<cr>
noremap <localleader>o :make publish<cr>
noremap <localleader>p :make package<cr>
noremap <localleader>q :VimuxCloseRunner<cr>
noremap <localleader>r :call VimuxSlime("pytest ".bufname("%"))<cr>
noremap <localleader>s :make sort<cr>
noremap <localleader>t :make test<cr>
noremap <localleader>u :VimuxPromptCommand<cr>
noremap <localleader>v :make venv<cr>
noremap <localleader>w :make work<cr>
noremap <localleader>x :call VimuxOpenRunner()<cr>
noremap <localleader>y :make type<cr>
noremap <localleader>z :call VimuxSlime("mypy ".bufname("%"))<cr>
noremap <localleader><space> :VimuxRunLastCommand<cr>
" AUTOCMD
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
" FUNCTION
function! VimuxSlime(text)
    call VimuxSendText(a:text)
    if a:text !~ '\n$'
        call VimuxSendKeys("Enter")
    endif
endfunction
