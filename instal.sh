#!/usr/bin/env bash

ansible-playbook site.yml --ask-become
chsh -s /usr/bin/fish
