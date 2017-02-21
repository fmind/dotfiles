#!/bin/bash

echo "[*] Installing git, ansible and python3."
sudo apt-get install -y git ansible python3

echo
read -p "Do you want to clone my-dev-tools to dev-tools now [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[*] Cloning my-dev-tools."
    git clone https://fmind@github.com/fmind/my-dev-tools dev-tools
fi

echo
read -p "Do you want to install pip3 (user) [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    wget -O get-pip.py https://bootstrap.pypa.io/get-pip.py
    python3 get-pip.py --user
    rm get-pip.py
fi

echo "[*] Done."
