# ADR 0004: autenticación inicial de usuario único

- Estado: aceptado para implementación futura
- Fecha: 2026-08-06

## Contexto

Command Center tendrá un único usuario inicialmente, pero manejará errores y datos
operativos de aplicaciones reales. La autenticación debe ser simple sin tratar el
frontend como un almacén seguro de credenciales.

## Decisión

- Watchman será dueño de autenticación y autorización.
- La contraseña se almacenará únicamente como hash Argon2id.
- El login creará una sesión opaca, aleatoria y revocable.
- PostgreSQL guardará solamente el hash del token de sesión.
- El navegador recibirá el token mediante cookie `HttpOnly`, `Secure` y
  `SameSite=Lax`; no se guardará en `localStorage`.
- Las sesiones tendrán expiración absoluta y podrán revocarse.
- La API privada rechazará requests sin sesión válida.
- La status page pública usará endpoints explícitamente públicos y sin detalles
  técnicos sensibles.

Se preferirá servir frontend y API bajo el mismo sitio para reducir complejidad de
CORS y CSRF. Si el despliegue final exige sitios distintos, se escribirá un ADR que
revise cookies, dominios permitidos y protección CSRF.

## Consecuencias

- No necesitamos OAuth ni un proveedor de identidad para un único usuario.
- Una filtración de PostgreSQL no revela contraseñas ni tokens de sesión activos.
- Watchman deberá incorporar rate limiting de login, rotación de sesión y limpieza
  de sesiones vencidas cuando se implemente la autenticación.

