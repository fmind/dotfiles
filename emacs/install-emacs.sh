#!/bin/bash

set -e

readonly version="24.4"

# install dependencies
sudo apt-get install -y build-essential
sudo apt-get build-dep -y emacs24

# download source package
if [[ ! -d emacs-"$version" ]]; then
   wget http://ftp.gnu.org/gnu/emacs/emacs-"$version".tar.gz
   tar xzvf emacs-"$version".tar.gz
fi

# buil and install
cd emacs-"$version"
./configure
make
sudo make install

# remove build files
cd ..
rm -rf "emacs-${version}" "emacs-${version}.tar.gz"
