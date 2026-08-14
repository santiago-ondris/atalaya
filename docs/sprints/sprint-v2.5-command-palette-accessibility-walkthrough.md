# Sprint V2.5 — Paleta de comandos, modos y accesibilidad

## Resultado

La sesión privada incorpora una única paleta global de navegación basada en `cmdk`, disponible desde el faro, el shell fino y la vista clásica. La validación manual en Safari real confirmó apertura con `⌘K`, cierre con `Escape`, búsqueda y operación general. El checkpoint fue aprobado el 14 de agosto de 2026 y V2.5 quedó cerrado.

## Catálogo y navegación

La definición existente del faro continúa siendo la fuente única para los nueve destinos. Cada entrada incluye ahora grupo, contexto y palabras de búsqueda, reutilizando los aliases del catálogo de aplicaciones y términos operativos previsibles.

La paleta presenta dos grupos:

- Aplicaciones: Farmami, Wheels House, Prensap, Notizap y Atalaya.
- Operación: Eventos, Bitácora, Reportes y Estado del sistema.

Cada resultado muestra nombre y contexto, sin estado vivo ni polling adicional. La selección cierra el diálogo y navega mediante React Router sin alterar el modo vigente.

## Integración y resiliencia

`CommandPaletteProvider` se monta una sola vez dentro del área autenticada, por encima de rutas, shells, `Suspense` y el límite de errores de Three.js. Por eso la paleta no depende del Canvas y continúa disponible durante la carga o recuperación de la escena.

Los disparadores “Ir a…” están presentes en las acciones del faro, el header fino y el sector de guardia del sidebar clásico. Los nueve enlaces DOM accesibles del faro permanecen intactos con su estado y frescura.

La preferencia `immersive|classic` continúa en `sessionStorage` bajo `atalaya:view-mode`; cambiar de modo conserva la navegación y cerrar sesión elimina la preferencia.

## Teclado, foco y movimiento

La paleta abre con `⌘K`, `Ctrl+K` o sus botones visibles. El atajo previene la acción nativa, el buscador recibe foco inicial y `Escape`, selección o interacción exterior cierran el diálogo. Después del cierre el foco vuelve al elemento que originó la apertura, incluido el foco previo al atajo global.

`cmdk` aporta semántica de diálogo, foco atrapado, grupos, listbox, navegación con flechas y selección con Enter. La interfaz muestra selección índigo, estado “Sin resultados” y rótulos accesibles. Con `prefers-reduced-motion: reduce` se eliminan las animaciones y transiciones decorativas de la paleta; la escena conserva la congelación incorporada en V2.4.

## Lenguaje visual

La paleta aplica la identidad Atalaya: papel cálido, regla de latón, superficies planas, radios contenidos, tipografía operativa Public Sans, contexto en IBM Plex Mono e índigo exclusivamente para foco y selección. El overlay usa el oscurecimiento modal del sistema y el diálogo mantiene una anchura máxima de 560 px.

## Pruebas y validación

La suite final cubre catálogo, rutas únicas, grupos, aliases, botón, `Meta+K`, `Ctrl+K`, rechazo de `K` sin modificador, foco inicial, restauración de foco, búsqueda normalizada, estado vacío, flechas, Enter, Escape, interacción exterior e integración con los tres shells.

La validación automatizada completó 51 tests, lint, build y formato. `npm audit` informó cero vulnerabilidades. El build conserva Three.js en un chunk diferido `LighthouseScene` separado del bundle principal.

La revisión visual confirmó la paleta sobre faro y vista clásica, además de su presencia en el shell fino. La aprobación manual en Safari cerró los recorridos esenciales; la preferencia de movimiento reducido queda respaldada por la implementación CSS y la cobertura existente de congelación de escena.

## Deuda posterior

- V2.6 incorporará perfiles explícitos de calidad, presupuesto de render y recuperación por rendimiento.
- V2.7 añadirá recorridos browser automatizados y la regresión final entre navegadores.
