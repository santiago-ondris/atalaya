# Sprint V2.0 — Rutas y cimientos de navegación

Command Center usa `BrowserRouter` y URLs reales. `/status` permanece pública; las rutas privadas conservan el deep link durante la verificación y el login. Las rutas funcionales son `/events`, `/events/:eventId`, `/operations`, `/reports`, `/architecture/:appSlug?`, `/system` y `/apps/:appSlug`; `/overview` conserva el shell clásico.

El catálogo único vive en `src/catalog/applications.ts`. Los aliases redirigen con `replace` y los slugs desconocidos muestran 404.

La sesión elige vista inmersiva en desktop (901 px o más y puntero fino) y clásica en los demás dispositivos, salvo preferencia de `sessionStorage`. “Vista clásica” abre `/overview`; “Volver al faro” abre `/`. El logout elimina la preferencia.

Verificación: `npm run test`, `npm run lint`, `npm run build` y `npm run format:check` desde `apps/command-center`.

V2.1 incorporará salud y reportes a las fichas. V2.2 reemplazará la portada mínima. V2.5 incorporará la command palette; este sprint no agrega 3D, Canvas ni datos de backend.
