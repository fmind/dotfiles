" INIT
set shortmess=I
" FILE
set hidden
set path+=**
set linebreak
" SHELL
set shell=/bin/bash
set clipboard=unnamedplus
" MENUS
set wildmode=list:longest,full
set completeopt=menuone,preview
" INDENT
set expandtab
" NUMBER
set number
set relativenumber
" WINDOW
set scrolloff=10
set statusline=\ %n:\ \%f\ %y%=%r\ %l\ :\ %c\ (%p%%)\
" COLORS
colorscheme zellner
" LEADER
let mapleader = "\<CR>"
" KEYMAPS
noremap <space> :
nnoremap gl :nohl<CR>
nnoremap g. :edit $MYVIMRC<CR>
" AUTOCMD
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
