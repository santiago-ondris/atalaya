# Atalaya — Centro de Comando y Observabilidad

Atalaya es un centro de comando y observabilidad para monitorear aplicaciones en producción (**Farmami**, **Wheels House**, **Prensap** y **Notizap**).

Combina ingesta multifuente (Sentry & Azure Application Insights), interpretación estructurada mediante LLM (OpenRouter), alertas inteligentes por Telegram, diagramas de arquitectura interactivos, reportes diarios de sesiones/aperturas, seguimiento de incidentes y deploys, observabilidad de costos, status page pública y meta-observabilidad sobre el propio sistema de monitoreo.

El Command Center v2 incorpora en desktop un faro 3D como navegación principal:
cinco ventanas llevan a las fichas operativas y cuatro protagonistas marítimos a
Eventos, Bitácora, Reportes y Estado del sistema. La command palette (`⌘K`/`Ctrl+K`)
y la vista clásica ofrecen caminos directos que no dependen del Canvas.

---

## Estructura del Monorepo

```text
atalaya/
├── apps/
│   ├── watchman/         # Go: Núcleo, ingesta, scheduler, API REST y automatizaciones
│   ├── interpreter/      # Python FastAPI: Interpretación de errores mediante LLM (OpenRouter)
│   └── command-center/   # React + TypeScript: Interfaz de operación y dashboard
├── deploy/
│   └── scripts/          # Scripts de respaldos (backup.sh, restore.sh)
├── docs/
│   ├── adr/              # Architecture Decision Records (ADRs 0001 al 0011)
│   ├── runbooks/         # Runbook operativo de producción
│   └── sprints/          # Documentación técnica y walkthroughs de los Sprints 0 al 11
└── compose.yaml          # Entorno dockerizado reproducible
```

---

## Requisitos y Toolchain

- **Go**: 1.26.5
- **Python**: 3.13.14 (administrado con `uv` 0.12.2)
- **Node.js**: 24.19.0 LTS
- **PostgreSQL**: 18.4
- **Docker & Docker Compose**

---

## Desarrollo Local

```bash
make up
```

El comando construye y levanta PostgreSQL, aplica migraciones con Goose y arranca los 3 servicios:

- **Command Center**: [http://localhost:5173](http://localhost:5173)
- **Watchman API**: [http://localhost:8080/health](http://localhost:8080/health)
- **Interpreter API**: [http://localhost:8000/health](http://localhost:8000/health)

Para detener el entorno:

```bash
make down
```

---

## Características Principales

1. **Ingesta Multifuente**:
   - Polling cada 2 minutos a **Sentry** (Farmami, Wheels House, Prensap) y **Azure Application Insights** (Notizap).
   - Sanitización de datos sensibles y checkpoints independientes con deduplicación e idempotencia.

2. **Interpretación Inteligente (LLM / OpenRouter)**:
   - Los eventos pasan por la cola durable de PostgreSQL y son procesados por el Interpreter.
   - Generación automática de resumen en español, explicación técnica, severidad, actionabilidad y sugerencias de resolución.

3. **Control y Presupuesto de Costos LLM**:
   - Monitoreo en tiempo real de tokens consumidos (prompt/completion), latencia y costo en USD acumulado por aplicación y por modelo.
   - Presupuesto mensual configurable (`LLM_MONTHLY_BUDGET_USD`, por defecto `$5.00 USD`).
   - Alertas por Telegram al alcanzar el 80% o 100% del presupuesto mensual.

4. **Alertas Inteligentes por Telegram**:
   - Envíos agrupados por aplicación y fingerprint con ventana de deduplicación y rate limiting anti-ruido.

5. **Reportes Diarios de Actividad**:
   - Generación automática a las 20:00 ARG (`America/Argentina/Buenos_Aires`) combinando volumen de excepciones con conteo real de sesiones/aperturas.

6. **Incidentes y Deploy Markers**:
   - Bitácora interactiva que superpone despliegues exitosos (Railway, GitHub Actions, Vercel, Cloudflare, Azure) con picos de errores.
   - Agrupación de errores en incidentes auditables con notas, conclusiones y cambios de estado.

7. **Status Page Pública y Meta-Observabilidad**:
   - Página `/status` pública sin información sensible.
   - Meta-observabilidad interna que detecta pollers detenidos, backlog creciente, degradación del Interpreter y dead man's switch con Healthchecks.io.

8. **Hardening, Retención y Respaldos**:
   - Limpieza automática en segundo plano de eventos mayores a `EVENT_RETENTION_DAYS` (90 días por defecto).
   - Cabeceras de seguridad HTTP, rate limiters, validación estricta de esquemas y scripts de backup/restore en `deploy/scripts/`.

9. **Experiencia Operativa v2**:
   - Faro desktop con nueve destinos, cinco fichas y clima derivado de la peor severidad.
   - Cambio de sesión entre modo inmersivo y clásico; bajo 901 px o sin puntero fino se usa temporalmente la vista clásica.
   - Calidad 3D adaptativa y retorno seguro al modo clásico ante bajo rendimiento, WebGL ausente, error de carga/render o pérdida irrecuperable del contexto.
   - Diagnóstico reproducible con `npm run build:report`; detalles operativos y de recuperación en `docs/runbooks/operational-runbook.md`.

El Corte 1 desktop está cerrado. La evidencia se encuentra en
`docs/sprints/sprint-v2.7-corte-1-closure-walkthrough.md`; Firefox real y la
automatización Playwright cross-browser permanecen como deuda conocida.

---

## Licencia y Uso

Proyecto personal para monitoreo de aplicaciones en producción. Desarrollado con estándares de ingeniería de software y portfolio técnico profesional.
