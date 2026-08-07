# Sprint 02 — Interpretación con OpenRouter

## Resultado

Cada evento nuevo de Prensap crea automáticamente un job durable. Watchman lo envía al
Interpreter, clasifica fallos y persiste la explicación, severidad, actionabilidad, acciones,
modelo, tokens, latencia y costo reportado por OpenRouter.

## Piezas principales

- `contracts/openapi/interpreter.v1.yaml`: contrato HTTP versionado entre servicios.
- `apps/interpreter/models.py`: schemas estrictos de entrada, análisis y respuesta.
- `apps/interpreter/prompt.py`: prompt versionado y truncado determinista del stack trace.
- `apps/interpreter/openrouter.py`: integración, timeout, retries y validación del proveedor.
- `apps/watchman/internal/interpreter`: cliente HTTP y worker durable.
- `apps/watchman/internal/store/postgres.go`: claim, finalización, retry y recuperación de leases.

## Probar localmente

1. Copiar `.env.example` a `.env` y completar `OPENROUTER_API_KEY` y las credenciales Sentry.
2. Ejecutar `make up`.
3. Importar o provocar un evento de Prensap.
4. Consultar `GET http://localhost:8080/internal/events` y luego
   `GET http://localhost:8080/internal/events/{id}`.

El detalle devuelve `interpretation` cuando el job terminó. Si OpenRouter no está configurado,
`/health` continúa confirmando que el proceso vive y `/ready` devuelve 503; los jobs quedan
durables y se reintentan de manera acotada.

## Validación automatizada

```bash
cd apps/watchman && go test ./...
cd apps/interpreter && uv run ruff check . && uv run pytest
```

Los tests cubren el contrato del cliente Go, la clasificación de un 503, el JSON Schema enviado
a OpenRouter, el cálculo de tokens y el rechazo de respuestas estructuradas inválidas.

## Validación real completada

El 2026-08-07 se importó un error real de Prensap y se completó automáticamente su job. La
interpretación quedó almacenada con severidad, acciones sugeridas, modelo, 597 tokens, costo
reportado de USD 0.0004368 y 4.902 ms de latencia.

La prueba descubrió además que OpenRouter reserva crédito contra la salida máxima cuando
`max_tokens` no está presente. Se fijó un límite inicial configurable de 900 tokens, suficiente
para el contrato y necesario para acotar el costo teórico de cada solicitud. Ningún secreto fue
incorporado al repositorio.
