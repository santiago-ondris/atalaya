# Atalaya Command Center v2

Frontend React + TypeScript construido con Vite.

En desktop, la ruta `/` presenta el faro inmersivo con nueve destinos. Las cinco
ventanas abren fichas operativas y Eventos, Bitácora, Reportes y Estado del sistema
se representan en el mar. `⌘K` o `Ctrl+K` abre la command palette desde el faro,
el shell fino o la vista clásica.

## Modos y recuperación

- “Vista clásica” conserva overview, sidebar y pantallas funcionales; “Volver al
  faro” recupera la experiencia inmersiva.
- La preferencia `immersive|classic` dura la sesión y se limpia al cerrar sesión.
- Viewports menores a 901 px o dispositivos sin puntero fino entran temporalmente
  en modo clásico.
- Bajo rendimiento sostenido reduce primero la calidad. Si el perfil reducido
  permanece bajo 30 FPS durante 10 segundos, se abre la vista clásica.
- WebGL ausente, errores de carga/render o una pérdida irrecuperable del contexto
  ofrecen o activan una salida clásica recuperable. El aviso permite reintentar el
  faro sin crear un loop automático.

## Desarrollo y diagnóstico

```bash
npm run test
npm run lint
npm run format:check
npm run build
npm audit --audit-level=high
npm run build:report
```

En desarrollo, `?scene=still` fija la escena, `?performanceDebug=1` muestra el
perfil y FPS medidos, y `?performance=degraded|critical` reproduce las transiciones
de calidad. Los detalles de operación están en
`../../docs/runbooks/operational-runbook.md`.
