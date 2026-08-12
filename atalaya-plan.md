# Atalaya — Plan de arquitectura y desarrollo
 
## 1. Contexto y decisiones confirmadas
 
Atalaya será una herramienta de uso personal para observar aplicaciones reales en producción pertenecientes a clientes. Aunque inicialmente tendrá un único usuario (yo mismo), se construirá con estándares de producto: autenticación, persistencia confiable, trazabilidad, manejo de fallos y despliegues reproducibles.
 
### Aplicaciones monitoreadas
 
| Aplicación | Fuente de observabilidad | Deploy actual |
|---|---|---|
| Farmami | Sentry | Push a GitHub: backend en Railway y frontend en Vercel |
| Wheels House | Sentry | Push a GitHub: backend en Railway y frontend en Vercel |
| Prensap | Sentry | Push a GitHub: backend en Railway y frontend en Cloudflare Pages |
| Notizap | Application Insights | Frontend mediante GitHub Actions a Azure Web App; backend mediante `Deploy to Web App` desde VS Code |
 
### Decisiones confirmadas
 
- PostgreSQL será la fuente de verdad.
- PostgreSQL 18.4 será la versión inicial de desarrollo; se mantendrá fija la
  versión mayor y las actualizaciones menores se validarán mediante PR.
- El toolchain local inicial queda fijado en Go 1.26.5, Python 3.13.14
  administrado con `uv` 0.12.2 y Node.js 24.19.0 LTS.
- La frecuencia inicial de polling será de 2 minutos.
- Prensap será la primera integración utilizada para construir el flujo vertical.
- Hay acceso administrativo a Sentry, Application Insights, Telegram, OpenRouter y las plataformas de deploy.
- Go, Python y el entorno local serán dockerizados.
- No se incorporará Turborepo al inicio: cada aplicación tendrá su toolchain y
  despliegue independiente. La decisión se revisará si aparecen paquetes
  TypeScript compartidos o tareas cross-project que lo justifiquen.
- El detalle de los health endpoints de las cuatro aplicaciones queda pendiente de inventario.
- El diagrama interactivo y el conteo diario de aperturas continúan confirmados; su implementación específica se definirá en sus respectivos sprints.
---
 
## 2. Recomendación de organización: monorepo
 
Se recomienda un **monorepo**, manteniendo servicios claramente separados dentro de él.
 
Atalaya es un solo producto desplegado como varias unidades. Go, Python y React compartirán contratos, configuración de desarrollo, migraciones, documentación y pruebas end-to-end. Tenerlos en un mismo repositorio permite realizar un cambio transversal en una única rama y validarlo con un solo pipeline.
 
Esto no convierte el sistema en un monolito: cada servicio conserva su propio runtime, dependencias, Dockerfile, tests y despliegue independiente.
 
### Ventajas para este proyecto
 
- Un solo lugar para comprender y mostrar el sistema completo.
- Cambios coordinados entre la API de Go, el contrato del interpreter y el frontend.
- Un único `docker compose up` para levantar el entorno local.
- CI centralizado con checks específicos según las carpetas modificadas.
- Menos mantenimiento operativo para un único desarrollador.
- Mejor presentación como proyecto de portfolio, sin sacrificar separación técnica.
### Estructura inicial propuesta
 
```text
atalaya/
├── apps/
│   ├── watchman/              # Go: núcleo, ingesta, API y automatizaciones
│   ├── interpreter/           # Python: interpretación mediante LLM
│   └── command-center/        # React + TypeScript
├── contracts/
│   ├── openapi/               # Contratos HTTP públicos e internos
│   └── examples/              # Payloads representativos y fixtures
├── database/
│   └── migrations/            # Migraciones PostgreSQL
├── deploy/
│   ├── docker/                # Configuración auxiliar de imágenes
│   └── railway/               # Notas/configuración de despliegue
├── docs/                      # ADRs, diagramas y documentación operativa
├── compose.yaml
├── .env.example
└── README.md
```
 
La estructura podrá ajustarse al crear el repositorio base, pero la separación conceptual debe mantenerse.
 
---
 
## 3. Responsabilidad de cada componente
 
### 3.1 Watchman — Go
 
Go será el **dueño del flujo de negocio y de los datos**. Será el único componente que escriba en PostgreSQL durante la primera versión.
 
Responsabilidades:
 
- Ejecutar el scheduler de polling cada 2 minutos.
- Consultar Sentry y Application Insights mediante adapters.
- Normalizar eventos de distintas fuentes al modelo común de Atalaya.
- Garantizar idempotencia para no importar dos veces el mismo evento externo.
- Persistir eventos, ocurrencias, interpretaciones, incidentes, reportes y costos.
- Mantener una cola durable de trabajos en PostgreSQL.
- Solicitar interpretaciones al servicio Python.
- Aplicar deduplicación, agrupamiento, severidad y rate limiting.
- Enviar alertas y reportes mediante Telegram.
- Reintentar operaciones transitorias sin perder trabajos.
- Exponer la API HTTP consumida por el command center.
- Resolver autenticación y autorización del usuario único.
- Proveer health/readiness endpoints y métricas sobre la propia Atalaya.
- Servir o exponer la configuración de los diagramas de arquitectura.
#### Interfaces de fuentes
 
El dominio no deberá depender directamente de Sentry ni de Azure. Una interfaz conceptual inicial será:
 
```go
type ErrorSource interface {
    FetchEvents(ctx context.Context, cursor Cursor) (EventBatch, error)
    FetchActivity(ctx context.Context, period Period) (ActivitySummary, error)
}
```
 
Los adapters iniciales serán:
 
- `SentrySource`, reutilizado por Farmami, Wheels House y Prensap.
- `ApplicationInsightsSource`, utilizado por Notizap mediante consultas KQL.
La forma final de la interfaz se decidirá al implementar el primer adapter; este ejemplo expresa el límite arquitectónico, no un contrato cerrado.
 
### 3.2 Interpreter — Python
 
Python será un servicio interno, pequeño y preferentemente stateless. Su especialidad será convertir información técnica estructurada en una interpretación útil mediante OpenRouter.
 
Responsabilidades:
 
- Recibir un evento normalizado, con datos sensibles previamente filtrados.
- Construir y versionar prompts.
- Consultar OpenRouter con timeouts y manejo explícito de errores.
- Validar la respuesta del modelo contra un esquema estricto.
- Devolver resumen, explicación, severidad, actionabilidad y acciones sugeridas.
- Reportar modelo, tokens consumidos, latencia y costo estimado.
- Permitir reemplazar el proveedor o modelo sin alterar el dominio de Go.
- Proveer health/readiness endpoints.
No será responsabilidad de Python:
 
- Pollear Sentry o Application Insights.
- Acceder directamente a PostgreSQL.
- Decidir si corresponde enviar una alerta.
- Enviar mensajes de Telegram.
- Ser la fuente de verdad de costos o interpretaciones.
### 3.3 Command Center — React + TypeScript
 
Responsabilidades:
 
- Mostrar el estado general de las cuatro aplicaciones.
- Explorar eventos, ocurrencias, interpretaciones e incidentes.
- Permitir marcar incidentes como resueltos y guardar notas.
- Mostrar actividad diaria, tendencias, deploy markers y costos de OpenRouter.
- Renderizar los diagramas interactivos y sus flujos.
- Mostrar el estado interno de Atalaya y sus integraciones.
- Proveer la vista pública de status sin exponer datos privados.
El navegador solo se comunicará con la API de Go. Nunca accederá directamente a Python, Sentry, Azure u OpenRouter.
 
---
 
## 4. Flujo principal del sistema
 
```text
Sentry / Application Insights
            │
            ▼
     Poller y adapter Go
            │
            ▼
 Normalización + idempotencia
            │
            ▼
        PostgreSQL
            │
            ▼
 Cola durable de interpretación
            │
            ▼
 Interpreter Python → OpenRouter
            │
            ▼
 Resultado persistido por Go
            │
            ▼
 Deduplicación + política de alerta
            │
      ┌─────┴─────┐
      ▼           ▼
  Telegram    Command Center
```
 
### Cola inicial
 
No se incorporará RabbitMQ, Kafka ni Redis al comienzo. Una tabla de jobs en PostgreSQL permitirá:
 
- conservar trabajos ante reinicios;
- reintentar interpretaciones fallidas;
- observar su estado desde el command center;
- evitar infraestructura innecesaria para el volumen inicial.
Si las métricas reales demuestran que PostgreSQL ya no alcanza, se registrará una decisión arquitectónica para introducir otra tecnología.
 
---
 
## 5. Estrategia Docker
 
Docker se utilizará tanto como herramienta de aprendizaje como para lograr paridad entre desarrollo y producción.
 
### Entorno local
 
`compose.yaml` levantará inicialmente:
 
- `postgres`
- `watchman`
- `interpreter`
- `command-center`
Se agregarán health checks y dependencias condicionadas por salud, sin depender únicamente del orden de arranque.
 
### Imágenes
 
- Cada aplicación tendrá un Dockerfile multi-stage propio.
- Los contenedores de producción se ejecutarán como usuarios no root.
- Las imágenes contendrán solo las dependencias necesarias para runtime.
- Configuración y secretos se inyectarán mediante variables de entorno.
- `.env.example` documentará nombres y valores seguros de ejemplo; ningún secreto real entrará al repositorio.
### Producción
 
- Railway desplegará Watchman e Interpreter como servicios independientes desde el mismo repositorio.
- PostgreSQL podrá ser provisto por Railway.
- El command center se desplegará en Cloudflare Pages.
- `compose.yaml` será principalmente la experiencia local y de pruebas integrales, no el mecanismo obligatorio de producción.
---
 
## 6. Forma de trabajo por sprint
 
Los sprints serán **orientados a resultados demostrables**, no a acumular componentes incompletos. 
 
### Inicio de cada sprint
 
1. Confirmar objetivo y criterios de aceptación.
2. Revisar decisiones pendientes que realmente bloqueen ese objetivo.
3. Dividir el trabajo en issues o tareas pequeñas.
4. Identificar riesgos, secretos o accesos necesarios.
### Durante el sprint

1. Trabajar en ramas cortas por feature o corrección.
2. El asistente implementa y valida los cambios, explicando en cada avance el
   propósito de las decisiones, librerías y archivos incorporados.
3. Mantener prolija la estructura del proyecto y preferir archivos enfocados antes
   que archivos excesivamente grandes.
4. Abrir un PR aunque exista un único desarrollador, para conservar contexto y evidencia.
5. Ejecutar lint, tests y build automáticamente.
6. Registrar decisiones importantes como ADRs breves en `docs/adr/`.
### Cierre de cada sprint
 
1. Ejecutar una demo con datos reales o fixtures representativos.
2. Validar todos los criterios de aceptación.
3. Actualizar documentación y variables de entorno.
4. Registrar deuda técnica descubierta sin esconderla dentro del sprint.
5. Hacer una retrospectiva breve: qué funcionó, qué trabó y qué cambia en el siguiente sprint.
### Definition of Done general
 
Una funcionalidad se considera terminada cuando:
 
- cumple sus criterios de aceptación;
- tiene tests proporcionales al riesgo;
- pasa lint, typecheck y build;
- maneja y registra sus errores relevantes;
- no expone secretos ni datos personales innecesarios;
- tiene documentación operativa suficiente;
- funciona mediante Docker en el entorno local;
- puede demostrarse desde la interfaz o mediante una prueba reproducible.
---
 
## 7. Roadmap completo por sprints
 
### Sprint 0 — Fundación y contratos

**Estado:** completado el 2026-08-06.
 
**Objetivo:** disponer de un repositorio ejecutable, verificable y preparado para crecer.
 
Alcance:
 
- Crear la estructura del monorepo.
- Inicializar Go, Python y React + TypeScript.
- Crear Dockerfiles y `compose.yaml` con PostgreSQL.
- Definir manejo de configuración y `.env.example`.
- Incorporar migraciones de base de datos.
- Definir logging estructurado y correlation IDs.
- Crear health/readiness endpoints en Go y Python.
- Configurar CI para lint, tests y builds.
- Escribir el primer ADR: monorepo y límites de servicios.
- Crear fixtures seguros de eventos Sentry para desarrollo.
**Demo:** un comando levanta todo el sistema; frontend, Go y Python reportan salud y Go se conecta a PostgreSQL.
 
 SPRINT 0 LISTO - INFORMACION EN DOCS/SPRINTS/SPRINT-00-FOUNDATION-WALKTHROUGH.md

### Sprint 1 — Primer evento real de Prensap

**Estado:** completado el 2026-08-07, incluida la prueba controlada con Sentry real.
 
**Objetivo:** importar de manera confiable errores reales de Prensap desde Sentry.
 
Alcance:
 
- Implementar cliente de Sentry.
- Implementar el primer `SentrySource`.
- Crear el modelo común de evento y ocurrencia.
- Persistir cursor/checkpoint del poller.
- Garantizar idempotencia frente al mismo evento externo.
- Ejecutar polling cada 2 minutos.
- Sanitizar tokens, headers y posibles datos sensibles.
- Exponer endpoints internos de consulta básica.
- Agregar pruebas con fixtures y una prueba controlada contra Sentry.
**Demo:** provocar o seleccionar un error de Prensap, importarlo una sola vez y verlo persistido con sus ocurrencias.

SPRINT 1 LISTO - INFORMACION EN DOCS/SPRINTS/SPRINT-01-PRENSAP-SENTRY-WALKTHROUGH.md
 
### Sprint 2 — Interpretación con OpenRouter

**Estado:** completado el 2026-08-07 con un evento real de Prensap y OpenRouter.
 
**Objetivo:** transformar un evento técnico en una explicación estructurada y útil.
 
Alcance:
 
- Definir contrato versionado entre Go y Python.
- Implementar schemas estrictos en Python.
- Integrar OpenRouter.
- Crear prompt inicial y política de truncado de stack traces.
- Devolver resumen, explicación, severidad, actionabilidad y acciones sugeridas.
- Registrar tokens, modelo, latencia y costo estimado.
- Implementar timeouts, retries acotados y clasificación de fallos.
- Crear la tabla de jobs durables en PostgreSQL.
- Hacer que Go persista los resultados.
**Demo:** un error de Prensap pasa automáticamente por el interpreter y queda almacenado con explicación y costo.

SPRINT 2 LISTO - INFORMACION EN docs/sprints/sprint-02-openrouter-walkthrough.md
 
### Sprint 3 — Telegram, deduplicación y resiliencia

**Estado:** completado el 2026-08-07 con entrega controlada al chat privado de Telegram.
 
**Objetivo:** convertir el pipeline en una herramienta operativa que alerte sin generar ruido.
 
Alcance:
 
- Integrar el bot de Telegram.
- Definir fingerprint inicial: aplicación, fuente, tipo y ubicación relevante.
- Implementar ventana de deduplicación y contador de ocurrencias.
- Definir política inicial de severidad y actionabilidad.
- Implementar rate limiting por aplicación y fingerprint.
- Enviar mensajes claros con vínculo al evento en Atalaya/Sentry.
- Reintentar envíos transitorios mediante jobs durables.
- Registrar cada intento y resultado de entrega.
- Alertar sobre fallos sostenidos del interpreter sin crear loops.
**Demo:** múltiples ocurrencias equivalentes generan una alerta agrupada, con contador, interpretación y trazabilidad.
 
**Hito:** al finalizar este sprint existe un MVP operativo real para Prensap.

SPRINT 3 LISTO - INFORMACION EN docs/sprints/sprint-03-telegram-walkthrough.md
 
### Sprint 4 — Las tres aplicaciones Sentry

**Estado:** completado el 2026-08-10. La validación final ocurrió con eventos reales:
Atalaya los interpretó correctamente y entregó las alertas en lenguaje natural por
Telegram.
 
**Objetivo:** monitorear Farmami, Wheels House y Prensap reutilizando el mismo adapter.
 
Alcance:
 
- Modelar aplicaciones e integraciones configurables.
- Incorporar proyectos de Farmami y Wheels House. (pertenecen a la misma organizacion en Sentry)
- Mantener checkpoints independientes por integración.
- Aislar fallos para que una fuente caída no bloquee las demás.
- Añadir filtros y políticas por aplicación.
- Incorporar métricas de última ejecución y último éxito de cada poller.
**Demo:** las tres aplicaciones son polleadas independientemente y producen eventos normalizados comparables.

SPRINT 4 LISTO - INFORMACION EN docs/sprints/sprint-04-multi-sentry-walkthrough.md
 
### Sprint 5 — Notizap y Application Insights

**Estado:** completado el 2026-08-10. Autenticación, consulta, checkpoint y flujo
de una excepción real fueron validados con alcance limitado al recurso
`notizap-insights`.
 
**Objetivo:** integrar Notizap sin contaminar el dominio común con detalles de Azure/KQL.
 
Alcance:
 
- Implementar autenticación contra Azure Monitor/Application Insights.
- Diseñar consultas KQL para excepciones y eventos relevantes.
- Implementar `ApplicationInsightsSource`.
- Mapear resultados al mismo modelo de evento y ocurrencia.
- Gestionar checkpoint temporal y solapamiento seguro de ventanas.
- Validar idempotencia y diferencias semánticas respecto de Sentry.
- Documentar permisos y consultas necesarias.
**Demo:** un error de Notizap recorre el mismo pipeline de persistencia, interpretación y alerta que uno de Sentry.

SPRINT 5 LISTO - INFORMACION EN docs/sprints/sprint-05-application-insights-walkthrough.md

 
### Sprint 6 — Command center privado

**Estado:** completado el 2026-08-10. Ver `docs/sprints/sprint-06-command-center-walkthrough.md`.
 
**Objetivo:** disponer de una interfaz segura para operar el sistema diariamente.
 
Alcance:
 
- Implementar login de usuario único con sesión segura.
- Crear layout y design tokens basados en la identidad náutica definida.
- Crear overview de las cuatro aplicaciones.
- Listar y filtrar eventos por aplicación, severidad, estado y período.
- Mostrar detalle técnico, interpretación y ocurrencias.
- Mostrar estado de pollers, interpreter y Telegram.
- Implementar paginación y estados de carga/error.
- Crear una API pública versionada en Go para el frontend.
**Demo:** iniciar sesión, ver el estado general y navegar desde una aplicación hasta el detalle de un error real.
 
### Sprint 7 — Reportes diarios y aperturas

**Estado:** completado el 2026-08-10. Ver `docs/sprints/sprint-07-daily-reports-walkthrough.md`.
 
**Objetivo:** recibir a las 20:00 ARG un resumen diario confiable y accionable.
 
Alcance:
 
- Implementar scheduler con zona horaria `America/Argentina/Buenos_Aires`.
- Agregar métricas diarias por aplicación.
- Investigar y definir la fuente de aperturas para cada stack.
- Incorporar sessions/page views o instrumentación adicional según corresponda.
- Generar el reporte con errores, ocurrencias, severidades y aperturas.
- Reintentar cada 30 minutos dentro del mismo día.
- No arrastrar reportes vencidos al día siguiente.
- Garantizar idempotencia para no enviar dos veces el mismo reporte.
- Mostrar historial de reportes en el command center.
 
### Sprint 8 — Incidentes y deploy markers

**Estado:** completado el 2026-08-11. La migración 6 está aplicada en producción y
los markers se validaron de extremo a extremo con Railway, GitHub Actions, Vercel,
Azure Static Web Apps y Cloudflare Pages. El Command Center permanece local por
decisión operativa y su despliegue no forma parte del cierre de este sprint. Ver
`docs/sprints/sprint-08-incidents-deployments-walkthrough.md`.
 
**Objetivo:** relacionar errores con cambios de producción y construir memoria operativa.
 
Alcance:
 
- Permitir convertir/agrupar eventos en incidentes.
- Marcar incidentes como investigando, resueltos o ruido.
- Guardar notas, resolución y timestamps.
- Mostrar historial de incidentes.
- Crear modelo común de deployment.
- Automatizar markers desde GitHub Actions/Railway cuando sea viable.
- Diseñar un mecanismo explícito para el deploy manual del backend de Notizap.
- Correlacionar visualmente deploys con spikes de errores.
 
### Sprint 9 — Diagramas interactivos de arquitectura
 
**Objetivo:** representar y explicar visualmente los flujos reales de las cuatro aplicaciones, cinco contando a la misma Atalaya.
 
Alcance: Preguntarme en el momento.

**Estado:** completado el 2026-08-11. 
 
### Sprint 10 — Status público y meta-observabilidad
 
**Objetivo:** hacer visible la salud de las aplicaciones y detectar cuando el propio observador falla.
 
Alcance:
 
- Inventariar o crear health endpoints en las cuatro aplicaciones.
- Implementar checks desde ubicaciones y frecuencias razonables.
- Calcular uptime y estado actual.
- Crear status page pública sin información sensible.
- Mostrar incidentes públicos seleccionados.
- Implementar heartbeats de pollers y jobs.
- Detectar polling detenido, backlog creciente e interpreter degradado.
- Definir un canal de alerta que reduzca el riesgo de fallo silencioso.
**Demo:** simular una aplicación caída y un poller detenido; ambos casos se reflejan y alertan de forma diferenciada.
 
### Sprint 11 — Costos, seguridad y preparación de producción
 
**Objetivo:** cerrar la primera versión completa con controles operativos y presentación profesional.
 
Alcance:
 
- Crear dashboard de tokens y costos por aplicación, período y modelo.
- Definir presupuestos y alertas de consumo.
- Revisar sanitización y retención de datos.
- Añadir límites de tamaño, timeouts y protección de endpoints.
- Ejecutar pruebas end-to-end y de recuperación ante reinicios.
- Revisar índices y consultas PostgreSQL.
- Completar backups, restore y runbooks.
- Configurar despliegues de Atalaya en Railway y Cloudflare.
- Preparar documentación técnica y recorrido de portfolio.
**Demo:** despliegue productivo reproducible, monitoreando las cuatro aplicaciones y mostrando sus controles operativos.
 
---
 
## 8. Entregas principales
 
| Entrega | Sprints | Resultado |
|---|---|---|
| Base técnica | 0 | Entorno reproducible y servicios saludables |
| MVP operativo | 1–3 | Prensap monitoreada, interpretada y alertada |
| Cobertura completa | 4–5 | Cuatro aplicaciones integradas |
| Producto operable | 6–8 | UI privada, reportes, incidentes y deploys |
| Diferenciadores | 9–10 | Arquitecturas interactivas, status y auto-observación |
| V1 productiva | 11 | Costos, seguridad, resiliencia y documentación |
 
El orden permite obtener valor real temprano sin abandonar las features de portfolio confirmadas.
 
---
 
## 9. Decisiones deliberadamente postergadas
 
Estas decisiones se tomarán cuando exista información suficiente:
 
- Fuente exacta del conteo de aperturas por aplicación.
- Modelo definitivo de nodos/aristas/flujos de los diagramas.
- Librería de diagramación del frontend.
- Ventana exacta de deduplicación y forma de resetear contadores.
- Modelo de OpenRouter y fallback inicial.
- Fuente automática de deploy markers para cada plataforma.
- Política de retención de eventos y stack traces.
- Necesidad real de Redis o un broker externo.
Cada decisión importante quedará documentada mediante un ADR.
 
---
 
## 10. Información pendiente antes de iniciar
 
No todo bloquea el Sprint 0, pero deberá resolverse en el sprint indicado:
 
- Confirmar health endpoints existentes en las cuatro aplicaciones.
- Identificar organización, proyecto y permisos API de cada integración Sentry.
- Identificar workspace, application ID y permisos de consulta de Application Insights.
- Definir el chat privado o grupo de Telegram que recibirá alertas.
- Elegir el modelo inicial de OpenRouter y un presupuesto mensual orientativo.
- Definir dominio/subdominios para command center, API y status público.
- Acordar política inicial de retención de datos.
- Definir qué campos pueden contener datos sensibles de clientes y cómo sanitizarlos.
---
