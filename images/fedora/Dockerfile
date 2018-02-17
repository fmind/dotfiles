FROM fedora
MAINTAINER fmind <fmind@users.noreply.github.com>

RUN dnf -y install git ansible libselinux-python
#RUN dnf -y update
RUN dnf clean all

RUN ansible-playbook --ask-become 

CMD /bin/bash
