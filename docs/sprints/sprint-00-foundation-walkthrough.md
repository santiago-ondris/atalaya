# Sprint 0 — Walkthrough de la fundación ejecutable

- Fecha: 2026-08-06
- Estado del sprint: completado
- Estado de este avance: completado

Este documento explica el avance en el que Atalaya pasó de ser una arquitectura
documentada a tener cuatro componentes ejecutables y saludables en Docker.

## 1. Qué buscábamos conseguir

El objetivo no era implementar todavía Sentry, OpenRouter ni Telegram. Buscábamos
probar que los límites fundamentales del sistema funcionan:

```text
Command Center
      │
      ▼
  Watchman ──────► PostgreSQL

  Interpreter (servicio interno independiente)
```

El criterio de éxito era que un único entorno pudiera:

1. iniciar PostgreSQL;
2. aplicar migraciones;
3. iniciar Watchman solamente después de migrar la base;
4. iniciar Interpreter;
5. construir y servir Command Center;
6. comprobar automáticamente la salud de todos los procesos.

## 2. Forma de trabajo adoptada

Se modificaron `atalaya-core.md` y `atalaya-plan.md` para registrar el acuerdo de
trabajo del proyecto:

- el asistente implementa, prueba y documenta;
- Santiago mantiene las decisiones de producto y arquitectura;
- cada avance explica el propósito de las decisiones, librerías y archivos;
- aprender significa comprender y poder evaluar el sistema, no transcribir código
  manualmente.

## 3. Watchman

### Por qué Go

Watchman ejecutará tareas continuas: polling, persistencia, reintentos, reglas de
deduplicación y alertas. Go encaja bien por su modelo de concurrencia, binarios
autocontenidos y bajo costo operativo.

### Dependencia incorporada

`github.com/jackc/pgx/v5` es el driver y toolkit PostgreSQL de Watchman.

Se eligió porque:

- habla PostgreSQL de forma nativa;
- ofrece un pool de conexiones mediante `pgxpool`;
- maneja correctamente contextos y cancelaciones de Go;
- evita introducir un ORM antes de entender nuestras consultas.

`go.mod` declara dependencias; `go.sum` fija sus checksums. Juntos hacen que CI y
Docker resuelvan el mismo árbol de paquetes.

### Archivos y responsabilidades

#### `apps/watchman/cmd/watchman/main.go`

Es el composition root: el único lugar que ensambla el proceso completo.

Responsabilidades:

- crear el logger JSON;
- cargar configuración;
- conectarse a PostgreSQL;
- construir el servidor HTTP;
- escuchar señales `SIGINT` y `SIGTERM`;
- apagar el servidor ordenadamente.

No contiene handlers ni consultas de negocio. Así `main` muestra cómo se conecta
el sistema sin convertirse en un archivo gigante.

#### `apps/watchman/internal/config/config.go`

Traduce variables de entorno a un `Config` tipado. Actualmente exige
`DATABASE_URL` y define valores seguros para dirección HTTP y timeouts.

Centralizar esto evita leer `os.Getenv` en cualquier parte del código. También
permite fallar inmediatamente si una configuración indispensable no existe.

#### `apps/watchman/internal/database/postgres.go`

Crea el pool de PostgreSQL y ejecuta un `Ping` inicial.

Un pool reutiliza conexiones en lugar de abrir una por request. Si PostgreSQL no
está disponible durante el arranque, Watchman termina con un error explícito en
lugar de quedar vivo pero inutilizable.

#### `apps/watchman/internal/httpserver/server.go`

Define el servidor y sus primeros endpoints:

- `/health`: comprueba que el proceso HTTP vive;
- `/ready`: comprueba que el proceso puede usar PostgreSQL.

La distinción importa en producción. Un proceso puede estar vivo pero no estar
listo para recibir tráfico porque perdió su base de datos.

El servidor depende de una interfaz `Database` con un solo método `Ping`. Gracias
a eso el test puede simular una base sana o caída sin iniciar PostgreSQL.

También configura límites como `ReadHeaderTimeout` e `IdleTimeout` para no aceptar
conexiones sin límites explícitos.

#### `apps/watchman/internal/httpserver/server_test.go`

Prueba dos comportamientos:

- health devuelve `200`;
- readiness devuelve `503` cuando PostgreSQL falla.

Esto verifica comportamiento observable, no detalles internos.

### Logging

Se usa `log/slog`, incluido en la biblioteca estándar de Go. Los logs salen como
JSON para que Railway o cualquier plataforma pueda filtrarlos por campos.

No agregamos una librería externa porque `slog` cubre lo necesario en esta etapa.

## 4. Interpreter

### Por qué Python y FastAPI

Python concentrará la interacción con modelos y validación de respuestas. FastAPI
integra naturalmente Pydantic, genera OpenAPI y ofrece un servicio HTTP pequeño sin
introducir un framework de aplicación más grande.

### Qué es `uv`

`uv` administra el proyecto Python:

- crea el entorno virtual `.venv`;
- resuelve dependencias;
- genera `uv.lock` con versiones exactas;
- ejecuta herramientas dentro del entorno correcto;
- instala el paquete durante el build de Docker.

Cumple el rol que antes solía repartirse entre `pip`, `venv`, `pip-tools` y otras
herramientas.

### Dependencias de runtime

- `fastapi`: rutas HTTP y validación integrada.
- `pydantic`: modelos y validación de datos.
- `pydantic-settings`: configuración tipada desde variables de entorno.
- `httpx`: futuro cliente HTTP para OpenRouter y cliente usado en tests.
- `uvicorn`: servidor ASGI que ejecuta FastAPI.

### Dependencias de desarrollo

- `ruff`: lint y formato Python.
- `pytest`: ejecución de tests.

Las dependencias de desarrollo no entran en la imagen final.

### Archivos y responsabilidades

#### `apps/interpreter/pyproject.toml`

Es el manifiesto del proyecto: nombre, versión mínima de Python, dependencias y
comando ejecutable.

#### `apps/interpreter/uv.lock`

Registra la resolución exacta de todas las dependencias transitivas. Docker usa
`uv sync --locked`, que falla si el manifiesto y el lockfile no coinciden.

#### `src/atalaya_interpreter/config.py`

Define `Settings` mediante Pydantic. El prefijo `ATALAYA_` hace que
`ATALAYA_ENVIRONMENT` alimente el campo `environment`.

`lru_cache` crea la configuración una sola vez por proceso.

#### `src/atalaya_interpreter/api.py`

Crea la aplicación FastAPI y define `/health` y `/ready`.

La documentación interactiva `/docs` solo está habilitada en desarrollo. Evitamos
exponerla automáticamente en producción.

Readiness todavía no consulta OpenRouter porque esa integración pertenece al
Sprint 2. Hoy no existe una dependencia externa indispensable para Interpreter.

#### `src/atalaya_interpreter/main.py`

Es el punto de entrada del proceso. Inicia Uvicorn en el puerto 8000 y escucha en
`0.0.0.0`, necesario para aceptar tráfico desde fuera del contenedor.

#### `tests/test_health.py`

Prueba ambos endpoints mediante el transporte ASGI de HTTPX. Esto ejercita el
contrato HTTP sin abrir un puerto real durante los tests.

Inicialmente usamos `TestClient`, pero la versión actual emitió una advertencia de
deprecación. Se cambió a `httpx.ASGITransport` para no esconder una advertencia ni
agregar una dependencia nueva innecesaria.

## 5. Command Center

### Generador y dependencias

Se usó el generador oficial actual de Vite con el template React + TypeScript.

Generó:

- React 19 para componentes;
- TypeScript 6 para chequeo estático;
- Vite 8 para desarrollo y build;
- plugin oficial de React para Vite;
- Oxlint para análisis estático.

Nuestro ADR mencionaba ESLint, pero la plantilla oficial actual usa Oxlint. Se
actualizó el ADR para evitar reemplazar una configuración moderna sin una razón
concreta.

### Lockfile

`package-lock.json` fija el árbol exacto. En Docker usamos `npm ci`, que instala
estrictamente ese lockfile en lugar de recalcular versiones.

### Archivos principales

#### `src/main.tsx`

Monta la aplicación React en el elemento `#root` de `index.html`.

#### `src/App.tsx`

Define la pantalla mínima del Sprint 0. No es todavía el dashboard funcional; es
una confirmación visual de que el frontend construye y se sirve correctamente.

#### `src/App.css` e `src/index.css`

Aplican el lenguaje visual definido en el documento core: verde profundo, amarillo
bandera, crema, latón y las tres familias tipográficas.

Las pequeñas figuras de estado ya siguen la idea de banderas náuticas, en lugar de
usar íconos genéricos.

#### `nginx.conf`

Nginx sirve los archivos estáticos generados por Vite dentro de Docker. La regla
`try_files` permite que una futura SPA resuelva rutas del navegador devolviendo
`index.html`.

También expone `/health` sin depender de JavaScript.

## 6. Dockerfiles

Los tres Dockerfiles son multi-stage: una etapa construye y otra etapa contiene
solamente lo necesario para ejecutar.

### Watchman

1. Descarga módulos Go usando primero `go.mod` y `go.sum` para aprovechar caché.
2. Compila un binario sin CGO.
3. Compila `goose` 3.27.1 excluyendo drivers de bases no utilizadas.
4. Copia binario, migraciones y certificados a Alpine.
5. Ejecuta como usuario sin privilegios.

`goose` no quedó en `go.mod`: es una herramienta de despliegue, no una dependencia
del proceso Watchman. Incluirlo allí había arrastrado drivers de MySQL, ClickHouse,
SQLite y otras bases; se detectó y corrigió con `go mod tidy`.

### Interpreter

1. Obtiene el binario de `uv` desde su imagen oficial fijada.
2. Instala primero dependencias para aprovechar caché.
3. Copia el código e instala el paquete.
4. Copia únicamente `.venv` a la imagen final.
5. Ejecuta como usuario sin privilegios.

La instalación final usa `--no-editable`. Una instalación editable conserva una
referencia a `/app/src`; al copiar solamente `.venv`, esa ruta desaparecía y el
contenedor fallaba con `ModuleNotFoundError`. La instalación no editable copia el
paquete dentro del entorno virtual.

### Command Center

1. Node instala mediante `npm ci` y ejecuta el build.
2. La imagen final Nginx recibe únicamente `dist/`.

Node no queda ejecutándose en producción: el resultado es HTML, CSS y JavaScript
estático, que encaja con Cloudflare Pages.

## 7. Docker Compose

`compose.yaml` modela cinco servicios:

### `postgres`

Base persistente con volumen y `pg_isready` como health check.

PostgreSQL 18 cambió el layout oficial de datos. El primer arranque falló porque el
volumen estaba montado en `/var/lib/postgresql/data`, válido históricamente pero no
recomendado en 18+. Se corrigió a `/var/lib/postgresql`.

### `migrate`

Job efímero que espera a PostgreSQL, ejecuta `goose up` y termina con código cero.

Separarlo evita que múltiples réplicas de Watchman intenten migrar simultáneamente.
En Railway este rol se trasladará a un pre-deploy command.

### `watchman`

Espera a que PostgreSQL esté saludable y `migrate` termine correctamente. Su
health check usa `/ready`, por lo que Docker no lo considera sano sin base.

### `interpreter`

Arranca independientemente y valida `/ready` desde dentro de su contenedor.

### `command-center`

Espera que los dos backends estén saludables. Su health check usa
`127.0.0.1`, no `localhost`: BusyBox resolvía `localhost` como IPv6 (`::1`) mientras
Nginx escuchaba en IPv4, aunque el endpoint funcionaba desde el host.

## 8. Configuración

`.env.example` documenta nombres de variables sin contener secretos reales.

Las credenciales `atalaya_local` dentro de Compose son deliberadamente locales y
no se reutilizarán en producción. Railway inyectará un `DATABASE_URL` secreto.

Las variables de OpenRouter, Telegram, Sentry y Azure aparecen vacías para mostrar
qué configuración llegará en sprints futuros.

## 9. Comandos de desarrollo

El `Makefile` ofrece una interfaz corta:

- `make up`: build y arranque del entorno;
- `make down`: detención sin borrar datos;
- `make logs`: seguimiento de logs;
- `make test`: checks de los tres proyectos.

Make no reemplaza los toolchains. Solo evita que una persona tenga que recordar
la ubicación y sintaxis de cada comando.

## 10. Validaciones realizadas

### Go

- `go test ./...`
- `go vet ./...`

### Python

- `ruff check`
- `ruff format --check`
- `pytest`: tres tests aprobados

### Frontend

- `oxlint`
- TypeScript build
- Vite production build
- auditoría npm sin vulnerabilidades reportadas

### Integración

- construcción desde cero de las cuatro imágenes;
- PostgreSQL saludable;
- migración `00001` aplicada por Goose;
- cuatro aplicaciones iniciales insertadas;
- Watchman saludable y conectado a PostgreSQL;
- Interpreter saludable;
- Command Center saludable;
- todos los endpoints devolviendo HTTP 200.

## 11. Correlation IDs

Watchman e Interpreter aceptan `X-Correlation-ID`. Si el valor contiene un UUID
válido, lo preservan; si falta o es inválido, generan uno nuevo. Ambos lo devuelven
en el header de respuesta y lo incluyen en sus logs estructurados.

Watchman usa `github.com/google/uuid`: una dependencia pequeña y enfocada que
evita mantener parsing y generación criptográficamente segura por cuenta propia.

Interpreter usa `uuid` y `contextvars` de la biblioteca estándar. `ContextVar`
mantiene el ID correcto aunque FastAPI procese varios requests concurrentemente.

La prueba integrada envió el UUID
`8f6f0961-6de7-47af-9ab7-0ad4b82e18d8` a ambos servicios. Los dos devolvieron el
mismo valor y sus logs JSON lo registraron.

## 12. Integración continua y actualizaciones

`.github/workflows/ci.yml` define cuatro trabajos:

- Go: tests, `vet` y formato;
- Python: sync bloqueado, Ruff y pytest;
- frontend: `npm ci`, Oxlint y build;
- contenedores: build de todas las imágenes después de aprobar los tres anteriores.

Los jobs separados hacen visible qué toolchain falló. La concurrencia cancela una
ejecución vieja cuando llega un push más nuevo a la misma rama.

`.github/dependabot.yml` revisa semanalmente módulos Go, paquetes administrados con
`uv`, npm, Dockerfiles y GitHub Actions. Dependabot abre PRs; nunca despliega una
actualización por sí solo. Esos PRs deberán pasar CI antes de integrarse.

El workflow está validado sintácticamente y se activará cuando se inicialice
Git, publique el repositorio y use `main` o abra un pull request.

## 13. Autenticación preparada

El ADR 0004 define sesiones opacas administradas por Watchman, contraseña con
Argon2id y cookies seguras `HttpOnly`. No se implementó login porque corresponde al
Sprint 6; se decidió ahora únicamente para evitar que la estructura futura dependa
de tokens almacenados en el navegador.

## 14. Estado final del Sprint 0

Completado:

- estructura y toolchains;
- contratos y modelo inicial;
- proyectos ejecutables;
- lockfiles;
- Dockerfiles y Compose;
- migraciones;
- health/readiness;
- pruebas locales e integradas.

También quedó guardado un fixture sintético de Sentry, sin usuarios, requests,
URLs ni credenciales reales, para desarrollar el primer adapter sin copiar un
evento de producción al repositorio.

Todos los criterios del Sprint 0 están completos. CI todavía no puede ejecutarse en
GitHub porque el repositorio no fue inicializado ni publicado; su configuración y
los mismos comandos que ejecutará sí fueron validados localmente.
