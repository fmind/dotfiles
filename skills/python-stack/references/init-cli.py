"""<description>."""

from __future__ import annotations

import structlog
from pydantic_settings import BaseSettings, SettingsConfigDict

__version__ = "0.1.0"


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    environment: str = "development"


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


def main() -> None:
    """Entry point invoked by the `<slug>` console script or `uv run`."""
    logger.info("hello from <slug>", version=__version__)
