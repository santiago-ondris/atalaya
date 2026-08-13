# Sprint 11 — Costos, seguridad y preparación de producción (V1 Release)

## Resultado

Atalaya alcanza la versión **V1 Release**, completando la observabilidad de costos de LLM en OpenRouter con control presupuestario, purga periódica de datos (retención de 90 días), hardening de seguridad HTTP, sanitización de secretos en Interpreter, índices de rendimiento PostgreSQL, scripts de backup/restore y el manual operativo para producción.

## Recorrido técnico

- **Control y Observabilidad de Costos (LLM / OpenRouter)**:
  - Implementación de `GetCostSummary` en `store/postgres.go` agregando gasto ($ USD), tokens (prompt y completion), conteo de peticiones y latencia promedio.
  - Implementación de `costs.BudgetMonitor` en Watchman con presupuesto mensual configurable (`LLM_MONTHLY_BUDGET_USD=5.00`).
  - Alertas automáticas por Telegram al superar el 80% o 100% del presupuesto.
  - Componente `CostCenter.tsx` en el Command Center (`/system`) con barra de avance presupuestario y desgloses tabulares por Aplicación y Modelo de LLM.
- **Retención y Purga de Datos**:
  - `PurgeOldEvents` en `store/postgres.go` y worker `retention.Worker` en Go para limpiar eventos y ocurrencias de más de `EVENT_RETENTION_DAYS=90` días en segundo plano, omitiendo eventos vinculados a incidentes.
- **Sanitización & Hardening de Seguridad**:
  - Expresiones regulares en `prompt.py` del Interpreter para sanitizar tokens Bearer, contraseñas, API keys y connection strings antes de consultar a OpenRouter.
  - Cabeceras de seguridad HTTP (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`) y middleware `MaxBytesReader` en Watchman.
- **Optimización de PostgreSQL**:
  - Migración `00008_performance_indexes_and_retention.sql` añadiendo índices compuestos sobre `events`, `interpretations`, `deployments`, `error_groups` y `job_queue`.
- **Infraestructura & Documentación**:
  - Scripts de backup y restore: `deploy/scripts/backup.sh` y `restore.sh`.
  - Manual operativo de producción: `docs/runbooks/operational-runbook.md`.
  - Registro de arquitectura: `docs/adr/0011-costos-seguridad-y-preparacion-produccion.md`.
  - Actualización completa del `README.md` principal.
