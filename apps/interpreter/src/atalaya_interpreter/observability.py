import json
import logging
from contextvars import ContextVar
from datetime import UTC, datetime
from uuid import UUID, uuid4

CORRELATION_ID_HEADER = "X-Correlation-ID"
correlation_id: ContextVar[str] = ContextVar("correlation_id", default="")


class JSONFormatter(logging.Formatter):
    def format(self, record: logging.LogRecord) -> str:
        payload = {
            "time": datetime.now(UTC).isoformat(),
            "level": record.levelname.lower(),
            "message": record.getMessage(),
        }
        current_correlation_id = correlation_id.get()
        if current_correlation_id:
            payload["correlation_id"] = current_correlation_id
        return json.dumps(payload)


def configure_logging() -> logging.Logger:
    handler = logging.StreamHandler()
    handler.setFormatter(JSONFormatter())
    logger = logging.getLogger("atalaya.interpreter")
    logger.handlers = [handler]
    logger.setLevel(logging.INFO)
    logger.propagate = False
    return logger


def normalize_correlation_id(value: str | None) -> str:
    if value is None:
        return str(uuid4())
    try:
        return str(UUID(value))
    except ValueError:
        return str(uuid4())
