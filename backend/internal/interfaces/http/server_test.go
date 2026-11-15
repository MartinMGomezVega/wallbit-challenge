package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

var testNow = time.Date(2026, 11, 5, 9, 0, 0, 0, clock.ARLocation())

func testDeps() Deps {
	return Deps{Clock: clock.NewFixed(testNow), CORSOrigin: "http://localhost:5173"}
}

func do(t *testing.T, h http.Handler, method, path, user string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if user != "" {
		req.Header.Set("X-Demo-User", user)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeErr(t *testing.T, rec *httptest.ResponseRecorder) errorPayload {
	t.Helper()
	var b errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("body no es error JSON: %s", rec.Body.String())
	}
	return b.Error
}

func TestHealth(t *testing.T) {
	h := NewRouter(testDeps())
	rec := do(t, h, "GET", "/health", "", nil)
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "ok" || body["now"] != "2026-11-05T09:00:00-03:00" {
		t.Fatal(body)
	}
}

func testRouterWithPing() http.Handler {
	rt := newRouter(testDeps())
	testPing := func(w http.ResponseWriter, r *http.Request) {
		id, _ := FromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]string{"role": id.Role, "clientId": id.ClientID})
	}
	rt.accountant("GET /v1/periods/ping", testPing)
	rt.clientResource("GET /v1/clients/{clientId}/ping", testPing)
	return rt.handler()
}

func TestPermissions(t *testing.T) {
	h := testRouterWithPing()
	cases := []struct {
		name, path, user string
		status           int
		code             string
	}{
		{"accountant sin header", "/v1/periods/ping", "", 401, "UNAUTHENTICATED"},
		{"cliente en accountant", "/v1/periods/ping", "client:cli_x", 403, "FORBIDDEN"},
		{"accountant en client resource", "/v1/clients/cli_x/ping", "accountant", 200, ""},
		{"cliente propio en client resource", "/v1/clients/cli_x/ping", "client:cli_x", 200, ""},
		{"cliente ajeno en client resource", "/v1/clients/cli_y/ping", "client:cli_x", 403, "FORBIDDEN"},
		{"client resource sin header", "/v1/clients/cli_x/ping", "", 401, "UNAUTHENTICATED"},
		{"header inválido", "/v1/clients/cli_x/ping", "root", 401, "UNAUTHENTICATED"},
		{"cliente vacío", "/v1/clients/cli_x/ping", "client:", 401, "UNAUTHENTICATED"},
		{"ruta inexistente", "/v1/nada", "accountant", 404, "NOT_FOUND"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := do(t, h, "GET", c.path, c.user, nil)
			if rec.Code != c.status {
				t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
			}
			if c.code != "" && decodeErr(t, rec).Code != c.code {
				t.Fatalf("code %s", rec.Body.String())
			}
		})
	}
	rec := do(t, h, "GET", "/v1/clients/cli_x/ping", "client:cli_x", nil)
	if !strings.Contains(rec.Body.String(), `"clientId":"cli_x"`) {
		t.Fatal(rec.Body.String())
	}
}

func TestRecover(t *testing.T) {
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }),
		Recover(slog.New(slog.NewTextHandler(io.Discard, nil))))
	rec := do(t, h, "GET", "/x", "", nil)
	if rec.Code != 500 || decodeErr(t, rec).Code != "INTERNAL" {
		t.Fatal(rec.Code, rec.Body.String())
	}
}

func TestRequestLogNoBody(t *testing.T) {
	var buf strings.Builder
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	n := 0
	now := func() time.Time { n++; return testNow.Add(time.Duration(n) * time.Millisecond) }
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(201) }), RequestLog(log, now))
	do(t, h, "POST", "/v1/x", "", strings.NewReader(`{"secret":"20-12345678-9"}`))
	out := buf.String()
	if !strings.Contains(out, `"status":201`) || !strings.Contains(out, `"path":"/v1/x"`) || strings.Contains(out, "secret") {
		t.Fatal(out)
	}
}

func TestCORS(t *testing.T) {
	h := testRouterWithPing()
	req := httptest.NewRequest("OPTIONS", "/v1/periods/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 204 || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" ||
		!strings.Contains(rec.Header().Get("Access-Control-Allow-Headers"), "X-Demo-User") {
		t.Fatal(rec.Code, rec.Header())
	}
	req = httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("no debería permitir otro origen")
	}
}

func TestWriteErrorMapping(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrNotFound, 404, "NOT_FOUND"},
		{fmt.Errorf("x: %w", domain.ErrNotFound), 404, "NOT_FOUND"},
		{domain.ValidationError{Field: "a", Msg: "b"}, 422, "VALIDATION_ERROR"},
		{domain.ValidationErrors{{Field: "a", Msg: "b"}}, 422, "VALIDATION_ERROR"},
		{domain.ErrValidation, 422, "VALIDATION_ERROR"},
		{domain.ErrConflict, 409, "CONFLICT"},
		{domain.ErrInvalidTransition, 409, "INVALID_TRANSITION"},
		{domain.ErrForbidden, 403, "FORBIDDEN"},
		{domain.ErrUnauthenticated, 401, "UNAUTHENTICATED"},
		{domain.NewError(domain.ErrConflict, "VERSION_CONFLICT", "cambió"), 409, "VERSION_CONFLICT"},
		{domain.NewError(domain.ErrValidation, "EXTERNAL_INCOME_NOT_ANSWERED", "x"), 422, "EXTERNAL_INCOME_NOT_ANSWERED"},
		{errors.New("boom"), 500, "INTERNAL"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		writeError(rec, c.err)
		p := decodeErr(t, rec)
		if rec.Code != c.status || p.Code != c.code {
			t.Errorf("%v → %d %s", c.err, rec.Code, p.Code)
		}
	}
	rec := httptest.NewRecorder()
	writeError(rec, domain.ValidationErrors{{Field: "fxRule", Msg: "desconocido"}})
	if decodeErr(t, rec).Fields["fxRule"] != "desconocido" {
		t.Fatal(rec.Body.String())
	}
	rec = httptest.NewRecorder()
	writeError(rec, errors.New("secreto interno"))
	if strings.Contains(rec.Body.String(), "secreto") {
		t.Fatal("no debe filtrar el mensaje interno")
	}
}

func TestIdentityHelpers(t *testing.T) {
	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("no debería haber identidad")
	}
	if id, ok := ParseIdentity(" accountant "); !ok || id.Role != RoleAccountant {
		t.Fatal(id)
	}
	if _, ok := ParseIdentity("client:a b"); ok {
		t.Fatal("id con espacio")
	}
}

func TestExhaustiveRouteAuthorization(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, period.Location())
	clk := clock.NewFixed(now)
	clientRepo := memory.NewClientRepository()
	closingRepo := memory.NewClosingRepository()
	paramRepo := memory.NewParametersRepository()
	movRepo := memory.NewMovementRepository()
	classRepo := memory.NewClassificationRepository()
	invRepo := memory.NewInvoiceRepository()
	purchRepo := memory.NewPurchaseInvoiceRepository()
	inputRepo := memory.NewClientInputRepository()
	notifRepo := memory.NewNotificationRepository()
	outboxRepo := memory.NewOutboxRepository()
	idGen := idgen.NewSequential()

	cuit1, _ := client.ParseCUIT("20-11111111-2")
	cuit2, _ := client.ParseCUIT("27-22222222-3")
	_ = clientRepo.Save(ctx, client.Client{ID: "cli_demo_01", Name: "Cliente 1", CUIT: cuit1, Regime: client.RegimeMonotributo})
	_ = clientRepo.Save(ctx, client.Client{ID: "cli_demo_02", Name: "Cliente 2", CUIT: cuit2, Regime: client.RegimeRI})

	h := NewRouter(Deps{
		Clock:               clk,
		ClientRepo:          clientRepo,
		ClosingRepo:         closingRepo,
		ParametersRepo:      paramRepo,
		MovementRepo:        movRepo,
		ClassificationRepo:  classRepo,
		InvoiceRepo:         invRepo,
		PurchaseInvoiceRepo: purchRepo,
		ClientInputRepo:     inputRepo,
		NotificationRepo:    notifRepo,
		OutboxRepo:          outboxRepo,
		IDGen:               idGen,
		ClosingQuery:        memory.NewClosingQuery(closingRepo, clientRepo),
		StatsQuery:          memory.NewStatsQuery(closingRepo),
	})

	t.Run("Rutas públicas sin autenticación", func(t *testing.T) {
		for _, path := range []string{"/health", "/v1/openapi.yaml"} {
			rec := do(t, h, "GET", path, "", nil)
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s sin auth: status = %d, want 200", path, rec.Code)
			}
		}
	})

	t.Run("Rutas protegidas sin autenticación devuelven 401 UNAUTHENTICATED", func(t *testing.T) {
		routes := []struct {
			method string
			path   string
		}{
			{"GET", "/v1/periods"},
			{"POST", "/v1/periods/2026-10/closing-runs"},
			{"GET", "/v1/periods/2026-10/closings"},
			{"GET", "/v1/periods/2026-10/closings/summary"},
			{"GET", "/v1/periods/2026-10/closings/ids"},
			{"GET", "/v1/periods/2026-10/stats"},
			{"GET", "/v1/closings/clo_1"},
			{"POST", "/v1/closings/clo_1/approvals"},
			{"POST", "/v1/periods/2026-10/approvals"},
			{"PUT", "/v1/closings/clo_1/payment"},
			{"PUT", "/v1/movements/mov_1/classification"},
			{"GET", "/v1/clients/cli_demo_01"},
			{"GET", "/v1/clients/cli_demo_01/closings"},
			{"GET", "/v1/clients/cli_demo_01/invoices"},
			{"POST", "/v1/clients/cli_demo_01/invoices"},
			{"POST", "/v1/clients/cli_demo_01/invoice-imports"},
			{"POST", "/v1/clients/cli_demo_01/purchase-invoices"},
			{"PUT", "/v1/clients/cli_demo_01/periods/2026-10/external-income"},
			{"POST", "/v1/clients/cli_demo_01/periods/2026-10/confirmation"},
			{"POST", "/v1/clients/cli_demo_01/regime-changes"},
			{"PUT", "/v1/clients/cli_demo_01/ri-profile"},
			{"GET", "/v1/clients/cli_demo_01/notifications"},
			{"PATCH", "/v1/clients/cli_demo_01/notifications/notif_1"},
			{"GET", "/v1/parameters"},
			{"PUT", "/v1/parameters"},
			{"GET", "/v1/signal-types"},
			{"POST", "/v1/reminder-runs"},
		}
		for _, r := range routes {
			rec := do(t, h, r.method, r.path, "", nil)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s %s sin auth: status = %d, want 401", r.method, r.path, rec.Code)
			}
		}
	})

	t.Run("Cliente accediendo a rutas exclusivas de contador devuelve 403 FORBIDDEN", func(t *testing.T) {
		accountantRoutes := []struct {
			method string
			path   string
		}{
			{"GET", "/v1/periods"},
			{"POST", "/v1/periods/2026-10/closing-runs"},
			{"GET", "/v1/periods/2026-10/closings"},
			{"GET", "/v1/periods/2026-10/closings/summary"},
			{"GET", "/v1/periods/2026-10/closings/ids"},
			{"GET", "/v1/periods/2026-10/stats"},
			{"POST", "/v1/closings/clo_1/approvals"},
			{"POST", "/v1/periods/2026-10/approvals"},
			{"PUT", "/v1/movements/mov_1/classification"},
			{"POST", "/v1/clients/cli_demo_01/regime-changes"},
			{"PUT", "/v1/clients/cli_demo_01/ri-profile"},
			{"GET", "/v1/parameters"},
			{"PUT", "/v1/parameters"},
			{"GET", "/v1/signal-types"},
			{"POST", "/v1/reminder-runs"},
		}
		for _, r := range accountantRoutes {
			rec := do(t, h, r.method, r.path, "client:cli_demo_01", nil)
			if rec.Code != http.StatusForbidden {
				t.Errorf("%s %s como client:cli_demo_01: status = %d, want 403", r.method, r.path, rec.Code)
			}
		}
	})

	t.Run("Cliente accediendo a recursos de otro cliente devuelve 403 FORBIDDEN", func(t *testing.T) {
		otherClientRoutes := []struct {
			method string
			path   string
		}{
			{"GET", "/v1/clients/cli_demo_02"},
			{"GET", "/v1/clients/cli_demo_02/closings"},
			{"GET", "/v1/clients/cli_demo_02/invoices"},
			{"POST", "/v1/clients/cli_demo_02/invoices"},
			{"POST", "/v1/clients/cli_demo_02/invoice-imports"},
			{"POST", "/v1/clients/cli_demo_02/purchase-invoices"},
			{"PUT", "/v1/clients/cli_demo_02/periods/2026-10/external-income"},
			{"POST", "/v1/clients/cli_demo_02/periods/2026-10/confirmation"},
			{"GET", "/v1/clients/cli_demo_02/notifications"},
			{"PATCH", "/v1/clients/cli_demo_02/notifications/notif_1"},
		}
		for _, r := range otherClientRoutes {
			rec := do(t, h, r.method, r.path, "client:cli_demo_01", nil)
			if rec.Code != http.StatusForbidden {
				t.Errorf("%s %s como client:cli_demo_01: status = %d, want 403", r.method, r.path, rec.Code)
			}
		}
	})

	t.Run("Cliente propio accede a sus recursos", func(t *testing.T) {
		// Cliente puede ver su propio perfil
		recCli := do(t, h, "GET", "/v1/clients/cli_demo_01", "client:cli_demo_01", nil)
		if recCli.Code != http.StatusOK {
			t.Errorf("GET /v1/clients/cli_demo_01 como client:cli_demo_01: status = %d, want 200", recCli.Code)
		}
	})

	t.Run("Contador tiene acceso a recursos operativos y clientes", func(t *testing.T) {
		for _, path := range []string{
			"/v1/periods",
			"/v1/parameters",
			"/v1/signal-types",
			"/v1/clients/cli_demo_01",
			"/v1/clients/cli_demo_02",
		} {
			rec := do(t, h, "GET", path, "accountant", nil)
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s como accountant: status = %d, want 200", path, rec.Code)
			}
		}
	})
}
