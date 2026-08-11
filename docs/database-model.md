# Modelo de datos inicial

## Relaciones

```text
applications
	├── daily_report_applications
	└── integrations
          ├── source_checkpoints
          └── error_groups
                ├── error_events
                │     └── interpretation_jobs
                │           └── interpretations
                └── alert_windows
                      └── notification_jobs
                            └── notification_delivery_attempts

daily_reports
	├── daily_report_applications
	└── daily_report_delivery_attempts

applications
	├── incidents
	│     ├── incident_error_groups ── error_groups
	│     └── incident_entries
	└── deployments
```

## Decisiones importantes

### Evento y grupo son conceptos distintos

`error_events` conserva una ocurrencia externa individual e idempotente.
`error_groups` agrupa eventos equivalentes por integración, entorno y fingerprint.
Esto permite investigar una ocurrencia concreta sin perder el contador agregado.

### Idempotencia

La restricción única `(integration_id, source_event_id)` evita importar dos veces
el mismo evento aunque un poller repita una ventana temporal. El código debe tratar
un conflicto de esa restricción como evento ya observado, no como fallo del poller.

### Checkpoints

Cada integración mantiene un cursor JSON independiente porque Sentry y Application
Insights no necesariamente expresan su progreso de la misma forma. El adapter es
responsable de validar la estructura que escribe y lee.

Una aplicación puede tener varias integraciones de la misma fuente, diferenciadas
por `component` (`frontend` o `backend`). `monitoring_started_at` evita que una
integración recién habilitada recorra e interprete historia anterior a su alta.
`last_attempt_at`, `last_success_at` y `last_error` alimentan el estado operativo.

### Políticas por aplicación

`applications.alert_policy` conserva la política efectiva de alertas. Frontend y
backend mantienen fingerprints y deduplicación independientes, pero comparten el
rate limit de su aplicación. Los filtros de ambientes pertenecen a la integración,
porque Farmami usa `production` en backend y `vercel-production` en frontend.

### Jobs durables

Los jobs pendientes se ordenan por `available_at` y `created_at`. Un worker los
reclamará dentro de una transacción usando `FOR UPDATE SKIP LOCKED`, marcará
`processing` e identificará su lease mediante `locked_at` y `locked_by`.

Los workers recuperan leases `processing` con más de cinco minutos. Tanto las
interpretaciones como las notificaciones conservan su contador de intentos,
último error y próximo instante disponible.

### Ventanas y entregas

`alert_windows` representa una ventana de deduplicación operativa para un grupo.
Mantiene el primer y último instante y el contador que alimenta el mensaje
consolidado. `notification_jobs` separa la decisión de alertar de la entrega a
Telegram. `notification_delivery_attempts` conserva cada resultado sin guardar el
token del bot ni otro secreto.

### Costos

Los costos usan `numeric`, no punto flotante. `NULL` significa que el proveedor no
dio información suficiente para estimarlos; cero significa costo conocido igual a
cero.

### Reportes diarios

`daily_reports.report_date` es la clave idempotente del día calendario argentino.
Cada reporte congela sus límites UTC, estado, próximo intento y resultado de
Telegram. `daily_report_applications` conserva la fotografía de errores y actividad
por aplicación; una actividad `unavailable` mantiene `activity_count` en `NULL`
para distinguir una fuente caída de un día con cero sesiones.

`daily_report_delivery_attempts` audita cada envío. Los fallos transitorios se
reprograman 30 minutos y los reportes que alcanzan `period_end` pasan a `expired`.

### Incidentes y deploys

Un incidente pertenece a una aplicación y agrupa uno o más `error_groups`. Un
grupo puede formar parte de distintos incidentes históricos, pero las
transacciones y advisory locks impiden que quede en dos investigaciones activas.
`incident_entries` es append-only y conserva notas, transiciones y cambios de
grupos; cerrar o reabrir siempre requiere una explicación.

`deployments` normaliza markers manuales, Railway y GitHub Actions. Solo se
persisten éxitos de producción y `(provider, external_id)` hace idempotentes los
reintentos de webhooks.
