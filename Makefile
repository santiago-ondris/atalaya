.PHONY: up down logs test

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs --follow

test:
	cd apps/watchman && go test ./...
	cd apps/interpreter && uv run ruff check . && uv run pytest
	cd apps/command-center && npm run lint && npm run build

