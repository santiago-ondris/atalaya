# Sprint 08 — Incidentes y deploy markers

## Resultado

Atalaya incorpora una Bitácora por aplicación: grafica ocurrencias, superpone
deploys exitosos y permite construir memoria operativa mediante incidentes con
varios grupos de error, notas y conclusiones obligatorias.

## Recorrido técnico

- `00006_incidents_deployments.sql` incorpora incidentes, entradas inmutables,
  relaciones con grupos y el modelo común de deployments.
- `internal/store/operations.go` aplica las invariantes dentro de transacciones,
  incluyendo advisory locks para evitar investigaciones activas duplicadas.
- La API privada expone incidentes, búsqueda de grupos, carga manual y timeline.
- Los hooks separados para GitHub Actions y Railway tienen credenciales distintas.
- `OperationsPage.tsx` carga bajo demanda y usa Recharts para el gráfico responsive,
  con una tabla textual equivalente.

## Configuración externa completada

Se generaron valores distintos para `DEPLOYMENT_INGEST_TOKEN` y
`RAILWAY_WEBHOOK_TOKEN`, y se configuraron markers exitosos para:

- los backends de Prensap, Farmami y Wheels House mediante webhooks de Railway;
- los frontends de Farmami y Wheels House mediante eventos de Vercel y workflows
  de GitHub Actions;
- el frontend de Notizap desde su workflow de Azure Static Web Apps;
- el frontend de Prensap después de su deploy con Wrangler a Cloudflare Pages.

Los workflows externos envían el contrato común únicamente después del éxito:

```yaml
- name: Registrar deploy en Atalaya
  if: success()
  run: |
    curl --fail-with-body -X POST "$ATALAYA_URL/hooks/v1/deployments" \
      -H "Authorization: Bearer $ATALAYA_DEPLOY_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"application\":\"prensap\",\"component\":\"frontend\",\"environment\":\"production\",\"external_id\":\"${GITHUB_RUN_ID}-${GITHUB_RUN_ATTEMPT}\",\"commit_sha\":\"${GITHUB_SHA}\",\"actor\":\"${GITHUB_ACTOR}\"}"
```

El backend manual de Notizap se registra desde “Bitácora → Registrar deploy”, que
queda precargado como `notizap/backend/production`.

El webhook Railway lleva su secreto en el query parameter `token`, no en el path,
para evitar que los access logs de Railway lo muestren. Los redeploys que no
incluyen commit usan el ID estable del deployment como versión de respaldo.

El decoder de webhooks acepta campos adicionales de Railway para mantener
compatibilidad hacia adelante; los contratos privados continúan usando
decodificación estricta.

El Command Center continúa ejecutándose localmente por decisión operativa. Su
despliegue es independiente del backend Watchman y no bloquea el cierre del sprint.

## Validación de producción del 2026-08-11

- PostgreSQL 18 y Watchman quedaron saludables en Railway, con Goose en versión 6.
- Los webhooks reales de los tres backends devolvieron `201`.
- Los workflows reales de los cuatro frontends devolvieron `201`.
- La Bitácora recibió los markers sin duplicados y muestra ambiente, proveedor,
  referencia, actor y enlaces disponibles.

## Validación local del 2026-08-11

- Goose aplicó la migración 6 sobre PostgreSQL 18.4.
- Se generó el marker manual `sprint8-local-validation` para Notizap.
- Un incidente de validación agrupó dos fingerprints de Wheels House, se cerró,
  reabrió con motivo obligatorio y volvió a cerrarse como ruido.
- La prueba real detectó y permitió corregir un cursor abierto de `pgx` durante la
  reapertura; el recorrido completo pasó después de reconstruir Watchman.
- Pasaron tests, vet y formato de Go; Vitest, lint, TypeScript, build y Prettier
  del Command Center; además de los builds Docker.
