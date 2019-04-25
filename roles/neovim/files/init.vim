" INIT
set shortmess=I
" PROG
set path+=**
set shell=/bin/bash
set clipboard=unnamedplus
if executable('ag')
  set grepprg=ag\ --nogroup\ --nocolor
endif
" MENU
set completeopt=menuone
set wildmode=list:longest,full
" SPELL
set spelllang=en,fr
set thesaurus+=/usr/share/dict/theses
set dictionary+=/usr/share/dict/words
" BUFFER
set hidden
set confirm
set autoread
set autowrite
" SEARCH
set hlsearch
set incsearch
" INDENT
set expandtab
set shiftround
set shiftwidth=4
" WINDOW
set number
set linebreak
set scrolloff=10
set relativenumber
" FOLDER
set foldmethod=syntax
set foldlevelstart=99
" PLUGIN
let g:loaded_netrw = 1
let g:loaded_matchparen=0
let g:loaded_netrwPlugin = 1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'benmills/vimux'
Plug 'itchyny/lightline.vim'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
Plug 'justinmk/vim-sneak'
Plug 'tomasr/molokai'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-rsi'
Plug 'tpope/vim-surround'
Plug 'Valloric/YouCompleteMe', {'do': './install.py'}
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
noremap <leader>. :set 
noremap <leader>/ :History/<cr>
noremap <leader>\ :BLines<cr>
noremap <leader><cr> :make<cr>
noremap <leader><tab> :b#<cr>
noremap <leader><space> :make 
noremap <localleader>; :make %<cr>
noremap <localleader>a :make add<cr>
" noremap <localleader>b :make <cr>
noremap <localleader>c :make clean<cr>
" noremap <localleader>d :make <cr>
" noremap <localleader>e :make <cr>
noremap <localleader>f :make format<cr>
" noremap <localleader>g :make <cr>
noremap <localleader>h :make hook<cr>
noremap <localleader>i :PlugInstall<cr>
"noremap <localleader>j :call VimuxSlime(@v)<cr>
noremap <localleader>k :make doc<cr>
noremap <localleader>l :make lint<cr>
noremap <localleader>m :make commit<cr>
noremap <localleader>n :PlugClean<cr>
noremap <localleader>o :make publish<cr>
noremap <localleader>p :make package<cr>
" noremap <localleader>q :make <cr>
noremap <localleader>r :make cover<cr>
noremap <localleader>s :make sort<cr>
noremap <localleader>t :make test<cr>
noremap <localleader>u :PlugUpdate<cr>
noremap <localleader>v :make venv<cr>
noremap <localleader>w :make watch<cr>
" noremap <localleader>x :make <cr>
noremap <localleader>y :make type<cr>
" noremap <localleader>z :make <cr>
" AUTOCMD
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
" FUNCTION
function! VimuxSlime(text)
    call VimuxSendText(a:text)
    call VimuxSendKeys("Enter")
endfunction
