# Guion de Demo (5 minutos)

> Este guion permite recorrer el flujo completo de punta a punta demostrando la propuesta de valor: **el contador pasa de armar liquidaciones cliente por cliente a supervisar y aprobar por excepción.**

---

### Preparación
1. En la terminal ejecutar:
   ```bash
   make dev
   ```
2. Abrir el navegador en [http://localhost:5173](http://localhost:5173).
3. Notar que la barra superior incluye el logo oficial de Wallbit y a la derecha el selector **"Ver como"** para alternar roles sin autenticación real.

---

### Paso 1: Tablero de octubre y gestión por excepción (Contador)
- **Rol activo:** `Contador`.
- En el encabezado se observa el período `octubre 2026` y los chips de resumen:
  - **🔴 21 Requiere acción** (7,0%)
  - **🟡 59 Revisar** (19,7%)
  - **🟢 220 Todo ok** (73,3%)
- **Mensaje:** Más del 70% de los clientes se resuelven automáticamente. El contador concentra su atención únicamente en los casos anómalos o de riesgo tributario.

---

### Paso 2: Filtros y navegación de cola (Carla Benítez)
- En la barra de filtros, abrir el selector de **Señal** y elegir **"Faltan facturas"** (código `MISSING_INVOICES`).
- En la tabla filtrada, hacer click en **Carla Benítez** (o en su botón **Revisar**).
- Se abre el modal de revisión con la cola filtrada congelada.
- Navegar con el botón **Siguiente →** (o tecla `ArrowRight` / `k`) para revisar secuencialmente los clientes con facturas faltantes sin regresar a la tabla.
- Cerrar el modal (`Esc`).

---

### Paso 3: Alerta crítica y transición a RI (Ignacio Sosa)
- En el chip rojo superior, hacer click en **"Revisar"** para abrir directamente la cola de casos críticos 🔴.
- El primer cliente es **Ignacio Sosa**:
  - En la sección de situación se observa la alerta: `Excede tope K (pasa a RI)`.
  - El sistema detectó que el 14/10/2026 sus ingresos acumulados superaron el límite anual del Monotributo.
- Hacer click en el botón **"Confirmar pase a RI"**:
  - Se abre el diálogo de confirmación indicando la fecha de exclusión (14/10/2026).
  - Al confirmar, el cierre se recalcula de inmediato como Responsable Inscripto (calculando Débito/Crédito Fiscal de IVA, autónomos y anticipos).

---

### Paso 4: Reclasificación de movimientos y notas (Tomás Álvarez)
- Buscar a **Tomás Álvarez** en el buscador de la tabla o en la cola de amarillos 🟡.
- Tomás presenta la señal `Cobro fuera de lo habitual` (ingreso atípico para su histórico).
- En la pestaña/sección **Movimientos**:
  - Se visualizan los movimientos clasificados con su regla de negocio.
  - Hacer click en **"Cambiar a no ingreso"** (o modificar la clasificación) para corregir el criterio.
  - El total facturado y el semáforo se recalculan automáticamente.
- En el campo de notas, ingresar: *"Verificado comprobante de reintegro; no computa como honorarios"* y hacer click en **Aprobar**.

---

### Paso 5: Aprobación en lote con auditoría (Cierres Verdes)
- Volver a la tabla y activar la casilla de selección múltiple en el encabezado para seleccionar clientes (o seleccionar varios verdes 🟢).
- Aparece la barra flotante de acciones masivas: hacer click en **"Aprobar verdes (N)"**.
- Se abre el diálogo modal mostrando el resumen de clientes que pasarán a estado `APPROVED`.
- Confirmar la acción: los clientes se aprueban en lote de manera concurrente.
- Abrir cualquiera de los clientes aprobados y verificar en la sección **Historial**:
  - Consta quién aprobó (`contador@estudio.com`), fecha/hora exacta, versión del cálculo y hash SHA-256 de integridad.

---

### Paso 6: Configuración de parámetros impositivos
- En la barra de navegación superior, ingresar a **Parámetros**.
- Desactivar la opción:
  - `¿La Factura E computa para el tope de Monotributo?` (por defecto activa con criterio conservador).
- Hacer click en **Guardar parámetros**.
- Regresar al **Tablero**: observar cómo los semáforos y los porcentajes acumulados de los clientes que emiten facturas E se actualizan instantáneamente en base a la nueva regla contable.

---

### Paso 7: Portal del cliente (Paula Rojas)
- En el selector **"Ver como"** (esquina superior derecha), seleccionar a la cliente **Paula Rojas** (`cli_demo_15`).
- La interfaz cambia a la vista mobile-first del cliente con lenguaje amigable y sin jerga técnica:
  - Se muestra el estado **"Acción requerida"** para octubre 2026.
- **Paso 1:** En la sección de facturas, descargar el archivo de muestra y subirlo (o arrastrar `sample_invoices.csv`). Se listan las facturas importadas con desglose.
- **Paso 2:** En la pregunta *¿Cobraste algo fuera de Wallbit en octubre 2026?*, seleccionar **"No, solo en Wallbit"**.
- **Paso 3:** Hacer click en **"Listo, ya cargué todo"**. El período pasa a **"En revisión"**.
- **Paso 4 (Aprobación):** Cambiar temporalmente a **Contador**, abrir a Paula Rojas (ahora en `READY_FOR_REVIEW`) y hacer click en **Aprobar**.
- **Paso 5 (Resumen final):** Regresar a **Paula Rojas**:
  - El período ahora está **"Listo"** mostrando el resumen: total facturado, cuota de Monotributo a pagar y fecha de vencimiento.
  - El botón **"Pagar en ARCA"** cuenta con el chip de estado inactivo **"Próximamente"** (explicando la futura integración con VEP).
  - Al hacer click en **"Ya pagué"**, se actualiza el estado a pagado con fecha actual.

---

### Paso 8: Recordatorios automáticos de vencimiento
- Para simular el scheduler de recordatorios automáticos de pago antes del vencimiento, ejecutar en una terminal:
  ```bash
  curl -X POST http://localhost:8080/api/v1/reminders/run
  ```
- En el navegador, observar la **campana de notificaciones** (junto al selector de rol):
  - Muestra el badge con el nuevo recordatorio generado para el cliente.
  - Al abrir la campana, se marca como leída y muestra el recordatorio sin exponer montos en la notificación push/email de acuerdo a los principios de privacidad.

