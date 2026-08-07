# ADR 0006 — Alertas Telegram y control de ruido

- Estado: aceptado
- Fecha: 2026-08-07

## Contexto

Prensap ya produce eventos normalizados e interpretaciones durables. El siguiente paso es
notificar problemas útiles sin convertir cada ocurrencia en un mensaje ni perder entregas ante
reinicios o fallos transitorios de Telegram.

## Decisión

- Una interpretación abre una alerta si es `critical`, `high`, o `medium` y accionable.
- El fingerprint normalizado del grupo identifica el problema equivalente. Incluye integración,
  entorno, tipo y ubicación relevante provista por el adapter.
- La primera ocurrencia elegible abre una ventana de 15 minutos y crea una alerta inmediata.
- Toda ocurrencia posterior del mismo grupo dentro de la ventana incrementa su contador, aunque
  esa interpretación aislada no resulte accionable.
- Al cerrar la ventana se envía una actualización sólo cuando el contador es mayor que uno.
- Se permiten hasta 10 alertas nuevas por aplicación en una ventana móvil de 10 minutos. Las
  alertas excedentes permanecen pendientes; los resúmenes y avisos internos no consumen ese cupo.
- Los envíos son jobs PostgreSQL con cinco intentos y backoff exponencial de 1, 2, 4 y 8 minutos.
- Cada llamada a Telegram queda registrada como exitosa, transitoria o permanente.
- Un job de interpretación permanentemente fallido crea un aviso `interpreter_degraded`. Su clave
  temporal limita estos avisos a uno cada 30 minutos y evita alertas recursivas.

Todos los tiempos y límites son configurables mediante variables de entorno, aunque estos valores
son los defaults operativos iniciales.

## Consecuencias

PostgreSQL continúa siendo la única cola y fuente de verdad. Un reinicio no pierde ventanas ni
entregas. La entrega es al menos una vez: si Telegram acepta un mensaje y Watchman cae antes de
confirmarlo en PostgreSQL, un reintento excepcional puede duplicarlo. Telegram Bot API no ofrece
una clave de idempotencia para `sendMessage`; el historial permite detectar ese caso.

Los eventos no elegibles permanecen consultables y podrán formar parte de los reportes diarios,
pero no generan ruido inmediato.
