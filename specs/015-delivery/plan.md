# 015 — Plan

## Punto de partida (estado real del repo)
- `README.md`, `NOTA_DE_DECISIONES.md`, `backend/api/openapi.yaml`, `docs/index.html`, `.github/workflows/pages.yml` y los favicons en `frontend/public/` llegan en el patch del usuario.
- El router registra `GET /api/v1/accountant/ping` y `GET /api/v1/me/ping` (de 000, para testear permisos). No están en el OpenAPI: moverlas a un router de test (o registrarlas solo en tests) para que el contrato quede limpio. No agregarlas al OpenAPI.
- Limpieza: borrar `backend/signal.out` y agregar `*.out` al `.gitignore`; borrar `docs/demo.md` (el README ya no tiene recorrido de demo); borrar `specs/012-delivery/` (reemplazada por 015).
- `frontend/index.html`: enlazar `favicon-32.png`, `favicon-64.png` y `apple-touch-icon` (`favicon-180.png`).

- Los documentos ya están escritos; este feature es de **verificación y ajuste**.
- `docs/index.html` detecta si corre en `/docs` (local, spec en `/api/openapi.yaml`) o en Pages (spec al lado).
- El workflow de Pages copia `backend/api/openapi.yaml`, `docs/index.html`, el logo blanco y el favicon a `_site`.
- Para la nota: contar líneas renderizadas en el preview de GitHub, no líneas del archivo.
