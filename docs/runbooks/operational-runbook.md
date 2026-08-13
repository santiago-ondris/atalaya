# Atalaya — Runbook Operativo de Producción

Este manual contiene los procedimientos operativos para la administración, mantenimiento, diagnóstico y recuperación de **Atalaya** en producción.

---

## 1. Arquitectura y Componentes Desplegados

- **Watchman (Go API & Workers)**: Deployed en Railway. Administra polling, jobs durables, alertas de Telegram, reportes diarios y la API REST del Command Center.
- **Interpreter (Python FastAPI)**: Deployed en Railway. Recibe peticiones de Go, sanitiza stack traces y consulta a OpenRouter.
- **Command Center (React + TypeScript)**: Deployed en Cloudflare Pages / Railway local.
- **PostgreSQL 18**: Provisto en Railway. Fuente de verdad principal.
- **Status Page Pública**: Deployed en Cloudflare Pages.

---

## 2. Variables de Entorno Críticas

| Variable | Descripción | Ejemplo / Default |
|---|---|---|
| `DATABASE_URL` | String de conexión a PostgreSQL | `postgres://user:pass@host:5432/db` |
| `ATALAYA_ADMIN_PASSWORD_HASH` | Hash bcrypt de la contraseña de admin | `$2a$10$...` |
| `OPENROUTER_API_KEY` | Key de OpenRouter en Interpreter | `sk-or-v1-...` |
| `TELEGRAM_BOT_TOKEN` | Token del bot de alertas de Telegram | `123456:ABC-DEF...` |
| `TELEGRAM_CHAT_ID` | ID del chat privado o grupo de alertas | `-10012345678` |
| `LLM_MONTHLY_BUDGET_USD` | Presupuesto mensual de LLM (USD) | `5.00` |
| `EVENT_RETENTION_DAYS` | Días de retención de eventos | `90` |
| `HEALTHCHECKS_PING_URL` | Dead man's switch heartbeat URL | `https://hc-ping.com/...` |

---

## 3. Respaldos y Restauración de Base de Datos

### Backup
```bash
export DATABASE_URL="postgres://..."
./deploy/scripts/backup.sh
```

### Restore
```bash
export DATABASE_URL="postgres://..."
./deploy/scripts/restore.sh ./backups/atalaya_db_20260812_200000.sql.gz
```

---

## 4. Diagnóstico de Problemas Frecuentes

### 4.1 Polling Detenido o Fallos de Sentry/Azure
1. Verificar en Command Center (`/system`) el estado de los pollers y checkpoints.
2. Revisar logs de Watchman en Railway buscando errores `failed to poll Sentry` o `Azure Monitor`.
3. Verificar la validez de los tokens `SENTRY_AUTH_TOKEN` o credenciales `AZURE_CLIENT_SECRET`.

### 4.2 Alerta de Presupuesto LLM / Fallo en Interpreter
1. Revisar `/system` en Command Center la barra de presupuesto consumido.
2. Si el Interpreter no responde, verificar `INTERPRETER_URL` y la validez de `OPENROUTER_API_KEY`.
3. Si los tokens se agotan, ajustar `LLM_MONTHLY_BUDGET_USD` o cambiar el modelo en `OPENROUTER_MODEL`.

### 4.3 Reinicios de Servicio y Resiliencia
- La cola de trabajos (`job_queue`) es durable en PostgreSQL.
- Si Watchman o el Interpreter se reinician, los trabajos en estado `pending` o `retrying` se retoman automáticamente sin pérdida de eventos ni alertas duplicadas.

---

## 5. Mantenimiento y Purga de Datos
- El worker `retention` purga automáticamente en segundo plano eventos y ocurrencias más antiguas que `EVENT_RETENTION_DAYS` (90 días por defecto), protegiendo aquellos asociados a incidentes activos.
