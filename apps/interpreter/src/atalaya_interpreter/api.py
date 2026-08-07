from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from atalaya_interpreter.config import get_settings
from atalaya_interpreter.models import InterpretationRequest, InterpretationResponse
from atalaya_interpreter.observability import (
    CORRELATION_ID_HEADER,
    configure_logging,
    correlation_id,
    normalize_correlation_id,
)
from atalaya_interpreter.openrouter import OpenRouterClient, ProviderError


def create_app() -> FastAPI:
    settings = get_settings()
    logger = configure_logging()
    app = FastAPI(
        title="Atalaya Interpreter",
        version="0.1.0",
        docs_url="/docs" if settings.environment == "development" else None,
    )
    provider = OpenRouterClient(settings)

    @app.exception_handler(RequestValidationError)
    async def validation_error_handler(_request: Request, _exc: RequestValidationError):
        return problem(
            422, "Invalid request", "The payload does not satisfy the v1 contract"
        )

    @app.middleware("http")
    async def correlation_middleware(request: Request, call_next):
        request_correlation_id = normalize_correlation_id(
            request.headers.get(CORRELATION_ID_HEADER)
        )
        token = correlation_id.set(request_correlation_id)
        try:
            response = await call_next(request)
            response.headers[CORRELATION_ID_HEADER] = request_correlation_id
            logger.info(
                "HTTP request method=%s path=%s status=%s",
                request.method,
                request.url.path,
                response.status_code,
            )
            return response
        finally:
            correlation_id.reset(token)

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.get("/ready")
    async def ready():
        if not settings.openrouter_api_key:
            return problem(503, "Provider unavailable", "OpenRouter is not configured")
        return {"status": "ok"}

    @app.post("/v1/interpretations", response_model=InterpretationResponse)
    async def create_interpretation(payload: InterpretationRequest):
        if payload.prompt_version != settings.prompt_version:
            return problem(
                422,
                "Unsupported prompt version",
                "The requested prompt version is not deployed",
            )
        if not settings.openrouter_api_key:
            return problem(503, "Provider unavailable", "OpenRouter is not configured")
        try:
            return await provider.interpret(payload)
        except ProviderError as exc:
            logger.warning(
                "interpretation provider failure kind=%s retryable=%s detail=%s",
                exc.kind,
                exc.retryable,
                exc.detail,
            )
            status = 504 if exc.kind == "timeout" else 503 if exc.retryable else 502
            return problem(status, "Interpretation failed", exc.detail)

    return app


def problem(status: int, title: str, detail: str) -> JSONResponse:
    return JSONResponse(
        status_code=status,
        media_type="application/problem+json",
        content={
            "type": f"/problems/interpreter/{status}",
            "title": title,
            "status": status,
            "detail": detail,
            "correlation_id": correlation_id.get(),
        },
    )


app = create_app()
