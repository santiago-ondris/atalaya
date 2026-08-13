# ADR 0011 — Costos de LLM, retención de datos, hardening de seguridad y preparación V1

## Contexto
En la versión productiva de Atalaya, es crítico controlar el costo de consumo de LLM en OpenRouter, evitar fugas de datos sensibles (PII/Secretos), garantizar retención de datos sustentable en PostgreSQL y proteger los endpoints HTTP expuestos.

## Decisiones

1. **Gestión de Presupuesto y Observabilidad de Costos LLM**:
   - Se establece un presupuesto mensual por defecto de `$5.00 USD` (`LLM_MONTHLY_BUDGET_USD=5.00`).
   - Watchman agrega el consumo en tiempo real (`input_tokens`, `output_tokens`, `total_tokens`, `estimated_cost_usd`, latencia) agrupado por aplicación y por modelo.
   - Si el gasto del mes actual alcanza o supera el 80% o 100% del presupuesto, `BudgetMonitor` encola automáticamente una alerta por Telegram con un cooldown de 24 horas.
   - El Command Center visualiza la barra de progreso del presupuesto y tablas de desglose en `/system`.

2. **Retención y Purga Automática de Datos**:
   - Se establece una política de retención de `90 días` (`EVENT_RETENTION_DAYS=90`).
   - Un worker en segundo plano en Watchman ejecuta una purga diaria en PostgreSQL eliminando eventos y ocurrencias más antiguas que la retención configurada, protegiendo aquellos eventos vinculados a incidentes activos en `incident_groups`.

3. **Hardening de Seguridad API y Sanitización**:
   - **Sanitización de Datos**: El Interpreter (Python) sanitiza mediante expresiones regulares tokens Bearer, passwords, API keys, connection strings y secretos en mensajes y stack traces antes de enviar peticiones a OpenRouter.
   - **Hardening HTTP**: Watchman incorpora middleware de `MaxBytesReader` (límite de payload), rate limiting en sesión y cabeceras de seguridad estrictas (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`).

4. **Optimización de PostgreSQL**:
   - Se aplica la migración `00008_performance_indexes_and_retention.sql` añadiendo índices compuestos sobre `events`, `interpretations`, `deployments`, `error_groups` y `job_queue`.

5. **Runbooks y Respaldos**:
   - Se crean scripts reproducibles `deploy/scripts/backup.sh` y `restore.sh` y el manual operativo `docs/runbooks/operational-runbook.md`.

## Consecuencia
Atalaya V1 queda completamente estabilizada, segura, costo-eficiente y documentada para producción.
