# 007 — Tasks

- [x] **T007.1** `domain/closing`: tipos, máquina de estados, `Approval`. Tests de todas las transiciones válidas e inválidas.
- [x] **T007.2** Hash canónico + tests (mismo input → mismo hash; cambio de un centavo → otro hash; orden de maps no afecta).
- [x] **T007.3** `closing.Apply` (filtros, orden, desempate estable, paginación aparte). Tests por cada filtro y cada sort.
- [x] **T007.4** `ClosingRepository` en memoria (+ `LatestExclusionDate`). Tests de concurrencia.
- [x] **T007.5** `ComputeClosing` con fakes: MONO verde, MONO con exclusión, RI, cambio de régimen en el período, vencimientos del mes siguiente (`p.Next()`), versión no cambia si no hay cambios, re-apertura de aprobado.
- [x] **T007.6** `RunMonthlyClosing` con worker pool + idempotencia + error en un cliente. Test con `-race`.
- [x] **T007.7** `RecalcOnChange` (reemplaza el no-op) + `RecalcPeriod` conectado a PUT parameters + confirm input → READY.
- [x] **T007.8** Scheduler + test con fake clock.
- [x] **T007.9** Handlers: run, list, summary, detail. Tests httptest (filtros combinados, paginación, 422 por período inválido, 404).
- [x] **T007.10** Wiring en `main.go`: seed → job del mes anterior → servidor.
- [x] **T007.11** Test end-to-end con seed (criterio 17) + test de performance (criterio 10).
- [x] **T007.12** `make coverage` ≥ 90%.
