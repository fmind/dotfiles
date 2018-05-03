FROM fedora

WORKDIR /root

MAINTAINER fmind <fmind@users.noreply.github.com>

RUN git clone https://github.com/fmind/dotfiles .dotfiles

RUN cd .dotfiles && make init console develop

RUN usermod -s /usr/bin/zsh root

RUN dnf -y update

RUN dnf clean all

ENTRYPOINT /usr/bin/zsh
