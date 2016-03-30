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

echo "[*] Done."
