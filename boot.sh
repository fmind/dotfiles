#!/bin/bash

echo "[*] Installing git and ansible ..."
sudo apt-get install -y git ansible

echo
read -p "Do you want to clone my-dev-tools now [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[*] Cloning my-dev-tools ..."
    git clone https://fmind@github.com/fmind/my-dev-tools
fi

echo
read -p "Do you want to install pip3 [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    wget -O get-pip.py https://bootstrap.pypa.io/get-pip.py
    sudo python3 get-pip.py
    rm get-pip.py
fi


echo "[*] Done."
