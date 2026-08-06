from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="ATALAYA_", env_file=".env", extra="ignore"
    )

    environment: str = "development"
    service_name: str = "interpreter"


@lru_cache
def get_settings() -> Settings:
    return Settings()
