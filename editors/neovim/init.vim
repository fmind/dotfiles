" vim: fdm=marker
" STARTING {{{
let $VIMDIR=$HOME.'/.config/nvim'
source ~/.vimrc
"}}}
" PLUGING {{{
syntax enable
filetype plugin on
filetype indent on
if !empty(glob($VIMDIR.'/autoload/plug.vim'))
call plug#begin($VIMDIR.'/plugged')
"" ale {{{
Plug 'w0rp/ale'
nnoremap <leader>E :ALEToggle<CR>
nnoremap <leader>e :ALENextWrap<CR>
nnoremap <leader><C-e> :ALEPreviousWrap<CR>
" }}}
"" fzf {{{
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
nnoremap <leader>/ :Ag<CR>
nnoremap <leader>A :Tags<CR>
nnoremap <leader>a :BTags<CR>
nnoremap <leader>m :Marks<CR>
nnoremap <leader>L :Lines<CR>
nnoremap <leader>C :Colors<CR>
nnoremap <leader>l :BLines<CR>
nnoremap <leader>F :GFiles<CR>
nnoremap <leader>f :Files .<CR>
nnoremap <leader>b :Buffers<CR>
nnoremap <leader>B :Filetypes<CR>
nnoremap <leader>y :BCommits<CR>
nnoremap <leader>Y :Commits<CR>
nnoremap <leader>w :Windows<CR>
nnoremap <leader>i :Snippets<CR>
nnoremap <leader>h :History<CR>
nnoremap <leader>: :History:<CR>
nnoremap <leader>; :History/<CR>
nnoremap <leader>? :Helptags<CR>
nnoremap <leader><Space> :Commands<CR>
let $FZF_DEFAULT_COMMAND = 'ag --hidden -p ~/.agignore -g ""'
" }}}
"" plug {{{
nnoremap <leader>pd :PlugDiff<CR>
nnoremap <leader>pc :PlugClean<CR>
nnoremap <leader>ps :PlugStatus<CR>
nnoremap <leader>pu :PlugUpdate<CR>
nnoremap <leader>pg :PlugUpgrade<CR>
nnoremap <leader>pi :PlugInstall<CR>
" }}}
"" test {{{
Plug 'janko-m/vim-test'
let test#strategy = "vimux"
let test#python#runner = 'pytest'
let g:test#preserve_screen = 1
nnoremap <leader>tf :TestFile<CR>
nnoremap <leader>tl :TestLast<CR>
nnoremap <leader>ts :TestSuite<CR>
nnoremap <leader>tv :TestVisit<CR>
nnoremap <leader>tt :TesgNearest<CR>
" }}}
"" codi {{{
Plug 'metakirby5/codi.vim'
nnoremap <leader>c :Codi!!<CR>
" }}}
"" sneak {{{
Plug 'justinmk/vim-sneak'
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
" }}}
"" vimux {{{
Plug 'benmills/vimux'
nnoremap <Leader>ro :VimuxOpenRunner<CR>
nnoremap <Leader>rc :VimuxCloseRunner<CR>
nnoremap <Leader>rr :VimuxPromptCommand<CR>
nnoremap <Leader>ri :VimuxInspectRunner<CR>
nnoremap <Leader>rl :VimuxRunLastCommand<CR>
nnoremap <Leader>rx :VimuxInterruptRunner<CR>
nnoremap <Leader>rz :call VimuxZoomRunner()<CR>
" }}}
"" tagbar {{{
Plug 'majutsushi/tagbar'
let g:tagbar_autofocus = 1
nnoremap <leader>j :TagbarToggle<CR>
" }}}
"" pencil {{{
let g:pencil#textwidth = 80
nnoremap <leader>W :HardPencil<CR>
" }}}
"" vimtex {{{
Plug 'lervag/vimtex', { 'for': 'tex' }
" }}}
"" tabular {{{
Plug 'godlygeek/tabular'
noremap <leader>X :Tabularize 
noremap <leader>x, :Tabularize /,<CR>
noremap <leader>x, :Tabularize /;<CR>
" }}}
"" molokai {{{
Plug 'tomasr/molokai'
colorscheme molokai
let g:molokai_original = 1
" }}}
"" airline {{{
Plug 'bling/vim-airline' 
let g:airline_powerline_fonts=1
let g:airline#extensions#ale#enabled = 1
let g:airline#extensions#branch#enabled = 1
let g:airline#extensions#tagbar#enabled = 1
let g:airline#extensions#tabline#enabled = 1
let g:airline#extensions#wordcount#enabled = 1
let g:airline#extensions#virtualenv#enabled = 1
" }}}
"" fugitive {{{
Plug 'tpope/vim-fugitive'
noremap <Leader>G Git 
noremap <Leader>gj :Glcd 
noremap <Leader>gh :Gpush<CR>
noremap <Leader>gl :Gpull<CR>
noremap <Leader>gm :Gmove
noremap <Leader>gw :Gwrite<CR>
noremap <Leader>gc :Gcommit<CR>
noremap <Leader>gs :Gstatus<CR>
noremap <Leader>go :Gbrowse<CR>
noremap <Leader>gb :Gblame<CR>
noremap <Leader>gd :Gvdiff<CR>
noremap <Leader>gr :Gremove<CR>
" }}}
"" deoplete {{{
Plug 'Shougo/deoplete.nvim', { 'do': ':UpdateRemotePlugins', 'for': 'python' }
Plug 'zchee/deoplete-jedi'
Plug 'wellle/tmux-complete.vim'
let g:deoplete#enable_at_startup = 1
" }}}
"" startify {{{
Plug 'mhinz/vim-startify'
let g:startify_session_dir = $VIMDIR.'/session/'
nnoremap <leader>S :Startify<CR>
nnoremap <leader>sl :SLoad<CR>
nnoremap <leader>ss :SSave<CR>
nnoremap <leader>sc :SClose<CR>
nnoremap <leader>sd :SDelete<CR>
" }}}
"" nerdtree {{{
Plug 'scrooloose/nerdtree' 
Plug 'Xuyuanp/nerdtree-git-plugin'
Plug 'tiagofumo/vim-nerdtree-syntax-highlight'
let g:NERDTreeQuitOnOpen = 1
let NERDTreeIgnore=['\.pyc$', '\~$']
nnoremap <leader>~ :NERDTreeFind<CR>
nnoremap <leader>` :NERDTreeToggle<CR>
" }}}
"" gitgutter {{{
Plug 'airblade/vim-gitgutter'
let g:gitgutter_grep = 'ag'
let g:gitgutter_map_keys = 0
nnoremap ]g <Plug>GitGutterNextHunk
nnoremap [g <Plug>GitGutterPrevHunk
" }}}
"" ultisnips {{{
Plug 'SirVer/ultisnips'
nnoremap <leader>I :UltiSnipsEdit<CR>
let g:UltiSnipsEditSplit = 'context'
let g:UltiSnipsExpandTrigger="<tab>"
let g:UltiSnipsListSnippets="<s-tab>"
let g:UltiSnipsJumpForwardTrigger="<c-n>"
let g:UltiSnipsJumpBackwardTrigger="<c-p>"
let g:UltiSnipsSnippetsDir = $VIMDIR.'/snippets/'
" }}}
"" thesaurus {{{
Plug 'beloglazov/vim-online-thesaurus'
let g:online_thesaurus_map_keys = 0
nnoremap <leader>U :Thesaurus 
nnoremap <leader>u :OnlineThesaurusCurrentWord<CR>
" }}}
"" autoformat {{{
Plug 'Chiel92/vim-autoformat'
let g:formatters_python = ['yapf']
nnoremap <leader>= :Autoformat<CR>
" }}}
"" easy-align {{{
Plug 'junegunn/vim-easy-align'
xmap ga <Plug>(EasyAlign)
nmap ga <Plug>(EasyAlign)
" }}}
"" projectionist {{{
Plug 'tpope/vim-projectionist'
nnoremap <leader>k :A<CR>
" }}}
"" tmux-navigator {{{
Plug 'christoomey/vim-tmux-navigator'
let g:tmux_navigator_save_on_switch = 1
" }}}
Plug 'mattn/emmet-vim'
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-eunuch'
Plug 'reedes/vim-pencil'
Plug 'tpope/vim-abolish'
Plug 'lepture/vim-jinja'
Plug 'fisadev/vim-isort'
Plug 'wellle/targets.vim'
Plug 'tpope/vim-surround'
Plug 'sheerun/vim-polyglot'
let g:polyglot_disabled = ['latex']
Plug 'tpope/vim-commentary'
Plug 'Raimondi/delimitMate'
Plug 'tpope/vim-speeddating'
Plug 'vim-scripts/matchit.zip'
Plug 'farmergreg/vim-lastplace'
Plug 'christoomey/vim-sort-motion'
Plug 'michaeljsmith/vim-indent-object'
Plug 'coderifous/textobj-word-column.vim'
Plug 'aklt/plantuml-syntax', {'for': 'plantuml'}
Plug 'scrooloose/vim-slumlord', {'for': 'plantuml'}
call plug#end()
endif
""}}}
" LANGS {{{
"" python {{{
"autocmd BufWritePost *.py :Isort
autocmd BufWritePost *.py :Autoformat
autocmd BufNewFile,BufRead *.py setlocal foldmethod=indent
" }}}
" }}}
