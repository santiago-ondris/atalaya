# ADR 0007 — Reportes diarios y métricas de actividad

## Estado

Aceptado el 2026-08-10.

## Contexto

Atalaya debe enviar a las 20:00 de Argentina un resumen del día con errores,
ocurrencias, severidades y actividad de las cuatro aplicaciones. El envío debe
resistir reinicios, no duplicarse y no continuar al día siguiente.

## Decisión

- Watchman calcula los límites del día con `America/Argentina/Buenos_Aires`; no
  resta 24 horas ni depende de la zona horaria del contenedor.
- Sentry Release Health aporta `sum(session)` para Farmami, Wheels House y
  Prensap.
- Notizap usa sesiones activas, definidas como `dcountif(SessionId,
  isnotempty(SessionId))` sobre `AppPageViews`. Las page views se calculan con
  `sum(ItemCount)` para respetar sampling, pero no sustituyen silenciosamente a
  las sesiones.
- Una apertura en el reporte significa una sesión con al menos una page view
  durante el día calendario; no representa usuarios únicos ni cada montaje de
  la SPA.
- `daily_reports` es la identidad durable e idempotente por fecha local.
  `daily_report_applications` congela la fotografía utilizada para el mensaje y
  el historial. Los intentos de Telegram quedan auditados por separado.
- Un fallo reintentable se programa cada 30 minutos. Al alcanzar el siguiente
  día local, el reporte pasa a `expired` y nunca se vuelve a enviar.
- Si una fuente de actividad falla, el reporte conserva la métrica como no
  disponible y continúa incluyendo los errores persistidos. La ausencia de
  telemetría no se presenta como cero.

## Consecuencias

Los reportes son reproducibles y observables aun con proveedores degradados. La
entrega sigue siendo al menos una vez: Telegram no ofrece una clave de
idempotencia y una caída después de aceptar el mensaje, pero antes del commit,
puede producir un duplicado excepcional.
