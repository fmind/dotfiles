#!/usr/bin/env python
# -*- coding: utf-8 -*-

from fabric.api import *
import shutil
import sys
import os

_fpkg = lambda x: ' '.join(x)
_curdir = os.path.dirname(os.path.realpath(__file__))
_proxy = os.getenv('http_proxy') or os.getenv('HTTP_PROXY')


def _link(src, dst):
    """ Create a link from the git conf. (src) to the program conf. (dst) """
    # returns if exists and already a link
    if os.path.exists(dst) and os.path.islink(dst):
        return

    if not os.path.exists(os.path.dirname(dst)):
        print "\t>> Please launch the application related to this file, then press Enter: {0}".format(dst)
        raw_input()

    print "\tLinking {0} with {1}".format(src, dst)

    # removes if not a link
    if os.path.exists(dst) and not os.path.islink(dst):
        if os.path.isfile(dst):
            os.remove(dst)
        else:
            shutil.rmtree(dst)

    os.symlink(src, dst)


def apt():
    print "[*] Installing new system packages (apt) ..."
    packages = ['vim', 'vim-gui-common', 'byobu', 'python-dev', 'python-pip',
                'python-zmq', 'ipython', 'python-matplotlib', 'curl', 'git', 'zsh']
    local('sudo apt-get install {packages}'.format(packages=_fpkg(packages))) 


def pip():
    print "[*] Installing new Python libraries (pip) ..."
    packages = ['flake8']
    proxy = '--proxy {0}'.format(_proxy) if _proxy else ''
    local('pip install --user --upgrade {proxy} {packages}'.format(packages=_fpkg(packages), proxy=proxy))


def bash():
    print "[*] Deploying bash ..."
    bash_aliases_src = os.path.join(_curdir, 'bash', 'bash_aliases')
    bash_aliases_dst = os.path.expanduser('~/.bash_aliases')
    bash_logout_src = os.path.join(_curdir, 'bash', 'bash_logout')
    bash_logout_dst = os.path.expanduser('~/.bash_logout')
    bashrc_src = os.path.join(_curdir, 'bash', 'bashrc')
    bashrc_dst = os.path.expanduser('~/.bashrc')
    profile_src = os.path.join(_curdir, 'bash', 'profile')
    profile_dst = os.path.expanduser('~/.profile')
    intputrc_src = os.path.join(_curdir, 'bash', 'inputrc')
    intputrc_dst = os.path.expanduser('~/.inputrc')
    face_src = os.path.join(_curdir, 'face')
    face_dst = os.path.expanduser('~/.face')

    _link(bash_aliases_src, bash_aliases_dst)
    _link(bash_logout_src, bash_logout_dst)
    _link(bashrc_src, bashrc_dst)
    _link(profile_src, profile_dst)
    _link(intputrc_src, intputrc_dst)
    _link(face_src, face_dst)


def byobu():
    print "[*] Deploying byobu ..."
    byobu_status_src = os.path.join(_curdir, 'byobu', 'status')
    byobu_status_dst = os.path.expanduser('~/.byobu/status')
    byobu_layouts_src = os.path.join(_curdir, 'byobu', 'layouts')
    byobu_layouts_dst = os.path.expanduser('~/.byobu/layouts')

    _link(byobu_status_src, byobu_status_dst)
    _link(byobu_layouts_src, byobu_layouts_dst)


def vim():
    print "[*] Deploying vim ..."
    vimdc_src = os.path.join(_curdir, 'vim')
    vimdc_dst = os.path.expanduser('~/.vim')
    vimrc_src = os.path.join(vimdc_src, 'vimrc')
    vimrc_dst = os.path.expanduser('~/.vimrc')

    _link(vimdc_src, vimdc_dst)
    _link(vimrc_src, vimrc_dst)

    # install/update all plugins
    local('vim +PluginInstall +qall')
    local('vim +PluginUpdate +qall')


def git():
    print "[*] Deploying git ..."
    gitconfig_src = os.path.join(_curdir, 'git', 'gitconfig')
    gitconfig_dst = os.path.expanduser('~/.gitconfig')
    gitignore_src = os.path.join(_curdir, 'git', 'gitignore')
    gitignore_dst = os.path.expanduser('~/.gitignore')
    
    _link(gitconfig_src, gitconfig_dst)
    _link(gitignore_src, gitignore_dst)


def xfce():
    print "[*] Deploying xfce ..."
    terminal_src = os.path.join(_curdir, 'xfce', 'terminalrc')
    terminal_dst = os.path.expanduser('~/.config/xfce4/terminal/terminalrc')
    shortcuts_src = os.path.join(_curdir, 'xfce', 'xfce4-keyboard-shortcuts.xml')
    shortcuts_dst = os.path.expanduser('~/.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-keyboard-shortcuts.xml')

    _link(terminal_src, terminal_dst)
    _link(shortcuts_src, shortcuts_dst)


def ipython():
    print "[*] Deploying ipython ..."
    profile_freaxmind = os.path.expanduser('~/.config/ipython/profile_freaxmind/')
    profile_default = os.path.expanduser('~/.config/ipython/profile_default')
    profile_config_src = os.path.join(_curdir, 'ipython', 'profile_freaxmind', 'ipython_config.py')
    profile_config_dst = os.path.join(profile_freaxmind, 'ipython_config.py')

    # create a new profile (if it does not already exists)
    if not os.path.exists(profile_freaxmind):
        local('ipython profile create freaxmind')

    if os.path.exists(profile_default):
        shutil.rmtree(profile_default)

    _link(profile_config_src, profile_config_dst)

def fonts():
    fonts_src = os.path.join(_curdir, 'fonts/')
    fonts_dst = os.path.expanduser('~/.fonts')

    _link(fonts_src, fonts_dst)

def security():
    print "[*] Performing Security Checks ..."
    print "\t[-] Files wih references to /home"
    local('rgrep /home .')

    print "\t[-] Files wih 'Group' and 'Others' permissions set"
    local('find . -perm -011')

def deploy_packages():
    apt()
    pip()

def deploy_conf():
    bash()
    byobu()
    vim()
    git()
    xfce()
    ipython()
    fonts()

def deploy_all():
    deploy_packages()
    deploy_conf()
