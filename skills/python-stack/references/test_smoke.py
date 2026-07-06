import pytest
from litestar.testing import AsyncTestClient

from <slug> import __version__, app


def test_version() -> None:
    assert __version__


@pytest.mark.anyio
async def test_health_check() -> None:
    # Runs against a real Postgres container (see conftest.py); /health does SELECT 1.
    async with AsyncTestClient(app=app) as client:
        response = await client.get("/health")
        assert response.status_code == 200
        assert response.json() == {"status": "healthy", "database": "connected"}
