# ADR 0008 — Incidentes y deploy markers

## Decisión

Los incidentes agrupan `error_groups` de una sola aplicación y conservan una
bitácora append-only. Un grupo solo puede pertenecer a una investigación activa,
pero puede reutilizarse si el problema reaparece después de un cierre.

Los deploys se normalizan en PostgreSQL y solo representan cambios exitosos de
producción. Railway notifica mediante webhook nativo, GitHub Actions mediante un
contrato autenticado y el Command Center cubre despliegues manuales.

## Consecuencias

- La correlación mostrada es temporal, nunca una afirmación automática de causa.
- Resolver, marcar ruido o reabrir exige una conclusión y queda auditado.
- Los reintentos de proveedores no duplican markers.
- Las URLs y tokens de ingesta son secretos operativos, no datos persistidos.
