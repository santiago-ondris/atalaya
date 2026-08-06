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
