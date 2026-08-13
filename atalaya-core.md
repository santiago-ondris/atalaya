# Atalaya — Documento Core

Sistema de observabilidad para 4 aplicaciones en producción (Farmami, Wheels House, Prensapp, Notizap). No es solo un tracker de errores: es un command center completo, pensado también como proyecto de portfolio que demuestre madurez técnica y ownership real sobre producción.

---

## 1. Objetivo del proyecto

- Centralizar monitoreo y error tracking de las 4 apps en un solo lugar.
- Traducir errores técnicos (Sentry / Application Insights) a lenguaje natural vía LLM.
- Alertar en tiempo real y resumir el día por Telegram.
- Visualizar la arquitectura de cada app de forma interactiva.
- Servir como pieza de portfolio: mostrar a recruiters/founders capacidad de diseño de sistemas y cuidado de producción real.

---

## 2. Arquitectura general

Tres servicios, dockerizados:

- **Go — "el watchman"**: pollea las fuentes de error de cada app, normaliza los datos, aplica deduplicación/rate limiting, y dispara alertas y reportes por Telegram.
- **Python — "el interpreter"**: recibe el error ya estructurado, consulta un LLM vía OpenRouter, devuelve una interpretación en lenguaje natural + clasificación de severidad/actionabilidad.
- **React + TS — "el command center"**: web app, incluye el diagrama de arquitectura interactivo y el panel de flujos.

**Deploy:**
- Backend (Go + Python) → Railway.
- Frontend (React) → Cloudflare.

---

## 3. Apps monitoreadas y fuentes de error

| App | Stack | Fuente de errores |
|---|---|---|
| Farmami | — | Sentry (ya configurado) |
| Wheels House | — | Sentry (ya configurado) |
| Prensapp | — | Sentry (ya configurado) |
| Notizap | — | Application Insights (KQL) |

Como las fuentes hablan "idiomas" distintos (Sentry vs. KQL/App Insights), el servicio Go necesita una capa de interpretación/normalización antes de mandar el error al Python service — probablemente vía un `ErrorSource` interface con un adapter por tipo de fuente, unificando todo a un esquema común antes de que llegue al interpreter.

---

## 4. Alertas por Telegram

- **Deduplicación y rate limiting**: si el mismo error (misma app + mismo tipo) se repite dentro de una ventana de tiempo, no se manda una alerta por cada ocurrencia — se agrupa y se informa con un contador ("este error ocurrió X veces en los últimos N minutos").
- Definir en la implementación: tamaño de la ventana, y si el contador resetea o se acumula durante el día.

---

## 5. Reportes diarios por Telegram

- **Horario**: 20:00 ARG, todos los días.
- **Resiliencia**: si el envío falla (el servicio está caído en ese momento), reintentar cada 30 minutos, siempre dentro del mismo día. Si no se logra enviar antes de que termine el día, no se arrastra al día siguiente.
- **Contenido del reporte**: incluye cantidad de aperturas de la app en el día (conteo de sesiones, no de usuarios únicos). Confirmado en scope; el detalle de si sale de Sentry Release Health / session tracking, de Application Insights (page views vía KQL), o si hace falta instrumentación nueva, se resuelve en el momento de implementar (ver sección 9).

---

## 6. LLM / OpenRouter

- El Python service consulta OpenRouter para interpretar cada error.
- Devuelve interpretación en lenguaje natural + severidad + si es accionable o ruido.
- *(Idea aceptada, a definir en implementación)*: trackear consumo/costo de tokens por interpretación y mostrarlo en el command center.

---

## 7. Command Center — Diagrama de arquitectura interactivo

Por cada una de las 4 apps:

- Diagrama de arquitectura interactivo (nodos + aristas).
- Panel de flujos a la derecha.
- Al seleccionar un flujo, se resalta la ruta completa en el diagrama.
- Tooltips por componente.
- Diseño limpio, profesional y responsive.

Confirmado en scope. El modelo de datos que alimenta estos diagramas (config por app con nodos/aristas/flujos) y de dónde lo sirve el backend se define más adelante con Claude Code (ver sección 9) — pero la feature en sí no está en duda.

---

## 8. Features adicionales aceptadas (para diferenciar el proyecto)

- **Deploy markers**: marcar en la línea de tiempo cuándo se deployó cada versión, correlacionado con spikes de errores posteriores.
- **Status page público**: vista simple tipo statuspage.io con estado de las 4 apps (up/down, uptime %, último incidente).
- **Historial de incidentes**: poder marcar un error como resuelto con una nota, armando con el tiempo un log de incidentes/postmortems.
- **Meta-observabilidad**: que Atalaya vigile su propia salud (si el watchman deja de pollear, si el interpreter falla con el LLM).
- **Costo de OpenRouter visible en el command center** (ver punto 6).
- **Login básico** en el command center, aunque sea de un solo usuario, para que se vea como producto terminado.

---

## 9. En scope, detalle pendiente de definir en implementación

Estos dos puntos SÍ son parte del proyecto — están confirmados y no se negocian. Lo único que queda abierto es el detalle técnico de cómo se implementan, que se define al momento de codear (no ahora):

- **Diagrama de arquitectura interactivo**: confirmado que va. Falta definir el modelo de datos exacto (config por app con nodos/aristas/flujos) y de dónde lo sirve el backend.
- **Conteo de aperturas diarias en el reporte**: confirmado que va. Falta definir la fuente exacta (Sentry sessions vs. App Insights page views vs. instrumentación nueva).

---

## 10. Estética — Paleta y tipografía

Dirección: "bitácora náutica" — carta de navegación vieja, banderas de señal, instrumentos de latón. El crema es superficie/papel, no fondo dominante; el verde profundo es el color principal de la marca.

**Paleta:**
- Verde profundo `#005838` — color principal / estado saludable-OK.
- Amarillo bandera `#FFE000` — atención / warning.
- Crema pergamino `#FAFAE4` — superficie, no fondo dominante.
- Índigo `#4750A8` — informativo / flow seleccionado en el diagrama.
- Ocre/latón `#BA8A40` — bordes, metadata, elementos secundarios.
- Rojo ladrillo `#C1432E` — crítico/alerta.
- Negro y blanco reales — máximo contraste de texto donde haga falta.

**Tipografía:**
- Display: **Instrument Serif** — títulos, nombres de apps, hero. Uso restringido.
- Body: **Public Sans** — texto general de la UI.
- Mono: **IBM Plex Mono** — timestamps, stack traces, KQL, IDs de error.

**Firma visual**: los estados de las 4 apps se representan como banderas de señal náutica (formas geométricas simples, no íconos genéricos) — verde = ok, amarillo = atención, rojo ladrillo = crítico.

---
