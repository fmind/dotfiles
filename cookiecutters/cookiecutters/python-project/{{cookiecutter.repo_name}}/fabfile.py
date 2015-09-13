#!/usr/bin/env python
# -*- coding: utf-8 -*-

from fabric.api import local


def freeze():
    local('pip freeze > requirements.txt')


def install():
    local('pip install -r requirements.txt')


def coverage():
    local('py.test --cov={{cookiecutter.project_name}}')
