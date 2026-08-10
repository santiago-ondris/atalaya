# Sprint 4 — Tres aplicaciones y seis proyectos Sentry

## Estado

- Sprint completado el 2026-08-10.
- Implementación, pruebas automatizadas, migración y conectividad con los seis
  proyectos: validadas.
- Validación end-to-end: completada con eventos reales que fueron interpretados
  correctamente y entregados en lenguaje natural por Telegram.

## Resultado

Watchman monitorea frontend y backend de Prensap, Farmami y Wheels House mediante
el mismo adapter. Cada proyecto tiene integración, checkpoint, filtro de ambiente
y métricas de ejecución independientes. PostgreSQL continúa siendo la fuente de
verdad operativa y los secretos permanecen exclusivamente en variables de entorno.

```text
Prensap ─────── backend ──┐
              frontend ──┤
Farmami ─────── backend ──┼── compuerta Sentry (concurrencia 1) ── adapter común
              frontend ──┤
Wheels House ── backend ──┤
              frontend ──┘
```

La compuerta fue necesaria porque la validación real reveló que el endpoint de
eventos de Sentry limita a una consulta concurrente por token. Solo serializa la
llamada HTTP: schedulers, fallos, persistencia y checkpoints siguen aislados.

## Configuración

`apps/watchman/config/sentry-integrations.json` contiene datos no sensibles:
aplicación, componente, proyecto, ambientes y overrides opcionales de política.
Las seis integraciones comparten `SENTRY_ORG_SLUG` y `SENTRY_AUTH_TOKEN`.

Ambientes habilitados:

| Aplicación | Componente | Proyecto | Ambiente |
|---|---|---|---|
| Prensap | Backend | `prensap-backend` | `production` |
| Prensap | Frontend | `prensap-frontend` | `production` |
| Farmami | Backend | `farmami-backend` | `production` |
| Farmami | Frontend | `farmami-frontend` | `vercel-production` |
| Wheels House | Backend | `wheelshouse-backend` | `production` |
| Wheels House | Frontend | `wheelshouse-frontend` | `production` |

## Migración y arranque seguro

La migración reemplaza la unicidad `(application_id, source)` por
`(application_id, source, component)`. La fila existente de Prensap backend no se
recrea: conserva UUID, cinco eventos históricos y checkpoint. Las otras cinco
integraciones fijan su alta como `monitoring_started_at`; el adapter descarta
eventos anteriores y detiene la paginación cuando cruza ese límite.

Una configuración inválida impide arrancar con un catálogo ambiguo. En cambio,
un fallo de red o proveedor se registra solo en el checkpoint afectado y los demás
pollers continúan.

## Políticas y observabilidad

La política default conserva el comportamiento del Sprint 3: severidades críticas
y altas siempre alertan; las medias solo cuando son accionables; deduplicación de
15 minutos y 10 alertas nuevas por aplicación cada 10 minutos. El catálogo admite
overrides por aplicación. El límite combina frontend y backend.

`GET /internal/integrations` devuelve los seis estados y sus timestamps. El listado
de eventos acepta `application` y `component`, y tanto la API como Telegram indican
el componente de origen.

## Validación realizada

- `go test ./...` y `go vet ./...` pasan.
- La imagen Docker contiene migraciones y catálogo.
- Goose aplicó la migración 3 sobre la base local existente.
- Los seis proyectos respondieron exitosamente usando el token actual.
- Los cinco proyectos nuevos quedaron con cero eventos históricos importados.
- Los seis pollers terminaron en estado `ok` y limpiaron errores transitorios.

## Validación final

La validación final ocurrió de manera orgánica durante el fin de semana posterior
a la implementación. Atalaya detectó eventos reales, completó sus interpretaciones
mediante OpenRouter y envió a Telegram mensajes útiles en lenguaje natural. Santiago
confirmó la recepción y la calidad de las interpretaciones el 2026-08-10.

Esto verifica el recorrido operativo completo construido en los Sprints 1–4:

```text
Sentry → poller aislado → PostgreSQL → interpreter → política de alertas → Telegram
```

La prueba controlada que estaba prevista dejó de ser necesaria porque el mismo
comportamiento fue observado con tráfico real.
