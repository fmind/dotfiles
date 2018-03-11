FROM fedora

MAINTAINER fmind <fmind@users.noreply.github.com>

RUN dnf -y install git ansible libselinux-python

RUN dnf -y update

WORKDIR /root

RUN git clone https://github.com/fmind/dotfiles .dotfiles

RUN cd .dotfiles && make console

RUN usermod -s /usr/bin/zsh root

RUN dnf clean all

ENTRYPOINT /usr/bin/zsh
