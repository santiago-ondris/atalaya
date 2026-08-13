# Sprint V2.1 — Fichas operativas de aplicaciones

## Resultado

`/apps/:appSlug` es ahora el punto de entrada operativo de cada aplicación. Las fichas de Farmami, Wheels House, Prensap y Notizap mantienen separados los estados de disponibilidad y observabilidad, y reúnen uptime, componentes, integraciones, último parte diario y una tendencia accesible de siete reportes.

Atalaya usa una variante de plataforma: muestra salud interna, señales, colas y consumo mensual, y aclara que no genera reportes diarios de producto.

## Rutas y navegación contextual

- Fichas: `/apps/farmami`, `/apps/wheels_house`, `/apps/prensap`, `/apps/notizap` y `/apps/atalaya`.
- Contexto operativo: `/events?application=:slug`, `/operations?application=:slug` y `/reports?application=:slug`.
- Arquitectura: `/architecture/:slug`.
- Atalaya enlaza sólo a `/system` y `/architecture/atalaya` porque no participa de los datos operativos de producto.

Los aliases continúan resolviéndose con redirección `replace` y los slugs desconocidos mantienen el 404. Eventos, Bitácora y Reportes toman la aplicación desde la URL; sus selectores actualizan el mismo parámetro, por lo que refresh y atrás/adelante restauran el contexto.

## Fuentes y actualización

No se agregó un endpoint agregado ni se modificó el backend. La ficha compone:

- `GET /api/v1/public/status` para disponibilidad, componentes y uptime.
- `GET /api/v1/integrations` para salud de observabilidad y entornos.
- `GET /api/v1/reports?limit=30` para último parte y tendencia.
- `GET /api/v1/system/health` y `GET /api/v1/system/costs` para Atalaya.

Disponibilidad e integraciones se consultan al entrar y cada 60 segundos; el intervalo se cancela al desmontar. Reportes, salud interna y costos se cargan una vez por visita. Cada panel conserva su propio estado de carga y error para evitar que una fuente caída bloquee el resto de la ficha.

Los valores ausentes se presentan como “Sin datos”. También hay estados explícitos para integración deshabilitada, actividad no disponible, ausencia de reportes, uptime desconocido y errores operativos. La tendencia ordena cronológicamente los siete reportes disponibles más recientes, no convierte actividad ausente en cero y ofrece una tabla equivalente para lectores de pantalla y pantallas pequeñas.

## Validación manual

1. Abrir los cinco deep links de ficha en modo clásico y fino.
2. Probar un producto sin reportes, uptime nulo, integración deshabilitada y actividad no disponible.
3. Interrumpir individualmente status, integrations y reports para confirmar los errores parciales.
4. Seguir cada acceso contextual y comprobar el selector, refresh y navegación atrás/adelante.
5. Revisar el gráfico y su tabla accesible en escritorio, tablet y móvil.

## Deuda posterior

- V2.2: alertas y acciones operativas sobre las señales de ficha.
- V2.5: configuración editable y evolución de recolectores/fuentes. La ficha V2.1 permanece deliberadamente de sólo lectura.
