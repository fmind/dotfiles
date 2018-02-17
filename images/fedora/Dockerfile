FROM fedora
MAINTAINER fmind <fmind@users.noreply.github.com>

RUN dnf -y install git ansible libselinux-python
RUN dnf -y update

WORKDIR /root

RUN git clone https://github.com/fmind/dotfiles .dotfiles
RUN cd .dotfiles && ansible-playbook tui-sys.yml tui-user.yml
RUN cd .dotfiles && ansible-playbook prog-sys.yml prog-user.yml
RUN cd .dotfiles && ansible-playbook data-sys.yml data-user.yml

RUN dnf clean all

CMD /bin/zsh
