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

| Variable                      | Descripción                            | Ejemplo / Default                   |
| ----------------------------- | -------------------------------------- | ----------------------------------- |
| `DATABASE_URL`                | String de conexión a PostgreSQL        | `postgres://user:pass@host:5432/db` |
| `ATALAYA_ADMIN_PASSWORD_HASH` | Hash bcrypt de la contraseña de admin  | `$2a$10$...`                        |
| `OPENROUTER_API_KEY`          | Key de OpenRouter en Interpreter       | `sk-or-v1-...`                      |
| `TELEGRAM_BOT_TOKEN`          | Token del bot de alertas de Telegram   | `123456:ABC-DEF...`                 |
| `TELEGRAM_CHAT_ID`            | ID del chat privado o grupo de alertas | `-10012345678`                      |
| `LLM_MONTHLY_BUDGET_USD`      | Presupuesto mensual de LLM (USD)       | `5.00`                              |
| `EVENT_RETENTION_DAYS`        | Días de retención de eventos           | `90`                                |
| `HEALTHCHECKS_PING_URL`       | Dead man's switch heartbeat URL        | `https://hc-ping.com/...`           |

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

### 4.4 Faro v2, rendimiento y recuperación

1. Si el faro no carga, usar “Abrir vista clásica”. La command palette
   (`⌘K`/`Ctrl+K`) permanece disponible porque no depende del Canvas.
2. Revisar el aviso del modo clásico: distingue WebGL no disponible, rendimiento
   insuficiente, error de carga/render y pérdida de contexto.
3. Usar “Reintentar faro” sólo después de resolver o descartar la causa. La
   aplicación evita reintentos automáticos para no alternar indefinidamente entre
   modos.
4. En desarrollo, agregar `?performanceDebug=1` para observar perfil, FPS y
   ventanas restantes; `?performance=degraded` simula 45 FPS y
   `?performance=critical` simula 20 FPS.
5. Ejecutar `npm run build:report` en `apps/command-center` para confirmar que el
   stack Three.js siga en el chunk diferido `LighthouseScene` y revisar el peso
   inicial y de escena.

El perfil baja de `normal` a `reduced` después de tres ventanas bajo 50 FPS. Diez
ventanas consecutivas bajo 30 FPS en el perfil reducido abren el modo clásico.
Con movimiento reducido se congelan las animaciones continuas, pero no se eliminan
estado ni navegación.

La preferencia de modo vive en `sessionStorage` y se elimina al cerrar sesión. En
viewports menores a 901 px o sin puntero fino, el modo clásico es el resguardo
temporal hasta los cortes mobile.

---

## 5. Mantenimiento y Purga de Datos

- El worker `retention` purga automáticamente en segundo plano eventos y ocurrencias más antiguas que `EVENT_RETENTION_DAYS` (90 días por defecto), protegiendo aquellos asociados a incidentes activos.
