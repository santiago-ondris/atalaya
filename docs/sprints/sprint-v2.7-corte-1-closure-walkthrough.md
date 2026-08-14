# Sprint V2.7 — Cierre documental del Corte 1

## Resultado

El Corte 1 desktop quedó cerrado el 14 de agosto de 2026 sin cambios de producto
ni una nueva ronda de validación manual. V2.7 consolida la evidencia publicada de
V2.0–V2.6, actualiza la documentación operativa y congela el roadmap. No incorpora
Playwright ni atribuye cobertura a navegadores que no fueron verificados.

## Decisiones de cierre

- Las aprobaciones incrementales en Safari real y Chrome se consideran evidencia
  suficiente para cerrar este corte.
- No se repitieron auditorías visuales, accesibles, funcionales ni Safari ya
  aprobadas.
- Las cifras de FPS de la comprobación humana de V2.6 no quedaron registradas; su
  resultado se conserva como cualitativo y no se reconstruyen valores.
- Firefox real, Playwright en Chromium/Firefox/WebKit y una suite browser
  transversal quedan como deuda explícita.
- No cambiaron rutas, APIs, tipos, componentes, dependencias ni comportamiento de
  producción.

## Evidencia por sprint

| Sprint | Evidencia consolidada                                                                                                                         |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------- |
| V2.0   | Rutas reales, guard de autenticación, retorno a deep link, catálogo único, shell fino y modo clásico; suite, lint, build y formato aprobados. |
| V2.1   | Cinco fichas, salud normalizada, frescura, reportes y tolerancia a errores parciales cubiertos por tests.                                     |
| V2.2   | Escena y chunk diferido, encuadres Chrome aprobados y carga, resize, órbita, zoom y animación validados en Safari real.                       |
| V2.3   | Exactamente nueve destinos y correspondencia única de rutas; mouse, teclado, foco y animaciones validados en Safari real.                     |
| V2.4   | Cinco ventanas conectadas a sus fuentes, tres climas, stale y recuperación; 34 tests y checkpoint Safari real aprobados.                      |
| V2.5   | Command palette con nueve entradas, ambos modos, tres shells, foco y movimiento reducido; 51 tests y checkpoint Safari real aprobados.        |
| V2.6   | Calidad normal/reducida, umbrales, fallbacks y retry; comprobación humana cualitativa en Safari real y Chrome.                                |

Los walkthroughs individuales en este directorio son la fuente detallada de cada
afirmación. Las cifras históricas de tests reflejan el momento de cada sprint; la
suite consolidada actual se informa abajo.

## Confirmación funcional consolidada

- Nueve destinos: cinco fichas —Farmami, Wheels House, Prensap, Notizap y
  Atalaya— más Eventos, Bitácora, Reportes y Estado del sistema.
- Cinco fichas: las cuatro aplicaciones de negocio separan disponibilidad,
  observabilidad y reporte; Atalaya muestra su salud interna y colas.
- Tres climas: despejado, bruma y tormenta, derivados de la peor severidad de las
  cinco ventanas.
- Command palette: las mismas nueve rutas, búsqueda por nombre/alias, teclado y
  devolución de foco desde faro, shell fino y modo clásico.
- Ambos modos: la selección dura la sesión, logout la limpia y el resguardo bajo
  901 px o sin puntero fino abre la vista clásica.
- Fallbacks: WebGL ausente, error de import/render, pérdida de contexto y bajo
  rendimiento conservan navegación, sesión y command palette y permiten retry.

Esta confirmación reutiliza tests y validaciones publicadas; no representa una
nueva prueba browser ejecutada en V2.7.

## Verificación automatizada final

Ejecutada el 14 de agosto de 2026 en `apps/command-center`:

| Comando                        | Resultado                                                                                                                      |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `npm run test`                 | 14 archivos y 59 tests aprobados. jsdom informó su aviso esperado al consultar `HTMLCanvasElement.getContext`; no hubo fallos. |
| `npm run lint`                 | Aprobado.                                                                                                                      |
| `npm run format:check`         | Aprobado.                                                                                                                      |
| `npm run build`                | Aprobado; 1.283 módulos transformados. Vite mantuvo el warning informativo de chunks mayores a 500 kB.                         |
| `npm audit --audit-level=high` | 0 vulnerabilidades.                                                                                                            |
| `npm run build:report`         | Aprobado. Inicial: 725,7 kB bruto / 211,8 kB gzip. `LighthouseScene`: 899,5 kB bruto / 235,7 kB gzip.                          |

El stack 3D permanece separado del bundle inicial y el chunk de escena queda muy
por debajo del presupuesto de 3 MB transferidos.

## Validación browser acumulada y deuda

Safari real respaldó progresivamente composición, cámara, navegación por mouse y
teclado, foco, estados, climas, recuperación, command palette y movimiento
reducido. Chrome respaldó los checkpoints visuales deterministas y la comprobación
cualitativa de rendimiento/recuperación de V2.6.

No se validó Firefox real y no existe suite Playwright en Chromium, Firefox o
WebKit. Tampoco existe aún una regresión browser transversal de rutas, accesibilidad
y fallbacks. Esta cobertura debe planificarse como mejora futura y no se presume a
partir de Safari o jsdom.

## Cierre

El roadmap `v2-corte-1-plan.md` queda congelado como registro histórico. Los
aprendizajes y preguntas que condicionan el trabajo mobile están en
`docs/handoffs/cortes-2-3-handoff.md`; ese documento no define sprints nuevos.
