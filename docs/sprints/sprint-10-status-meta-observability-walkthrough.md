# Sprint 10 — Status y meta-observabilidad

## Configuración

1. Completar las ocho variables `*_URL` de `.env.example`. Las de backend incluyen el path de salud.
2. Configurar `HEALTHCHECKS_PING_URL` como secreto en Railway.
3. Crear Cloudflare Pages con raíz `apps/command-center`, comando `npm run build`, salida `dist` y `WATCHMAN_ORIGIN` apuntando a Watchman.
4. Aplicar la migración `00007_status_meta_observability.sql` y desplegar Watchman antes del frontend.

## Verificación

- Abrir `/status` sin cookies; comprobar cuatro aplicaciones, componentes e incidentes.
- Cambiar un objetivo a un servidor controlado que falle. El segundo ciclo confirma la caída y envía Telegram. Restaurarlo durante dos ciclos para confirmar recuperación.
- Atrasar `integrations.last_attempt_at` más de tres intervalos y consultar, con sesión, `/api/v1/system/health`.
- Detener Watchman y comprobar que Healthchecks.io notifica tras período y gracia configurados.
- Publicar un incidente con `PATCH /api/v1/incidents/{id}/publication`; resolverlo y comprobar que permanece en la bitácora. Marcarlo como ruido para confirmar su retiro automático.

La demo nunca debe utilizar un endpoint real de cliente para provocar una caída ni publicar información técnica o de clientes.
