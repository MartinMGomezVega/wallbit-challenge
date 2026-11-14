# 001 — Tasks

- [x] **T001.1** `domain/money`: Money, Currency, Rate, BasisPoints, conversión half-even, Parse/Format decimal y localizado. Tests table-driven (incluir overflow-safe con montos de 10^12 centavos, half-even `0.5→0`, `1.5→2`, `2.5→2`).
- [x] **T001.2** `domain/period`: Parse, Start/End en AR, Prev/Next, ventana 12 meses, business days. Tests (cambio de año, febrero, fin de semana).
- [x] **T001.3** `domain/client/cuit.go`: ParseCUIT con dígito verificador + `Terminacion()`. Tests con CUITs válidos/inválidos.
- [x] **T001.4** `domain/tax/scale.go` + tests (`CategoryFor` en bordes exactos de cap, exceso de K).
- [x] **T001.5** Ports de §plan en `application/ports`.
- [x] **T001.6** Fixture `monotributo_scales.json` con **exactamente** los valores de `domain-context.md §2.1 y §2.2`. Test que verifica 3 valores puntuales por escala (A, H, K) contra la tabla.
- [x] **T001.7** Mock `arca` `MonotributoScale` + opciones (latencia, failEvery, ctx). Tests criterios 5 y 10.
- [x] **T001.8** Fixture `due_dates.json` (noviembre 2026 real) + `DueDates` con patrón referencial para otros meses. Tests criterio 7 (incluir un mes donde el día 20 cae sábado).
- [x] **T001.9** Completar `usd_bna.json`: octubre 2026 desde `domain-context.md §2.4`; oct-2025→sep-2026 y nov-2026 (1–13) desde dolarhistorico.com. **No inventar valores.** Test: el fixture está ordenado, sin duplicados, `buy < sell`, y `2026-10-01` = `1486.11/1539.87`.
- [x] **T001.10** Mock `fx` `USDQuote` con fallback al día hábil anterior. Tests criterio 8 y 10.
- [x] **T001.11** `domain/parameters` (Defaults, Validate) + `memory.ParametersRepository` + use cases Get/Update + handlers GET/PUT. Tests de validación y de handler (200, 422 con `fields`, 403 si rol cliente).
- [x] **T001.12** Wiring en `main.go`. `make coverage` ≥ 90%.
