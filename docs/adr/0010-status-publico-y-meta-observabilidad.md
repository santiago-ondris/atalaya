# ADR 0010 — Status público y meta-observabilidad

## Decisión

Watchman consulta un catálogo versionado de ocho objetivos HTTP cada minuto. Dos resultados consecutivos confirman tanto caída como recuperación. Cada prueba y cada fotografía agregada quedan persistidas; el uptime de 30 días solo se publica cuando existe una ventana completa.

`GET /api/v1/public/status` es el único contrato operativo sin sesión. Utiliza DTOs públicos deliberadamente pequeños y nunca serializa URLs, errores, proveedores, grupos o notas. Los incidentes requieren publicación explícita y pasan a retirados al marcarlos como ruido.

El evaluador interno combina heartbeats de procesos, frescura de pollers y profundidad/antigüedad de colas. Sus transiciones se guardan para deduplicar Telegram. Solo después de una evaluación y un `Ping` correcto a PostgreSQL se llama a `HEALTHCHECKS_PING_URL`; Healthchecks.io actúa como dead man's switch del proceso completo.

Cloudflare Pages sirve el mismo bundle y su Function `/api/*` reenvía hacia `WATCHMAN_ORIGIN`, conservando un único origen para cookies y API.

## Operación

- Intervalo HTTP: 60 s; timeout: 5 s; éxito: cualquier `2xx`.
- Poller detenido: tres intervalos sin intento. Worker detenido: 90 s sin heartbeat.
- Healthchecks.io: período 2 min, gracia 3 min. La URL se trata como secreto.
- Retener `availability_probes` y `availability_snapshots` al menos 30 días. La política definitiva queda para Sprint 11.
