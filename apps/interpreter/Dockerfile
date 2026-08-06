FROM ghcr.io/astral-sh/uv:0.12.2 AS uv

FROM python:3.13.14-slim AS build

COPY --from=uv /uv /uvx /bin/
WORKDIR /app
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy
COPY pyproject.toml uv.lock ./
RUN uv sync --locked --no-dev --no-install-project
COPY src ./src
RUN uv sync --locked --no-dev --no-editable

FROM python:3.13.14-slim

RUN addgroup --system atalaya && adduser --system --ingroup atalaya atalaya
WORKDIR /app
COPY --from=build --chown=atalaya:atalaya /app/.venv /app/.venv
ENV PATH="/app/.venv/bin:$PATH" PYTHONUNBUFFERED=1
USER atalaya
EXPOSE 8000
CMD ["atalaya-interpreter"]
