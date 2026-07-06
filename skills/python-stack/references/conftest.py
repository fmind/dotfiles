"""Shared pytest fixtures and integration-test wiring.

A real PostgreSQL container (matching the app's async driver) backs the tests, per
the "integration over mocks" standard. `pytest_configure` starts it and exports
`DATABASE_URL` *before* any test module imports the app, so `Settings()` binds to
the container; `pytest_unconfigure` tears it down. For suites with many DB-free
tests, gate this behind a marker so the container only starts when needed.
"""

from __future__ import annotations

import os

import pytest
from testcontainers.postgres import PostgresContainer

_postgres = PostgresContainer("postgres:17-alpine", driver="asyncpg")


def pytest_configure() -> None:
    _postgres.start()
    os.environ["DATABASE_URL"] = _postgres.get_connection_url()


def pytest_unconfigure() -> None:
    _postgres.stop()


@pytest.fixture
def anyio_backend() -> str:
    return "asyncio"
