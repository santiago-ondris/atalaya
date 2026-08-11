# Sprint 07 — Reportes diarios y aperturas

## Resultado

Atalaya genera una única fotografía diaria a las 20:00 de Argentina, consulta la
actividad de las cuatro aplicaciones, agrega los errores persistidos y entrega el
resumen por Telegram. El historial se puede consultar desde la API privada y desde
la sección Reportes del Command Center.

## Fuentes de actividad

- Farmami, Wheels House y Prensap: `sum(session)` de Sentry Release Health. El
  token requiere `org:read` además de lectura de eventos.
- Notizap: sesiones activas calculadas con `dcountif(SessionId,
  isnotempty(SessionId))` en `AppPageViews`, filtrando
  `AppRoleName == "notizap-frontend"`.

La métrica representa sesiones con actividad durante el día, no usuarios únicos.
Una fuente que falla queda como no disponible; nunca se transforma en un cero
engañoso.

## Recorrido del código

- `migrations/00005_daily_reports.sql`: identidad por fecha, fotografía por
  aplicación e historial de intentos.
- `internal/reporting/reporting.go`: límites del día argentino, orquestación de
  fuentes, worker de entrega y formato Telegram. Importa `time/tzdata` para que la
  zona IANA funcione también dentro de la imagen Alpine mínima.
- `internal/sentry/client.go`: consulta Release Health por proyecto.
- `internal/applicationinsights/client.go`: consulta sesiones de Notizap mediante
  KQL sin acceder a identidades ni URLs individuales.
- `internal/store/postgres.go`: snapshot transaccional, claim concurrente,
  idempotencia, expiración y reintento de 30 minutos.
- `GET /api/v1/reports`: historial autenticado.
- `features/reports/ReportsPage.tsx`: bitácora responsive en el Command Center.

## Resiliencia

`report_date` es único. Reinicios o ticks repetidos encuentran el mismo reporte y
no crean otro envío. Los workers reclaman con `FOR UPDATE SKIP LOCKED`. Un fallo
transitorio de Telegram vuelve a estar disponible 30 minutos después; si ese
instante cae fuera del día local, el reporte vence y no se arrastra.

Telegram no ofrece claves de idempotencia: como en las alertas, existe una ventana
excepcional de entrega al menos una vez si el proceso cae después de que Telegram
acepta el mensaje y antes del commit PostgreSQL.

## Demo real del 2026-08-10

La migración 5 se aplicó sobre PostgreSQL 18.4 y Watchman generó el reporte al
arrancar después de las 20:00 ARG. Telegram aceptó el primer intento con HTTP 200 y
message ID 11. La fotografía persistida fue:

| Aplicación | Sesiones | Fuente |
|---|---:|---|
| Farmami | 178 | Sentry |
| Notizap | 1 | Application Insights |
| Prensap | 4 | Sentry |
| Wheels House | 432 | Sentry |

No hubo eventos de error dentro del período. Los ticks posteriores conservaron un
único reporte enviado y un único intento.

## Validaciones

- `go test ./...`
- `npm run lint`
- `npm run build`
- `npm run format:check`
- `git diff --check`
- Migración real con Goose hasta versión 5.
- Consulta real de Sentry Release Health con HTTP 200.
- Consulta real de `AppPageViews` con sesión y rol `notizap-frontend`.
- Entrega real de Telegram y auditoría PostgreSQL.

Queda como comprobación externa posterior al sprint desplegar los cambios ya
subidos del backend de Notizap y validar la correlación W3C completa. No afecta la
fuente de sesiones ni el reporte diario ya operativo.
