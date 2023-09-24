# https://docs.docker.com/engine/reference/builder/

FROM python:3.11
ARG USER=fmind

RUN useradd -m ${USER}

USER ${USER}
WORKDIR /home/${USER}/dotfiles
ENV PATH="/home/${USER}/.local/bin:$PATH"

RUN python3 -m pip install pipx \
    && pipx install ansible --include-deps \
    && pipx inject ansible pipx

COPY --chown=${USER}:${USER} . .

RUN ansible-playbook site.yml

CMD ["xonsh"]
