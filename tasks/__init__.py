"""Task collection."""

# %% IMPORTS

from invoke import Collection

from . import docker, install

# %% NAMESPACES

ns = Collection()

# %% COLLECTIONS

ns.add_collection(docker)
ns.add_collection(install, default=True)
