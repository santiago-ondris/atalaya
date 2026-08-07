import asyncio
import json

import httpx

from atalaya_interpreter.api import app
from atalaya_interpreter.config import Settings
from atalaya_interpreter.models import InterpretationRequest
from atalaya_interpreter.openrouter import OpenRouterClient, ProviderError

REQUEST = InterpretationRequest.model_validate(
    {
        "event": {
            "id": "018f47a8-7b2a-7a68-aeb3-2fcb95ea1031",
            "source": "sentry",
            "source_event_id": "event-1",
            "application": "prensap",
            "environment": "production",
            "occurred_at": "2026-08-06T14:30:00Z",
            "error_type": "TypeError",
            "message": "undefined value",
            "fingerprint": "typeerror:load",
        },
        "prompt_version": "error-analysis-v1",
    }
)


def test_parses_strict_openrouter_response() -> None:
    async def handler(request: httpx.Request) -> httpx.Response:
        sent = json.loads(request.content)
        assert sent["response_format"]["type"] == "json_schema"
        assert sent["max_tokens"] == 900
        return httpx.Response(
            200,
            json={
                "model": "fixture/model",
                "choices": [
                    {
                        "message": {
                            "content": json.dumps(
                                {
                                    "summary": "Resumen",
                                    "explanation": "Explicación",
                                    "severity": "medium",
                                    "actionable": True,
                                    "suggested_actions": ["Validar el dato"],
                                }
                            )
                        }
                    }
                ],
                "usage": {"prompt_tokens": 10, "completion_tokens": 4, "cost": 0.0001},
            },
        )

    client = OpenRouterClient(
        Settings(openrouter_api_key="test"), httpx.MockTransport(handler)
    )
    result = asyncio.run(client.interpret(REQUEST))
    assert result.usage.total_tokens == 14
    assert result.estimated_cost_usd == 0.0001


def test_classifies_invalid_provider_output_as_permanent() -> None:
    async def handler(_: httpx.Request) -> httpx.Response:
        return httpx.Response(200, json={"choices": []})

    client = OpenRouterClient(
        Settings(openrouter_api_key="test"), httpx.MockTransport(handler)
    )
    try:
        asyncio.run(client.interpret(REQUEST))
    except ProviderError as exc:
        assert exc.kind == "invalid_response"
        assert not exc.retryable
    else:
        raise AssertionError("expected ProviderError")


def test_api_returns_contract_problem_for_invalid_payload() -> None:
    async def send() -> httpx.Response:
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(
            transport=transport, base_url="http://test"
        ) as client:
            return await client.post("/v1/interpretations", json={"unexpected": True})

    response = asyncio.run(send())
    assert response.status_code == 422
    assert response.headers["content-type"].startswith("application/problem+json")
    assert response.json()["correlation_id"]
