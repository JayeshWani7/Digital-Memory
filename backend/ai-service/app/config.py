import os
from typing import Optional


class Config:
    """Application configuration"""
    
    def __init__(self):
        self.database_url: str = os.getenv(
            "DATABASE_URL",
            "postgres://user:password@localhost:5432/digital_memory"
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
        
        # Validate required settings
        if not self.openai_api_key:
            raise ValueError("OPENAI_API_KEY environment variable is required")


class DatabaseConfig:
    """Database configuration"""
    pool_size = 5
    max_overflow = 10
    pool_timeout = 30
    pool_recycle = 3600
