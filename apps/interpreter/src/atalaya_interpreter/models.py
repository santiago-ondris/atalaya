from datetime import datetime
from typing import Annotated, Any, Literal
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field, model_validator


class StrictModel(BaseModel):
    model_config = ConfigDict(extra="forbid")


class NormalizedError(StrictModel):
    id: UUID
    source: Literal["sentry", "application_insights"]
    source_event_id: str = Field(min_length=1, max_length=255)
    application: Literal["farmami", "wheels_house", "prensap", "notizap"]
    environment: str = Field(min_length=1, max_length=100)
    occurred_at: datetime
    error_type: str = Field(min_length=1, max_length=255)
    message: str = Field(min_length=1, max_length=10_000)
    stack_trace: str | None = Field(default=None, max_length=50_000)
    release: str | None = Field(default=None, max_length=255)
    fingerprint: str = Field(min_length=1, max_length=255)
    metadata: dict[str, Any] = Field(default_factory=dict)


class InterpretationRequest(StrictModel):
    event: NormalizedError
    prompt_version: str = Field(min_length=1, max_length=50)


class Analysis(StrictModel):
    summary: str = Field(min_length=1, max_length=500)
    explanation: str = Field(min_length=1, max_length=5000)
    severity: Literal["critical", "high", "medium", "low"]
    actionable: bool
    suggested_actions: list[Annotated[str, Field(min_length=1, max_length=500)]] = (
        Field(max_length=5)
    )


class TokenUsage(StrictModel):
    input_tokens: int = Field(ge=0)
    output_tokens: int = Field(ge=0)
    total_tokens: int = Field(ge=0)

    @model_validator(mode="after")
    def total_matches_parts(self) -> "TokenUsage":
        if self.total_tokens != self.input_tokens + self.output_tokens:
            raise ValueError("total_tokens must equal input_tokens + output_tokens")
        return self


class InterpretationResponse(Analysis):
    model: str = Field(min_length=1, max_length=255)
    prompt_version: str = Field(min_length=1, max_length=50)
    usage: TokenUsage
    estimated_cost_usd: float | None = Field(default=None, ge=0)
    latency_ms: int = Field(ge=0)
