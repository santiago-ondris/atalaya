# Watchman

Servicio Go que concentra la lógica operativa y el acceso a PostgreSQL.

- `GET /health`: confirma que el proceso HTTP responde.
- `GET /ready`: confirma que PostgreSQL está disponible.
- `GET /internal/events?limit=50`: lista los eventos importados más recientes.
- `GET /internal/events/{id}`: muestra el evento normalizado y su stack trace.

## Pollers Sentry

Watchman activa seis pollers independientes cuando existen `SENTRY_AUTH_TOKEN` y
`SENTRY_ORG_SLUG`: frontend y backend de Prensap, Farmami y Wheels House. El
catálogo versionado `config/sentry-integrations.json` define proyecto, componente,
ambientes productivos y overrides opcionales de alertas. Las consultas comparten
una compuerta porque Sentry admite una sola llamada concurrente a este endpoint.

Cada integración conserva su propio checkpoint, momento de habilitación, última
ejecución, último éxito y último error. Una integración nueva solo importa eventos
posteriores a su alta. Sin token, el resto del servicio permanece operativo y
registra que los pollers están desactivados.

El token necesita solamente permisos de lectura sobre eventos. Nunca se persiste:
se usa como Bearer token en la llamada a Sentry y cualquier header, token,
contraseña o secreto encontrado en los campos de texto se reemplaza por
`[REDACTED]` antes de guardar.

- `GET /internal/integrations`: estado operativo de cada integración.
- `GET /internal/events?application=farmami&component=frontend`: filtra eventos
  por aplicación y componente; ambos filtros son opcionales.

## Alertas por Telegram

El worker se activa sólo cuando `TELEGRAM_BOT_TOKEN` y `TELEGRAM_CHAT_ID` están
configurados juntos. Las interpretaciones críticas, altas, y medias accionables
abren ventanas durables de deduplicación. La primera alerta sale inmediatamente;
si el problema se repite, se envía un único resumen al cerrar la ventana.

Los defaults son 15 minutos de deduplicación, 10 alertas nuevas por aplicación
cada 10 minutos, cinco intentos por entrega y un cooldown de 30 minutos para
avisos de degradación del interpreter. Todos pueden ajustarse mediante las
variables documentadas en `.env.example`; el catálogo puede sobrescribir la
política para una aplicación. El límite se comparte entre su frontend y backend.

Watchman nunca registra el token ni el chat ID. Los intentos guardan resultado,
código HTTP, clase de error y, en caso exitoso, el ID de mensaje de Telegram.
