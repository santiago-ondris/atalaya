# Handoff — Cortes 2 y 3

## Punto de partida

El Corte 1 desktop cerró con una navegación inmersiva que convive con una ruta
operativa directa: command palette, URLs reales y modo clásico. Esa redundancia es
una propiedad de resiliencia y debería preservarse en mobile.

## Aprendizajes desktop

- El Canvas puede ser reemplazable si catálogo, rutas, sesión, paleta y capa DOM
  viven fuera de la escena.
- La semántica de cinco ventanas y tres climas funciona mejor centralizada y no
  debe duplicarse en una variante mobile.
- La degradación escalonada —reducir calidad antes de abandonar el 3D— evita que
  un problema visual bloquee la operación.
- Semilla y animación deterministas facilitan checkpoints, pero no reemplazan una
  suite browser transversal.
- Las métricas reproducibles de bundle son confiables; las cifras manuales de FPS
  deben registrarse durante la sesión o conservarse sólo como evidencia cualitativa.

## Restricciones mobile heredadas

- El desvío actual a clásico bajo 901 px o sin puntero fino es una protección
  temporal, no el diseño del Corte 2.
- Los hotspots, órbita, hover y densidad visual desktop no deben trasladarse sin
  adaptación táctil.
- El modo clásico actual tampoco debe asumirse como responsive definitivo.
- La elección persistente post-login y la detección de capacidad pertenecen al
  trabajo futuro; hoy la preferencia sólo dura la sesión.
- Salud, rutas, autenticación y páginas funcionales existentes no necesitan nuevos
  contratos para iniciar el diseño mobile.

## Preguntas pendientes

- ¿Qué jerarquía y navegación necesita la experiencia clásica mobile rediseñada?
- ¿Qué gestos, tamaños de hitbox y cámara hacen al faro táctil comprensible sin
  depender de hover?
- ¿Qué señales de GPU y dispositivo justifican ofrecer o recomendar el modo
  inmersivo?
- ¿Cuándo y dónde se presenta la elección post-login, y cómo se persiste sin
  encerrar al usuario en un modo fallido?
- ¿Qué presupuesto de bundle, memoria y batería se adopta para cada perfil mobile?
- ¿La deuda browser transversal se resuelve antes o dentro del próximo corte?

Este handoff no crea alcance ni sprints. El roadmap de Cortes 2 y 3 deberá
escribirse cuando estas preguntas tengan decisiones explícitas.
