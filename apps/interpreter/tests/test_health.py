import asyncio

import httpx

from atalaya_interpreter.api import app
from atalaya_interpreter.observability import CORRELATION_ID_HEADER


def test_health() -> None:
    response = asyncio.run(get("/health"))
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}
    assert response.headers[CORRELATION_ID_HEADER]


def test_ready() -> None:
    response = asyncio.run(get("/ready"))
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_preserves_valid_correlation_id() -> None:
    expected = "8f6f0961-6de7-47af-9ab7-0ad4b82e18d8"
    response = asyncio.run(get("/health", {CORRELATION_ID_HEADER: expected}))
    assert response.headers[CORRELATION_ID_HEADER] == expected


async def get(path: str, headers: dict[str, str] | None = None) -> httpx.Response:
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://test") as client:
        return await client.get(path, headers=headers)
