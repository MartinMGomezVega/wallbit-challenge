# 015 — Tasks

- [x] **T015.0** Limpieza: rutas `/ping` fuera del router de producción, borrar `backend/signal.out`, `docs/demo.md` y `specs/012-delivery/`, `*.out` en `.gitignore`, favicons enlazados en `frontend/index.html`.
- [x] **T015.1** `backend/api/embed.go` + rutas `GET /api/openapi.yaml` y `GET /docs`. Tests httptest (200, content-type, sin auth).
- [x] **T015.2** Router: `Routes() []string` + test de contrato contra `openapi.yaml`. Si falla, corregir el **código o el spec** según cuál se aparte de las specs de features (no borrar rutas del spec para que pase).
- [x] **T015.3** Recorrer `openapi.yaml` contra los handlers reales: nombres de campos, enums, códigos de error, endpoint `stats` (014). Ajustar el YAML donde difiera.
- [x] **T015.4** README: completar `⟨xx⟩%` con `make coverage`, verificar comandos, diagrama de secuencia y tabla de clientes de demo.
- [x] **T015.5** `NOTA_DE_DECISIONES.md`: ≤ 10 líneas renderizadas y todo verdadero.
- [ ] **T015.6** GitHub Pages: Settings → Pages → Source: "GitHub Actions"; correr el workflow y verificar que el Swagger carga con el logo y el spec.
- [x] **T015.7** Verificación desde carpeta limpia (criterio 9). `make lint` sin warnings; sin `TODO` pendientes; sin datos reales de personas.
- [x] **T015.8** Fixtures reales del backend en `frontend/src/test/fixtures/api/` validados con Zod (`make fixtures` incluido en `make test`) y workflow de GitHub Pages con `actions/configure-pages@v5` copiando `wallbit-logo-white.svg` y `favicon.svg`.
