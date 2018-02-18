FROM fedora
MAINTAINER fmind <fmind@users.noreply.github.com>

RUN dnf -y install git ansible libselinux-python
RUN dnf -y update

WORKDIR /root

RUN git clone https://github.com/fmind/dotfiles .dotfiles
RUN cd .dotfiles && ansible-playbook tui-sys.yml
RUN cd .dotfiles && ansible-playbook prog-sys.yml
RUN cd .dotfiles && ansible-playbook data-sys.yml
RUN cd .dotfiles && ansible-playbook tui-user.yml
RUN cd .dotfiles && ansible-playbook prog-user.yml
RUN cd .dotfiles && ansible-playbook data-user.yml

RUN dnf clean all # cleanup old package

RUN usermod -s /usr/bin/zsh root

ENTRYPOINT /usr/bin/zsh
