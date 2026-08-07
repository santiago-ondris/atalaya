import json
import time
from dataclasses import dataclass

import httpx
from pydantic import ValidationError

from atalaya_interpreter.config import Settings
from atalaya_interpreter.models import (
    Analysis,
    InterpretationRequest,
    InterpretationResponse,
    TokenUsage,
)
from atalaya_interpreter.prompt import SYSTEM_PROMPT, event_prompt


@dataclass
class ProviderError(Exception):
    detail: str
    kind: str
    retryable: bool

    def __str__(self) -> str:
        return self.detail


class OpenRouterClient:
    def __init__(
        self, settings: Settings, transport: httpx.AsyncBaseTransport | None = None
    ):
        self.settings = settings
        self.client = httpx.AsyncClient(
            base_url=settings.openrouter_base_url,
            timeout=settings.openrouter_timeout_seconds,
            transport=transport,
        )

    async def interpret(self, request: InterpretationRequest) -> InterpretationResponse:
        started = time.monotonic()
        payload = {
            "model": self.settings.openrouter_model,
            "messages": [
                {"role": "system", "content": SYSTEM_PROMPT},
                {
                    "role": "user",
                    "content": event_prompt(
                        request.event, self.settings.max_stack_trace_chars
                    ),
                },
            ],
            "temperature": 0.1,
            "max_tokens": self.settings.openrouter_max_output_tokens,
            "response_format": {
                "type": "json_schema",
                "json_schema": {
                    "name": "atalaya_error_analysis",
                    "strict": True,
                    "schema": Analysis.model_json_schema(),
                },
            },
        }
        response = await self._post_with_retries(payload)
        try:
            body = response.json()
            content = body["choices"][0]["message"]["content"]
            analysis = Analysis.model_validate(json.loads(content))
            raw_usage = body.get("usage", {})
            input_tokens = int(raw_usage.get("prompt_tokens", 0))
            output_tokens = int(raw_usage.get("completion_tokens", 0))
            usage = TokenUsage(
                input_tokens=input_tokens,
                output_tokens=output_tokens,
                total_tokens=input_tokens + output_tokens,
            )
            cost = raw_usage.get("cost")
            return InterpretationResponse(
                **analysis.model_dump(),
                model=body.get("model", self.settings.openrouter_model),
                prompt_version=request.prompt_version,
                usage=usage,
                estimated_cost_usd=float(cost) if cost is not None else None,
                latency_ms=int((time.monotonic() - started) * 1000),
            )
        except ValidationError as exc:
            locations = ", ".join(
                ".".join(str(part) for part in error["loc"])
                for error in exc.errors(include_input=False)
            )
            raise ProviderError(
                f"OpenRouter response failed schema validation at: {locations}",
                "invalid_response",
                False,
            ) from exc
        except (
            KeyError,
            IndexError,
            TypeError,
            ValueError,
            json.JSONDecodeError,
        ) as exc:
            raise ProviderError(
                f"OpenRouter returned an invalid structured response ({type(exc).__name__})",
                "invalid_response",
                False,
            ) from exc

    async def _post_with_retries(self, payload: dict) -> httpx.Response:
        for attempt in range(self.settings.openrouter_max_retries + 1):
            try:
                response = await self.client.post(
                    "/chat/completions",
                    headers={
                        "Authorization": f"Bearer {self.settings.openrouter_api_key}"
                    },
                    json=payload,
                )
            except httpx.TimeoutException as exc:
                if attempt < self.settings.openrouter_max_retries:
                    continue
                raise ProviderError("OpenRouter timed out", "timeout", True) from exc
            except httpx.HTTPError as exc:
                if attempt < self.settings.openrouter_max_retries:
                    continue
                raise ProviderError(
                    "Could not connect to OpenRouter", "unavailable", True
                ) from exc
            if response.status_code in {408, 429} or response.status_code >= 500:
                if attempt < self.settings.openrouter_max_retries:
                    continue
                raise ProviderError(
                    "OpenRouter is temporarily unavailable", "unavailable", True
                )
            if response.status_code >= 400:
                raise ProviderError(
                    "OpenRouter rejected the request", "rejected", False
                )
            return response
        raise AssertionError("unreachable")
