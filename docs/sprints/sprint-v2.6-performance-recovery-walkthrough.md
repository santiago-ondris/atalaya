# Sprint V2.6 — Rendimiento, adaptación y recuperación

## Resultado

El faro mide frames en ventanas de un segundo después de tres segundos de calentamiento. Tres ventanas bajo 50 FPS cambian el perfil de `normal` a `reduced`; diez ventanas consecutivas bajo 30 FPS abren la vista clásica. La recuperación de una ventana reinicia su contador y el perfil nunca mejora automáticamente.

Los parámetros de desarrollo son:

- `?performanceDebug=1`: perfil, FPS y ventanas restantes.
- `?performance=degraded`: fuerza 45 FPS medidos y reproduce `normal → reduced`.
- `?performance=critical`: fuerza 20 FPS y reproduce `normal → reduced → classic`.

## Presupuestos visuales

| Perfil  | DPR | Sombras |   Mar | Nubes | Estrellas | Flota decorativa |
| ------- | --: | ------- | ----: | ----: | --------: | ---------------: |
| normal  | 1.5 | 1024 px | 48×48 |     5 |        90 |                6 |
| reduced |   1 | no      | 24×24 |     3 |        45 |                2 |

El faro, cinco ventanas, cuatro destinos narrativos, clima, estado y los nueve enlaces DOM permanecen completos. `prefers-reduced-motion` sólo congela movimiento.

## Recuperación

WebGL se comprueba antes de montar el import diferido. La ausencia de WebGL abre `/overview` y conserva una explicación en `sessionStorage`. Los errores de import/render y la pérdida irrecuperable de contexto muestran acciones para reintentar o abrir la vista clásica. Una restauración WebGL dentro de tres segundos remonta el Canvas una vez. Logout limpia modo y motivo.

## Verificación automatizada

- Vitest: máquina de ventanas, calentamiento, pausa, recuperación, ambos umbrales, perfiles, persistencia de motivos y nueve destinos.
- `lint`, `build` y `format:check`: aprobados.
- `build:report`: bundle inicial 211.8 kB gzip; `LighthouseScene` 235.6 kB gzip; límite de escena 3 MB aprobado. Three.js permanece en el chunk diferido `LighthouseScene`.
- `npm audit --audit-level=high`: 0 vulnerabilidades.

## Validación de referencia

Equipo: MacBook Air M5, 16 GB, GPU 8 núcleos. Viewport: 1440×900.

La validación humana de cinco minutos en Safari real y Chrome se completó como
checkpoint de publicación el 14 de agosto de 2026. Confirmó la transición crítica,
retry, navegación y `⌘K`. No quedaron cifras de FPS registradas, por lo que el
cierre conserva esta evidencia como cualitativa y no inventa valores para los
perfiles normal o reducido.
