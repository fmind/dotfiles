" INIT
set shortmess=I
" FILE
set path+=**
" SHELL
set shell=/bin/bash
set clipboard=unnamedplus
" MENUS
set wildmode=list:longest,full
set completeopt=menuone,preview
" INDENT
set expandtab
" BUFFER
set hidden
set confirm
set linebreak
" NUMBER
set number
set relativenumber
" WINDOW
set scrolloff=10
set statusline=\ %n:\ \%f\ %y%=%r\ %l\ :\ %c\ (%p%%)\
" COLORS
colorscheme zellner
" PLUGINS
let g:loaded_netrw = 1
let g:loaded_matchparen=0
let g:loaded_netrwPlugin = 1
" LEADERS
let mapleader = "\<SPACE>"
" KEYMAPS
noremap j gj
noremap k gk
noremap <CR> :
nnoremap gl :nohl<CR>
nnoremap g. :edit $MYVIMRC<CR>
" AUTOCMD
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
