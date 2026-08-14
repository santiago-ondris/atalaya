# Sprint V2.3 — Destinos narrativos e interacción

## Resultado

El diorama funciona como navegación principal hacia nueve destinos: las cinco fichas de aplicación, Eventos, Bitácora, Reportes y Estado del sistema. El checkpoint visual final fue aprobado el 13 de agosto de 2026.

La validación manual posterior en Safari real confirmó hover y rótulos, navegación por click, prevención de activaciones durante la órbita, recorrido completo con `Tab`, activación con teclado, salida de foco del Canvas y estabilidad de las animaciones. Con esa comprobación V2.3 quedó completado.

## Destinos e interacción

Una definición centralizada relaciona cada identificador estable, rótulo, ruta y clase narrativa. Las cinco ventanas conducen a `/apps/:slug`; el patrullero abre `/events`, el mercante `/operations`, el paquebote `/reports` y la boya `/system`.

Cada protagonista conserva un hitbox transparente separado de su geometría. Los decorativos, el mar, la isla, el haz y los detalles no participan del raycasting. Un click sólo navega si el puntero no se desplazó más de cuatro píxeles, evitando activaciones accidentales durante la órbita.

El hover cambia el cursor y presenta un único rótulo discreto. La capa DOM contiene nueve enlaces reales en el mismo orden narrativo: permanecen visualmente ocultos hasta recibir foco, muestran entonces nombre y foco visible, y permiten recorrer toda la escena sin depender del Canvas.

## Lenguaje visual

- Patrullero angular con casco perfilado, puente, ventanas, barandas, radar, antenas y reflector de búsqueda.
- Mercante con casco perfilado, contenedores, puente, barandas y radar.
- Paquebote con dos cubiertas, ojos de buey, chimenea, botes laterales y equipamiento.
- Boya marítima con cuerpo flotante, defensas, escalera, estructura y linterna protegida.
- Cinco ventanas arqueadas empotradas con marco, visera, alféizar, parteluz y vidrio neutral iluminado. El estado real se incorporará en V2.4.

Los protagonistas se distribuyen en sectores separados del mar. La ambientación añade oleaje facetado, variación tonal, espuma costera, rocas, estelas y seis barcos decorativos pequeños en diferentes profundidades. Todo es procedural y de autoría propia.

## Accesibilidad y movimiento

`prefers-reduced-motion` congela oleaje, barcos, reflector, boya y haz sin retirar destinos ni información. `?scene=still` mantiene semilla, cámara y animaciones congeladas para capturas reproducibles.

Los tests verifican que existan exactamente nueve destinos, que sus identificadores resuelvan una ruta única y que la capa DOM exponga los nueve enlaces en orden con rótulo al foco.

## Deuda posterior

- V2.4 completó la aplicación de estado real a las ventanas y el clima reactivo sin modificar rutas ni modelos.
- V2.6 incorporará perfiles de calidad y degradación de elementos ambientales.
- V2.7 añadirá recorridos browser automatizados con Playwright para mouse, teclado y foco.
