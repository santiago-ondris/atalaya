# Sprint 06: Command Center privado

## Resultado

Atalaya ya dispone de un flujo operativo privado y demostrable:

1. El usuario inicia sesión con una contraseña cuyo hash Argon2id vive en configuración.
2. Watchman crea una sesión opaca y entrega una cookie `HttpOnly`, `SameSite=Lax` y `Secure` en producción.
3. El Command Center carga el estado de las cuatro aplicaciones y las integraciones reales.
4. La bitácora permite filtrar y paginar por aplicación, severidad, estado y período.
5. Desde cualquier fila se abre el detalle técnico, la interpretación y las acciones sugeridas.
6. La vista Sistema muestra checkpoints, último éxito y errores de cada integración.

## Responsabilidad de los archivos principales

- `apps/watchman/internal/auth/service.go`: verifica Argon2id, genera tokens aleatorios y transforma cada token en SHA-256 antes de persistirlo.
- `apps/watchman/migrations/00004_command_center_auth.sql`: incorpora sesiones expirables y revocables sin almacenar el token original.
- `apps/watchman/internal/httpserver/server.go`: expone `/api/v1`, escribe la cookie y protege las rutas privadas.
- `apps/watchman/internal/store/postgres.go`: consulta páginas de eventos con filtros y compone los estados usados por la interfaz.
- `contracts/openapi/watchman.v1.yaml`: documenta el contrato público del Watchman para el navegador.
- `apps/command-center/src/api.ts`: concentra transporte HTTP y tipos del contrato; ningún componente conoce URLs externas.
- `apps/command-center/src/App.tsx`: coordina sesión y vistas de overview, eventos, detalle y sistema.
- `apps/command-center/src/App.css`: aplica el lenguaje visual Atalaya aprobado mediante `hue`.
- `apps/command-center/nginx.conf`: sirve la SPA y reenvía `/api/` al Watchman para conservar un único sitio y evitar CORS.

## Decisiones de seguridad

- La contraseña y el token de sesión nunca se almacenan en texto.
- El navegador no puede leer la cookie de sesión.
- Las sesiones vencen de forma absoluta y el logout las revoca en PostgreSQL.
- Todos los endpoints con información operativa requieren sesión, incluidos los endpoints internos heredados.
- `ATALAYA_COOKIE_SECURE=false` existe únicamente para HTTP local. Producción debe usar `true` y HTTPS.
- La respuesta de login no distingue entre contraseña inexistente e incorrecta.

## Diseño

La primera impresión concentra la composición cartográfica expresiva. Dentro del producto, el diseño baja la teatralidad: navegación verde profunda, papel cálido, reglas de latón, interacción índigo, tipografía mono para datos y banderas geométricas para estados. El sistema fuente vive en `.agents/skills/atalaya-design`.

## Ejecución local

```bash
docker compose up -d --build
```

Abrir `http://localhost:5173`. Si no se define `ATALAYA_ADMIN_PASSWORD_HASH`, Compose utiliza el hash de desarrollo cuya contraseña es `atalaya_local`. Ese valor es solo para desarrollo.

Para producción, generar un hash Argon2id fuera del repositorio y configurar:

```text
ATALAYA_ADMIN_PASSWORD_HASH=<hash-argon2id>
ATALAYA_COOKIE_SECURE=true
ATALAYA_SESSION_DURATION_SECONDS=86400
```

## Validación realizada

- Suite completa de Go.
- Lint y build de React/TypeScript.
- Build de las cuatro imágenes Docker.
- Migración real sobre PostgreSQL local.
- Rechazo `401` sin cookie.
- Login real, cookie persistida y overview autenticado con datos reales.
- Revisión visual de la pantalla de login a 1440×900.

## Pendientes conscientes

- La rotación periódica de sesiones y un rate limiter distribuido se evaluarán antes de producción pública. La mitigación inicial del login incorpora respuesta uniforme y demora fija ante credenciales inválidas.
- El modo oscuro está definido en la design skill, pero no se expone todavía como preferencia en el producto real porque la versión aprobada del sprint es clara.
