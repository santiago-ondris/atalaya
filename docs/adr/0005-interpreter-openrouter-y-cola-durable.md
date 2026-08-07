# ADR 0005 — Interpreter con OpenRouter y cola durable en PostgreSQL

- Estado: aceptado
- Fecha: 2026-08-07

## Contexto

La interpretación depende de un proveedor externo que puede responder lento, aplicar rate
limits o devolver contenido inválido. El evento importado no puede perderse ni mantener
abierta la transacción del poller mientras espera al LLM.

## Decisión

Watchman crea un `interpretation_job` en la misma transacción que persiste cada evento.
Workers Go reclaman jobs mediante `FOR UPDATE SKIP LOCKED`, llaman al Interpreter y son los
únicos que persisten el resultado. Un lease de cinco minutos permite recuperar trabajos que
quedaron en proceso durante un reinicio.

Interpreter usa OpenRouter detrás de un cliente propio, solicita JSON Schema estricto y
valida nuevamente la respuesta con Pydantic. El prompt tiene versión explícita y el stack
trace se trunca a un límite configurable antes de abandonar el servicio.
La salida se limita inicialmente a 900 tokens para acotar costo y evitar que OpenRouter
reserve saldo contra el máximo teórico del modelo.

Los fallos de red, timeout, HTTP 408/429 y 5xx son transitorios. Watchman los reprograma con
backoff exponencial hasta `max_attempts`. Payloads inválidos y rechazos 4xx son permanentes.

El modelo se configura por entorno. El costo se toma de `usage.cost` cuando OpenRouter lo
informa; no se inventa una estimación cuando el proveedor no devuelve ese dato.

## Consecuencias

- Reiniciar Watchman o Interpreter no pierde eventos ni jobs.
- No se necesita un broker adicional para el volumen inicial.
- PostgreSQL concentra estado y coordinación; habrá que observar el backlog y los leases.
- Cambiar de modelo no exige modificar el contrato ni recompilar los servicios.
