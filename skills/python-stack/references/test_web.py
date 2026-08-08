from __future__ import annotations

from typing import cast
from unittest.mock import Mock

import pytest
from litestar.testing import AsyncTestClient
from pydantic import SecretStr
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from <package> import Settings, app, check_readiness, create_app, create_server, main, settings


class HealthySession:
    async def execute(self, _statement: object) -> None:
        return None


class UnhealthySession:
    async def execute(self, _statement: object) -> None:
        raise SQLAlchemyError("database unavailable")


def test_cors_is_closed_by_default() -> None:
    assert app.cors_config
    assert app.cors_config.allow_origins == []


@pytest.mark.anyio
async def test_health_check_is_dependency_free() -> None:
    config = Settings(database_url=SecretStr("postgresql+asyncpg://unavailable.invalid/test"), environment="test")

    async with AsyncTestClient(app=create_app(config)) as client:
        response = await client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "healthy"}


@pytest.mark.anyio
async def test_readiness_check_success() -> None:
    response = await check_readiness(cast(AsyncSession, HealthySession()))

    assert response.status_code == 200
    assert response.content == {"status": "ready", "database": "connected"}


@pytest.mark.anyio
async def test_readiness_check_failure() -> None:
    response = await check_readiness(cast(AsyncSession, UnhealthySession()))

    assert response.status_code == 503
    assert response.content == {"status": "not_ready", "database": "disconnected"}


def test_server_targets_the_import_package() -> None:
    server = create_server(settings)

    assert server.target == "<package>:app"


def test_main_starts_the_server(monkeypatch: pytest.MonkeyPatch) -> None:
    server = Mock()
    monkeypatch.setattr("<package>.create_server", lambda _config: server)

    main()

    server.serve.assert_called_once_with()
