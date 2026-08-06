from fastapi import FastAPI, Request

from atalaya_interpreter.config import get_settings
from atalaya_interpreter.observability import (
    CORRELATION_ID_HEADER,
    configure_logging,
    correlation_id,
    normalize_correlation_id,
)


def create_app() -> FastAPI:
    settings = get_settings()
    logger = configure_logging()
    app = FastAPI(
        title="Atalaya Interpreter",
        version="0.1.0",
        docs_url="/docs" if settings.environment == "development" else None,
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
    async def ready() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
