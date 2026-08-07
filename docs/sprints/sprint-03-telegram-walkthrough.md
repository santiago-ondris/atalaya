# Sprint 03 — Telegram, deduplicación y resiliencia

## Resultado

Las interpretaciones relevantes generan alertas durables al chat privado configurado. Las
ocurrencias equivalentes se agrupan durante 15 minutos, existe rate limiting por aplicación y
cada intento de entrega queda trazado en PostgreSQL.

## Flujo

```text
interpretation completed
        │ misma transacción
        ▼
alert_window + notification_jobs
        │
        ▼
notification worker → Telegram Bot API
        │
        ├── éxito → delivery_attempt + completed
        └── fallo → delivery_attempt + retry/backoff o failed
```

## Piezas principales

- `migrations/00002_telegram_alerting.sql`: ventanas, jobs e historial de intentos.
- `internal/domain/event.go`: elegibilidad y contratos internos de notificación.
- `internal/notification`: formato seguro de mensajes y worker durable.
- `internal/telegram`: cliente HTTP y clasificación de errores.
- `internal/store/postgres.go`: transacciones, deduplicación, rate limiting y leases.
- `docs/adr/0006-alertas-telegram-y-control-de-ruido.md`: política y tradeoffs.

## Configuración

El bot no puede iniciar un chat privado: primero el usuario debe enviarle `/start`. Luego se
configuran `TELEGRAM_BOT_TOKEN` y `TELEGRAM_CHAT_ID`. Los límites opcionales y sus defaults están
en `.env.example`; ningún secreto se persiste en la base ni se incluye en logs.

## Resiliencia

Los códigos 429 y 5xx, junto con fallos de red, son transitorios. Se reintentan con backoff
exponencial. Los 4xx restantes son permanentes. Un lease abandonado durante cinco minutos puede
ser recuperado por otro worker.

La semántica de entrega es al menos una vez. Existe una ventana excepcional de duplicación si
Telegram acepta un mensaje y el proceso cae antes de registrar el éxito, porque `sendMessage` no
acepta claves de idempotencia.

## Validación

```bash
cd apps/watchman
go test ./...
docker compose config --quiet
docker compose up -d --build migrate watchman
```

La prueba real inicial confirmó que el worker arrancó con las credenciales locales y que el bot
pudo entregar un mensaje controlado al chat privado configurado.
