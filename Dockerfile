# syntax=docker/dockerfile:1.7
FROM debian:bookworm-slim

ENV HOME=/home/fmind \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    LANGUAGE=C.UTF-8 \
    TERM=xterm-256color \
    DEBIAN_FRONTEND=noninteractive

ENV PATH="$HOME/.local/bin:$HOME/.local/share/mise/bin:$HOME/.local/share/mise/shims:$PATH"

RUN apt-get update -qq && \
    apt-get install -yq --no-install-recommends apt-utils sudo git curl ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -m -s /bin/bash fmind && \
    echo "fmind ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/fmind && \
    chmod 0440 /etc/sudoers.d/fmind

USER fmind
WORKDIR /home/fmind

COPY --chown=fmind:fmind . dotfiles

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN=${GITHUB_TOKEN}

RUN dotfiles/install.sh

CMD ["/usr/bin/fish"]
