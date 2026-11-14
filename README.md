<p align="center">
  <img src="frontend/src/assets/brand/wallbit-logo-dark.png" alt="Wallbit" height="40">
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

**Con Docker (recomendado).** Solo necesitás Docker.

```bash
git clone https://github.com/MartinMGomezVega/wallbit-challenge.git
cd wallbit-challenge
docker compose up --build
```

| Qué | Dónde |
|---|---|
| App completa | http://localhost:5173 |
| API | http://localhost:8080/api/v1 |
| Swagger local | http://localhost:8080/docs |
| Swagger publicado | https://martinmgomezvega.github.io/wallbit-challenge/ |

En el primer arranque se crea el esquema en **PostgreSQL**, se generan **300 clientes (mockeados)** con 13 meses de movimientos y facturas (datos inventados y determinísticos) y se corre el cierre de **octubre 2026**. La "fecha de hoy" del prototipo está fijada en el 5/11/2026 para que siempre se vea lo mismo. Los arranques siguientes reutilizan los datos. Para empezar de cero: `make reset-db`.

<details>
<summary><b>Sin Docker para la app, tests y variables</b></summary>

Con Go ≥ 1.23 y Node ≥ 20:

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
| Arma el borrador del aviso y los totales del mes | Escribe una nota para el cliente si hace falta | Recibe un recordatorio antes del vencimiento |

Lo dimensioné así: con 500 clientes, si el 70% sale verde y se revisa en ~2 minutos, y el 30% necesita ~10, son unas **37 horas por mes**. Entra en una semana de trabajo del contador.

### Por qué reglas y no IA para clasificar
Las reglas son **explicables**: cada movimiento muestra qué regla lo clasificó, por ejemplo "Empresa del exterior → ingreso", y el contador puede auditarlas y corregirlas. Además respetan que **los datos no salgan de Wallbit**, porque no se mandan a un modelo externo. Cubren el caso típico (cobros de empresas del exterior, transferencias propias, gastos con tarjeta, conversiones a stablecoins). Lo dudoso queda "sin clasificar" y aparece en amarillo: prefiero pedirle 10 segundos al contador antes que adivinar un número impositivo.

### Sistema de semáforo que pense:
- 🔴 **Requiere acción:** excede el tope de la categoría K, facturas C después de la exclusión, cobros sin factura, depósitos por encima del tope.
- 🟡 **Revisar:** cerca del tope K o de su categoría, debe recategorizar, cobro fuera de lo habitual, movimientos sin clasificar, cobros declarados fuera de Wallbit, facturas sin cobro, gastos con tarjeta altos, puede bajar de categoría.
- 🟢 **Todo ok.**: No requiere cambios.

El color es la situación **más grave** del cliente, y la descripción explica esa situación en una línea. El color nunca va solo: siempre lo acompañan un ícono y un texto, para quien no distingue colores. El **estado del cierre** (esperando al cliente, para revisar, aprobado, notificado) va en otra columna. Mezclarlo con el semáforo haría que "todavía no subió facturas" se viera igual que "tiene un problema real".

### El modal de revisión
Se abre desde una fila, desde un chip de color o desde una selección, y recorre **exactamente la lista que el contador tenía filtrada**, en el mismo orden, con Anterior / Siguiente y atajos de teclado. La cola se congela al abrir: si un cliente cambia de color al aprobarlo, el índice no salta. Todo queda en la URL, así que recargar, compartir el link o usar el botón Atrás del navegador vuelve al mismo lugar.

### Resumen del cierre
La pestaña **Resumen** muestra, de arriba hacia abajo:
- **Estado de los clientes:** tarjetas Rojos, Amarillos y Verdes con su porcentaje del total y link directo a la tabla filtrada;
- **Indicadores del cierre:** avance de aprobación con estimación de horas de revisión, facturado del mes y total a pagar con próximo vencimiento;
- **Gráficos:** situaciones detectadas, clientes por categoría (y cerca del tope K) y vencimientos próximos;
- **¿Está funcionando el servicio?:** sección separada con % de verdes, % avisados antes del día 10 y % de pagos en término (con su detalle `X de Y`).

El selector de período permite alternar entre **octubre 2026** (el período en curso, con los cierres listos para revisar y la estimación de tiempo) y **septiembre 2026** (período cerrado y aprobado, donde se aprecian las métricas de éxito históricas de avisos y pagos en término). No es decorativo: cada número lleva a la tabla filtrada que lo explica.

### Aprobación: rápida pero auditada
- **En lote, solo los verdes.** Los amarillos y rojos los firma el contador uno por uno.
- Cada aprobación guarda **quién, cuándo, la versión del cálculo y un hash del resultado**. Si después cambia un dato, el cierre vuelve a "para revisar" con una versión nueva y la aprobación anterior queda en el historial.
- **Nunca se firma un número que no se vio:** si el cierre cambió mientras el contador lo miraba, la aprobación se rechaza y se le pide refrescar.

### Cómo se entera el usuario
- **Acción requerida** (subir facturas, declarar cobros externos): sale sin aprobación, porque no es asesoramiento.
- **Cierre listo:** solo después de la firma. Dentro de la app ve cuánto facturó, cuánto paga y hasta cuándo.
- **Email y push sin montos ni CUIT**, solo "Tenés novedades en Wallbit". El detalle se ve dentro de la app. Un test falla si un mensaje saliente contiene un monto.
- **Recordatorio** unos días antes del vencimiento si no marcó "Ya pagué".

En este prototipo, la parte del cliente es la API `/api/v1/me/**` (lista para que la consuma la app de Wallbit, se puede probar desde Swagger), y el contador ve en el modal exactamente el aviso que va a recibir el cliente.

### Monotributo y Responsable Inscripto
Incluí RI porque **es a donde va el freelancer que crece**: si factura más que el tope de la categoría K, queda excluido del Monotributo desde ese día. La herramienta calcula el acumulado de 12 meses día por día y detecta **la fecha exacta** en que se cruzó el tope. Esa fecha importa: desde ahí debería haber facturado como RI. El contador confirma el pase y la herramienta empieza a calcular IVA, autónomos y anticipos.

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
    └── parameters/           parámetros del contador
```

TanStack Query y TanStack Table, React Router (con el estado en la URL), Tailwind con los tokens de marca, Radix UI para accesibilidad, Recharts para el Resumen y Zod para validar las respuestas.

### Estados del cierre

```
PENDING_CLIENT_INPUT ──cliente confirma──► READY_FOR_REVIEW ──contador aprueba──► APPROVED ──aviso──► NOTIFIED
          └──────────── el contador puede aprobar igual ─────────────┘        ▲                       │
                                                         cambio de datos ─────┴───────────────────────┘
```

---

## API

La documentación completa está en **[Swagger](https://martinmgomezvega.github.io/wallbit-challenge/)** (requiere habilitar Pages una vez en *Settings → Pages → Source: GitHub Actions* en repo público o plan pago). En GitHub Pages es solo documentación; para probar los endpoints usá `http://localhost:8080/docs`. El contrato vive en [`backend/api/openapi.yaml`](backend/api/openapi.yaml) y un test verifica que cada ruta del router esté documentada.

Identidad de demo por header: `X-Demo-User: accountant` o `X-Demo-User: client:<clientId>`.

| Para | Endpoints principales |
|---|---|
| Contador | `GET /accountant/periods/{period}/closings` (+ `/summary`, `/ids`, `/stats`) · `GET /accountant/closings/{id}` · `POST .../approve` · `POST .../approve-batch` · `PATCH /accountant/movements/{id}/classification` · `POST /accountant/clients/{id}/regime-change` · `GET/PUT /accountant/parameters` |
| Cliente | `GET /me/closings` · `POST /me/invoices/import` · `PUT /me/periods/{period}/external-income` · `POST /me/periods/{period}/confirm` · `POST /me/closings/{id}/paid` · `GET /me/notifications` |

---

## Datos de demo

300 clientes generados con semilla fija (85% Monotributo, 15% RI, cobros de USD 800 a 9.000 por mes de empresas y plataformas del exterior) y **15 clientes armados a mano**, uno por cada situación. Los datos semilla cargan dos períodos: **septiembre 2026** como mes cerrado y aprobado (con notificaciones y pagos registrados para alimentar las métricas de éxito del servicio) y **octubre 2026** como mes en curso listo para revisar en el tablero. Los encontrás con el buscador del tablero:

| Cliente | Situación |
|---|---|
| Lucía Fernández | 🟢 Todo ok (Monotributo D) |
| Ignacio Sosa | 🔴 Superó el tope K el 14/10 → pase a RI |
| Carla Benítez | 🔴 Cobros sin factura |
| Diego Romero | 🟡 Tiene que recategorizar (C → E) |
| Valentina Ruiz | 🟡 Cerca del tope de su categoría |
| Tomás Álvarez | 🟡 Cobro fuera de lo habitual |
| Sofía Medina | 🟡 Movimiento sin clasificar |
| Joaquín Herrera | 🟡 Cerca del tope K |
| Camila Torres | 🟢 RI exportadora con saldo a favor de IVA |
| Federico Díaz | 🟡 RI con factura sin cobro + IVA a pagar |
| Agustina Castro | 🔴 Facturas C después de pasar a RI |
| Nicolás Peralta | 🔴 Depósitos por encima del tope K |
| Julieta Molina | 🟡 Declaró cobros fuera de Wallbit |
| Bruno Acosta | 🟡 Podría bajar de categoría |
| Paula Rojas | Esperando que suba sus facturas (su flujo se prueba por API: importar el CSV de `frontend/public/samples/` como `client:cli_demo_15` desde Swagger) |

Todos los nombres, CUITs y empresas son inventados.

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

### Vista del cliente
La experiencia del cliente (subir facturas, ver el resumen, marcar "Ya pagué") va dentro de la app de Wallbit. Dejé lista la API que la alimenta y la vista previa del aviso en el modal del contador.

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

## Calidad

- Tests unitarios con **coverage ≥ 90%** en backend y frontend (`make coverage` falla por debajo). Último resultado: backend 90.1% · frontend 98.2%.
- Backend: tests table-driven del dominio, `httptest` para los handlers, `-race` para el cierre concurrente, un benchmark del cálculo de Monotributo y un test end-to-end que verifica que cada cliente de demo cae en su situación.
- Integración: los repositorios Postgres pasan las mismas suites de contrato que los de memoria, más tests de transacciones y de concurrencia optimista.
- Frontend: Vitest + Testing Library + MSW, fixtures reales generados por el backend validados con Zod (`make fixtures`), y `vitest-axe` para accesibilidad en las pantallas principales.

---

## Si tuviera las 6 semanas

| Semana | Foco |
|---|---|
| 1 | Movimientos reales y reglas de clasificación calibradas con el contador; escalas y vencimientos reales. |
| 2 | Carga de facturas y conciliación; piloto interno con 20 clientes. |
| 3 | Liquidación de Monotributo, señales, tablero y modal de revisión. |
| 4 | RI y transición; aprobación auditada; avisos in-app y por email. |
| 5 | Piloto con 100 clientes reales; ajustar umbrales; medir con el Resumen. |
| 6 | Base gestionada, backups, observabilidad y lanzamiento. |

---

## Cómo trabajé

Trabajé con **Spec-Driven Development**: antes de escribir código, escribí las specs en [`specs/`](specs/). Ahí está el contexto del dominio, la arquitectura, las guías de UI y una carpeta por feature con `spec.md`, `plan.md` y `tasks.md`. Usé **Claude** para investigar la normativa 2026 y escribir las specs conmigo, y **Gemini en Google Antigravity** para implementarlas tarea por tarea. Las decisiones de producto, la revisión del código y los tests son míos.
