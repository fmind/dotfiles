FROM ubuntu
MAINTAINER fmind

RUN apt-add-repository ppa:ansible/ansible
RUN apt-get install git ansible

RUN git clone https://github.com/fmind/dotfiles .dotfiles
RUN cd .dotfiles && ansible-playbook tui-sys.yml tui-user.yml
RUN cd .dotfiles && ansible-playbook prog-sys.yml prog-user.yml
RUN cd .dotfiles && ansible-playbook data-sys.yml data-user.yml

RUN apt-get clean

CMD /bin/zsh
