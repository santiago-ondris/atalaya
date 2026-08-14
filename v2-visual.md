# Atalaya v2 — Faro 3D, homepage y navegación principal

## Estado del documento

Este documento es la fuente de verdad conceptual y de producto para la evolución visual de Atalaya v2. Las decisiones del **Corte 1 — Desktop** están cerradas. Su ejecución se divide en sprints en `v2-corte-1-plan.md`.

Los Cortes 2 y 3 permanecen deliberadamente fuera de ese roadmap. Cuando termine el Corte 1 se escribirá un plan nuevo para ambos, incorporando lo aprendido durante la implementación y validación real de la escena desktop.

## Contexto y excepción de autoría

El principio general de Atalaya es que Santiago escribe el código y los agentes actúan como guía. La v1 se construyó así para adquirir experiencia real en Go, Python y Docker.

**La experiencia 3D es una excepción deliberada, no un abandono del principio.** Three.js no forma parte del conjunto de habilidades objetivo y el faro no introduce lógica de negocio. Por eso un agente puede implementar de punta a punta la escena, los shaders y su integración, manteniendo explicadas las decisiones relevantes y validando el resultado como producto.

## Evolución del concepto

El faro nació como una pieza decorativa dentro del layout clásico de v1. En v2 pasa a ser la homepage y la navegación principal: es la puerta de entrada a las aplicaciones monitoreadas y a las herramientas funcionales de Atalaya.

La experiencia acepta conscientemente una navegación algo más exploratoria a cambio de una identidad visual fuerte. Esa decisión no debe perjudicar la operación diaria: la command palette y el modo clásico permanecen disponibles como caminos rápidos y seguros.

## Dirección artística confirmada

La escena será un **diorama nocturno low-poly**:

- Mar oscuro, cielo nocturno, luces cálidas y niebla controlada.
- Geometrías estilizadas y siluetas claras, sin buscar realismo cinematográfico.
- Uso reconocible de la paleta de `atalaya-core.md`: verde profundo `#005838`, amarillo bandera `#FFE000`, índigo `#4750A8`, ocre/latón `#BA8A40` y rojo ladrillo `#C1432E`.
- Faro y embarcaciones construidos principalmente con geometrías primitivas o procedurales propias.
- Texturas mínimas. Si algún recurso externo resulta imprescindible, deberá tener licencia compatible y quedar documentado.
- WebGL como renderer del Corte 1. WebGPU queda fuera por compatibilidad y complejidad innecesaria para este diorama.

La composición debe sentirse como una maqueta viva y legible, no como un videojuego ni como una simulación marítima realista.

## Estructura de la escena

### El faro y sus cinco ventanas

El faro representa el conjunto del sistema. Tiene cinco ventanas funcionales:

1. Farmami.
2. Wheels House.
3. Prensap.
4. Notizap.
5. Atalaya.

Cada ventana:

- Cambia de color según el estado real de su destino.
- Se puede activar con mouse o teclado.
- Navega a una ficha operativa con URL propia.
- Dispone de una hitbox estable, independiente de la geometría visible.

El faro completo funciona como overview implícito del sistema. El overview clásico no desaparece: permanece dentro del modo v1 alternativo.

El haz de luz completa una vuelta cada **12 segundos**. Debe ser sutil y detenerse cuando el usuario prefiera movimiento reducido.

### Fichas operativas

Las ventanas no abren simplemente la página de Arquitectura. Cada aplicación tendrá un espacio dedicado en `/apps/:appSlug`.

Para Farmami, Wheels House, Prensap y Notizap, la ficha contiene:

- Estado de disponibilidad general.
- Estado frontend/backend y última comprobación.
- Bloque correspondiente al último reporte diario disponible.
- Diagrama de arquitectura interactivo existente.
- Accesos a las pantallas globales relacionadas.

La ficha de Atalaya es deliberadamente distinta porque Atalaya no pertenece al modelo de reportes diarios de las cuatro aplicaciones. Contiene:

- Estado interno agregado.
- Resumen de señales y colas.
- Diagrama de arquitectura de Atalaya.
- Acceso al detalle de Estado del sistema.

### El mar y los destinos funcionales

Las cuatro páginas funcionales de v1 conservan este mapeo exacto:

| Destino            | Representación                      | Comportamiento narrativo                                         |
| ------------------ | ----------------------------------- | ---------------------------------------------------------------- |
| Eventos            | Patrullero o pesquero con reflector | Nave activa con patrón de búsqueda                               |
| Bitácora           | Mercante o barco de carga           | Sigue una ruta de entrada, como un cargamento que llega a puerto |
| Reportes           | Paquebote o barco correo            | Transporta la correspondencia y el parte diario                  |
| Estado del sistema | Boya señalizadora                   | Emite una señal de seguridad periódica                           |

Los barcos decorativos pertenecen a una familia homogénea: embarcaciones pequeñas, blancas, de menor jerarquía y movimiento ambiental simple.

Los protagonistas no se distinguen únicamente por color. Cada uno debe resaltar por una combinación de:

- Silueta y escala.
- Equipamiento reconocible.
- Patrón de movimiento.
- Iluminación funcional.
- Posición y ritmo dentro de la composición.

Al posar el cursor o alcanzar un protagonista mediante teclado aparece un cursor de enlace y una etiqueta pequeña con su destino. No se usará un glow genérico como mecanismo principal de diferenciación. Los objetos decorativos nunca interceptan la interacción de un destino.

### Mar

El oleaje será sutil y low-poly, implementado con desplazamiento de vértices y ruido liviano. No se incorporará una simulación física.

El shader puede adaptar una implementación existente únicamente si su licencia es compatible, se conserva la atribución necesaria y la fuente queda documentada. También puede escribirse una versión pequeña específica para Atalaya si eso produce un resultado más simple y mantenible.

### Cámara

En desktop se utilizará órbita limitada:

- Rotación horizontal y vertical acotadas alrededor del faro.
- Zoom limitado para evitar perder la composición o atravesar geometrías.
- Damping habilitado.
- Sin desplazamiento lateral o `pan`.
- Posición inicial que permita leer el faro y los destinos principales sin interacción obligatoria.

## Semántica del estado real

La codebase v1 expone tres conceptos distintos y no deben mezclarse accidentalmente:

- `/api/v1/overview`: salud de las integraciones de ingesta de las cuatro aplicaciones.
- `/api/v1/public/status`: disponibilidad frontend/backend de las cuatro aplicaciones.
- `/api/v1/system/health`: salud interna de Atalaya, sus procesos y colas.

### Ventanas de aplicaciones

Farmami, Wheels House, Prensap y Notizap usan `/api/v1/public/status`:

| Estado API     | Color visual |
| -------------- | ------------ |
| `operational`  | Verde        |
| `degraded`     | Amarillo     |
| `unknown`      | Amarillo     |
| `major_outage` | Rojo         |

### Ventana de Atalaya

Atalaya usa `/api/v1/system/health`:

| Estado                                      | Color visual |
| ------------------------------------------- | ------------ |
| `healthy`                                   | Verde        |
| `degraded` o señal interna no saludable     | Amarillo     |
| Sin dato previo o señal interna desconocida | Amarillo     |

La imposibilidad de consultar una fuente no debe bloquear la escena ni inutilizar sus enlaces. Se conserva el último dato conocido cuando corresponda y se comunica que está desactualizado; si no hay dato previo, se representa como desconocido amarillo, sin inventar una caída roja.

Los estados se consultan en paralelo y se refrescan cada 30 segundos. La v2 no requiere un endpoint agregado específico para el faro ni una migración de base de datos.

### Clima de tres niveles

El clima representa el peor estado de las cinco ventanas:

1. **Despejado:** todas las ventanas verdes; cielo nocturno limpio y estrellas visibles.
2. **Bruma:** existe al menos una ventana amarilla y ninguna roja; niebla y nubes suaves.
3. **Tormenta contenida:** existe al menos una ventana roja; mayor densidad atmosférica y señales visuales de tensión sin comprometer la legibilidad.

El clima nunca debe esconder los hotspots, impedir el uso de la cámara ni convertir el estado crítico en una animación estridente.

## Navegación y URLs

La navegación privada deja de vivir únicamente en estado React. Las rutas confirmadas son:

| Ruta                      | Destino                           |
| ------------------------- | --------------------------------- |
| `/`                       | Faro inmersivo                    |
| `/overview`               | Overview clásico                  |
| `/apps/:appSlug`          | Ficha operativa de una aplicación |
| `/events`                 | Eventos                           |
| `/events/:eventId`        | Detalle de evento                 |
| `/operations`             | Bitácora                          |
| `/reports`                | Reportes                          |
| `/architecture/:appSlug?` | Arquitectura                      |
| `/system`                 | Estado del sistema                |
| `/status`                 | Status público existente          |

Las rutas privadas soportan refresco, historial atrás/adelante y deep links. Si una sesión no está autenticada, el login debe conservar el destino pedido y recuperarlo después del acceso.

### Shell fino

Al salir del faro hacia una pantalla privada se muestra un shell mínimo con:

- Regreso al faro.
- Nombre de la sección actual.
- Acceso visible a la command palette.
- Cerrar sesión.

Eventos, Bitácora, Reportes y Estado del sistema conservan por ahora su contenido y comportamiento de v1. No se rediseñan sus layouts en el Corte 1.

### Modo clásico

La experiencia v1 completa —overview y sidebar— permanece disponible como alternativa estable:

- Desde el faro existe un control visible “Vista clásica”.
- Desde el modo clásico existe “Volver al faro”.
- La elección se guarda en `sessionStorage`, dura solamente durante la pestaña/sesión y se elimina al cerrar sesión.
- No reemplaza todavía la preferencia persistente post-login prevista para el Corte 3.

Durante el Corte 1, dispositivos sin puntero fino o viewports menores a 901 px entran temporalmente al modo clásico. Esto es una protección transitoria, no la implementación del Corte 2.

## Command palette

`⌘K` en macOS y `Ctrl+K` en otros sistemas abren una paleta disponible desde el faro, el shell fino y el modo clásico.

Contiene exactamente:

- Las cinco fichas de aplicación.
- Eventos.
- Bitácora.
- Reportes.
- Estado del sistema.

Debe ser operable con teclado, tener búsqueda por nombre y alias, conservar foco correctamente y cerrar con `Escape`. La implementación elegida es `cmdk`, estilizada con la identidad de Atalaya.

## Accesibilidad

El Canvas no será la única representación navegable de los destinos:

- Los nueve destinos tendrán equivalentes DOM para teclado y tecnologías de asistencia.
- Los elementos interactivos tendrán nombres accesibles y foco visible.
- Las etiquetas y estados mantendrán contraste suficiente.
- `prefers-reduced-motion` detendrá oleaje, barcos, clima y haz continuos sin eliminar información ni navegación.
- La command palette debe funcionar aunque la escena no llegue a cargar.

## Herramientas confirmadas

- React Router en modo declarativo para rutas y layouts.
- Three.js con WebGL.
- `@react-three/fiber` 9, compatible con React 19.
- `@react-three/drei` para controles y herramientas de rendimiento.
- `cmdk` para la command palette.
- Vitest y Testing Library para lógica y componentes.
- Playwright para navegación, fallbacks y pruebas en navegadores reales.

No se incorporarán motor de física, WebGPU ni un sistema de estado global salvo que aparezca una necesidad demostrable durante la implementación.

## Carga, rendimiento y recuperación

La optimización se divide en tres problemas independientes:

1. Descarga inicial del bundle y recursos.
2. Costo de renderizado por frame.
3. Consultas de datos de la aplicación.

Requisitos del Corte 1:

- Three.js, R3F, Drei y la escena se cargan mediante un chunk diferido y no forman parte del bundle inicial.
- Mientras carga el chunk se muestra una portada 2D coherente con Atalaya.
- La escena inicial transfiere menos de 3 MB de recursos.
- Se limitan DPR, luces, sombras y costo del shader.
- Las geometrías repetidas se comparten o instancian.
- Existen perfiles de calidad normal y reducido.
- El objetivo es 60 FPS en equipos desktop normales y al menos 45 FPS en el perfil reducido.
- Ante degradación sostenida se baja primero la calidad.
- Si el perfil reducido permanece por debajo de 30 FPS durante 10 segundos, se cambia automáticamente al modo clásico.
- WebGL no disponible, error del chunk, error de render o pérdida irrecuperable del contexto también activan el modo clásico.
- Safari debe validarse en un dispositivo real; Playwright WebKit no lo sustituye por completo.

## Implementación por cortes

### Corte 1 — Desktop

Incluye la escena, cinco ventanas, fichas operativas, barcos y boya funcionales, clima, rutas, command palette, shell fino, modo clásico, accesibilidad, performance y fallbacks desktop.

Su roadmap exclusivo está en `v2-corte-1-plan.md`.

### Corte 2 — Mobile clásica

Rediseñará la experiencia clásica para mobile. No será simplemente el responsive actual de v1. Su alcance y sprints se definirán después de cerrar el Corte 1.

### Corte 3 — Mobile inmersiva y elección post-login

Adaptará el faro a controles táctiles, hitboxes mayores y perfiles según capacidad. Incorporará detección de GPU, elección explícita post-login y preferencia persistente. Se planificará junto con el Corte 2 en un archivo nuevo posterior.

## Backlog visual posterior

El Corte 1 mantiene el contenido de Eventos, Bitácora, Reportes y Estado del sistema. Queda registrado como trabajo futuro el rediseño de sus layouts para que toda la aplicación alcance el mismo nivel de identidad que el faro y las fichas operativas.

Ese rediseño no debe introducirse lateralmente durante los sprints del Corte 1.

## No-objetivos del Corte 1

- No agregar lógica de negocio, tablas ni endpoints exclusivos para la escena.
- No sumar Atalaya a reportes, eventos o deploys como si fuera una quinta aplicación de negocio.
- No rediseñar los layouts funcionales de v1.
- No implementar mobile clásica ni mobile inmersiva.
- No incorporar elección persistente post-login.
- No buscar realismo cinematográfico.
- No usar mocks permanentes como fuente de salud.

## Instrucción para quien implemente

El detalle técnico del 3D puede ser resuelto autónomamente, pero no están abiertos a reinterpretación:

- El mapeo entre ventanas, embarcaciones y destinos.
- La semántica de salud definida en este documento.
- La dirección de diorama nocturno low-poly.
- La identidad diferenciada de los protagonistas frente a los barcos decorativos.
- Las rutas y fichas operativas.
- La command palette como vía rápida siempre disponible.
- El modo clásico y los fallbacks.
- La accesibilidad equivalente fuera del Canvas.
- Los límites entre Corte 1 y los cortes mobile posteriores.
