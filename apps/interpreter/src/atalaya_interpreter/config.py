from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="ATALAYA_", env_file=".env", extra="ignore"
    )

    environment: str = "development"
    service_name: str = "interpreter"
    openrouter_api_key: str = ""
    openrouter_base_url: str = "https://openrouter.ai/api/v1"
    openrouter_model: str = "openai/gpt-4.1-mini"
    openrouter_timeout_seconds: float = 20.0
    openrouter_max_retries: int = 2
    openrouter_max_output_tokens: int = 900
    max_stack_trace_chars: int = 12000
    prompt_version: str = "error-analysis-v1"


@lru_cache
def get_settings() -> Settings:
    return Settings()
