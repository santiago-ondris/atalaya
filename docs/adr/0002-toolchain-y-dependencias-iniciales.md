# ADR 0002: toolchain y dependencias iniciales

- Estado: aceptado
- Fecha: 2026-08-06

## Contexto

El Sprint 0 necesita una base actual, reproducible y compatible con los entornos
de build de Railway y Cloudflare, evitando dependencias antes de necesitarlas.

## Decisión

### Runtimes

- Go 1.26.5.
- Python 3.13.14 administrado con `uv` 0.12.2.
- Node.js 24.19.0 LTS con npm.
- PostgreSQL 18.4 para desarrollo local.

Se fijan versiones exactas en desarrollo y CI. Las actualizaciones llegan mediante
PRs y se validan antes de desplegar. PostgreSQL mantiene fija la versión mayor y
recibe actualizaciones menores controladas.

### Watchman

- `net/http` para HTTP y routing mientras cubra las necesidades del proyecto.
- `pgx/v5` para PostgreSQL.
- Migraciones SQL explícitas con una herramienta CLI, elegida al crear la primera
  migración.
- `log/slog` para logging estructurado.
- Biblioteca estándar `testing` como base de tests.

No se incorpora ORM. Las consultas SQL son parte importante del comportamiento
operacional y deben permanecer visibles.

### Interpreter

- FastAPI como framework HTTP.
- Pydantic para contratos y validación.
- `pydantic-settings` para configuración.
- HTTPX para OpenRouter.
- Uvicorn como servidor ASGI.
- Ruff y pytest como herramientas de desarrollo.

Las versiones exactas quedarán resueltas en `uv.lock` al inicializar el proyecto.

### Command Center

- React + TypeScript.
- Vite para desarrollo y build.
- Oxlint para análisis estático, siguiendo la plantilla oficial actual de Vite.

No se eligen todavía router, manejo de estado remoto, componentes visuales ni
diagramación. Se incorporarán cuando el primer caso de uso defina sus requisitos.

## Consecuencias

- Hay pocas abstracciones al comienzo y cada dependencia tiene un uso concreto.
- Los lockfiles forman parte del repositorio.
- Los deploys nunca ejecutarán actualizaciones abiertas o globales.
- Las actualizaciones automáticas crearán PRs; no modificarán producción por sí
  solas.
