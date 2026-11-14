# 012 — Entrega: README, nota de decisiones y guion de demo

## Objetivo
Que alguien que nunca vio el proyecto lo corra en < 5 minutos, entienda las decisiones, y que la demo cuente la historia de punta a punta.

## Criterios de aceptación
### README.md (raíz)
1. **Requisitos**: Go ≥ 1.23, Node ≥ 20. **Correr**: `make dev` → front `http://localhost:5173`, API `http://localhost:8080`. **Tests**: `make test`, `make coverage`.
2. **Qué es** (3 líneas) y **cómo usar la demo** (ver guion abajo).
3. **Decisiones clave** (gestión por excepción, reglas determinísticas, aprobación auditada, avisos sin montos, mocks en la frontera externa).
4. **Supuestos** (copiar `challenge.md §5` + los de `domain-context.md` marcados como supuesto + régimen al último día del período + formato CSV).
5. **Arquitectura**: diagrama ASCII de capas y flujo del cierre; link a `specs/`.
6. **Próximamente — cómo integraría**:
   - **Emitir factura en ARCA**: el botón abriría un formulario prellenado (cliente del exterior, monto, moneda, TC) y emitiría la Factura E vía el web service de factura electrónica de exportación de ARCA (WSFEX), con el cliente delegando a Wallbit el servicio de facturación con su Clave Fiscal. El CAE vuelve y la factura entra sola al cierre (desaparece el CSV). Validar con el contador y con legales.
   - **Pagar en ARCA**: generar el VEP (Volante Electrónico de Pago) de la obligación aprobada y pagarlo desde el saldo de Wallbit, marcando `paidAt` automáticamente con el comprobante.
   - **ARCA/tipo de cambio reales**: reemplazar los adaptadores `mock` por clientes HTTP detrás de los mismos puertos (`ARCA_CLIENT=http`).
7. **Qué haría en las 6 semanas** (tabla semana a semana, una línea por semana).
8. Uso de IA (una línea).

### NOTA_DE_DECISIONES.md (≤ 10 líneas, viñetas)
9. Exactamente estas viñetas, ≤ 10 líneas en total:
   - Qué hace la v1.
   - Qué quedó afuera.
   - Riesgo más grande.
   - Cómo sabríamos si funcionó.
   - Uso de IA (una línea).

### Guion de demo (`docs/demo.md`, 5 minutos)
10. Pasos:
    1. Contador: tablero de octubre, chips 🔴/🟡/🟢, explicar gestión por excepción.
    2. Filtrar "Faltan facturas" → abrir Carla Benítez → recorrer con Siguiente.
    3. "Revisar 🔴": Ignacio Sosa (excede K) → Confirmar pase a RI.
    4. Tomás Álvarez (cobro atípico): reclasificar o aprobar con nota.
    5. Seleccionar verdes → aprobar en lote → ver auditoría en Historial.
    6. Parámetros: desactivar "Factura E computa para el tope" → ver cómo cambian los semáforos.
    7. Cambiar a Paula Rojas (cliente): subir CSV de ejemplo, declarar "No", confirmar → volver al contador, aprobar → Paula ve el resumen, "Pagar en ARCA" muestra "Próximamente".
    8. `POST /reminders/run` → recordatorio en la campana.

### Verificación final
11. Clonar en carpeta limpia, seguir el README al pie de la letra, y que `make dev`, `make test` y `make coverage` pasen.
