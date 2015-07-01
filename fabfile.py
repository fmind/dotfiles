#!/usr/bin/env python
# -*- coding: utf-8 -*-

from fabric.api import *
import shutil
import os

_hidden = lambda s: '.' + s
_pkg_line = lambda l: ' '.join(l)
_homedir = os.path.expanduser('~')
_curdir = os.path.dirname(os.path.realpath(__file__))
_proxy = os.getenv('http_proxy') or os.getenv('HTTP_PROXY')


def _link(src, dst):
    """ Create a link from the git repository (src) to the program directory (dst) """
    # returns if the link is already created
    if os.path.exists(dst) and os.path.islink(dst):
        return

    # ask to launch the application so the base directories can be created
    if not os.path.exists(os.path.dirname(dst)):
        print "\t>> Please launch the application associated to this file. Then press Enter: {0}".format(dst)
        raw_input()

    print "\tLinking {0} with {1}".format(src, dst)

    # removes the current directory/file if it's not  a link
    if os.path.exists(dst) and not os.path.islink(dst):
        if os.path.isfile(dst):
            os.remove(dst)
        else:
            shutil.rmtree(dst)

    # create the symbolic link
    os.symlink(src, dst)


def apt():
    print "[*] Installing new system packages (using apt) ..."
    packages = ['vim', 'vim-gui-common', 'byobu', 'python-dev', 'python-pip', 'python-flake8',
                'python-zmq', 'ipython', 'python-matplotlib', 'curl', 'git', 'zsh', 'exuberant-ctags']
    local('sudo apt-get install {packages}'.format(packages=_pkg_line(packages)))


def pip():
    print "[*] Installing new Python libraries (using pip) ..."
    packages = ['flake8']
    proxy = '--proxy {0}'.format(_proxy) if _proxy else ''
    local('pip install --user --upgrade {proxy} {packages}'.format(packages=_pkg_line(packages), proxy=proxy))


def zsh():
    print "[*] Deploying zsh ..."
    # base
    src = os.path.join(_curdir, 'zsh')
    dst = os.path.join(_homedir)
    # file configs
    zshrc = 'zshrc'
    inputrc = 'inputrc'
    # directory configs
    ohmyzsh = 'oh-my-zsh'

    # change the default shell to zsh
    #if not 'zsh' in os.getenv('SHELL'):
    #    local('chsh -s /usr/bin/zsh')

    _link(os.path.join(src, zshrc), os.path.join(dst, _hidden(zshrc)))
    _link(os.path.join(src, inputrc), os.path.join(dst, _hidden(inputrc)))
    _link(os.path.join(src, ohmyzsh), os.path.join(dst, _hidden(ohmyzsh)))


def byobu():
    print "[*] Deploying byobu ..."
    # base
    dirname = 'byobu'
    src = os.path.join(_curdir, dirname)
    dst = os.path.join(_homedir, _hidden(dirname))
    # file configs
    status = 'status'
    backend = 'backend'
    # directory configs
    layouts = 'layouts'

    _link(os.path.join(src, status), os.path.join(dst, status))
    _link(os.path.join(src, backend), os.path.join(dst, backend))
    _link(os.path.join(src, layouts), os.path.join(dst, layouts))


def vim(skip_plugins=False):
    print "[*] Deploying vim ..."
    # base
    src = os.path.join(_curdir, 'vim')
    dst = os.path.join(_homedir)
    # file configs
    vimrc = 'vimrc'
    # directory configs
    vimdc = 'vim'
    # executable
    font_installer = os.path.join(src, vimdc, 'bundle', 'fonts', 'install.sh')

    _link(os.path.join(src, vimdc), os.path.join(dst, _hidden(vimdc)))
    _link(os.path.join(src, vimrc), os.path.join(dst, _hidden(vimrc)))

    # install/update all vim plugins
    if not skip_plugins:
        local('vim +PluginInstall +qall')
        local('vim +PluginUpdate +qall')

    # install powerline fonts
    local(font_installer)


def git():
    print "[*] Deploying git ..."
    # base
    src = os.path.join(_curdir, 'git')
    dst = os.path.join(_homedir)
    # file configs
    gitconfig = 'gitconfig'
    gitignore = 'gitignore'

    _link(os.path.join(src, gitconfig), os.path.join(dst, _hidden(gitconfig)))
    _link(os.path.join(src, gitignore), os.path.join(dst, _hidden(gitignore)))


def xfce(pull=False):
    print "[*] Deploying xfce ..."
    # base
    src = os.path.join(_curdir, 'xfce')
    dst = os.path.join(_homedir, '.config', 'xfce4')
    # file configs
    terminal = os.path.join('terminal', 'terminalrc')
    shortcuts = os.path.join('xfconf', 'xfce-perchannel-xml', 'xfce4-keyboard-shortcuts.xml')

    if pull:
        print "\tPulling xfce files ..."
        shutil.copy(os.path.join(dst, terminal), os.path.join(src, terminal))
        shutil.copy(os.path.join(dst, shortcuts), os.path.join(src, shortcuts))
    else:
        print "\tPushing xfce files ..."
        shutil.copy(os.path.join(src, terminal), os.path.join(dst, terminal))
        shutil.copy(os.path.join(src, shortcuts), os.path.join(dst, shortcuts))


def ipython():
    print "[*] Deploying ipython ..."
    # base
    src = os.path.join(_curdir, 'ipython')
    dst = os.path.join(_homedir, '.config', 'ipython')
    # profile informations
    profile_name = 'freaxmind'
    profile_dir = 'profile_{0}'.format(profile_name)
    # file configs
    config = 'ipython_config.py'
    # directory configs
    profile_dir_src = os.path.join(src, profile_dir)
    profile_dir_dst = os.path.join(dst, profile_dir)

    # create a new profile if it does not already exist
    if not os.path.exists(profile_dir_dst):
        local('ipython profile create {profile_name}'.format(profile_name=profile_name))

    _link(os.path.join(profile_dir_src, config), os.path.join(profile_dir_dst, config))


def fonts():
    print "[*] Deploying fonts ..."
    # directory configs
    dirname = 'fonts'
    src = os.path.join(_curdir, dirname)
    dst = os.path.join(_homedir, _hidden(dirname))

    _link(src, dst)


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
    zsh()
    byobu()
    vim()
    git()
    xfce()
    ipython()
    fonts()


def deploy_all():
    deploy_packages()
    deploy_conf()
