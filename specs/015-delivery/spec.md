# 015 — Entrega: README, nota de decisiones, Swagger

## Objetivo
Que alguien que nunca vio el proyecto lo levante con **un comando**, entienda las decisiones y pueda explorar la API. Todo lo que dicen los documentos tiene que ser cierto en la app.

## Base ya escrita (no reescribir: completar y ajustar a lo implementado)
- `README.md`, `NOTA_DE_DECISIONES.md` (raíz), `backend/api/openapi.yaml`, `docs/index.html`, `.github/workflows/pages.yml` ya existen. Placeholder pendiente: `⟨xx⟩%` (coverage real).
- Si lo implementado difiere de lo que dicen (un endpoint, un nombre, un número, un comando), se corrige el documento. **El README no puede prometer nada que la app no haga.**

## Criterios de aceptación
### README
1. Las instrucciones de "Cómo correrlo" funcionan tal cual en una carpeta limpia: `docker compose up --build` (opción principal) y `make dev` / `make dev-memory` (opción local).
2. El diagrama de secuencia (Mermaid) refleja el flujo real: job del día 1, carga del cliente, revisión, aprobación transaccional con outbox, aviso.
3. La tabla de clientes de demo coincide con el e2e de 007.
4. La sección "Lo que dejé afuera" está completa y no contradice lo construido.

### Nota de decisiones
5. ≤ 10 líneas renderizadas, en viñetas: qué hace la v1, qué quedó afuera, riesgo más grande, cómo sabría si funcionó, una línea de IA. Todo lo que afirma es cierto en la app.

### API documentada
6. `backend/api/embed.go` (`package api`, `//go:embed openapi.yaml`, `var Spec []byte`). `GET /api/openapi.yaml` sirve el spec (`application/yaml`) y `GET /docs` sirve `docs/index.html` (embebido) sin auth.
7. **Test de contrato** (`internal/interfaces/http/contract_test.go`): extrae de `openapi.yaml` las líneas `^  (/api/v1/[^:]+):$` y los métodos debajo (`^    (get|post|put|patch|delete):$`), las compara con `Routes() []string` del router (`"GET /api/v1/..."`) y falla si sobra o falta alguna. Sin dependencias YAML.
8. GitHub Pages publica el Swagger con el workflow incluido (si el repo lo permite; si no, queda documentado el Swagger local).

### Verificación final
9. Carpeta limpia: `docker compose up --build`, `make test`, `make coverage` (≥ 90%) y `make test-integration` pasan.
