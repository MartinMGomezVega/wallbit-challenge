# UI guidelines

> Sistema visual del prototipo. Todas las features de frontend (008–011) lo respetan. Si un componente no está acá, se construye con estos tokens; no se inventan colores ni tamaños nuevos.

## 1. Principios

1. **Herramienta de trabajo, no landing.** El contador pasa horas en la tabla: interfaz sobria, densa y calma. El color se reserva para **lo que requiere atención** (el semáforo). Todo lo demás, en grises.
2. **El estado se entiende de un vistazo**, y nunca solo por color: siempre color + ícono + texto (accesibilidad y daltonismo).
3. **Números primero.** Montos alineados a la derecha, con `tabular-nums`, moneda explícita (`US$` / `$`).
4. **Dos tonos para dos usuarios.** Contador: denso y técnico, desktop-first. Cliente: aireado, amable, sin jerga, mobile-first.
5. **Marca Wallbit.** Logo oficial en el header (ver §2.6) y azul de marca `#0D99FF`. Como ese azul no llega a contraste AA con texto blanco ni sobre fondos claros, los **controles** usan una variante más oscura del mismo azul (`--color-primary`); el azul de marca se usa en el logo, en el header oscuro y en acentos decorativos.

## 2. Tokens

Definidos como variables CSS en `src/shared/ui/tokens.css` y mapeados en `tailwind.config.ts` (`theme.extend.colors`). Los componentes usan las clases de Tailwind mapeadas, **nunca hex sueltos**.

### 2.1 Color

| Token | Valor | Uso |
|---|---|---|
| `--color-bg` | `#F1F1F1` | Fondo de la app |
| `--color-surface` | `#FFFFFF` | Tarjetas, tabla, modal |
| `--color-surface-muted` | `#F7F7F7` | Encabezado de tabla, hover, zonas secundarias dentro de superficies blancas |
| `--color-border` | `#E5E5E5` | Bordes y divisores |
| `--color-text` | `#171717` | Texto principal |
| `--color-text-muted` | `#6B6B6B` | Texto secundario (CUIT, ARS bajo USD, ayudas). 5,3:1 sobre blanco, 4,7:1 sobre `#F1F1F1` |
| `--color-header-bg` | `#0A0A0A` | Header (negro de marca) |
| `--color-header-text` | `#FFFFFF` | Texto e íconos del header |
| `--color-brand` | `#0D99FF` | Azul de marca: logo, indicador de pestaña activa del header, acentos decorativos. **No** para texto ni fondos de botones con texto blanco (2,99:1) |
| `--color-primary` | `#006FC0` | Controles: botón primario, links, checkbox activo. 5,2:1 con blanco |
| `--color-primary-hover` | `#005FA8` | |
| `--color-primary-soft` | `#E6F4FF` | Fila seleccionada, chips activos |
| `--color-focus` | `#006FC0` | Anillo de foco (2 px + offset 2 px). En el header oscuro usar `--color-brand` |

**Semáforo** (contraste verificado sobre blanco: ícono ≥ 3:1, texto ≥ 4.5:1):

| Severidad | Formas / Fill | Ícono / Texto | Fondo suave | Borde | Ícono (lucide) | Label |
|---|---|---|---|---|---|---|
| RED | `#DC2626` | `#B91C1C` | `#FEF2F2` | `#FECACA` | `OctagonAlert` | "Requiere acción" |
| YELLOW | `#EAB308` | `#A16207` | `#FEFCE8` | `#FDE047` | `TriangleAlert` | "Revisar" |
| GREEN | `#059669` | `#047857` | `#ECFDF5` | `#A7F3D0` | `CircleCheck` | "Todo ok" |

> Regla de aplicación para el amarillo:
> - Usar **fill** (`#EAB308` / `--color-severity-yellow-fill`) en formas geométricas y rellenos: punto del semáforo en la tabla, borde izquierdo de tarjetas de severidad, barras del gráfico de situaciones y borde de alerta en distribución.
> - Usar **icon / text** (`#A16207` / `--color-severity-yellow-icon`, `--color-severity-yellow-text`) en íconos vectoriales y textos (chips, labels, StatusBadge).
> - Nunca usar amarillo `#F59E0B` para texto o íconos sobre blanco (no llega a contraste).
> - Los elementos del semáforo (íconos, chips, textos) van **siempre sobre superficies blancas** (tabla, tarjetas, modal), no directamente sobre el fondo `#F1F1F1`, donde el amarillo pierde contraste.

**Barra de progreso de acumulado:**
- Umbral normal (< 80%): verde (`--color-severity-green-icon`: `#059669`).
- Umbral de advertencia (80–99%): naranja (`--color-progress-warning`: `#EA580C`). Se utiliza naranja para diferenciar nítidamente este umbral de límite fiscal del amarillo del semáforo.
- Umbral crítico (≥ 100%): rojo (`--color-severity-red-icon`: `#DC2626`).

**Estados del cierre** (badges neutros, sin competir con el semáforo):

| Estado | Fondo | Texto | Texto visible |
|---|---|---|---|
| `PENDING_CLIENT_INPUT` | `#F5F5F5` | `#525252` | Esperando al cliente |
| `READY_FOR_REVIEW` | `#E6F4FF` | `#005FA8` | Para revisar |
| `APPROVED` | `#F0F9FF` | `#0369A1` | Aprobado |
| `NOTIFIED` | `#ECFDF5` | `#047857` | Notificado |
| `paidAt` | borde `#A7F3D0`, texto `#047857` | | Pagado dd/MM |

### 2.2 Tipografía
- Fuente: **Inter**, self-hosted con `@fontsource-variable/inter` (sin CDN externo). Fallback `system-ui, sans-serif`.
- Números: `font-variant-numeric: tabular-nums` en toda celda/valor monetario o porcentaje.

| Token | Tamaño / alto de línea / peso | Uso |
|---|---|---|
| `text-xs` | 12 / 16 / 500 | Encabezados de tabla (MAYÚSCULAS, `tracking-wide`), badges, ayudas |
| `text-sm` | 13 / 20 / 400 | **Base del contador** (celdas, formularios) |
| `text-base` | 15 / 22 / 400 | **Base del cliente** |
| `text-lg` | 18 / 26 / 600 | Títulos de sección, nombre en el modal |
| `text-xl` | 22 / 28 / 600 | Título de página |
| `text-2xl` | 28 / 34 / 700 | Cifras destacadas (A pagar en la vista del cliente, KPIs) |

### 2.3 Espaciado, radios, sombras
- Escala de 4 px (Tailwind por defecto). Padding de página: 24 px (contador), 16 px (cliente mobile) / 24 px (cliente ≥ 640).
- Radios: `6px` controles y badges, `10px` tarjetas y modal, `9999px` chips y puntos.
- Sombras: solo dos — `shadow-sm` (tarjetas) y `shadow-xl` (modal, popovers). Nada de gradientes.

### 2.4 Movimiento
- Transiciones de 150 ms `ease-out` (hover, apertura de popovers). Modal: fade + scale 0.98 → 1 en 150 ms.
- Respetar `prefers-reduced-motion: reduce` (sin animaciones).
- Resaltado de la fila al cerrar el modal (009): fondo `--color-primary-soft` que se desvanece en 2 s.

### 2.5 Íconos
- **lucide-react**, 16 px en tabla y botones, 20 px en encabezados, `stroke-width 2`. Siempre con `aria-hidden` cuando hay texto al lado; con `aria-label` cuando están solos.

### 2.6 Logo y assets de marca
Archivos en `specs/assets/brand/` → copiar a `frontend/src/assets/brand/` (logos) y `frontend/public/` (favicons):

| Archivo | Uso |
|---|---|
| `wallbit-logo-white.svg` (isotipo azul + texto blanco) | **Header oscuro** del contador. Variante principal. |
| `wallbit-logo-dark.svg` (isotipo azul + texto `#0A0A0A`) | Fondos claros: estados vacíos, página `/_ui`. |
| `wallbit-isotype.svg` (solo isotipo azul) | Espacios chicos: modal en mobile, loader. |
| `favicon.svg` | `<link rel="icon" type="image/svg+xml">` (principal). |
| `favicon-32.png`, `favicon-64.png` | `<link rel="icon">` fallback PNG. |
| `favicon-180.png` (isotipo sobre negro) | `apple-touch-icon`. |

Reglas:
- Alto del logo en el header: **24 px** (ancho automático). Isotipo solo: mínimo 20 px.
- Espacio libre alrededor del logo: al menos la mitad de su alto.
- No recolorear, estirar, rotar ni agregar sombras. Solo las variantes SVG oficiales provistas.
- `alt="Wallbit"` en el logo con texto; el isotipo decorativo lleva `alt=""`.
- Separación de **32 px** (`gap-8`) entre el logo y la primera pestaña; sin texto "Contabilidad" en el header (queda en el `<title>` de cada página).

## 3. Layout

### 3.1 Contador (desktop-first, mínimo soportado 1280 px)
```
┌──────────────────────────────────────────────────────────────────────────┐
│ [logo wallbit 24px]   Cierres  Resumen  Parámetros   Período [octubre 2026 ▾]  (C) Contador │  ← header 56 px, fondo #0A0A0A, logo blanco 24 px; navegación activa con subrayado 2 px #0D99FF
├──────────────────────────────────────────────────────────────────────────┤
│ Cierre de octubre 2026                                   300 clientes     │  ← título + subtítulo muted
│ [● Rojos 12  Revisar] [▲ Amarillos 47  Revisar] [✓ Verdes 241]           │  ← chips de resumen (toggle)
│ [Situación ▾] [Régimen ▾] [Estado ▾] [🔍 Buscar nombre o CUIT ] Limpiar  │  ← barra de filtros
│ ┌──────────────────────────────────────────────────────────────────────┐ │
│ │ ☐ │ ● │ SITUACIÓN            │ CLIENTE │ RÉG. │ FACT. │ ACUM. │ … │   │ │  ← tabla full width
│ └──────────────────────────────────────────────────────────────────────┘ │
│ Mostrando 1–50 de 300                         ‹ 1 2 3 … 6 ›   50 ▾       │
└──────────────────────────────────────────────────────────────────────────┘
```
- Contenido a ancho completo (la tabla necesita el espacio). Fondo `--color-bg`; tabla dentro de una tarjeta `surface`.
- Barra de selección: flotante, centrada abajo, `shadow-xl`, aparece con ≥ 1 seleccionado.

### 3.2 Cliente
Fuera del alcance del prototipo: la vista del cliente vive en la app de Wallbit.

## 4. Componentes

### 4.1 Tabla de cierres (008)
- Encabezado sticky, fondo `surface-muted`, `text-xs` MAYÚSCULAS muted, alto 40 px. Todas las columnas de datos son ordenables server-side con ícono `ArrowUpDown` / `ArrowUp` / `ArrowDown` y `aria-sort`.
- Filas de 52 px, borde inferior `--color-border`, hover `surface-muted` como resaltado visual, seleccionada `primary-soft`. Sin cursor pointer general en la fila: el texto se puede seleccionar y copiar libremente.
- Apertura del modal de revisión: únicamente a través del botón "Revisar" o haciendo click en el nombre del cliente (como link).
- **Fila roja:** borde izquierdo de 3 px `#DC2626` en la celda fija izquierda (no se pinta la fila entera: el rojo se mantiene como acento, no como ruido).
- Columnas fijas a la izquierda (sticky): Checkbox (44 px), Semáforo (44 px, punto de 10 px + ícono de 16 px) y Cliente (180 px con borde divisor derecho). Al hacer scroll horizontal en pantallas pequeñas (1024 px), estas tres columnas permanecen visibles.
- Columna Situación: flexible, 2 líneas máximo (`line-clamp-2`), tooltip con el texto completo, chip de la señal principal y chip `+N` neutro.
- Columna Régimen: ancho confortable (`min-w-[140px]`, `whitespace-nowrap`), muestra `Monotributo [Categoría]` o `Resp. Inscripto` con tooltip "Responsable Inscripto".
- Montos y porcentajes: `whitespace-nowrap`, `tabular-nums`, alineados a la derecha y con anchos que nunca cortan cifras grandes (ej. "$ 1.614.446,04" entra completo).
- Acum. 12 meses: barra de 6 px de alto y 96 px de ancho, fondo `border`, relleno según umbral (verde < 80%, naranja `--color-progress-warning` #EA580C 80–99%, rojo `#DC2626` ≥ 100%), porcentaje a la derecha.
- Skeleton de carga: 8 filas grises con shimmer (desactivado con reduced-motion).

### 4.2 Chips de resumen
- Altura 36 px, radio completo, borde de su severidad, fondo blanco; **activo** = fondo suave de su severidad + borde 2 px.
- Contador en `font-semibold tabular-nums`. Botón "Revisar" interno como link discreto separado por un divisor vertical.

### 4.3 Botones
| Variante | Estilo | Uso |
|---|---|---|
| Primary | fondo `primary`, texto blanco | "Aprobar y siguiente", "Guardar" |
| Secondary | fondo blanco, borde, texto `text` | "Aprobar", "Revisar", "Cancelar" |
| Ghost | sin fondo ni borde | acciones de tabla, "Limpiar filtros" |
| Danger-soft | fondo `#FEF2F2`, texto `#B91C1C` | "Confirmar pase a RI" |

Altura: 32 px. Foco visible siempre (anillo `--color-focus`).

### 4.4 Modal de revisión (009)
- Tamaño `min(1120px, 94vw)` × `88vh`, radio 10 px, `shadow-xl`, overlay `rgba(10,10,10,0.5)`.
- **Header sticky** (64 px): `← Anterior` | semáforo + nombre (`text-lg`) + CUIT muted + badge de régimen y estado | "Cliente 3 de 47" + filtro activo en muted | `Siguiente →` | `✕`. Barra de progreso de 2 px abajo (revisados / total).
- **Cuerpo**: tabs (Resumen, Señales, Movimientos, Facturas, Aviso al cliente, Historial) con scroll interno; la pestaña Resumen abre por defecto.
- **Footer sticky** (64 px): "Nota para el cliente" (input) a la izquierda; a la derecha "Confirmar pase a RI" (si aplica), "Aprobar", "Aprobar y siguiente" (primary). Atajos visibles como `kbd` chicos (`←` `→` `A` `Esc`).
- Aviso "Ahora 🟢 Verde" / "Este cliente ya no cumple el filtro": banner de 36 px con fondo suave de la severidad nueva, debajo del header.

### 4.5 Tooltips, toasts y diálogos
- Tooltip (Radix): fondo `#171717`, texto blanco `text-xs`, delay 300 ms, también en foco.
- Toast (`@radix-ui/react-toast`): abajo a la derecha, 4 s, variantes success / error.
- Diálogos de confirmación: 440 px de ancho, título + 1 párrafo + 2 botones.

### 4.6 Vista del cliente (011)
Fuera del alcance del prototipo: la vista del cliente vive en la app de Wallbit.

### 4.7 Resumen (dashboard, 014)
- Grilla de 12 columnas, gap 16 px, ordenada de arriba hacia abajo en 4 bloques:
  1. **Estado de los clientes (`SeverityOverview`):** 3 tarjetas de 4 columnas (**Rojos · Amarillos · Verdes**) con borde izquierdo de 4 px del color fill de su severidad (`#EAB308` para amarillo), ícono, cantidad grande, "% del total" y link "Ver en la tabla →" (`/contador?period=…&severity=…`).
  2. **Indicadores del cierre:** 3 tarjetas de 4 columnas (Progreso del cierre, Facturado del mes, A pagar estimado con próximo vencimiento real). Sin KPI "Todo ok".
  3. **Gráficos (Recharts):** "Situaciones detectadas" 6 columnas, "Distribución por categoría" 6, "Vencimientos próximos" 12.
  4. **Sección separada "¿Está funcionando el servicio?" (`ServiceHealth`):** separador superior, título `text-lg`, subtítulo muted y 3 tarjetas de 4 columnas (% en verde, % avisados antes del día 10, % pagos en término), cada una con **"X de Y"** debajo.
- **KPI:** tarjeta blanca, padding 20 px, label `text-xs` MAYÚSCULAS muted arriba, valor `text-2xl` `tabular-nums`, línea secundaria `text-sm` muted. Toda la tarjeta clickeable si tiene drill-down (cursor pointer, hover `surface-muted`, foco visible).
- **Barra de avance apilada:** alto 8 px, radio completo, segmentos en este orden: Notificado `#047857`, Aprobado `#0369A1`, Para revisar `#006FC0` al 40% de opacidad, Esperando al cliente `#D4D4D4`.
- **Gráficos (Recharts):** sin fondo propio (van dentro de una tarjeta blanca), grilla horizontal `#E5E5E5` punteada, sin grilla vertical, ejes `text-xs` `#6B6B6B` sin línea de eje, barras con radio 4 px en el extremo, ancho máximo de barra 28 px, tooltip con el estilo de §4.5. Colores: barras de situaciones con el color de **fill** de su severidad (`#EAB308` para amarillo); categorías y vencimientos en `--color-primary`. Nada de gradientes, 3D ni tortas.
- Título de tarjeta de gráfico `text-lg`, subtítulo muted de una línea que explica qué mirar ("Click en una barra para ver esos clientes").

## 5. Estados vacíos, carga y error
- Vacío: ícono de 32 px muted + título corto + una acción ("Limpiar filtros").
- Carga: skeletons con la forma del contenido; nunca spinners de página completa.
- Error: tarjeta con borde rojo suave, mensaje en lenguaje humano + "Reintentar". El `code` técnico solo en `title` (tooltip).

## 6. Redacción
- Español rioplatense con voseo ("Subí", "Tenés", "Revisá").
- **Contador:** términos técnicos permitidos (categoría, tope K, débito fiscal, RG).
- **Cliente:** sin jerga. "Cuota del Monotributo", no "impuesto integrado"; "lo que facturaste", no "ingresos brutos devengados".
- Botones con verbo + objeto ("Aprobar y siguiente", "Subir facturas"). Sin signos de exclamación salvo en la confirmación final del cliente.
- Fechas `dd/MM/yyyy` (en tabla `dd/MM`); meses en minúscula sin `capitalize` en encabezados ("Cierre de octubre 2026", "Resumen de octubre 2026"); montos `$ 1.234.567,89` y `US$ 1.234,56`.

## 7. Accesibilidad (checklist por componente)
- Contraste AA en texto; 3:1 en íconos y bordes significativos.
- Todo operable con teclado; orden de foco lógico; foco atrapado en modal y diálogos.
- `aria-live="polite"` para el resultado de importación y para el contador "Cliente 3 de 47".
- Tests de accesibilidad con `vitest-axe` en DashboardPage, ReviewModal y SummaryPage (0 violaciones serias).

## 8. Fuera de alcance
Modo oscuro (los tokens quedan preparados como variables), responsive del tablero del contador por debajo de 1280 px, internacionalización.
