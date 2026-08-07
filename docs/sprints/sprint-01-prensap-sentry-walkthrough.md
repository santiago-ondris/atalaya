# Sprint 1 — Primer evento de Prensap desde Sentry

- Fecha: 2026-08-07
- Estado de implementación: completado
- Validación controlada contra Sentry: completada el 2026-08-07

## Resultado

Watchman puede consultar los eventos de error de Prensap, normalizarlos al modelo
común de Atalaya y persistir cada evento externo una sola vez. La primera consulta
ocurre al iniciar el proceso y el polling continúa cada 120 segundos.

```text
Sentry API
   │  Bearer token (solo en memoria)
   ▼
SentrySource ──► Event común ──► transacción PostgreSQL
                                      ├── error_groups
                                      ├── error_events
                                      └── source_checkpoints
```

## Responsabilidad de los archivos

- `internal/domain/event.go`: modelo neutral y límite `ErrorSource`; no conoce
  payloads ni URLs de Sentry.
- `internal/sentry/client.go`: cliente HTTP, paginación por cursor y traducción
  del payload de Sentry.
- `internal/sentry/sanitize.go`: remueve credenciales y headers sensibles antes
  de persistir texto o incluir respuestas del proveedor en errores.
- `internal/store/postgres.go`: crea la integración, importa lotes de forma
  transaccional y resuelve las consultas internas.
- `internal/poller/poller.go`: ejecuta inmediatamente y cada dos minutos, registra
  éxito o fallo y conserva el checkpoint.
- `internal/httpserver/server.go`: expone lista y detalle de eventos importados.

No se agregó una librería HTTP ni un scheduler: `net/http` y `time.Ticker` de la
biblioteca estándar cubren este volumen y mantienen explícitos timeouts y
cancelación. Se reutilizan `pgx` y PostgreSQL para transacciones e idempotencia.

## Idempotencia y checkpoint

`(integration_id, source_event_id)` es único. Antes de actualizar el grupo se
comprueba esa identidad, por lo que reimportar el mismo evento no incrementa el
contador. Eventos nuevos actualizan primera/última aparición y ocurrencias.

El lote y el cursor se guardan dentro de la misma transacción. Si cualquier insert
falla, el cursor no avanza. La siguiente corrida vuelve a intentar el lote.
Cuando la paginación llega al final, el cursor vacío hace que la siguiente corrida
vuelva al inicio y detecte eventos recientes; los ya vistos se descartan por su ID.

## Demo reproducible con datos reales

1. Crear `.env` desde `.env.example` y completar organización, proyecto y un token
   de Sentry de solo lectura.
2. Ejecutar `make up`.
3. Provocar o elegir un error de Prensap y esperar como máximo dos minutos.
4. Consultar `curl http://localhost:8080/internal/events`.
5. Copiar el `id` interno y consultar
   `curl http://localhost:8080/internal/events/{id}`.
6. Esperar otra corrida y comprobar que el mismo `source_event_id` aparece una
   sola vez.

## Validación automatizada

Los tests cubren normalización, autorización, cursor, sanitización y la interacción
poller/checkpoint sin red real. `go test ./...` y `go vet ./...` deben pasar.

La llamada real no forma parte del test automático para evitar colocar secretos o
depender de Sentry en CI. Queda como prueba controlada local una vez provistas las
variables del paso anterior.

## Validación real realizada

La prueba real detectó que el endpoint de eventos de Sentry devolvía cero resultados al usar
`query=event.type:error`, aunque el payload sin filtro identificaba correctamente esos eventos
con `type=error`. El adapter ahora consulta el feed sin ese filtro incompatible y descarta del
lado de Atalaya cualquier evento cuyo tipo no sea `error`.

## Deuda y siguiente paso

- Los endpoints `/internal` todavía no tienen autenticación; el login y la API
  privada están planificados para el Sprint 6. No deben publicarse sin una regla de
  red o autenticación previa.
- El fingerprint inicial de Sentry usa tipo + culprit (o mensaje como fallback).
  La política más sofisticada de agrupamiento se ajustará con datos reales en el
  Sprint 3.
- Sprint 2 consumirá estos eventos mediante jobs durables para interpretarlos con
  OpenRouter.
