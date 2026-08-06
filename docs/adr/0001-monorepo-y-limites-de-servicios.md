# ADR 0001: monorepo y límites de servicios

- Estado: aceptado
- Fecha: 2026-08-06

## Contexto

Atalaya tiene tres aplicaciones con runtimes distintos que forman un único
producto. Necesitamos cambios coordinados, desarrollo local reproducible y
despliegues independientes en Railway y Cloudflare.

## Decisión

Usaremos un monorepo aislado con estas unidades desplegables:

- `apps/watchman`: Go; dueño del dominio, persistencia y API pública.
- `apps/interpreter`: Python; servicio interno y stateless para LLM.
- `apps/command-center`: React + TypeScript; cliente de la API de Watchman.

Los contratos viven en `contracts/` y las migraciones en `database/migrations/`.
Cada aplicación mantiene sus dependencias, tests, Dockerfile y ciclo de deploy.

No usaremos Turborepo inicialmente. Railway y Cloudflare pueden construir cada
directorio de manera independiente y no existe todavía código TypeScript
compartido que justifique un task runner adicional.

## Consecuencias

- Un cambio de contrato puede revisarse junto con sus consumidores.
- Cada servicio puede desplegarse y revertirse por separado.
- CI deberá seleccionar checks según las rutas modificadas.
- Las tareas cross-language se coordinarán con Docker Compose y comandos simples.
- La incorporación futura de Turbo, Nx u otra herramienta requerirá un ADR nuevo
  basado en una necesidad observada.

