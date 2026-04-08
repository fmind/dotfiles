FROM debian:bookworm-slim
ENV HOME=/home/fmind \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    LANGUAGE=C.UTF-8 \
    TERM=xterm-256color \
    DEBIAN_FRONTEND=noninteractive \
    MISE_YES=1 \
    MISE_TRUSTED_CONFIG_PATHS="/home/fmind/.local/share/chezmoi:/home/fmind/.config/mise" \
    PATH="/home/fmind/.local/bin:/home/fmind/.local/share/mise/bin:/home/fmind/.local/share/mise/shims:$PATH"
RUN apt-get update -qq && \
    apt-get install -yq --no-install-recommends git curl sudo ca-certificates libatomic1 && \
    rm -rf /var/lib/apt/lists/* && \
    useradd -m -G sudo -s /bin/bash fmind && \
    echo "fmind ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/fmind && \
    chmod 0440 /etc/sudoers.d/fmind
USER fmind
WORKDIR /home/fmind
RUN mkdir -p .local/share/chezmoi .local/bin .config/mise
COPY --chown=fmind:fmind . .local/share/chezmoi
RUN cd .local/share/chezmoi && \
    git init --initial-branch=main && \
    git remote add origin https://github.com/fmind/dotfiles.git && \
    git remote set-url --push origin no_push && \
    ./install.sh && \
    mise run toolchain
CMD ["fish"]
