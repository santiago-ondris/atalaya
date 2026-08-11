# Atalaya

Command center de observabilidad para Farmami, Wheels House, Prensapp y Notizap.

## Estructura

```text
apps/
  watchman/         Servicio Go: ingesta, normalización y alertas
  interpreter/      Servicio Python: interpretación de errores mediante LLM
  command-center/   Aplicación React + TypeScript
infra/              Configuración local y de despliegue
```

Cada aplicación es una unidad independiente de build y despliegue. Esta forma de
monorepo permite configurar un directorio raíz y watch paths distintos para cada
servicio en Railway y Cloudflare.

## Toolchain

- Go 1.26.5
- Python 3.13.14, administrado con `uv`
- Node.js 24.19.0 LTS
- Docker y Docker Compose

Los manifiestos y lockfiles de cada aplicación mantienen sus dependencias aisladas
y reproducibles.

## Desarrollo local

```bash
make up
```

El comando construye y levanta PostgreSQL, aplica migraciones y arranca Watchman,
Interpreter y Command Center. Cuando estén saludables:

- Command Center: <http://localhost:5173>
- Watchman: <http://localhost:8080/health>
- Interpreter: <http://localhost:8000/health>

Para detener el entorno:

```bash
make down
```

## Importar eventos desde Sentry

Copiar `.env.example` a `.env` y completar `SENTRY_ORG_SLUG`,
`SENTRY_AUTH_TOKEN` y, si se cambia su ubicación, `SENTRY_CATALOG_PATH`. Los seis
proyectos se definen en `apps/watchman/config/sentry-integrations.json`; la variable
histórica `SENTRY_PROJECT_SLUG` ya no se utiliza. El token necesita lectura de
eventos y `org:read` para consultar las sesiones de Release Health. Al ejecutar
`make up`, Watchman consulta Sentry de inmediato y luego cada dos minutos.

```bash
curl http://localhost:8080/internal/events
```

El poller queda desactivado de forma explícita si el token está vacío, por lo que
el entorno local también puede ejecutarse sin credenciales.

## Reportes diarios

Watchman genera a las 20:00 de `America/Argentina/Buenos_Aires` un reporte durable
con errores y sesiones de las cuatro aplicaciones. Sentry Release Health alimenta
Farmami, Wheels House y Prensap; `AppPageViews.SessionId` alimenta Notizap. Un fallo
transitorio de Telegram se reintenta cada 30 minutos, únicamente dentro del mismo
día. El historial privado está disponible en `GET /api/v1/reports` y en la sección
Reportes del Command Center.

## Incidentes y deploy markers

La sección Bitácora relaciona ocurrencias con deploys exitosos y permite agrupar
tipos de error en investigaciones auditables. Los deploys manuales se registran
desde la interfaz. GitHub Actions usa `POST /hooks/v1/deployments` con
`Authorization: Bearer $DEPLOYMENT_INGEST_TOKEN`.

Railway se configura por servicio con una URL de esta forma:

```text
https://ATALAYA/hooks/v1/deployments/railway/APLICACION/COMPONENTE?token=RAILWAY_WEBHOOK_TOKEN
```

Atalaya descarta eventos fallidos o que no pertenezcan al ambiente `production`.

## Interpretar eventos

Completar `OPENROUTER_API_KEY` en `.env`. Cada evento importado crea un job durable;
Watchman lo procesa automáticamente y el resultado aparece en
`GET /internal/events/{id}`. `OPENROUTER_MODEL` permite cambiar el modelo sin tocar
el código. La guía y el comportamiento ante fallos están documentados en
`docs/sprints/sprint-02-openrouter-walkthrough.md`.
