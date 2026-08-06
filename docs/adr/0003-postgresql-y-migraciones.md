# ADR 0003: PostgreSQL y estrategia de migraciones

- Estado: aceptado
- Fecha: 2026-08-06

## Contexto

Watchman será dueño de los datos y necesita idempotencia, polling incremental y
trabajos durables. El esquema debe evolucionar sin depender de un ORM y las mismas
migraciones deben funcionar en desarrollo, CI y Railway.

## Decisión

- Usar PostgreSQL 18.4 en desarrollo local.
- Escribir migraciones SQL incrementales en `apps/watchman/migrations/`, para que
  formen parte del contexto de build aislado de Watchman en Railway.
- Ejecutarlas con `goose` y numeración secuencial.
- Ejecutar cada migración dentro de una transacción salvo excepción documentada.
- Usar `timestamptz` y almacenar todos los instantes en UTC.
- Usar UUID para identidades internas y claves externas únicas para idempotencia.
- Representar estados y proveedores con `text` más restricciones `CHECK`; evitamos
  enums nativos hasta conocer su ritmo real de evolución.
- Mantener secretos fuera de las tablas de configuración. La base solo almacena
  referencias e identificadores no sensibles.

En Railway, las migraciones se ejecutarán como pre-deploy command de Watchman. No
se ejecutarán silenciosamente durante el arranque de cada réplica.

## Límites de la primera migración

Incluye:

- aplicaciones e integraciones;
- checkpoints de polling;
- grupos y eventos normalizados;
- jobs de interpretación e interpretaciones terminadas.

No incluye todavía usuarios, sesiones, notificaciones, reportes, incidentes,
deploy markers ni diagramas. Esas entidades se agregarán junto con la feature que
permita validar su modelo.

## Consecuencias

- El esquema y sus cambios son revisables como SQL.
- Los rollbacks de desarrollo están disponibles mediante bloques `Down`.
- En producción se priorizarán migraciones correctivas hacia adelante cuando un
  rollback pueda destruir datos.
- Los workers podrán reclamar jobs concurrentemente mediante
  `FOR UPDATE SKIP LOCKED` sin introducir otro broker.
