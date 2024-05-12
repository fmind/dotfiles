"""Docker tasks."""

# %% IMPORTS

from invoke import task
from invoke.context import Context

# %% TASKS

@task
def run(ctx: Context) -> None:
    """Run the docker image."""
    ctx.run(f"docker run --rm -it {ctx.docker.image}", pty=True)


@task
def build(ctx: Context) -> None:
    """Build the docker image."""
    ctx.run(f"docker build -t {ctx.docker.image} .")


@task
def push(ctx: Context) -> None:
    """Push the image to Docker Hub."""
    ctx.run(f"docker push {ctx.docker.image}")


@task(pre=[build, push], default=True)
def default(_: Context) -> None:
    """Run all docker tasks."""
