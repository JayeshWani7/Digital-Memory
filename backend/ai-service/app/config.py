import logging
import os
from typing import Optional

logger = logging.getLogger(__name__)


class Config:
    """Application configuration"""

    def __init__(self):
        # ---------------------------------------------------------------
        # DATABASE — default uses correct local credentials + sslmode=disable
        # ---------------------------------------------------------------
        self.database_url: str = os.getenv(
            "DATABASE_URL",
            "postgres://postgres:postgres@localhost:5432/digital_memory?sslmode=disable"
        )

        self.redis_url: str = os.getenv(
            "REDIS_URL",
            "redis://localhost:6379"
        )

        self.openai_api_key: str = os.getenv("OPENAI_API_KEY", "")
        self.openai_model: str = os.getenv("OPENAI_MODEL", "gpt-3.5-turbo")
        self.openai_embedding_model: str = os.getenv(
            "OPENAI_EMBEDDING_MODEL",
            "text-embedding-3-small"
        )
        self.log_level: str = os.getenv("LOG_LEVEL", "INFO")
        self.environment: str = os.getenv("ENV", "development")
        self.batch_size: int = int(os.getenv("BATCH_SIZE", "10"))
        self.max_retries: int = int(os.getenv("MAX_RETRIES", "3"))

        # DEBUG: print resolved DB URL so you can confirm credentials at startup
        logger.info("[DEBUG] DATABASE_URL = %s", self.database_url)
        logger.info("[DEBUG] REDIS_URL    = %s", self.redis_url)

        # Warn (don't crash) if OpenAI key is missing — lets DB-only tests run
        if not self.openai_api_key:
            logger.warning(
                "OPENAI_API_KEY is not set. LLM processing will fail. "
                "Set it in your .env file to enable AI features."
            )


class DatabaseConfig:
    """Database configuration"""
    pool_size = 5
    max_overflow = 10
    pool_timeout = 30
    pool_recycle = 3600
