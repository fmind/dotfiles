"""Install tasks."""

# %% IMPORTS

import platform

from invoke import task
from invoke.context import Context

# %% TASKS

@task(default=True)
def site(ctx: Context, sudo: bool = False) -> None:
    """Install site.yml with ansible-playbook."""
    if platform.system() == "Darwin":
        ctx.run("ansible-playbook --become-user=$USER site.yml")
    elif sudo is True:
        ctx.run("ansible-playbook --ask-become site.yml")
    else:
        ctx.run("ansible-playbook site.yml")
