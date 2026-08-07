# Watchman

Servicio Go que concentra la lógica operativa y el acceso a PostgreSQL.

- `GET /health`: confirma que el proceso HTTP responde.
- `GET /ready`: confirma que PostgreSQL está disponible.
- `GET /internal/events?limit=50`: lista los eventos importados más recientes.
- `GET /internal/events/{id}`: muestra el evento normalizado y su stack trace.

## Poller de Prensap

Watchman activa el poller cuando existen `SENTRY_AUTH_TOKEN`, `SENTRY_ORG_SLUG` y
`SENTRY_PROJECT_SLUG`. Hace una primera consulta al arrancar y luego repite cada
`POLL_INTERVAL_SECONDS` (120 segundos por defecto). Sin token, el resto del
servicio permanece operativo y registra que la integración está desactivada.

El token necesita solamente permisos de lectura sobre eventos. Nunca se persiste:
se usa como Bearer token en la llamada a Sentry y cualquier header, token,
contraseña o secreto encontrado en los campos de texto se reemplaza por
`[REDACTED]` antes de guardar.

## Alertas por Telegram

El worker se activa sólo cuando `TELEGRAM_BOT_TOKEN` y `TELEGRAM_CHAT_ID` están
configurados juntos. Las interpretaciones críticas, altas, y medias accionables
abren ventanas durables de deduplicación. La primera alerta sale inmediatamente;
si el problema se repite, se envía un único resumen al cerrar la ventana.

Los defaults son 15 minutos de deduplicación, 10 alertas nuevas por aplicación
cada 10 minutos, cinco intentos por entrega y un cooldown de 30 minutos para
avisos de degradación del interpreter. Todos pueden ajustarse mediante las
variables documentadas en `.env.example`.

Watchman nunca registra el token ni el chat ID. Los intentos guardan resultado,
código HTTP, clase de error y, en caso exitoso, el ID de mensaje de Telegram.
