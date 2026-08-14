# Atalaya v2 — Roadmap del Corte 1 Desktop (cerrado)

> **Estado:** Corte 1 completado el 14 de agosto de 2026. Este roadmap queda
> congelado como registro histórico; sólo admite correcciones factuales. El cierre
> y su evidencia consolidada están en
> `docs/sprints/sprint-v2.7-corte-1-closure-walkthrough.md`.

## Propósito y límite del plan

Este archivo contiene **exclusivamente los sprints del Corte 1 de Atalaya v2**. Su objetivo es entregar la experiencia desktop completa del faro 3D sin mezclar trabajo mobile ni el rediseño general de las páginas funcionales.

El handoff documental para el trabajo posterior está en
`docs/handoffs/cortes-2-3-handoff.md`. Cuando se decida iniciar esa etapa se creará
un archivo de planificación nuevo para:

- **Corte 2:** experiencia mobile clásica rediseñada.
- **Corte 3:** experiencia mobile inmersiva, detección de capacidad y elección post-login.

No se deben agregar anticipadamente esos sprints a este documento. Las decisiones de producto y dirección visual del Corte 1 están en `v2-visual.md`.

## Principios de ejecución

- Cada sprint termina en un resultado demostrable, no en infraestructura aislada sin uso.
- Se preserva el comportamiento de la v1 mientras se reemplaza gradualmente su navegación interna.
- Los datos del faro provienen de APIs reales existentes; los fixtures se usan solo en tests y demos controladas.
- Los cambios de arquitectura se documentan cuando afectan contratos o decisiones futuras.
- La escena 3D se construye como una capa reemplazable: autenticación, rutas, paleta y modo clásico deben seguir funcionando si WebGL falla.
- No se rediseñan lateralmente Eventos, Bitácora, Reportes ni Estado del sistema.
- Cada sprint parte de una rama corta y concluye con tests, lint, build, demo y walkthrough breve.
- Super importante mantener ordenado el codigo y los archivos.

## Definition of Done general

Un sprint está terminado cuando:

- Cumple todos sus criterios de aceptación.
- Tiene tests proporcionales a sus riesgos.
- Pasa `npm run test`, `npm run lint` y `npm run build`.
- No rompe login, `/status` ni las pantallas funcionales existentes.
- Maneja estados de carga, error y ausencia de datos relevantes.
- No agrega assets sin licencia o atribución comprobada.
- Su demo puede repetirse desde un checkout limpio.
- Actualiza documentación y registra deuda descubierta sin incorporarla silenciosamente al siguiente sprint.

---

## Sprint V2.0 — Rutas y cimientos de navegación

### Estado

✅ Completado.

### Objetivo

Reemplazar la navegación privada mantenida únicamente en estado React por URLs reales y preparar la convivencia entre experiencia inmersiva y modo clásico.

### Dependencias

Ninguna. Es el primer sprint del Corte 1.

### Alcance

- Incorporar React Router en modo declarativo.
- Definir las rutas `/`, `/overview`, `/apps/:appSlug`, `/events`, `/events/:eventId`, `/operations`, `/reports`, `/architecture/:appSlug?`, `/system` y `/status`.
- Crear un guard de autenticación para las rutas privadas.
- Conservar y recuperar el deep link solicitado después del login.
- Migrar `activeView`, `selectedEventId` y `selectedArchApp` a parámetros o rutas.
- Crear un catálogo frontend único para las cinco aplicaciones, incluyendo nombres, aliases, diagrama, orden y metadatos visuales.
- Reutilizar ese catálogo en Arquitectura, formateo y navegación; eliminar listas duplicadas cuando sea seguro.
- Crear el shell fino con regreso al faro, sección actual, acceso a la paleta y logout.
- Conservar `AppLayout` y el overview actual como experiencia clásica completa.
- Implementar el modo `immersive|classic` en `sessionStorage` y limpiarlo durante logout.
- Enviar temporalmente a modo clásico los viewports menores a 901 px o sin puntero fino.
- Mantener el fallback SPA existente de Nginx y Cloudflare Pages para las rutas nuevas.
- Extraer el catálogo exportado desde `ArchitecturePage.tsx` y resolver el warning preexistente de Fast Refresh.

### Fuera de alcance

- Canvas o dependencias 3D.
- Fichas con datos operativos completos.
- Persistencia del modo entre sesiones.
- Rediseño mobile.

### Criterios de aceptación

- Cada ruta puede abrirse directamente y sobrevivir un refresco.
- Atrás y adelante del navegador recorren destinos reales.
- Un usuario no autenticado vuelve al deep link pedido después del login.
- `/status` continúa público y no renderiza el shell privado.
- Cerrar sesión elimina el modo guardado y bloquea nuevamente las rutas privadas.
- Las páginas v1 conservan su comportamiento.

### Pruebas y demo

- Tests del catálogo, aliases y slugs inválidos.
- Tests del guard de autenticación y retorno post-login.
- Tests de ambos shells y selección durante la sesión.
- Demo: abrir un evento y una arquitectura mediante URL, refrescar, usar atrás/adelante y alternar entre v1 clásica y shell fino.

---

## Sprint V2.1 — Salud normalizada y fichas operativas

### Estado

✅ Completado.

### Objetivo

Crear los cinco destinos funcionales del faro y una semántica de salud coherente antes de sumar la escena 3D.

### Dependencias

Sprint V2.0 completado.

### Alcance

- Definir tipos frontend para catálogo, salud normalizada, severidad visual, freshness y nivel de clima.
- Consultar en paralelo `/api/v1/public/status`, `/api/v1/system/health` y `/api/v1/reports`.
- Normalizar disponibilidad de las cuatro aplicaciones a verde, amarillo o rojo.
- Normalizar salud interna de Atalaya con reglas propias.
- Calcular la peor severidad agregada sin mezclar salud de ingesta de `/overview`.
- Refrescar salud cada 30 segundos y cancelar trabajo obsoleto al desmontar.
- Conservar último dato conocido y marcarlo desactualizado ante errores posteriores.
- Crear ficha común de Farmami, Wheels House, Prensap y Notizap con componentes, última comprobación, último reporte y diagrama.
- Crear ficha específica de Atalaya con señales, colas, diagrama y acceso a Sistema.
- Resolver slugs desconocidos mediante una pantalla 404 o redirección explícita, sin seleccionar silenciosamente la primera aplicación.
- Mantener los diagramas existentes dentro de sus iframes en este corte.

### Fuera de alcance

- Nuevo endpoint agregado o migración de base de datos.
- Incorporar Atalaya a los reportes diarios.
- Escena o clima visual.

### Criterios de aceptación

- Las cinco fichas funcionan con datos reales.
- Un fallo en reportes no oculta disponibilidad ni diagrama.
- Un fallo en salud no elimina enlaces ni bloquea la ficha.
- El último reporte se selecciona correctamente por aplicación.
- Atalaya no muestra una tarjeta de reporte vacía: muestra salud interna útil.

### Pruebas y demo

- Tests puros de normalización y peor severidad.
- Tests de frescura, errores parciales y selección de reportes.
- Tests de las dos variantes de ficha.
- Demo con datos reales y fixtures controlados para `operational`, `degraded`, `unknown`, `major_outage` y API inaccesible.

---

## Sprint V2.2 — Fundaciones visuales del diorama

### Estado

✅ Completado.

### Objetivo

Entregar una primera escena desktop cargada de forma diferida, con composición y lenguaje visual aprobables antes de invertir en todos los objetos funcionales.

### Dependencias

Sprint V2.0. Puede consumir el modelo de V2.1, pero su validación artística no depende de datos vivos.

### Alcance

- Incorporar `three`, `@react-three/fiber` 9 y `@react-three/drei` con versiones compatibles con React 19.
- Crear un límite de carga y error alrededor de un chunk 3D diferido.
- Diseñar una portada 2D para el tiempo de carga.
- Construir Canvas, cámara, iluminación, fondo y composición nocturna.
- Construir el faro low-poly y reservar cinco ventanas identificables.
- Implementar el haz con una vuelta de 12 segundos.
- Implementar mar low-poly con desplazamiento de vértices liviano.
- Documentar origen y licencia si se adapta código de shader.
- Configurar `OrbitControls` con damping, órbita y zoom limitados y sin pan.
- Crear la familia base de barcos decorativos pequeños y blancos.
- Establecer DPR máximo, política de sombras, cantidad de luces y reutilización de geometrías.
- Añadir un modo determinista de escena para tests y capturas.

### Decisiones no reabribles durante el sprint

- Diorama nocturno low-poly.
- WebGL, no WebGPU.
- Geometrías propias como primera opción.
- Sin motor de física.
- Sin cámara libre.

### Criterios de aceptación

- La escena presenta una composición clara al cargar sin mover la cámara.
- El usuario no puede perder el faro ni atravesar la escena con los controles.
- El stack 3D no aparece en el chunk inicial.
- La animación se mantiene estable sin datos reales.
- Ningún asset carece de licencia documentada.

### Pruebas y demo

- Test del límite de carga/error sin depender de WebGL real.
- Análisis de chunks del build.
- Capturas deterministas del encuadre inicial.
- Validación manual temprana en Chrome, Firefox, WebKit y Safari real.
- Demo: carga limpia, órbita limitada, zoom y presentación artística base.

---

## Sprint V2.3 — Destinos narrativos e interacción

### Estado

✅ Completado.

### Objetivo

Convertir el diorama en la navegación principal completa de Atalaya.

### Dependencias

Sprints V2.0 y V2.2 completados. Las rutas de V2.1 deben estar disponibles para cerrar la navegación de ventanas.

### Alcance

- Construir las cinco ventanas funcionales del faro.
- Construir el patrullero o pesquero de Eventos con reflector y patrón de búsqueda.
- Construir el mercante de Bitácora con ruta de entrada.
- Construir el paquebote de Reportes con comportamiento propio.
- Construir la boya de Estado del sistema con señal periódica.
- Diferenciar protagonistas mediante silueta, escala, equipamiento, iluminación y movimiento.
- Mantener homogénea y secundaria la flota decorativa.
- Crear hitboxes simples, estables y separadas de las mallas visibles.
- Evitar que objetos decorativos intercepten raycasts o clicks.
- Añadir cursor de enlace y etiqueta discreta al hover/foco.
- Conectar clicks y activaciones con las nueve rutas.
- Crear una capa DOM equivalente para teclado y tecnologías de asistencia.
- Asegurar que el foco no quede atrapado en el Canvas.

### Fuera de alcance

- Color dinámico de ventanas.
- Clima reactivo.
- Command palette completa.

### Criterios de aceptación

- Los nueve destinos pueden abrirse con mouse.
- Los mismos nueve destinos pueden abrirse con teclado sin depender de raycasting.
- Los tres barcos funcionales se distinguen por diseño aunque se vean en escala de grises.
- Las embarcaciones decorativas no parecen destinos ni bloquean interacción.
- Las etiquetas no ensucian la escena cuando no existe hover o foco.

### Pruebas y demo

- Tests de la definición de hotspots y su correspondencia uno a uno con las rutas.
- Tests de la navegación DOM accesible.
- Pruebas browser de mouse, teclado y foco.
- Demo: recorrido completo desde el diorama a cinco fichas y cuatro páginas funcionales.

---

## Sprint V2.4 — Estado vivo y clima operativo

### Estado

✅ Completado y validado en Safari real el 13 de agosto de 2026.

### Objetivo

Hacer que el faro comunique disponibilidad y salud reales sin convertir el 3D en una fuente adicional de lógica de negocio.

### Dependencias

Sprints V2.1 y V2.3 completados.

### Alcance

- Conectar las cuatro ventanas de aplicaciones a `/public/status` normalizado.
- Conectar la ventana de Atalaya a `/system/health` normalizado.
- Aplicar verde, amarillo y rojo según las reglas de `v2-visual.md`.
- Derivar clima despejado, bruma o tormenta del peor estado.
- Implementar transiciones ambientales discretas entre estados.
- Mantener hotspots y etiquetas legibles bajo todos los climas.
- Actualizar materiales y atmósfera sin reconstruir la escena completa.
- Representar datos desconocidos y desactualizados de forma distinguible.
- Evitar que un error de fetch provoque un error del Canvas.

### Criterios de aceptación

- Las cinco ventanas reflejan la fuente correcta y no la salud de ingesta de `/overview`.
- Una caída mayor domina el clima sobre cualquier warning.
- `unknown` produce atención, no una falsa señal verde.
- La escena continúa navegable cuando una o ambas APIs fallan.
- El refresco periódico no reinicia cámara ni animaciones.

### Pruebas y demo

- Matriz de estados y clima mediante fixtures.
- Test de actualizaciones sucesivas y datos desactualizados.
- Prueba de error parcial y recuperación posterior.
- Demo controlada recorriendo los tres climas y un cambio de estado en vivo.

---

## Sprint V2.5 — Command palette, modo clásico y accesibilidad

### Estado

✅ Completado y aprobado en Safari real el 14 de agosto de 2026.

### Objetivo

Garantizar que la identidad inmersiva nunca comprometa velocidad, orientación ni acceso.

### Dependencias

Sprints V2.0 y V2.3 completados.

### Alcance

- Incorporar `cmdk` y estilizarlo con la identidad Atalaya.
- Añadir exactamente cinco fichas y cuatro páginas funcionales.
- Incorporar nombres y aliases útiles para búsqueda.
- Abrir con `⌘K` o `Ctrl+K` y cerrar con `Escape`.
- Mantener la paleta disponible desde faro, shell fino y modo clásico.
- Implementar controles visibles “Vista clásica” y “Volver al faro”.
- Completar la persistencia de modo durante la sesión.
- Mantener command palette y modo clásico utilizables cuando el chunk 3D falla.
- Respetar `prefers-reduced-motion` en haz, barcos, oleaje y clima.
- Validar foco inicial, devolución de foco, orden de tabulación, nombres accesibles y contraste.
- Verificar que lectores de pantalla reciban destino y estado sin interpretar el Canvas.

### Criterios de aceptación

- Cualquier destino se alcanza desde el teclado en pocos pasos.
- La búsqueda encuentra las nueve entradas por nombre o alias esperado.
- El foco vuelve al disparador al cerrar la paleta.
- El modo elegido dura la sesión y se limpia al cerrar sesión.
- Reducir movimiento no elimina estado ni navegación.
- El modo clásico conserva sidebar, overview y páginas v1 completas.

### Pruebas y demo

- Tests de atajos, búsqueda, selección y foco de `cmdk`.
- Tests de persistencia y logout.
- Tests con `prefers-reduced-motion` simulado.
- Auditoría de teclado y lector de pantalla.
- Demo: operar toda la navegación sin mouse, alternar modos y repetir con movimiento reducido.

---

## Sprint V2.6 — Rendimiento, adaptación y recuperación

### Estado

✅ Completado el 14 de agosto de 2026.

### Objetivo

Mantener una carga inicial prolija y garantizar una salida recuperable cuando el dispositivo no pueda sostener la experiencia 3D.

### Dependencias

Escena funcional hasta V2.5.

### Alcance

- Medir bundle inicial, chunk 3D, recursos transferidos y frame rate.
- Mantener el stack 3D fuera del chunk inicial.
- Mantener recursos iniciales de escena por debajo de 3 MB transferidos.
- Implementar perfiles normal y reducido.
- Reducir en el perfil bajo DPR, flota decorativa, costo del mar, luces y efectos atmosféricos.
- Compartir o instanciar geometrías y materiales repetidos.
- Usar monitoreo de rendimiento con ventanas estables para evitar cambios de calidad espasmódicos.
- Bajar primero al perfil reducido ante degradación sostenida.
- Pasar al modo clásico si el perfil reducido permanece por debajo de 30 FPS durante 10 segundos.
- Implementar fallback ante WebGL ausente, error al importar el chunk, error de render y pérdida irrecuperable del contexto.
- Comunicar el fallback con un mensaje breve y permitir reintentar manualmente el faro.
- Evitar loops automáticos entre faro y clásico.

### Criterios de aceptación

- Objetivo de 60 FPS en desktop típico.
- El perfil reducido alcanza al menos 45 FPS en el equipo de referencia antes de considerar fallback.
- Un dispositivo no compatible llega siempre a una UI clásica operativa.
- Las condiciones de fallback pueden reproducirse en pruebas.
- La navegación, sesión y command palette sobreviven al desmontaje del Canvas.

### Pruebas y demo

- Test de selección de perfiles y temporización del umbral.
- Test de cada clase de fallo sin crear un contexto WebGL real en jsdom.
- Pruebas browser con WebGL deshabilitado o simulado.
- Reporte reproducible de bundle y recursos.
- Demo: degradación normal → perfil reducido → clásico, además de error de carga y reintento.

---

## Sprint V2.7 — Validación integral y cierre del Corte 1

### Estado

✅ Completado el 14 de agosto de 2026 como cierre documental.

### Objetivo

Consolidar la evidencia ya aprobada de V2.0–V2.6, cerrar documentalmente la
experiencia desktop y dejar explícita la cobertura browser pendiente antes de
planificar mobile.

### Dependencias

Todos los sprints anteriores completados.

### Alcance

- Consolidar los walkthroughs, tests y aprobaciones manuales de V2.0–V2.6.
- Confirmar documentalmente los nueve destinos, cinco fichas, tres climas,
  command palette, modos inmersivo y clásico y fallbacks recuperables.
- Reutilizar las validaciones Safari real y Chrome ya realizadas, sin abrir una
  nueva ronda visual, accesible o funcional.
- Ejecutar `test`, `lint`, `format:check`, `build`, `audit` y `build:report`.
- Revisar el presupuesto reproducible de transferencia sin inventar métricas de FPS.
- Actualizar README, documentación de operación y troubleshooting.
- Crear un walkthrough del Corte 1 con decisiones, evidencia, pruebas y deuda conocida.
- Crear un handoff breve con aprendizajes y preguntas para Cortes 2 y 3, sin
  definir nuevos sprints.
- Registrar Firefox real, Playwright Chromium/Firefox/WebKit y una suite browser
  transversal como deuda futura explícita.

### Criterios de aceptación

- Las suites existentes pasan y sus resultados quedan publicados en el walkthrough.
- La evidencia acumulada confirma los nueve destinos, las cinco fichas y los tres climas.
- Command palette, ambos modos y fallbacks quedan respaldados por tests y
  validaciones ya aprobadas.
- Las métricas de bundle son reproducibles y mantienen el stack 3D diferido.
- La cobertura browser no ejecutada queda declarada como deuda, no como validación.
- README, runbook, walkthrough y handoff permiten operar y continuar el proyecto.

### Evidencia de cierre

El recorrido funcional, la accesibilidad, los estados vivos, los tres climas y la
recuperación fueron validados incrementalmente en V2.0–V2.6. V2.7 no repite esa
demo: consolida sus walkthroughs y ejecuta únicamente las suites existentes. La
trazabilidad completa está en el walkthrough final del Corte 1.

### Cierre formal

Con el walkthrough publicado:

- El Corte 1 y V2.7 quedan completados.
- Este roadmap queda congelado como registro histórico, salvo correcciones factuales.
- El handoff no constituye un nuevo roadmap ni define sprints para Cortes 2 y 3.
- La planificación futura deberá usar los aprendizajes desktop para decidir
  controles táctiles, perfiles de GPU, selector post-login y persistencia mobile.

---

## Entregas del Corte 1

| Entrega             | Sprints   | Resultado                                              |
| ------------------- | --------- | ------------------------------------------------------ |
| Navegación v2       | V2.0      | Rutas, autenticación, catálogo y doble shell           |
| Destinos operativos | V2.1      | Cinco fichas y semántica de salud consistente          |
| Diorama navegable   | V2.2–V2.3 | Dirección visual, cámara y nueve destinos interactivos |
| Observación viva    | V2.4      | Ventanas y clima conectados a estado real              |
| Operación accesible | V2.5      | Paleta, teclado, movimiento reducido y modo clásico    |
| Resiliencia visual  | V2.6      | Calidad adaptativa, presupuestos y fallbacks           |
| Corte 1 cerrado     | V2.7      | Consolidación documental, checks y deuda browser       |

## Trabajo deliberadamente posterior

- Diseño e implementación de la mobile clásica del Corte 2.
- Faro táctil y perfil mobile reducido del Corte 3.
- Detección de GPU y elección persistente post-login.
- Rediseño de layouts de Eventos, Bitácora, Reportes y Estado del sistema.
- Evaluación de transiciones cinematográficas entre escena y destinos.
- Incorporación de modelos `.glb` únicamente si las geometrías propias demuestran ser insuficientes.
- Validación en Firefox real.
- Playwright en Chromium, Firefox y WebKit.
- Suite browser transversal para navegación, accesibilidad y recuperación.

Estos puntos no deben ampliar informalmente ningún sprint del Corte 1.
