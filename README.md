<p align="center">
  <img src="frontend/src/assets/brand/wallbit-logo-dark.svg" alt="Wallbit" height="40">
</p>

<h1 align="center">Contabilidad · Cierre de mes</h1>

<p align="center">
  Prototipo de la herramienta con la que el contador de Wallbit puede atender a cientos de clientes por mes.<br>
  Challenge Product Engineer — Martín Gómez Vega
</p>

<p align="center">
  <a href="#cómo-correrlo">Cómo correrlo</a> ·
  <a href="#decisiones-de-producto">Decisiones</a> ·
  <a href="#arquitectura">Arquitectura</a> ·
  <a href="https://martinmgomezvega.github.io/wallbit-challenge/">API (Swagger)</a> ·
  <a href="#lo-que-dejé-afuera-y-cómo-lo-haría">Lo que dejé afuera</a> ·
  <a href="NOTA_DE_DECISIONES.md">Nota de decisiones</a>
</p>

---

## Descripción

El día 1 de cada mes, la herramienta arma sola el cierre de cada cliente y lo marca 🔴 / 🟡 / 🟢. El contador revisa solo lo que no es verde, aprueba el resto en lote, y el cliente recibe el aviso recién cuando el contador firmó.

---

## Cómo correrlo

> [!IMPORTANT]
> El prototipo vive en un "hoy" fijo: **05/11/2026** (`APP_NOW`). Así la demo muestra
> siempre el mismo momento: el cierre de octubre recién armado, septiembre ya cerrado y
> los vencimientos de noviembre por delante. Se puede cambiar con la variable `APP_NOW`
> (vacía = reloj real).

**Con Docker (recomendado).**

```bash
git clone https://github.com/MartinMGomezVega/wallbit-challenge.git
cd wallbit-challenge
docker compose up --build
```

| Qué | Dónde |
|---|---|
| App completa | http://localhost:5173 |
| API | http://localhost:8080/v1 |
| Swagger local | http://localhost:8080/docs |
| Swagger publicado | https://martinmgomezvega.github.io/wallbit-challenge/ |

En el primer arranque se crea el esquema en **PostgreSQL**, se generan **300 clientes (mockeados)** con 13 meses de movimientos y facturas (datos inventados y determinísticos) y se corre el cierre de **octubre 2026**. Los arranques siguientes reutilizan los datos. Para empezar de cero: `make reset-db`.

<details>
<summary><b>Sin Docker para la app, tests y variables</b></summary>

Con Go ≥ 1.25 y Node ≥ 20:

```bash
make install      # dependencias del frontend
make dev          # Postgres en Docker + backend y frontend locales
make dev-memory   # todo en memoria, sin Docker
```

```bash
make test               # tests unitarios de backend (-race) y frontend, sin Docker
make coverage           # falla si alguno queda por debajo de 90%
make test-integration   # repositorios y consultas contra Postgres
make lint
```

| Variable | Default | Para qué |
|---|---|---|
| `STORAGE` | `postgres` | `postgres` o `memory` |
| `DATABASE_URL` | `postgres://wallbit:wallbit@localhost:5432/accountant` | Conexión (credenciales solo de demo local) |
| `APP_NOW` | `2026-11-05T09:00:00-03:00` | Fecha "de hoy". Vacía = reloj real |
| `SEED_CLIENTS` | `300` | Clientes generados. Probalo con `1000` |
| `SEED_RESET` | `false` | Vacía la base y vuelve a sembrar |
| `CLOSING_WORKERS` | cantidad de CPUs | Workers del cierre en paralelo |
| `MOCK_LATENCY_MS` | `0` | Latencia simulada de ARCA y tipo de cambio |

</details>

---

## Decisiones de producto

### Quién hace qué

| La herramienta, sola | El contador | El cliente |
|---|---|---|
| Cierra el período el día 1 | Define las reglas y los parámetros | Sube sus facturas (CSV de ARCA o a mano) |
| Clasifica cada movimiento: ingreso / no ingreso / sin clasificar | Revisa rojos y amarillos | Declara si cobró algo fuera de Wallbit |
| Pesifica y concilia cobros con facturas | Corrige clasificaciones | Confirma que terminó |
| Calcula Monotributo o RI y detecta el pase a RI | Aprueba: individual o en lote (solo verdes) | Ve el resumen **después** de la aprobación |
| Marca cada cierre 🔴/🟡/🟢 con una descripción | Confirma el pase a RI | Marca "Ya pagué" |
| Arma el borrador del aviso y los totales del mes | Importa facturas y confirma en nombre del cliente si se las manda | Recibe un recordatorio antes del vencimiento |

Lo dimensioné así: con 500 clientes, si el 70% sale verde y se revisa en ~2 minutos, y el 30% necesita ~10, son unas **37 horas por mes**. Entra en una semana de trabajo del contador.

### Funcionamiento del semáforo
- 🔴 **Requiere acción:** excede el tope de la categoría K, facturas C después de la exclusión, cobros sin factura, depósitos por encima del tope.
- 🟡 **Revisar:** cerca del tope K o de su categoría, debe recategorizar, cobro fuera de lo habitual, movimientos sin clasificar, cobros declarados fuera de Wallbit, facturas sin cobro, gastos con tarjeta altos, puede bajar de categoría.
- 🟢 **Todo ok:** no requiere cambios.

El color es la situación **más grave** del cliente, y la descripción explica esa situación en una línea. El color nunca va solo: siempre lo acompañan un ícono y un texto, para quien no distingue colores. El **estado del cierre** (esperando al cliente, para revisar, aprobado, notificado) va en otra columna. Mezclarlo con el semáforo haría que "todavía no subió facturas" se viera igual que "tiene un problema real".

### El modal de revisión
Se abre desde una fila, desde un chip de color o desde una selección, y recorre **exactamente la lista que el contador tenía filtrada**, en el mismo orden, con Anterior / Siguiente y atajos de teclado. La cola se congela al abrir: si un cliente cambia de color al aprobarlo, el índice no salta. Todo queda en la URL, así que recargar, compartir el link o usar el botón Atrás del navegador vuelve al mismo lugar.

### Resumen del cierre
La pestaña **Resumen** muestra, de arriba hacia abajo:
- **Estado de los clientes:** tarjetas Rojos, Amarillos y Verdes con su porcentaje del total y link directo a la tabla filtrada;
- **Indicadores del cierre:** avance de aprobación con estimación de horas de revisión, facturado del mes y total a pagar con próximo vencimiento;
- **Gráficos:** situaciones detectadas, clientes por categoría (y cerca del tope K) y vencimientos próximos;
- **¿Está funcionando el servicio?:** sección separada con % de verdes, % avisados antes del día 10 y % de pagos en término (con su detalle `X de Y`).

Con el selector de período se compara **septiembre** (cerrado y aprobado, con métricas de éxito reales) con **octubre** (en curso). Cada número lleva a la tabla filtrada que lo explica.

### Aprobación: rápida pero auditada
- **En lote, solo los verdes.** Los amarillos y rojos los firma el contador uno por uno.
- Cada aprobación guarda **quién, cuándo, la versión del cálculo y un hash del resultado**. Si después cambia un dato, el cierre vuelve a "para revisar" con una versión nueva y la aprobación anterior queda en el historial.
- **Nunca se firma un número que no se vio:** si el cierre cambió mientras el contador lo miraba, la aprobación se rechaza y se le pide refrescar.

### La experiencia del cliente
Vive dentro de la app de Wallbit. Así la pensé:
1. **Día 1:** "Subí tus facturas de octubre" (push sin montos y aviso en la app).
2. **Carga en 3 pasos:** importar el CSV de "Mis Comprobantes" de ARCA (o cargar a mano) → "¿Cobraste algo fuera de Wallbit?" → "Listo, ya cargué todo". Si prefiere mandárselas al contador, él las importa desde el modal.
3. **Espera:** "Tu contador está revisando tu cierre".
4. **Con la firma:** "Facturaste US$ X. Pagás $ Y antes del 20/11", con "Ya pagué" y un recordatorio antes del vencimiento.

Email y push nunca llevan montos ni CUIT, y nada impositivo sale sin la firma del contador. En el prototipo, esta experiencia es la API `/v1/clients/{clientId}/**` (se puede probar desde Swagger), y el contador ve en el modal el aviso exacto que recibe el cliente. A futuro, la factura se emite desde Wallbit y el paso 2 desaparece. Para la demo, el switch **Cliente (demo)** muestra esta experiencia con 5 clientes de ejemplo.

### Monotributo y Responsable Inscripto
Incluí RI porque si facturás más que el tope de la categoría K, quedás excluido del Monotributo desde ese día. La herramienta calcula el acumulado de 12 meses día por día y detecta **la fecha exacta** en que se cruzó el tope. Esa fecha es clave: desde ahí debería haber facturado como RI. El contador confirma el pase y la herramienta empieza a calcular IVA, autónomos y anticipos.

### Decisiones que son del contador, no mías
Algunas cosas las fuentes no las resuelven igual, o dependen del criterio profesional. Esas quedaron como **parámetros editables**:
- si la **factura E** suma al tope del Monotributo (default: sí, el criterio conservador);
- qué **tipo de cambio** usar cuando la factura no lo trae;
- los **umbrales** de "cerca del tope" y de "cobro fuera de lo habitual";
- la tolerancia de la conciliación.

Cambiar un parámetro recalcula al instante todos los cierres no aprobados.

---

## Decisiones técnicas

- **Hay backend real.** El cierre tiene que correr solo el día 1, la aprobación tiene que auditarse del lado del servidor, los movimientos de todos los clientes no deberían bajarse al navegador y el aviso sale del servidor. **Los únicos mocks son los sistemas externos (ARCA y tipo de cambio), no mi propio backend.**
- **PostgreSQL en Docker.** Una base propia no es un servicio externo: es infraestructura del backend. Elegí Postgres y no DynamoDB porque el dominio es relacional (clientes, cierres, facturas, aprobaciones), necesita transacciones y el tablero y el Resumen son filtros y agregaciones, que se resuelven en SQL.
- **Aprobar es atómico.** Aprobar, publicar, crear la notificación y encolar el email/push pasan en **una sola transacción** (patrón outbox). Si algo falla, no queda un cliente notificado de un cierre sin aprobar.
- **Concurrencia optimista.** `UPDATE closings ... WHERE version = $n`: si dos personas aprueban a la vez, una gana y la otra recibe un 409.
- **Mocks con datos reales, detrás de puertos.** El dominio pide `TaxScaleProvider` y `ExchangeRateProvider`. Hoy los implementa un cliente mock con las escalas reales de febrero y agosto 2026, los vencimientos de noviembre 2026 y el dólar BNA día por día. Pasar a los servicios reales es cambiar el adaptador sin tocar el dominio.
- **Dinero en centavos (`int64` / `BIGINT`), nunca `float`.** Los tipos de cambio son enteros escalados y la conversión redondea half-even. Un test de arquitectura falla si aparece `float64` en el dominio.
- **Cierre concurrente e idempotente.** Un worker pool procesa los clientes en paralelo. Correrlo dos veces no duplica nada, porque cada cálculo tiene un hash y si no cambió no se crea versión.
- **Misma lógica, dos almacenamientos.** Los repositorios en memoria se usan en los tests unitarios. Los de Postgres pasan **las mismas suites de contrato** en los tests de integración, así el tablero ordena y filtra igual en los dos.
- **Tiempo inyectado.** Nada llama a `time.Now()` fuera del reloj, así que la demo y los tests son reproducibles.
- **Monorepo.** Backend y frontend son proyectos independientes que se hablan por HTTP. Están juntos para que se clone y se levante una sola vez, y para que una feature que toca las dos puntas sea un solo commit.

---

## Arquitectura

### Secuencia del cierre de mes

```mermaid
sequenceDiagram
    autonumber
    participant J as Job día 1
    participant API as API (Go)
    participant DB as PostgreSQL
    participant EXT as ARCA y tipo de cambio (mock)
    actor CL as Cliente
    actor CO as Contador
    participant N as Email / push

    J->>API: RunMonthlyClosing(octubre)
    loop por cliente, en paralelo
        API->>DB: movimientos, facturas e inputs
        API->>EXT: escala vigente, vencimientos, cotización
        API->>API: clasificar, conciliar, liquidar, evaluar señales
        API->>DB: guardar cierre (nueva versión solo si cambió el hash)
    end
    API->>DB: avisos "Subí tus facturas"

    CL->>API: importar CSV de ARCA, declarar cobros externos, confirmar
    API->>API: recalcular el cierre
    API->>DB: cierre en READY_FOR_REVIEW

    CO->>API: tablero filtrado y Resumen
    API->>DB: filtros, orden y agregaciones en SQL
    CO->>API: abrir cliente y recorrer con Siguiente
    CO->>API: aprobar (versión 3)
    rect rgba(13,153,255,0.08)
        Note over API,DB: una sola transacción
        API->>DB: actualizar cierre si la versión sigue siendo 3
        API->>DB: guardar aprobación, publicar, crear notificación, encolar salientes
    end
    API->>N: "Tenés novedades en Wallbit" (sin montos)
    CL->>API: ver mi cierre
    API-->>CL: lo publicado (facturado, a pagar y vencimientos)
```

### Backend (Go) — por capas, Clean Architecture

```
interfaces/http  ─┐
infrastructure   ─┼──►  application (casos de uso + puertos)  ──►  domain (reglas puras)
cmd (wiring)     ─┘
```

```
backend/internal/
├── domain/          money · period · client · movement · invoice · reconciliation · tax · signal · closing · parameters
├── application/     ports (interfaces) · usecase (un archivo por caso de uso)
├── infrastructure/  postgres · memory · arca/mock · fx/mock · notify/log · clock · seed · scheduler
└── interfaces/http  handlers · middlewares · errores
```

Las reglas impositivas viven en `domain`, sin HTTP ni base de datos, y se testean como funciones puras. Un test de arquitectura verifica que ninguna capa interna importe una externa.

### Frontend (React + TypeScript) — por features

```
frontend/src/
├── app/          router, providers, layout, selector global de período
├── shared/       api (fetch + zod), ui (tokens de marca, semáforo, botones), format (es-AR)
└── features/
    ├── closings-dashboard/   tablero, filtros, chips, aprobación en lote
    ├── closing-review/       modal con cola Anterior / Siguiente
    ├── summary-dashboard/    Resumen del cierre
    ├── client-demo/          vista demo del cliente (check-in y resumen validado)
    └── parameters/           parámetros del contador
```

TanStack Query y TanStack Table, React Router (con el estado en la URL), Tailwind con los tokens de marca, Radix UI para accesibilidad, Recharts para el Resumen y Zod para validar las respuestas.

### Estados del cierre

```mermaid
stateDiagram-v2
    direction LR

    state "Esperando al cliente<br/><small>PENDING_CLIENT_INPUT</small>" as PENDING
    state "Para revisar<br/><small>READY_FOR_REVIEW</small>" as READY
    state "Aprobado<br/><small>APPROVED</small>" as APPROVED
    state "Notificado<br/><small>NOTIFIED</small>" as NOTIFIED

    [*] --> PENDING: 1. Día 1
    PENDING --> READY: 2. Confirma
    PENDING --> APPROVED: 3. Aprueba igual
    READY --> APPROVED: 4. Aprueba
    APPROVED --> NOTIFIED: 5. Publica y avisa
    NOTIFIED --> READY: 6. Cambian los datos
    APPROVED --> READY: 6. Cambian los datos

    classDef pending fill:#F5F5F5,stroke:#A3A3A3,color:#171717
    classDef ready fill:#E6F4FF,stroke:#006FC0,color:#005FA8
    classDef approved fill:#F0F9FF,stroke:#0369A1,color:#0369A1
    classDef notified fill:#ECFDF5,stroke:#059669,color:#047857
    class PENDING pending
    class READY ready
    class APPROVED approved
    class NOTIFIED notified
```

| # | Transición | Quién la dispara |
|---|---|---|
| 1 | El job del día 1 arma el cierre. Si el cliente ya había confirmado, nace directo en *Para revisar*. | El sistema |
| 2 | "Ya cargué todo": facturas subidas y cobros externos declarados. | El cliente, o el contador en su nombre (con nota de cómo se lo confirmó) |
| 3 | El contador puede aprobar sin la confirmación del cliente, pero uno por uno y con aviso. Nunca en lote. | El contador |
| 4 | Aprobación individual, o en lote si el cierre es verde. | El contador |
| 5 | Se publica el resumen y se encola el aviso en la **misma transacción** que la aprobación, así que *Aprobado* dura un instante. | El sistema |
| 6 | Si cambia un dato (factura nueva, reclasificación), el cierre vuelve a *Para revisar* con **versión nueva**. La aprobación anterior queda en el historial y el cliente sigue viendo lo publicado hasta la nueva firma. | El sistema |

Tres reglas que no se ven en el diagrama:
- **Recalcular no siempre cambia el estado.** En *Esperando* o *Para revisar*, cada cambio de datos o de parámetros recalcula el cierre y genera una versión nueva, sin moverlo de estado. Cambiar parámetros nunca toca un cierre aprobado.
- **Aprobar exige ver la última versión.** Si el cierre cambió mientras el contador lo miraba, la aprobación se rechaza y se le pide refrescar.
- **"Ya pagué" no es un estado.** Es una marca (`paidAt`) sobre un cierre notificado. Así el estado cuenta en qué etapa está el cierre y el pago se mide aparte.

---

## API

La documentación está en **[Swagger](https://martinmgomezvega.github.io/wallbit-challenge/)** (solo lectura). Para probar los endpoints, usá el Swagger local: `http://localhost:8080/docs`. El contrato vive en [`backend/api/openapi.yaml`](backend/api/openapi.yaml), y un test verifica que cada ruta del router esté documentada.

Base: `http://localhost:8080/v1`. Diseño REST por recursos: el rol lo decide la identidad, no la URL. En la demo, la identidad va en el header `X-Demo-User`: `accountant` accede a todo y `client:<clientId>` solo a sus propios recursos.

| Recurso | Endpoints |
|---|---|
| Períodos y cierres | `GET /v1/periods` · `POST /v1/periods/{period}/closing-runs` · `GET /v1/periods/{period}/closings` (+ `/summary`, `/ids`) · `GET /v1/periods/{period}/stats` · `GET /v1/closings/{closingId}` |
| Aprobaciones y pagos | `POST /v1/closings/{closingId}/approvals` · `POST /v1/periods/{period}/approvals` (lote) · `PUT /v1/closings/{closingId}/payment` |
| Clientes | `GET /v1/clients/{clientId}` · `GET /v1/clients/{clientId}/closings` · `POST /v1/clients/{clientId}/regime-changes` · `PUT /v1/clients/{clientId}/ri-profile` |
| Facturas e inputs del período | `GET·POST /v1/clients/{clientId}/invoices` · `POST /v1/clients/{clientId}/invoice-imports` · `POST /v1/clients/{clientId}/purchase-invoices` · `PUT /v1/clients/{clientId}/periods/{period}/external-income` · `POST /v1/clients/{clientId}/periods/{period}/confirmation` |
| Movimientos | `PUT /v1/movements/{movementId}/classification` |
| Notificaciones | `GET /v1/clients/{clientId}/notifications` · `PATCH /v1/clients/{clientId}/notifications/{notificationId}` |
| Configuración y sistema | `GET·PUT /v1/parameters` · `GET /v1/signal-types` · `POST /v1/reminder-runs` · `GET /health` · `GET /v1/openapi.yaml` |

---

## Datos de demo

300 clientes generados con semilla fija (85% Monotributo, 15% RI, cobros de USD 800 a 9.000 por mes de empresas y plataformas del exterior) y **15 clientes armados a mano**, uno por cada situación. Los datos semilla cargan dos períodos: **septiembre 2026** como mes cerrado y aprobado (con notificaciones y pagos registrados para alimentar las métricas de éxito del servicio) y **octubre 2026** como mes en curso listo para revisar en el tablero. Los encontrás con el buscador del tablero:

| Cliente | Situación |
|---|---|
| Lucía Fernández (demo) | 🟢 Octubre aprobado y validado (Monotributo A) |
| Ignacio Sosa (demo) | 🔴 Superó el tope K el 14/10 → pase a RI |
| Carla Benítez (demo) | 🔴 Cobros sin factura |
| Diego Romero | 🟡 Tiene que recategorizar (C → E) |
| Valentina Ruiz | 🟡 Cerca del tope de su categoría |
| Tomás Álvarez | 🟡 Cobro fuera de lo habitual |
| Sofía Medina | 🟡 Movimiento sin clasificar |
| Joaquín Herrera | 🟡 Cerca del tope K |
| Camila Torres (demo) | 🟢 Octubre aprobado y validado (RI exportadora con saldo a favor de IVA) |
| Federico Díaz | 🟡 RI con factura sin cobro + IVA a pagar |
| Agustina Castro | 🔴 Facturas C después de pasar a RI |
| Nicolás Peralta | 🔴 Depósitos por encima del tope K |
| Julieta Molina | 🟡 Declaró cobros fuera de Wallbit |
| Bruno Acosta | 🟡 Podría bajar de categoría |
| Paula Rojas (demo) | Esperando sus facturas: check-in en la vista del cliente o en el modal → Facturas → Importar. |

Todos los nombres, CUITs y empresas son inventados.
La interfaz tiene un enfoque mobile-first porque está pensada para integrarse de forma nativa en la app de Wallbit. Para esta demo construí el flujo usando React y TypeScript, pero en un escenario a producción lo implementaría en Flutter (stack en el que tengo experiencia) o me adaptaría sin problemas a la tecnología mobile que ya utilice el equipo.

---

## Supuestos

- Se cierra el **mes calendario anterior**, en hora de Argentina.
- Clientes Monotributo: **prestación de servicios**.
- La categoría se calcula con **lo facturado**. Lo cobrado se usa para conciliar y para controlar depósitos.
- Si el período incluye el día del pase a RI, se liquida con el **régimen vigente al último día del período**. Lo decide el contador.
- El CSV sigue el formato del export de **"Mis Comprobantes"** de ARCA. Las notas de crédito se ignoran con una advertencia.
- Los **anticipos de Ganancias y los aportes de autónomos** de un RI los carga el contador, porque dependen de la declaración jurada anterior, que Wallbit no tiene.
- Wallbit liquida los cobros a **compra BNA** del día.
- Vencimientos: los de **noviembre 2026 son reales**. Para otros meses uso el patrón habitual, sin feriados.
- El tiempo estimado de revisión del Resumen usa 2 minutos por cliente verde y 10 por amarillo o rojo.

---

## Lo que dejé afuera (y cómo lo haría)

No entraba en el alcance del challenge, pero son los próximos pasos para llevarlo a producción.

### Login y roles
Hoy la identidad es un header de demo (`X-Demo-User`). Para producción:
- **Contador:** login real con MFA obligatorio y roles. El backend validaría el token en un middleware que reemplaza al de demo, y cada aprobación quedaría atada a la identidad del contador matriculado que firmó.
- **Cliente:** no crearía cuentas nuevas. El cliente ya tiene cuenta en Wallbit, así que esta parte viviría dentro de la app de Wallbit y usaría la sesión que ya tiene.

### Base de datos en la nube y datos reales
Hoy Postgres corre en Docker y los datos salen de un generador. Para producción:
- **Postgres gestionado** (Amazon RDS/Aurora o Cloud SQL en GCP) con backups automáticos, cifrado y acceso solo desde la red privada del backend.
- **Ingesta real de movimientos** desde el core de Wallbit (eventos o lectura de una réplica) en lugar del seed.
- **Réplica de lectura** para el tablero y el Resumen, y vistas materializadas por período si el volumen crece, para que las agregaciones no compitan con el cierre.
- Particionado de cierres y movimientos por período.
- Descarté DynamoDB por lo mismo que en el prototipo: transacciones entre varias entidades y agregaciones.

### Emitir facturas en ARCA
Lo haría así: un formulario prellenado con el cobro que acaba de entrar (cliente del exterior, monto, moneda, tipo de cambio) que emite la **Factura E** por el web service de facturación de exportación de ARCA, con el usuario delegando ese servicio a Wallbit desde su Clave Fiscal. La factura vuelve con su CAE y entra sola al cierre. Esto **elimina el riesgo más grande del producto**: que el usuario no cargue sus facturas.

### Pagar en ARCA
Generaría el VEP de la obligación ya aprobada y lo pagaría con el saldo de Wallbit en un click. `paidAt` se marca solo con el comprobante, y las métricas de pago en término dejan de depender de que el usuario marque "Ya pagué".

### Integraciones reales
- ARCA (escalas, vencimientos, padrón) y cotización oficial del BNA/BCRA, reemplazando los adaptadores mock detrás de los mismos puertos.
- Email y push reales (por ejemplo SES y el servicio de notificaciones de la app), reemplazando el notificador que solo registra en el log.

### Escala y operación
- Recálculo **asincrónico** con una cola (hoy es sincrónico para que el contador vea el cambio al instante).
- Observabilidad: métricas del job, tiempos de revisión por contador, alertas si el cierre no terminó.
- Más de un contador: asignación de clientes y colas por contador.

### Alcance impositivo
Ingresos Brutos (varía por provincia), declaraciones juradas anuales, notas de crédito, retenciones y percepciones, y reintegro de IVA para exportadores. Las dos cosas a validar con legales son la emisión de facturas y los pagos.

---

## Cómo trabajé

Apliqué **Spec-Driven Development (SDD)**: antes de escribir código, documenté en [`specs/`](specs/) el contexto del dominio, la arquitectura y las guías de UI. Luego estructuré cada feature en su propia carpeta con tres artefactos:
- `spec.md`: especificaciones funcionales y reglas de negocio.
- `plan.md`: decisiones técnicas, de producto y diseño de la solución.
- `tasks.md`: plan de implementación desglosado paso a paso.

Separar especificaciones, plan y tareas por feature me permitió tener control total sobre el alcance, los tests y el comportamiento de la herramienta antes de pasar al código.
