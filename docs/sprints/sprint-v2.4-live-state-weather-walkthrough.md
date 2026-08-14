# Sprint V2.4 — Estado vivo y clima operativo

## Resultado

El faro comunica la disponibilidad real de Farmami, Wheels House, Prensap y Notizap desde `/api/v1/public/status`, y la salud interna de Atalaya desde `/api/v1/system/health`. El checkpoint de los tres climas fue aprobado y la validación manual final en Safari real se completó el 13 de agosto de 2026. Con ambas comprobaciones V2.4 quedó cerrado.

## Estado y frescura

Un módulo puro normaliza las respuestas en severidad verde, amarilla o roja, frescura `loading`, `fresh` o `stale`, y clima despejado, bruma o tormenta. Un fallo inicial se representa como desconocido amarillo. Ante fallos posteriores cada fuente conserva de manera independiente su último resultado válido y lo marca como desactualizado; una respuesta exitosa recupera la frescura.

Las dos fuentes se consultan en paralelo al montar y cada 30 segundos. Los requests aceptan `AbortSignal`, se cancelan al desmontar o al comenzar una generación nueva y las respuestas obsoletas no pueden reemplazar datos más recientes. No se incorporaron consultas a `/overview`, endpoints ni cambios de backend.

## Comunicación visual

Las cinco ventanas reciben únicamente el estado normalizado. El vidrio interior usa el color de severidad, el hover o foco refuerza la ventana activa y la frescura desactualizada reduce la intensidad y añade un aro ámbar discontinuo. La capa DOM expone nombre, estado comprensible y la indicación “datos desactualizados”; los cuatro destinos narrativos conservan sus nombres sin estados ficticios.

El peor estado deriva uno de tres climas:

- Despejado, con cielo limpio, estrellas presentes, niebla baja y mar nocturno.
- Bruma, con estrellas atenuadas, nubes suaves, niebla media, luz difusa y agua desaturada.
- Tormenta contenida, con cielo y agua oscuros, nubes densas, niebla mayor, oleaje reforzado y luz fría, sin rayos ni flashes.

Los parámetros ambientales interpolan sin cambiar keys ni reconstruir Canvas, cámara, controles o protagonistas. Con movimiento reducido se aplica el estado final y se congelan las animaciones continuas.

## Pruebas y validación

La suite cubre la matriz de estados públicos, salud interna, peor severidad, clima, antigüedad por `generated_at`, fallo inicial, retención stale, recuperación, fetch paralelo, polling, cancelación y actualización independiente. Los tests DOM verifican nombres accesibles con severidad y frescura.

La validación automatizada final completó 34 tests, lint, build y formato. `npm audit --omit=dev` informó cero vulnerabilidades. El stack 3D continúa en un chunk diferido separado.

El modo de demo está disponible exclusivamente en desarrollo con `?scene=still&weather=clear|mist|storm`. Se revisaron capturas deterministas de los tres climas a 1440×900 y una captura complementaria a 1024×768. La comprobación final en Safari real confirmó colores, climas, polling, recuperación parcial, navegación, movimiento reducido y estabilidad de cámara.

## Deuda posterior

- V2.5 completará command palette y accesibilidad transversal.
- V2.6 incorporará perfiles explícitos de calidad y presupuesto de render.
- V2.7 añadirá recorridos browser automatizados para interacción y recuperación.
