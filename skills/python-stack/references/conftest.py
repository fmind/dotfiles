"""Opt-in integration-test fixtures."""

from __future__ import annotations

from collections.abc import Iterator

import pytest
from testcontainers.community.postgres import PostgresContainer


@pytest.fixture(scope="session")
def postgres_database_url() -> Iterator[str]:
    """Start PostgreSQL only for tests that request this fixture."""
    with PostgresContainer("postgres:17-alpine", driver="asyncpg") as postgres:
        yield postgres.get_connection_url()


@pytest.fixture
def anyio_backend() -> str:
    return "asyncio"
