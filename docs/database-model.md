# Modelo de datos inicial

## Relaciones

```text
applications
    └── integrations
          ├── source_checkpoints
          └── error_groups
                ├── error_events
                │     └── interpretation_jobs
                │           └── interpretations
                └── alert_windows
                      └── notification_jobs
                            └── notification_delivery_attempts
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
