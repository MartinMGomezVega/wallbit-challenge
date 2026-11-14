# 016 — Importar facturas desde el contador

## Objetivo
Que el contador pueda **cargar las facturas de un cliente en su nombre** desde el modal de revisión, con el mismo CSV de "Mis Comprobantes" de ARCA que usaría el cliente, y opcionalmente **confirmar el período en su nombre** cuando el cliente se lo pidió por otro canal.

## Por qué
- Pasa en la realidad: muchos clientes le mandan los comprobantes por mail o WhatsApp al contador en vez de cargarlos.
- Sin esto, un cierre "Esperando al cliente" solo se destraba desde la app del cliente. Con esto, el contador puede completarlo.
- Todo queda **auditado**: en cada factura y cada confirmación queda registrado quién la cargó.

## Historias
- Como contador, abro un cliente con "Faltan facturas", importo el CSV que me mandó y veo al instante cómo se recalcula el cierre.
- Como contador, si el cliente me confirmó que no tuvo cobros fuera de Wallbit, confirmo el período en su nombre y dejo registrado cómo me lo dijo.

## Criterios de aceptación
### Backend
1. `POST /api/v1/accountant/clients/{id}/invoices/import` (multipart, campo `file`, máx 2 MB). Reutiliza el caso de uso `ImportInvoicesCSV` (mismo parser, validación, deduplicación e idempotencia que `/me/invoices/import`). Responde el mismo `ImportResult`.
2. Cada factura guarda `uploadedBy`: `CLIENT`, `ACCOUNTANT` o `SEED`. Las importadas por este endpoint quedan como `ACCOUNTANT`. Migración Postgres nueva con default `CLIENT` para las existentes que no son del seed.
3. `POST /api/v1/accountant/clients/{id}/periods/{period}/confirm-input` con body `{ "externalIncome": { "none": true } | { "items": [...] }, "note": "Me lo confirmó por mail el 06/11" }`:
   - `note` es obligatoria (1 a 300 caracteres): es la evidencia de por qué confirma el contador. Si falta → 422 `NOTE_REQUIRED`;
   - setea el input del período como respondido y confirmado, con `confirmedBy: "ACCOUNTANT"` y la nota guardada;
   - mueve el cierre de `PENDING_CLIENT_INPUT` a `READY_FOR_REVIEW`;
   - si el período ya estaba confirmado → 409 `ALREADY_CONFIRMED`.
4. Ambas operaciones emiten `ClosingInputsChanged` y recalculan el cierre. Si el cierre estaba aprobado o notificado, queda en una versión nueva `READY_FOR_REVIEW` (regla de `domain-context.md §10`).
5. El detalle del cierre (`GET /accountant/closings/{id}`) expone `uploadedBy` en cada factura y `confirmedBy` y `confirmationNote` en el input del cliente.
6. OpenAPI: los dos endpoints nuevos, `uploadedBy` en `Invoice`, y `confirmedBy` y `confirmationNote` en `ClientInput`. Ambos entran en el test de contrato de rutas.

### Frontend (modal de revisión, pestaña "Facturas y conciliación")
7. Botón secundario **"Importar facturas (CSV de ARCA)"** arriba a la derecha de la pestaña. Abre un diálogo con:
   - zona para soltar o elegir el archivo (`.csv`, máx 2 MB) y el link "Descargar CSV de ejemplo" (`/samples/mis_comprobantes_ok.csv`);
   - botón "Importar";
   - resultado: importadas, duplicadas, advertencias y errores por número de línea.
8. Si el cierre está aprobado o notificado, el diálogo avisa antes de importar: "Este cierre ya fue aprobado. Si importás facturas, vuelve a 'Para revisar' con una versión nueva."
9. Al cerrar el diálogo con al menos una factura importada, se refrescan el detalle, la fila de la tabla y los chips. La cola del modal mantiene el mismo índice.
10. En la lista de facturas, las cargadas por el contador muestran el badge "Cargada por el contador".
11. Si el cierre está `PENDING_CLIENT_INPUT`, la pestaña muestra el bloque **"Confirmar en nombre del cliente"**:
    - opción "No tuvo cobros fuera de Wallbit" u opción "Tuvo cobros fuera de Wallbit" (lista de monto en US$ y descripción);
    - campo obligatorio "¿Cómo te lo confirmó?";
    - botón "Confirmar período".
    - Al confirmar, el estado pasa a "Para revisar" y el bloque desaparece. En "Historial" se ve "Período confirmado por el contador: ‹nota›".
12. Accesibilidad: el diálogo usa el mismo patrón que los demás (Radix, foco atrapado, resultado con `aria-live`).

### Demo
13. Flujo completo de Paula Rojas desde la UI:
    1. buscar "Paula";
    2. Revisar → Facturas → Importar el CSV de ejemplo → el cierre pasa a verde;
    3. Confirmar en nombre del cliente ("No tuvo cobros", nota) → "Para revisar";
    4. Aprobar → Aviso al cliente muestra lo publicado.
    Hay un test e2e del backend con este flujo.
