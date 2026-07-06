"""<description>."""

from __future__ import annotations

import structlog
from advanced_alchemy.extensions.litestar import SQLAlchemyAsyncConfig, SQLAlchemyPlugin
from litestar import Litestar, Response, get
from litestar.config.cors import CORSConfig
from litestar.di import NamedDependency
from pydantic import SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict
from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import DeclarativeBase

__version__ = "0.1.0"


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    database_url: SecretStr = SecretStr("postgresql+asyncpg://postgres:postgres@localhost:5432/<slug>")
    environment: str = "development"
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


class Base(DeclarativeBase):
    pass


db_config = SQLAlchemyAsyncConfig(
    connection_string=settings.database_url.get_secret_value(),
    metadata=Base.metadata,
    create_all=False,  # schema is managed by Alembic; flip to True only for throwaway/smoke runs
)
db_plugin = SQLAlchemyPlugin(config=db_config)


@get("/health")
async def health_check(db_session: NamedDependency[AsyncSession]) -> Response[dict[str, str]]:
    """Verify database connectivity and application health."""
    try:
        await db_session.execute(text("SELECT 1"))
        return Response({"status": "healthy", "database": "connected"}, status_code=200)
    except Exception as e:
        logger.error("Health check database error", error=str(e))
        return Response({"status": "unhealthy", "database": "disconnected"}, status_code=500)


app = Litestar(
    plugins=[db_plugin],
    route_handlers=[health_check],
    cors_config=CORSConfig(allow_origins=["*"]),
)


def main() -> None:
    """Entrypoint invoked by the `<slug>` console script or `uv run`."""
    from granian import Granian

    # Auto-detect context for hot-reloading imports
    target = f"{__name__}:app"

    logger.info(
        "Starting <slug> web server",
        host=settings.host,
        port=settings.port,
        environment=settings.environment,
    )

    server = Granian(
        target,
        interface="asgi",
        port=settings.port,
        address=settings.host,
        reload=settings.environment == "development",
    )
    server.serve()


if __name__ == "__main__":
    main()
