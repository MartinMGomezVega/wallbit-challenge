# 000 — Foundation

## Objetivo
Dejar el monorepo listo para construir features: backend Go con capas vacías y servidor HTTP funcionando, frontend React con layout del contador, tooling de tests/coverage y `make dev`.

## Historias
- Como engineer, quiero clonar el repo y correr `make dev` para tener backend y frontend levantados.
- Como engineer, quiero `make test` y `make coverage` para validar la calidad en cada feature.
- Como contador, quiero navegar entre Cierres, Resumen y Parámetros desde un header claro con selector global de período.

## Criterios de aceptación
1. `GET /api/v1/health` responde `200 {"status":"ok","now":"<APP_NOW>"}`.
2. Middleware `X-Demo-User`:
   - sin header en `/api/v1/accountant/**` → `401 {"error":{"code":"UNAUTHENTICATED",...}}`;
   - `client:cli_x` en `/api/v1/accountant/**` → `403 FORBIDDEN`;
   - `accountant` en `/api/v1/me/**` → `403`.
3. Middlewares: recover (500 JSON), request log (`slog`, sin body), CORS con `CORS_ORIGIN`.
4. `ports.Clock` con implementación fija (`APP_NOW`) y real.
5. Test de arquitectura: falla si `internal/domain/**` importa algo de `internal/application|infrastructure|interfaces`, o si `internal/application/**` importa `infrastructure|interfaces`, o si aparece `float64` en `internal/domain/**`.
6. Frontend: layout oscuro (`#0A0A0A`, 56 px) con logo SVG oficial de Wallbit (24 px, `alt="Wallbit"`), pestañas **Cierres · Resumen · Parámetros**, selector global de período en URL (`PeriodSelect`), identidad `(C) Contador` (tooltip `"Sesión de demo"`) y `<title>` dinámico por página. Favicons SVG/PNG en `index.html`. Ruta `/` y cualquier ruta desconocida redirigen a `/contador`.
7. El http client del front (`shared/api/http.ts`) envía siempre `X-Demo-User: accountant` como constante de demo.
8. `make coverage` falla si backend o frontend < 90%.

## Fuera de alcance
Lógica de negocio. Base de datos. Auth real.
