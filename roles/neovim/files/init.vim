" PATH
set path+=**
" INIT
set shortmess=I
" CLIP
set clipboard=unnamedplus
" MENUS
set wildmode=list:longest,full
set completeopt=menuone,preview
" SPELLS
set spelllang=en,fr
set thesaurus+=/usr/share/dict/theses
set dictionary+=/usr/share/dict/words
" BUFFER
set hidden
set confirm
set autoread
set autowrite
set linebreak
" SEARCH
set hlsearch
set incsearch
" INDENT
set expandtab
set shiftround
set shiftwidth=4
" NUMBER
set number
set relativenumber
" FOLDER
set foldmethod=syntax
set foldlevelstart=99
" WINDOW
set scrolloff=10
set statusline=\ %n:\ \%f\ %y%=%r\ %l\ :\ %c\ (%p%%)
" COLORS
try
    colorscheme molokai
catch
    colorscheme zellner
endtry
" PLUGINS
let g:loaded_netrw = 1
let g:loaded_matchparen=0
let g:loaded_netrwPlugin = 1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'w0rp/ale'
let g:ale_set_quickfix = 1
let g:ale_sign_column_always = 1
call plug#end()
" KEYMAPS
noremap <cr> :
noremap gl :nohl<cr>
" move
cnoremap <C-k> <Up>
inoremap <C-k> <Up>
cnoremap <C-h> <Left>
inoremap <C-h> <Left>
cnoremap <C-j> <Down>
inoremap <C-J> <Down>
cnoremap <C-l> <Right>
inoremap <C-l> <Right>
cnoremap <C-a> <Home>
inoremap <C-a> <Home>
cnoremap <C-e> <End>
inoremap <C-e> <End>
" LEADERS
let mapleader=" "
"noremap <leader>a :nohl<cr>
"noremap <leader>b :nohl<cr>
"noremap <leader>c :nohl<cr>
"noremap <leader>d :nohl<cr>
"noremap <leader>e :nohl<cr>
"noremap <leader>f :nohl<cr>
"noremap <leader>g :nohl<cr>
noremap <leader>h :bprevious<cr>
"noremap <leader>i :nohl<cr>
noremap <leader>j :tnext<cr>
noremap <leader>k :tprevious<cr>
noremap <leader>l :bnext<cr>
"noremap <leader>m :nohl<cr>
"noremap <leader>n :nohl<cr>
"noremap <leader>o :nohl<cr>
"noremap <leader>p :nohl<cr>
noremap <leader>q :bdelete<cr>
"noremap <leader>r :nohl<cr>
"noremap <leader>s :nohl<cr>
"noremap <leader>t :nohl<cr>
"noremap <leader>u :nohl<cr>
"noremap <leader>v :nohl<cr>
"noremap <leader>w :nohl<cr>
"noremap <leader>x :nohl<cr>
"noremap <leader>y :nohl<cr>
"noremap <leader>z :nohl<cr>
"noremap <leader>A :nohl<cr>
"noremap <leader>B :nohl<cr>
"noremap <leader>C :nohl<cr>
"noremap <leader>D :nohl<cr>
"noremap <leader>E :nohl<cr>
"noremap <leader>F :nohl<cr>
"noremap <leader>G :nohl<cr>
"noremap <leader>H :nohl<cr>
"noremap <leader>I :nohl<cr>
"noremap <leader>J :nohl<cr>
"noremap <leader>K :nohl<cr>
"noremap <leader>L :nohl<cr>
"noremap <leader>M :nohl<cr>
"noremap <leader>N :nohl<cr>
"noremap <leader>O :nohl<cr>
"noremap <leader>P :nohl<cr>
"noremap <leader>Q :nohl<cr>
"noremap <leader>R :nohl<cr>
"noremap <leader>S :nohl<cr>
"noremap <leader>T :nohl<cr>
"noremap <leader>U :nohl<cr>
"noremap <leader>V :nohl<cr>
"noremap <leader>W :nohl<cr>
"noremap <leader>X :nohl<cr>
"noremap <leader>Y :nohl<cr>
"noremap <leader>Z :nohl<cr>
"noremap <leader>` :edit $MYVIMRC<cr>
"noremap <leader>- :edit $MYVIMRC<cr>
"noremap <leader>= :edit $MYVIMRC<cr>
noremap <leader>[ :cprev<cr>
noremap <leader>] :cnext<cr>
"noremap <leader>' :edit $MYVIMRC<cr>
noremap <leader>, :set spell<cr>
noremap <leader>. :edit $MYVIMRC<cr>
"noremap <leader>/ :edit $MYVIMRC<cr>
noremap <leader><cr> :make<cr>
noremap <leader><tab> :b#<cr>
noremap <leader><space> :make
let maplocalleader=";"
" AUTOCMD
augroup tex
    autocmd!
    autocmd FileType tex setlocal spell
augroup end
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
