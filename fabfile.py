#!/usr/bin/env python
# -*- coding: utf-8 -*-

from fabric.api import local
import shutil
import os


_homedir = os.path.expanduser('~')
_curdir = os.path.dirname(os.path.realpath(__file__))
_proxy = os.getenv('http_proxy') or os.getenv('HTTP_PROXY')


def _hidden(s):
    return '.' + s


def _pkg_line(l):
    return ' '.join(l)


def _link(src, dst):
    """ Create a symbolic link from the git repository (src) to the host directory (dst) """
    # returning if the link is already created
    if os.path.exists(dst) and os.path.islink(dst):
        return

    # asking to launch the application so the base directories can be created
    if not os.path.exists(os.path.dirname(dst)):
        print("\t>> Please launch the application associated to this file: {0}".format(dst))

    print("\tLinking {0} to {1}".format(src, dst))

    # removing the current directory/file if it's not a link or if the link is broken
    if (os.path.exists(dst) and not os.path.islink(dst)) or (os.path.islink(dst) and not os.path.exists(dst)):
        if os.path.isfile(dst) or os.path.islink(dst):
            os.remove(dst)
        else:
            shutil.rmtree(dst)

    # creating a symbolic link
    os.symlink(src, dst)


def apt(update_packages=False):
    print("[*] Installing new system packages (using apt-get) ...")
    packages = ['zsh', 'byobu', 'vim', 'git', 'silversearcher-ag', 'fabric',            # shell
                'python3', 'python3-dev', 'python-dev', 'build-essential', 'cmake',     # programming
                'libpng12-dev', 'libfreetype6-dev', 'gfortran', 'python3-setuptools',   # dependencies
                'gfortran', 'libatlas-base-dev', 'liblapack-dev', 'libblas-dev']

    if update_packages:
        print("[*] Updating system packages (using apt-get) ...")
        local('sudo apt-get update')
        local('sudo apt-get upgrade')

    local('sudo apt-get install {packages}'.format(packages=_pkg_line(packages)))


def pip():
    print("[*] Installing/Updating Python libraries (using pip) ...")
    packages = ['cookiecutter', 'virtualenv', 'wheel', 'pytest', 'pytest-cov', 'sphinx',    # programming environment
                'frosted', 'pep8', 'py3kwarn',                                              # syntax checker
                'jupyter', 'pandas', 'seaborn', 'ipython',                                  # data analysis
                'pymongo', 'redis', 'mongoengine',                                          # databases
                'flask', 'requests', 'httpie', 'beautifulsoup4']                            # web

    # install a standalone version of pip
    local('wget https://bootstrap.pypa.io/get-pip.py')
    local('sudo -H python3 get-pip.py')
    local('rm get-pip.py')

    proxy = '--proxy {0}'.format(_proxy) if _proxy else ''
    local('pip3 install --user --upgrade {proxy} {packages}'.format(packages=_pkg_line(packages), proxy=proxy))


def zsh(deploy_shell=False):
    print("[*] Deploying zsh ...")
    # base
    src = os.path.join(_curdir, 'zsh')
    dst = os.path.join(_homedir)
    # file configs
    zshrc = 'zshrc'
    inputrc = 'inputrc'
    # directory configs
    ohmyzsh = 'oh-my-zsh'

    if deploy_shell and 'zsh' not in os.getenv('SHELL'):
        print("[*] Changing default shell to zsh ...")
        local('chsh -s /usr/bin/zsh')

    _link(os.path.join(src, zshrc), os.path.join(dst, _hidden(zshrc)))
    _link(os.path.join(src, inputrc), os.path.join(dst, _hidden(inputrc)))
    _link(os.path.join(src, ohmyzsh), os.path.join(dst, _hidden(ohmyzsh)))


def vim(update_plugins=False):
    print("[*] Deploying vim ...")
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

    local('vim +PluginInstall +qall')

    if update_plugins:
        print("[*] Updating vim plugins ...")
        local('vim +PluginUpdate +qall')
        local('python ~/.vim/bundle/YouCompleteMe/install.py --clang-completer --gocode-completer')

    # install powerline fonts
    print("[*] Installing powerline fonts ...")
    local(font_installer)


def git():
    print("[*] Deploying git ...")
    # base
    src = os.path.join(_curdir, 'git')
    dst = os.path.join(_homedir)
    # file configs
    gitconfig = 'gitconfig'
    gitignore = 'gitignore'

    _link(os.path.join(src, gitconfig), os.path.join(dst, _hidden(gitconfig)))
    _link(os.path.join(src, gitignore), os.path.join(dst, _hidden(gitignore)))


def ipython():
    print("[*] Deploying ipython ...")
    # base
    src = os.path.join(_curdir, 'ipython')
    dst = os.path.join(_homedir, '.ipython')
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


def cookie():
    print("[*] Deploying cookiecutter ...")
    # base
    src = os.path.join(_curdir, 'cookiecutters')
    dst = os.path.join(_homedir)
    # file configs
    cookierc = 'cookiecutterrc'
    # directory configs
    cookiedc = 'cookiecutters'

    _link(os.path.join(src, cookiedc), os.path.join(dst, _hidden(cookiedc)))
    _link(os.path.join(src, cookierc), os.path.join(dst, _hidden(cookierc)))


def fonts(update_cache=False):
    print("[*] Deploying fonts ...")
    # directory configs
    dirname = 'fonts'
    src = os.path.join(_curdir, dirname)
    dst = os.path.join(_homedir, _hidden(dirname))

    # update fonts cache
    if update_cache:
        print("[*] Updating font-cache ...")
        local('sudo fc0cache fv')

    _link(src, dst)


def security():
    print("[*] Performing Security Checks ...")
    print("\t[-] Files wih references to /home")
    local('rgrep /home .')

    print("\t[-] Files wih 'Group' and 'Others' permissions set")
    local('find . -perm -011')


def clean_pip():
    print("[*] Cleaning pip libs and bins ...")
    local('rm -r ~/.local/bin ~/.local/lib')


def deploy_packages():
    apt()
    pip()


def deploy_conf():
    zsh()
    vim()
    git()
    ipython()
    cookie()
    fonts()


def deploy_all():
    deploy_packages()
    deploy_conf()
