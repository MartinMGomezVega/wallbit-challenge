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
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
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
	rec := do(t, h, "GET", "/api/v1/health", "", nil)
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
	rt.accountant("GET /api/v1/accountant/ping", testPing)
	rt.client("GET /api/v1/me/ping", testPing)
	return rt.handler()
}

func TestPermissions(t *testing.T) {
	h := testRouterWithPing()
	cases := []struct {
		name, path, user string
		status           int
		code             string
	}{
		{"accountant sin header", "/api/v1/accountant/ping", "", 401, "UNAUTHENTICATED"},
		{"cliente en accountant", "/api/v1/accountant/ping", "client:cli_x", 403, "FORBIDDEN"},
		{"accountant en me", "/api/v1/me/ping", "accountant", 403, "FORBIDDEN"},
		{"me sin header", "/api/v1/me/ping", "", 401, "UNAUTHENTICATED"},
		{"header inválido", "/api/v1/me/ping", "root", 401, "UNAUTHENTICATED"},
		{"cliente vacío", "/api/v1/me/ping", "client:", 401, "UNAUTHENTICATED"},
		{"accountant ok", "/api/v1/accountant/ping", "accountant", 200, ""},
		{"cliente ok", "/api/v1/me/ping", "client:cli_x", 200, ""},
		{"ruta inexistente", "/api/v1/nada", "accountant", 404, "NOT_FOUND"},
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
	rec := do(t, h, "GET", "/api/v1/me/ping", "client:cli_x", nil)
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
	do(t, h, "POST", "/api/v1/x", "", strings.NewReader(`{"secret":"20-12345678-9"}`))
	out := buf.String()
	if !strings.Contains(out, `"status":201`) || !strings.Contains(out, `"path":"/api/v1/x"`) || strings.Contains(out, "secret") {
		t.Fatal(out)
	}
}

func TestCORS(t *testing.T) {
	h := testRouterWithPing()
	req := httptest.NewRequest("OPTIONS", "/api/v1/accountant/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 204 || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" ||
		!strings.Contains(rec.Header().Get("Access-Control-Allow-Headers"), "X-Demo-User") {
		t.Fatal(rec.Code, rec.Header())
	}
	req = httptest.NewRequest("GET", "/api/v1/health", nil)
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
