SHELL := /bin/bash

CONN=local
SYS=sys.yml
USER=user.yml
HOSTS='localhost,'
ANS=ansible-playbook -c$(CONN) -i$(HOSTS)

docker:
	docker build --no-cache -t fmind/shell .
	docker push fmind/shell

%/sys.yml:
	$(ANS) -k $@

%/user.yml:
	$(ANS) $@

# GROUPS

all: sys user;

sys: console-sys develop-sys graphic-sys;

user: console-user develop-user graphic-user;

console: console-sys console-user;

console-sys: editors-sys shells-sys tools-sys;

console-user: editors-user shells-user tools-user;

develop: develop-sys develop-user

develop-sys: languages-sys;

develop-user: languages-user;

graphic: graphic-sys graphic-user;

graphic-sys: applications-sys distributions-sys;

graphic-user: applications-user distribution-user;

# PROFILES

.PHONY: applications
applications: applications-user applications-sys;

applications-sys: applications/anki/$(SYS) applications/chrome/$(SYS) applications/deja-dup/$(SYS) applications/insync/$(SYS) applications/keepassx/$(SYS) applications/tlp/$(SYS) applications/xbacklight/$(SYS) applications/xsel/$(SYS)
	$(ANS) -K $^

applications-user: applications/zotero/$(USER)
	$(ANS) $^

.PHONY: distributions
distributions: distribution-sys distribution-user;

distributions-sys: distributions/fedora/$(SYS) distributions/gnome/$(SYS)
	$(ANS) -K $^

distribution-user: ;

.PHONY: editors
editors: editors-sys editors-user;

editors-sys: editors/emacs/$(SYS) editors/fonts/$(SYS) editors/neovim/$(SYS) editors/vim/$(SYS)
	$(ANS) -K $^

editors-user: editors/emacs/$(USER) editors/neovim/$(USER) editors/vim/$(USER)
	$(ANS) $^

.PHONY: languages
languages: languages-sys languages-user;

languages-sys: languages/clang/$(SYS) languages/graphviz/$(SYS) languages/groovy/$(SYS) languages/java/$(SYS) languages/nodejs/$(SYS) languages/python/$(SYS);
	$(ANS) -K $^

languages-user: languages/clojure/$(USER) languages/nodejs/$(USER) languages/plantuml/$(USER) languages/python/$(USER)
	$(ANS) $^

.PHONY: sciences
sciences: science-user science-sys;

sciences-sys: sciences/latexmk/$(SYS) sciences/pandoc/$(SYS) sciences/tex/$(SYS)
	$(ANS) -K $^

sciences-user: jupyter/jupyter/$(USER) sciences/latexmk/$(USER)
	$(ANS) $^

.PHONY: shells
shells: shells-sys shells-user;

shells-sys: shells/byobu/$(SYS) shells/check/$(SYS) shells/zsh/$(SYS)
	$(ANS) -K $^

shells-user: shells/aliases/$(USER) shells/bash/$(USER) shells/byobu/$(USER) shells/config/$(USER) shells/environ/$(USER) shells/input/$(USER) shells/zsh/$(USER)
	$(ANS) $^

.PHONY: tools
tools: tools-sys tools-user;

tools-sys: tools/ag/$(SYS) tools/ansible/$(SYS) tools/curl/$(SYS) tools/fasd/$(SYS) tools/git/$(SYS) tools/htop/$(SYS) tools/imagemagick/$(SYS) tools/jq/$(SYS) tools/ncdu/$(SYS) tools/parallel/$(SYS) tools/percol/$(SYS) tools/pigz/$(SYS) tools/pv/$(SYS) tools/pyped/$(SYS) tools/ranger/$(SYS) tools/rlwrap/$(SYS)
	$(ANS) -K $^

tools-user: tools/ag/$(USER) tools/ansible/$(USER) tools/cookiecutter/$(USER) tools/functools/$(USER) tools/git/$(USER) tools/httpie/$(USER) tools/percol/$(USER) tools/tldr/$(USER) tools/watchdog/$(USER)
	$(ANS) $^
