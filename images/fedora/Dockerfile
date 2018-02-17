FROM fedora
MAINTAINER fmind <fmind@users.noreply.github.com>

RUN dnf -y install git ansible libselinux-python
RUN dnf -y update
RUN dnf clean all

RUN git clone https://github.com/fmind/dotfiles .dotfiles
RUN cd .dotfiles && ansible-playbook --ask-become tui-sys.yml tui-user.yml
RUN cd .dotfiles && ansible-playbook --ask-become prog-sys.yml prog-user.yml
RUN cd .dotfiles && ansible-playbook --ask-become data-sys.yml data-user.yml

CMD /bin/bash
