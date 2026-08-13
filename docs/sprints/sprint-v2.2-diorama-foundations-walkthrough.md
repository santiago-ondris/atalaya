# Sprint V2.2 — Fundaciones visuales del diorama

## Arquitectura

La portada inmersiva carga `LighthouseScene` mediante un chunk diferido. Mientras se descarga conserva una portada 2D con la identidad de Atalaya; un error de importación o render queda contenido en un límite local y ofrece acceso a la vista clásica.

La escena usa React Three Fiber 9, Drei y Three.js con WebGL. Toda la geometría es procedural y propia: faro, isla, mar, estrellas y barcos decorativos se construyen con primitivas, sin modelos ni texturas externos.

## Composición y movimiento

- Cámara elevada de tres cuartos con órbita y zoom acotados, damping y pan deshabilitado.
- Faro central low-poly con cinco ventanas neutrales reservadas para V2.3/V2.4.
- Haz cálido con una vuelta cada doce segundos.
- Mar facetado con desplazamiento liviano de vértices.
- Tres embarcaciones blancas secundarias que reutilizan la misma familia geométrica.
- `prefers-reduced-motion` congela haz, oleaje y barcos.
- `/?scene=still` congela además cámara y controles para capturas reproducibles.

## Presupuesto

El Canvas limita DPR a 1–1.5. La escena usa iluminación ambiental, una luz direccional con sombras de 1024 px y una luz puntual del faro, sin postprocesado ni física. El build debe mantener el stack 3D fuera del chunk inicial y los recursos de escena por debajo de 3 MB transferidos.

Los perfiles de calidad, monitoreo de FPS y fallback automático por WebGL o rendimiento pertenecen a V2.6.

## Validación

La prueba automatizada cubre portada de carga y recuperación ante error sin crear WebGL en jsdom. La revisión visual usa capturas deterministas en Chrome. Antes de marcar V2.2 como completo se requiere aprobación del checkpoint visual y una comprobación manual en Safari real de carga, resize, órbita y zoom.

El checkpoint visual de 1440×900 y 1024×768 fue aprobado el 13 de agosto de 2026.

La validación manual posterior en Safari real confirmó carga, órbita limitada, zoom acotado, adaptación al redimensionar y animaciones estables. Con esa comprobación V2.2 quedó completado.

Firefox y Playwright WebKit se validarán en V2.7 junto con la suite browser integral.
