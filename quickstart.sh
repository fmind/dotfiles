#!/bin/bash

ANSIBLE_PPA="ansible/ansible"


echo "[*] Installing prerequisites ..."
sudo apt-get install -y software-properties-common

if ! grep -q "$ANSIBLE_PPA" /etc/apt/sources.list /etc/apt/sources.list.d/*; then
    echo "[*] Adding Ansible ppa ..."
    sudo apt-add-repository "ppa:$ANSIBLE_PPA"
    sudo apt-get update -y
fi

echo "[*] Installing git and ansible ..."
sudo apt-get install -y ansible git

echo "[*] Cloning my-tools-settings ..."
git clone https://freaxmind@github.com/freaxmind/my-tools-settings

echo
read -p "Do you want set /etc/ansible/hosts to localhost only [Y/y] ? " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[*] Replacing Ansible Inventory to localhost only ..."
    echo "localhost ansible_connection=local" | sudo tee /etc/ansible/hosts
fi

echo "[*] Done."
