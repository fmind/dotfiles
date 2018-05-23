SHELL := /bin/bash

CONN=local
SYS=sys.yml
USER=user.yml
HOSTS='localhost,'
ANS=ansible-playbook -c$(CONN) -i$(HOSTS)

docker:
	docker build --no-cache -t fmind/shell .
	docker push fmind/shell

init:
	sudo dnf -y install git python3-ansible

.PHONY: editors
editors: editors-sys editors-user;

editors-sys: editors/fonts/$(SYS) editors/neovim/$(SYS) editors/vim/$(SYS)
	$(ANS) -K $^

editors-user: editors/neovim/$(USER) editors/vim/$(USER)
	$(ANS) $^

.PHONY: shells
shells: shells-sys shells-user;

shells-sys: shells/byobu/$(SYS) shells/zsh/$(SYS)
	$(ANS) -K $^

shells-user: shells/aliases/$(USER) shells/bash/$(USER) shells/byobu/$(USER) shells/config/$(USER) shells/environ/$(USER) shells/input/$(USER) shells/zsh/$(USER)
	$(ANS) $^

.PHONY: tools
tools: tools-sys tools-user;

tools-sys: tools/ag/$(SYS) tools/curl/$(SYS) tools/fasd/$(SYS) tools/git/$(SYS) tools/htop/$(SYS) tools/imagemagick/$(SYS) tools/jq/$(SYS) tools/ncdu/$(SYS) tools/parallel/$(SYS)  tools/pigz/$(SYS) tools/pv/$(SYS) tools/ranger/$(SYS) tools/rlwrap/$(SYS)
	$(ANS) -K $^

tools-user: tools/ag/$(USER) tools/ansible/$(USER) tools/cookiecutter/$(USER) tools/functools/$(USER) tools/git/$(USER) tools/httpie/$(USER) tools/percol/$(USER) tools/pydf/$(USER) tools/pyped/$(USER) tools/tldr/$(USER) tools/watchdog/$(USER)
	$(ANS) $^
