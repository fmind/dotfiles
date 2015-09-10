"""
Tests for `{{ cookiecutter.repo_name }}` module.
"""
import pytest
from {{ cookiecutter.repo_name }} import {{ cookiecutter.repo_name }}


class Test{{ cookiecutter.repo_name|capitalize }}(object):

    @classmethod
    def setup_class(cls):
        pass

    def test_something(self):
        pass

    @classmethod
    def teardown_class(cls):
        pass
