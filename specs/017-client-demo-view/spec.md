# 017 — Vista del cliente (demo)

## Objetivo
Mostrar, dentro de la misma app y claramente marcada como **demo**, cómo vive el cliente el cierre de mes dentro de la app de Wallbit. Cierra el circuito de punta a punta:

> el cliente hace el check-in → el contador revisa y aprueba → el cliente ve el resumen **validado por el contador**.

Es una maqueta interactiva que funciona contra la **API real del cliente** (`/v1/clients/{clientId}/**`, spec 011). No es una segunda aplicación. El foco del prototipo sigue siendo la herramienta del contador.

## Decisiones
- **Switch en el header:** `Contador | Cliente (demo)`. La vista del cliente se muestra dentro de un **marco de celular** con la leyenda "Así se vería dentro de la app de Wallbit".
- **5 clientes de demo**, elegidos por el estado que muestran (lista fija en el front, sin endpoint nuevo):

  | Cliente | ID | Qué muestra |
  |---|---|---|
  | Paula Rojas | `cli_demo_15` | Check-in pendiente: carga de comprobantes en vivo |
  | Lucía Fernández | `cli_demo_01` | Octubre aprobado: tarjeta validada (Monotributo) |
  | Camila Torres | `cli_demo_09` | Octubre aprobado: tarjeta validada (Responsable Inscripto) |
  | Ignacio Sosa | `cli_demo_02` | "Tu contador está revisando octubre" + septiembre validado. Si el contador lo aprueba en la demo, ve el aviso de exclusión |
  | Carla Benítez | `cli_demo_03` | "Tu contador está revisando octubre" + septiembre validado |

- **Contador que valida:** nombre y matrícula por configuración. Default: `C.P. Mariana Quiroga · Matrícula de ejemplo` (persona y matrícula inventadas).
- **Carga de comprobantes** con el patrón de los formularios de ARCA (tabla de comprobantes + "Alta de comprobante"), con el estilo visual de Wallbit (**no** imitar la estética de ARCA). Alternativa: importar el CSV de "Mis Comprobantes".
- **No hay carga de "monto total"**: las facturas se cargan una por una (o por CSV), porque la conciliación, la pesificación por factura y la fecha de exclusión dependen de cada comprobante. Los cobros fuera de Wallbit sí se declaran como monto + descripción.

## Criterios de aceptación

### Backend
1. Config `ACCOUNTANT_DISPLAY_NAME` (default `C.P. Mariana Quiroga`) y `ACCOUNTANT_LICENSE` (default `Matrícula de ejemplo`).
2. Al publicar un cierre (aprobación), el snapshot `Published` guarda además:
   - `validatedBy { name, license }`;
   - `regime`;
   - `category` (si es Monotributo);
   - `pctOfCategoryCapBp` y `categoryCapARS` (si es Monotributo);
   - `billed12mARS`.
   El cliente **solo** ve estos datos del snapshot, nunca el cálculo vigente. Es JSONB: no hace falta migración, pero los cierres publicados antes de esta feature no tienen los campos nuevos y el front los trata como opcionales.
3. Seed: los cierres de **octubre** de `cli_demo_01` y `cli_demo_09` quedan **aprobados y notificados** (aprobación del 04/11/2026 10:00 con la nota interna "Cierre de octubre (dato semilla)"). El resto de los demos no cambia. Actualizar el e2e de 007 y la tabla del README.
4. `GET /v1/clients/{clientId}/closings` devuelve los períodos del cliente (octubre y septiembre) con `state` y `published`. Si hace falta, se completa la spec 011 (no se agregan endpoints).
5. Fixtures reales para el front: perfil, cierres (uno por cada estado), facturas, importación, external-income, confirmación, pago.

### Frontend: navegación
6. Switch `Contador | Cliente (demo)` en el header (a la derecha, antes del período). "Cliente (demo)" lleva a `/cliente/:clientId` (default Paula Rojas). Arriba del marco de celular, un selector con los 5 clientes de la tabla y a su derecha el badge de fecha simulada (`Hoy: dd/MM/yyyy · fecha simulada`).
7. **Identidad por ruta, no global:** las llamadas de `features/client-demo` usan `X-Demo-User: client:<clientId>` (opción `as` en `apiFetch`). Todo lo demás sigue como `accountant`. **No** se reintroduce `roleStore` ni un estado de rol global.
8. El período de la vista del cliente es siempre el último cerrado (octubre 2026). El selector de período del header se oculta en `/cliente/**`.

### Frontend: pantallas (dentro del marco de celular)
9. **Check-in** (`state = ACTION_REQUIRED`), título "Cierre de octubre":
   1. **"Comprobantes que emitiste en octubre"**:
      - tabla con Fecha · Tipo · Número (punto de venta–número) · Cliente · Monto;
      - botón **"Alta de comprobante"**, que abre un formulario con: Fecha, Tipo (Monotributo: Factura E / Factura C; RI: E / A / B), Punto de venta, Número, Cliente (denominación) y, opcional, su documento, Moneda (USD/ARS), Monto total y Tipo de cambio (opcional, "si lo informaste en la factura"). Validaciones iguales a `POST /v1/clients/{clientId}/invoices`; errores 422 bajo cada campo; 409 → "Ese comprobante ya está cargado";
      - link alternativo **"Importar CSV de Mis Comprobantes"** con el resultado de la importación;
      - no se ofrece borrar comprobantes (la API no lo tiene; queda fuera de alcance).
   2. **"¿Cobraste algo fuera de Wallbit en octubre?"**: No / Sí (líneas de monto US$ + descripción).
   3. **"Listo, ya cargué todo"**: habilitado cuando el paso 2 está respondido. Pasa a "En revisión".
10. **En revisión** (`IN_REVIEW`): "Tu contadora está revisando tu cierre de octubre. Te avisamos cuando esté listo." + resumen de lo cargado (cantidad de comprobantes y cobros externos). Debajo, la tarjeta validada del mes anterior si existe.
11. **Resumen validado** (`READY`), tarjeta estilo mobile/email:
    - **Facturaste en octubre:** US$ (grande) y $ (chico), con la aclaración "pesificado con el tipo de cambio de cada factura";
    - **Tu categoría:** Monotributo → "Categoría H · 65% del tope anual" con barra de progreso (mismos colores que el tablero); RI → "Responsable Inscripto";
    - **Qué tenés que pagar este mes:** cada concepto con monto y vencimiento, más el total;
    - el **cuerpo del mensaje publicado** (incluye los agregados, por ejemplo el de exclusión);
    - badge **"✓ Validado por C.P. Mariana Quiroga · Matrícula de ejemplo · 04/11/2026"**;
    - botón **"Ya pagué"** (`PUT /v1/closings/{closingId}/payment`) → badge "Pagado el dd/MM".
12. Cliente sin cierre del período: estado vacío "Todavía no hay un cierre para este mes".

### Demo en vivo (criterio de aceptación del circuito)
13. 1) Cliente (demo) → Paula: carga un comprobante con "Alta de comprobante" (o importa el CSV de ejemplo), responde "No", confirma → "En revisión". 2) Contador → buscar Paula: verde y "Para revisar" → Aprobar. 3) Cliente (demo) → Paula: aparece la tarjeta validada con el badge. Hay un test e2e del backend con este circuito y un test de integración del front con MSW.

### Calidad
14. Accesibilidad: el marco de celular no rompe el orden de foco; formularios con labels; `vitest-axe` sin violaciones serias.
15. Coverage ≥ 90% (back y front).
