FROM ubuntu
MAINTAINER fmind <fmind@users.noreply.github.com>

RUN apt-get -y update && apt-get -y upgrade
RUN apt-get -y install git software-properties-common

RUN add-apt-repository -y ppa:ansible/ansible
RUN apt-get -y update && apt-get install -y ansible

WORKDIR /root

RUN git clone https://github.com/fmind/dotfiles .dotfiles
RUN cd .dotfiles && ansible-playbook tui-sys.yml tui-user.yml
RUN cd .dotfiles && ansible-playbook prog-sys.yml prog-user.yml
RUN cd .dotfiles && ansible-playbook data-sys.yml data-user.yml

RUN apt-get clean

CMD /bin/zsh
