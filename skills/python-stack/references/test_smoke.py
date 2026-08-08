from <package> import __main__ as module_entrypoint
from <package> import __version__


def test_version() -> None:
    assert __version__
    assert module_entrypoint.main
