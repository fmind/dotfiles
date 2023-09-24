"""Install tasks."""

# %% IMPORTS

from invoke import task
from invoke.context import Context

# %% TASKS

@task(default=True)
def site(ctx: Context) -> None:
    """Install site.yml with ansible."""
    ctx.run("ansible-playbook site.yml")
