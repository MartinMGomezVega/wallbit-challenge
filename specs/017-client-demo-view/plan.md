# 017 — Plan

## Prerrequisitos
- 016 completa y commiteada (importador del contador, fixes del modal: schemas de RI y aprobación).
- La API del cliente (011) existe: `/me`, `/me/closings`, `/me/invoices` (GET/POST), `/me/invoices/import`, `/me/periods/{period}/external-income`, `/me/periods/{period}/confirm`, `/me/closings/{id}/paid`.

## Backend
- `internal/config`: `AccountantDisplayName`, `AccountantLicense`.
- `domain/closing`: ampliar `Published` con `ValidatedBy{Name, License}`, `Regime`, `Category *tax.Category`, `PctOfCategoryCapBp *int64`, `CategoryCapARS *money.Money`, `Billed12mARS money.Money`. `NotifyClosing` / `Publish` los completan desde el assessment de la versión aprobada y la config.
- Seed: aprobar y notificar octubre de `cli_demo_01` y `cli_demo_09` reutilizando `ApproveClosing` (no escribir estados a mano), con el reloj fijo en 2026-11-04 10:00.
- Fixtures de `/me/**`.
- OpenAPI: `Published` con los campos nuevos.

## Frontend
```
src/app/ViewSwitch.tsx                # Contador | Cliente (demo)
src/features/client-demo/
├── index.ts, api.ts (apiFetch con { as: clientId }), schemas.ts
├── demoClients.ts                    # los 5 de la tabla de la spec
├── ClientDemoPage.tsx                # selector de cliente + PhoneFrame
├── PhoneFrame.tsx                    # 390×780, radio 36px, leyenda
├── CheckIn/
│   ├── InvoicesStep.tsx              # tabla + Alta de comprobante + importar CSV
│   ├── InvoiceFormDialog.tsx
│   ├── ExternalIncomeStep.tsx
│   └── ConfirmStep.tsx
├── InReview.tsx
├── ValidatedSummaryCard.tsx
└── __tests__/
```
- `shared/api/http.ts`: `apiFetch(path, schema, init, { as?: string })` → header `client:<id>` si viene `as`; si no, `accountant`. Test de ambos casos.
- Router: `/cliente/:clientId` y `/cliente` → `/cliente/cli_demo_15`.
- Reutilizar `ImportResultView` y el parseo del importador de 016.
- Estilo: tokens de `ui-guidelines.md`, base 15 px, tarjetas blancas sobre `#F1F1F1` dentro del marco, botones de 40 px de alto.

## Documentación
- `ui-guidelines.md` §3.2 y §4.6: reemplazar "fuera de alcance" por la vista demo (marco de celular, tarjeta validada, check-in).
- `challenge.md` §4.4: "El prototipo incluye una vista **demo** del cliente (marco de celular) contra la API real."
- 011: mencionar que la consume la vista demo (017).
- README:
  - en "La experiencia del cliente", agregar al final: "Para la demo, el switch **Cliente (demo)** muestra esta experiencia con 5 clientes de ejemplo."
  - en "Datos de demo", marcar con (demo) los 5 clientes y actualizar Lucía y Camila (octubre aprobado).
- NOTA_DE_DECISIONES: en "Qué dejé afuera", cambiar "la vista del cliente (queda la API para la app de Wallbit)" por "la app real del cliente (hay una vista demo contra la API)". Seguir en ≤ 10 líneas.
