" BUFFER {{{
set hidden
set confirm
set autoread
set autowrite
" }}}
" FOLDER {{{
set foldmethod=marker
set foldlevelstart=99
" }}}
" INDENT {{{
set tabstop=4
set expandtab
set shiftround
set shiftwidth=4
set softtabstop=4
" }}}
" NUMBER {{{
set number
set relativenumber
" }}}
" OTHERS {{{
set shell=/bin/bash
set clipboard=unnamedplus
" }}}
" POPUPS {{{
set wildmode=list:longest,full
set completeopt=menuone,longest
" }}}
" SEARCH {{{
set gdefault
set hlsearch
set incsearch
set smartcase
set ignorecase
" }}}
" SPELLS {{{
set spell
set spelllang=en
" }}}
" WINDOW {{{
set linebreak
set lazyredraw
set shortmess=I
set scrolloff=10
" }}}
" PLUGIN {{{
let g:loaded_netrw = 1
let g:loaded_matchparen=1
let g:loaded_netrwPlugin = 1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'benmills/vimux'
Plug 'davidhalter/jedi-vim'
let g:jedi#completions_enabled = 0
let g:jedi#auto_vim_configuration = 0
Plug 'deoplete-plugins/deoplete-jedi'
Plug 'francoiscabrol/ranger.vim'
let g:ranger_map_keys = 0
Plug 'godlygeek/tabular'
Plug 'honza/vim-snippets'
Plug 'itchyny/lightline.vim'
Plug 'janko/vim-test'
let test#strategy = "vimux"
let test#python#runner = "pytest"
let g:test#preserve_screen = 1
Plug 'jiangmiao/auto-pairs'
let g:AutoPairsShortcutToggle='<M-a>'
Plug 'julienr/vim-cellmode'
let g:cellmode_tmux_panenumber='1'
let g:cellmode_default_mappings='0'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
Plug 'justinmk/vim-sneak'
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
Plug 'michaeljsmith/vim-indent-object'
Plug 'rbgrouleff/bclose.vim' " ranger dependency
Plug 'sheerun/vim-polyglot'
Plug 'Shougo/deoplete.nvim', {'do': ':UpdateRemotePlugins'}
let g:deoplete#enable_at_startup = 1
Plug 'SirVer/ultisnips'
Plug 'szw/vim-g'
Plug 'tomasr/molokai'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-eunuch'
Plug 'tpope/vim-fugitive'
Plug 'tpope/vim-projectionist'
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-rsi'
Plug 'tpope/vim-surround'
Plug 'tpope/vim-unimpaired'
Plug 'w0rp/ale'
let g:ale_set_quickfix = 1
let b:ale_fixers = {'python': ['black', 'isort']}
let b:ale_linters = {'python': ['mypy', 'pylint']}
let g:ale_python_pylint_options = '--error-only'
Plug 'wellle/tmux-complete.vim'
call plug#end()
" }}}
" COLORS {{{
try
    colorscheme molokai
catch
    colorscheme zellner
endtry
" }}}
" REMAPS {{{
nnoremap j gj
nnoremap k gk
nnoremap B g^
nnoremap E g$
nnoremap Y y$
xnoremap < <gv
xnoremap > >gv
nnoremap U <C-r>
nnoremap gl :nohl<CR>
vnoremap <silent> <C-p> :call RunTmuxPythonChunk()<CR>
nnoremap <silent> <C-p> :call RunTmuxPythonCell(1)<CR>
nnoremap <silent> <M-p> :call RunTmuxPythonCell(0)<CR>
nnoremap <silent> <C-M-p> :call RunTmuxPythonAllCellsAbove()<CR>
" }}}
" POPUPS {{{
inoremap <expr> <C-j> pumvisible() ? "\<C-n>" : "\<C-j>"
inoremap <expr> <C-k> pumvisible() ? "\<C-p>" : "\<C-k>"
" }}}
" LEADERS {{{
noremap <CR> :
let mapleader="\<space>"
noremap <leader>a :A<CR>
noremap <leader>b :Buffers<CR>
noremap <leader>c :Colors<CR>
noremap <leader>d :GFiles<CR>
noremap <leader>e :cnext<CR>
noremap <leader>f :Files<CR>
noremap <leader>g :Google 
noremap <leader>h :Helptags<CR>
noremap <leader>i :Lines<CR>
noremap <leader>j :bnext<CR>
noremap <leader>k :bprevious<CR>
noremap <leader>l :BLines<CR>
noremap <leader>m :Marks<CR>
noremap <leader>n :BCommits<CR>
noremap <leader>o :call VimuxOpenRunner()<CR>
noremap <leader>p :Commands<CR>
noremap <leader>q :bdelete<CR>:bnext<CR>
noremap <leader>r :Tags<CR>
noremap <leader>s :Ag<CR>
noremap <leader>t :BTags<CR>
noremap <leader>u :UltiSnipsEdit<CR>
noremap <leader>v "vy :call VimuxSlime(@v)<CR>
noremap <leader>w :Windows<CR>
noremap <leader>x :VimuxRunLastCommand<CR>
noremap <leader>y :VimuxInspectRunner<CR>
noremap <leader>z :Filetypes<CR>
noremap <leader><CR> :make<CR>
noremap <leader><tab> :b#<CR>
noremap <leader><space> :make
noremap <leader>; :call VimuxSlime(join(getline(1, '$'), "\n"))<CR>
noremap <leader>] :ALENextWrap<CR>
noremap <leader>[ :ALEPreviousWrap<CR>
noremap <leader>' :VimuxPromptCommand<CR>
noremap <leader>` :Locate 
noremap <leader>- :Ranger<CR>
noremap <leader>= :Tabularize 
noremap <leader>! :GitGutterToggle<CR>
noremap <leader>, :Gw<CR>
noremap <leader>? :Maps<CR>
noremap <leader>\ :History<CR>
noremap <leader>: :History:<CR>
noremap <leader>/ :History/<CR>
noremap <leader>. :edit $MYVIMRC<CR>
" }}}
" LLEADERS {{{
let localmapleader="\\"
" jedi {{{
let g:jedi#goto_command = "<localleader>jj"
let g:jedi#rename_command = "<localleader>jr"
let g:jedi#usages_command = "<localleader>ju"
let g:jedi#documentation_command = "<localleader>jd"
" }}}
" files {{{
noremap <localleader>ea :e ~/.agignore<CR>
noremap <localleader>eb :e ~/.bashrc<CR>
noremap <localleader>ec :e ~/.cookiecutterrc<CR>
noremap <localleader>eg :e ~/.gitconfig<CR>
noremap <localleader>ei :e ~/.gitignore<CR>
noremap <localleader>ej :e ~/.jupyter_notebook_config.py<CR>
noremap <localleader>ep :e ~/.ipython/profile_default/ipython_config.py<CR>
noremap <localleader>et :e ~/.ctags<CR>
noremap <localleader>ev :e ~/.config/nvim/init.vim<CR>
noremap <localleader>ex :e ~/.xonshrc<CR>
" }}}
" tests {{{
noremap <localleader>tf :TestFile<CR>
noremap <localleader>tl :TestLast<CR>
noremap <localleader>ts :TestSuite<CR>
noremap <localleader>tt :TestNearest<CR>
noremap <localleader>tv :TestVisit<CR>
" }}}
" spells {{{
noremap <localleader>sa :set spelllang=en,fr<CR>
noremap <localleader>se :set spelllang=en<CR>
noremap <localleader>sf :set spelllang=fr<CR>
noremap <localleader>sn :set nospell<CR>
" }}}
" python {{{
noremap <localleader>pb :!black %<CR>
noremap <localleader>pd :!pydoc3 
noremap <localleader>pi :!pip3 install 
noremap <localleader>pl :!pylint %<CR>
noremap <localleader>pm :!mypy %<CR>
noremap <localleader>pn :!pip3 install pynvim<CR>
noremap <localleader>po :!invoke 
noremap <localleader>pp :!python3 %<CR>
noremap <localleader>pr :!bowler 
noremap <localleader>ps :!isort %<CR>
noremap <localleader>pt :!pytest %<CR>
noremap <localleader>pu :!pip3 uninstall 
noremap <localleader>pv :!python3 -m venv 
noremap <localleader>py :!ipython -i %<CR>
" }}}
" windows {{{
noremap <localleader>wd :set background=dark<CR>
noremap <localleader>wl :set background=light<CR>
noremap <localleader>wn :highlight Normal guibg=NONE ctermbg=NONE<CR>
" }}}
" }}}
" FUNCTIONS {{{
function! VimuxSlime(text)
    call VimuxSendText(a:text)
    if a:text !~ '\n$'
        call VimuxSendKeys("Enter")
    endif
endfunction
" }}}
" FILE-TYPES {{{
autocmd FileType yaml setlocal tabstop=2 softtabstop=2 shiftwidth=2
" }}}
" AUTO-GROUPS {{{
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
" }}}
