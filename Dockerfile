FROM debian:bookworm-slim
ENV LANG=C.UTF-8
ENV HOME=/home/fmind
ENV PATH="${HOME}/.local/bin:${HOME}/.local/share/mise/bin:${HOME}/.local/share/mise/shims:${PATH}"
RUN apt-get update -qq \
    && apt-get install -yq --no-install-recommends \
    git curl sudo libatomic1 build-essential ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -m -G sudo -s /bin/bash fmind \
    && echo "fmind ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/fmind \
    && chmod 0440 /etc/sudoers.d/fmind
USER fmind
WORKDIR /home/fmind
COPY --chown=fmind:fmind . .local/share/chezmoi/
RUN .local/share/chezmoi/install.sh
CMD ["fish"]
