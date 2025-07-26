# https://docs.docker.com/engine/reference/builder/

FROM python:3.13
ARG USER=fmind
RUN apt update \
    && apt upgrade -y \
    && apt install -y sudo
RUN useradd -m ${USER} \
    && echo "${USER} ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/${USER}
USER ${USER}
WORKDIR /home/${USER}/dotfiles
ENV PATH="/home/${USER}/.local/bin:$PATH"
RUN python3 -m pip install pipx \
    && pipx install ansible --include-deps \
    && pipx inject ansible pipx
COPY --chown=${USER}:${USER} . /home/${USER}/dotfiles
RUN ansible-playbook site.yml
CMD ["zsh"]
