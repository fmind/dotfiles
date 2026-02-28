FROM python:3.13-slim-bookworm

ARG USER=fmind

ARG DEBIAN_FRONTEND=noninteractive
ENV PATH="/home/${USER}/.local/bin:$PATH"

RUN apt-get update && apt-get install -y --no-install-recommends \
    sudo git curl zsh unzip \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /usr/bin/zsh ${USER} \
    && echo "${USER} ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/${USER} \
    && chmod 0440 /etc/sudoers.d/${USER}

USER ${USER}
WORKDIR /home/${USER}/dotfiles

RUN python3 -m pip install --no-cache-dir --user pipx \
    && pipx install ansible --include-deps \
    && pipx inject ansible pipx

COPY --chown=${USER}:${USER} . .

RUN ansible-playbook -i inventory.ini site.yml

CMD ["zsh"]
