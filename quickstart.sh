#!/bin/bash

echo "[*] Installing git and ansible ..."
sudo apt-get install -y git ansible

echo
read -p "Do you want to clone my-tools-settings now [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[*] Cloning my-tools-settings ..."
    git clone https://freaxmind@github.com/freaxmind/my-tools-settings
fi

echo
read -p "Do you want set /etc/ansible/hosts to localhost only [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[*] Replacing Ansible Inventory to localhost only ..."
    echo "localhost ansible_connection=local" | sudo tee /etc/ansible/hosts
fi

echo
read -p "Do you want to install pip3 [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    wget -O get-pip.py https://bootstrap.pypa.io/get-pip.py
    python3 get-pip.py
    rm get-pip.py
fi


echo "[*] Done."
