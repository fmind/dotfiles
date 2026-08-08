"""<description>."""

from __future__ import annotations

from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from typing import Literal

import structlog
from granian import Granian
from granian.constants import Interfaces
from litestar import Litestar, Response, get
from litestar.config.cors import CORSConfig
from litestar.di import NamedDependency, Provide
from pydantic import Field, SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict
from sqlalchemy import text
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

__version__ = "0.1.0"


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    database_url: SecretStr
    cors_origins: list[str] = Field(default_factory=list)
    environment: Literal["development", "test", "production"] = "development"
    host: str = "127.0.0.1"
    port: int = 8000


settings = Settings()

structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.dev.ConsoleRenderer()
        if settings.environment == "development"
        else structlog.processors.JSONRenderer(),
    ],
)
logger = structlog.get_logger()


@get("/health")
async def health_check() -> dict[str, str]:
    """Report process liveness without touching external dependencies."""
    return {"status": "healthy"}


async def check_readiness(db_session: AsyncSession) -> Response[dict[str, str]]:
    """Report whether the database dependency can serve traffic."""
    try:
        await db_session.execute(text("SELECT 1"))
        return Response({"status": "ready", "database": "connected"}, status_code=200)
    except SQLAlchemyError:
        logger.exception("Readiness check database error")
        return Response({"status": "not_ready", "database": "disconnected"}, status_code=503)


@get("/ready")
async def readiness_check(db_session: NamedDependency[AsyncSession]) -> Response[dict[str, str]]:
    """Expose the dependency readiness check through HTTP."""
    return await check_readiness(db_session)


def create_app(config: Settings) -> Litestar:
    """Build an application whose database resources follow its lifespan."""
    engine = create_async_engine(config.database_url.get_secret_value())
    session_factory = async_sessionmaker(engine, expire_on_commit=False)

    async def provide_db_session() -> AsyncIterator[AsyncSession]:
        async with session_factory() as session:
            yield session

    @asynccontextmanager
    async def database_lifespan(_: Litestar) -> AsyncIterator[None]:
        try:
            yield
        finally:
            await engine.dispose()

    return Litestar(
        route_handlers=[health_check, readiness_check],
        dependencies={"db_session": Provide(provide_db_session)},
        cors_config=CORSConfig(allow_origins=config.cors_origins),
        lifespan=[database_lifespan],
    )


app = create_app(settings)


def create_server(config: Settings) -> Granian:
    """Configure Granian without starting a worker process."""
    # The import package is the ASGI target; the distribution/command may contain hyphens.
    target = f"{__name__}:app"

    return Granian(
        target,
        interface=Interfaces.ASGI,
        port=config.port,
        address=config.host,
        reload=config.environment == "development",
    )


def main() -> None:
    """Entrypoint invoked by the `<slug>` console script or `uv run`."""

    logger.info(
        "Starting <slug> web server",
        host=settings.host,
        port=settings.port,
        environment=settings.environment,
    )

    create_server(settings).serve()


if __name__ == "__main__":
    main()
