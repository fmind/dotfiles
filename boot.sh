#!/bin/bash

echo "[*] Installing git, ansible, python3"
sudo apt-get install -y git ansible python3

echo
read -p "Do you want to clone fmind/dotfiles [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[*] Cloning fmind/dotfiles"
    git clone https://fmind@github.com/fmind/dotfiles
fi

echo
read -p "Do you want to install pip3 (for current user) [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    wget -O get-pip.py https://bootstrap.pypa.io/get-pip.py
    python3 get-pip.py --user
    rm get-pip.py
fi

echo "[*] Done."
