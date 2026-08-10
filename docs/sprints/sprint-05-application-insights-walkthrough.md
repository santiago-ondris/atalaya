# Sprint 5 — Notizap y Application Insights

**Estado:** completado el 2026-08-10 con una excepción real del recurso
`notizap-insights` verificada de punta a punta.

## Resultado implementado

Watchman incorpora una segunda implementación de `domain.ErrorSource` para Azure
Monitor Logs. El resto del pipeline no conoce KQL ni tipos de Azure: recibe el mismo
`domain.Event` que ya recibía desde Sentry, lo persiste, crea el job durable de
interpretación y aplica las mismas políticas de Telegram.

El cliente realiza dos llamadas:

1. solicita un token mediante OAuth 2.0 client credentials a Microsoft Entra;
2. ejecuta una consulta `POST` sobre el workspace de Log Analytics.

El token se conserva en memoria hasta un minuto antes de su vencimiento. Los
secretos solo llegan por variables de entorno y nunca se escriben en
`integrations.external_identifier`; allí se guarda únicamente el Workspace ID.

## Consulta KQL

La consulta versionada en `internal/applicationinsights/client.go` usa la tabla
workspace-based de Application Insights:

```kusto
AppExceptions
| where _ResourceId =~ "<RESOURCE_ID_DE_NOTIZAP_INSIGHTS>"
| project TimeGenerated,
          ItemId=tostring(column_ifexists("ItemId", "")),
          OperationId=tostring(column_ifexists("OperationId", "")),
          AppRoleName=tostring(column_ifexists("AppRoleName", "")),
          AppVersion=tostring(column_ifexists("AppVersion", "")),
          ExceptionType=tostring(column_ifexists("ExceptionType", "Exception")),
          OuterMessage=tostring(column_ifexists("OuterMessage", "")),
          InnermostMessage=tostring(column_ifexists("InnermostMessage", "")),
          Details=tostring(column_ifexists("Details", "")),
          Properties=tostring(column_ifexists("Properties", ""))
| order by TimeGenerated asc
```

La API recibe además un `timespan` absoluto. Mantener la selección de columnas en
el adapter hace que cambios de nombres o tablas en Azure no alcancen al dominio.

## Checkpoint e idempotencia

Las columnas salvo `TimeGenerated` se resuelven con `column_ifexists`: algunos
workspaces no exponen `ItemId` u otros campos opcionales. Esto permite conservar
un único adapter sin acoplar el dominio al esquema particular de Notizap.

El cursor es el extremo superior UTC de la última ventana consultada. Cada nuevo
poll retrocede cinco minutos (configurable) respecto del cursor. Este solapamiento
permite capturar telemetría que llegó tarde a Log Analytics.

Volver a recibir filas ya vistas es intencional y seguro: `ItemId` se usa como
`source_event_id`, protegido por la unicidad `(integration_id, source_event_id)`.
Si una fila no incluye `ItemId`, se genera un SHA-256 determinístico con timestamp,
operación, tipo y mensaje.

## Configuración y permisos

Se requieren juntos:

- `AZURE_TENANT_ID`: tenant de Microsoft Entra;
- `AZURE_CLIENT_ID`: Application (client) ID del service principal;
- `AZURE_CLIENT_SECRET`: valor del secreto, no su identificador;
- `AZURE_LOG_ANALYTICS_WORKSPACE_ID`: GUID visible en las propiedades del workspace.
- `AZURE_APPLICATION_INSIGHTS_RESOURCE_ID`: Resource ID completo de
  `notizap-insights`, usado para excluir otros recursos conectados al workspace.

El service principal debe tener el rol **Reader** sobre el workspace de Log
Analytics de Notizap. La propagación de RBAC puede tardar y producir respuestas
403 transitorias inmediatamente después de asignar el rol.

Variables opcionales: `AZURE_QUERY_OVERLAP_SECONDS`, `NOTIZAP_COMPONENT`,
`NOTIZAP_DISPLAY_NAME` y `NOTIZAP_ENVIRONMENT`. Los endpoints alternativos solo
existen para pruebas o nubes soberanas y no necesitan configurarse en Azure público.

## Validación local

```bash
make test
make lint
make build
```

Las pruebas cubren autenticación, scope OAuth, formato de ventana, mapeo de filas,
cache del token, ID alternativo estable y sanitización de errores del proveedor.

## Validación real

El service principal fue validado contra Azure y Watchman completó polls reales.
El workspace de Notizap no expone `ItemId`, lo que motivó el uso de columnas
opcionales y activó el ID determinístico previsto por el adapter.

La primera consulta al workspace sin filtro reveló que `DefaultWorkspace` también
contenía telemetría del recurso `func-dairydelivery`. Dos eventos de ese recurso
recorrieron el pipeline antes de detectar el alcance incorrecto; luego fueron
eliminados junto con sus registros derivados.

La consulta quedó limitada mediante `_ResourceId` al recurso
`notizap-insights`. El primer poll con ese filtro completó correctamente con cero
eventos recibidos y la validación final con una excepción real confirmó el flujo
de persistencia, interpretación y alerta con el alcance definitivo.

## Retrospectiva breve

- Funcionó bien reutilizar el contrato `ErrorSource`: el dominio y los workers no
  necesitaron conocer KQL ni tipos de Azure.
- La prueba real descubrió que un workspace puede reunir varios recursos de
  Application Insights. Consultar solamente por Workspace ID no aísla una app.
- A partir de este sprint, cada integración de Application Insights debe declarar
  también su Resource ID y filtrar `_ResourceId` explícitamente.
- El solapamiento temporal más la unicidad del evento continúan siendo la defensa
  frente a retrasos de ingestión y filas repetidas.

Si Notizap todavía utiliza Application Insights clásico y la tabla disponible es
`exceptions` en lugar de `AppExceptions`, solo debe cambiar la consulta y el mapeo
dentro del adapter antes de la prueba real.
