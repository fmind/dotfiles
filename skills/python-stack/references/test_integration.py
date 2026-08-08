from __future__ import annotations

import pytest
from litestar.testing import AsyncTestClient
from pydantic import SecretStr

from <package> import Settings, create_app


@pytest.mark.integration
@pytest.mark.anyio
async def test_readiness_check(postgres_database_url: str) -> None:
    config = Settings(database_url=SecretStr(postgres_database_url), environment="test")

    async with AsyncTestClient(app=create_app(config)) as client:
        response = await client.get("/ready")

    assert response.status_code == 200
    assert response.json() == {"status": "ready", "database": "connected"}
