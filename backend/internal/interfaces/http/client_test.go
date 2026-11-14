package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func setupClientTestHandler(t *testing.T) (http.Handler, *memory.ClientRepository, *memory.ClosingRepository, *memory.NotificationRepository, *memory.ClientInputRepository, *memory.InvoiceRepository) {
	t.Helper()
	ctx := context.Background()

	clk := clock.NewFixed(time.Date(2026, 11, 2, 10, 0, 0, 0, clock.ARLocation()))
	clientRepo := memory.NewClientRepository()
	closingRepo := memory.NewClosingRepository()
	notifRepo := memory.NewNotificationRepository()
	inputRepo := memory.NewClientInputRepository()
	invRepo := memory.NewInvoiceRepository()
	idGen := idgen.NewSequential()

	cuit1, _ := client.ParseCUIT("20-12345678-6")
	cuit2, _ := client.ParseCUIT("27-32111110-0")

	cli1 := client.Client{
		ID:                  "cli_demo_15",
		Name:                "Paula Rojas",
		CUIT:                cuit1,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
	}
	cli2 := client.Client{
		ID:     "cli_demo_01",
		Name:   "Lucía Fernández",
		CUIT:   cuit2,
		Regime: client.RegimeRI,
	}
	_ = clientRepo.SaveAll(ctx, []client.Client{cli1, cli2})

	deps := Deps{
		Clock:            clk,
		ClientRepo:       clientRepo,
		ClosingRepo:      closingRepo,
		NotificationRepo: notifRepo,
		ClientInputRepo:  inputRepo,
		InvoiceRepo:      invRepo,
		IDGen:            idGen,
	}
	h := NewRouter(deps)
	return h, clientRepo, closingRepo, notifRepo, inputRepo, invRepo
}

func TestHTTP_DemoClients(t *testing.T) {
	h, _, _, _, _, _ := setupClientTestHandler(t)

	// GET /api/v1/demo/clients es público
	req := httptest.NewRequest("GET", "/api/v1/demo/clients", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var clients []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &clients)
	if len(clients) != 2 {
		t.Fatalf("len = %d, want 2", len(clients))
	}
	// cli_demo_01 va antes de cli_demo_15
	if clients[0]["id"] != "cli_demo_01" || clients[1]["id"] != "cli_demo_15" {
		t.Errorf("orden incorrecto: %+v", clients)
	}
}

func TestHTTP_GetMe(t *testing.T) {
	h, _, _, _, _, _ := setupClientTestHandler(t)

	// 1. Sin autenticación -> 401
	reqAnon := httptest.NewRequest("GET", "/api/v1/me", nil)
	recAnon := httptest.NewRecorder()
	h.ServeHTTP(recAnon, reqAnon)
	if recAnon.Code != 401 {
		t.Fatalf("status = %d, want 401", recAnon.Code)
	}

	// 2. Con rol contador -> 403
	reqAcc := httptest.NewRequest("GET", "/api/v1/me", nil)
	reqAcc.Header.Set("X-Demo-User", "accountant")
	recAcc := httptest.NewRecorder()
	h.ServeHTTP(recAcc, reqAcc)
	if recAcc.Code != 403 {
		t.Fatalf("status = %d, want 403", recAcc.Code)
	}

	// 3. Con rol cliente Paula Rojas -> 200 OK con CUIT enmascarado
	reqPaula := httptest.NewRequest("GET", "/api/v1/me", nil)
	reqPaula.Header.Set("X-Demo-User", "client:cli_demo_15")
	recPaula := httptest.NewRecorder()
	h.ServeHTTP(recPaula, reqPaula)
	if recPaula.Code != 200 {
		t.Fatalf("status = %d, want 200", recPaula.Code)
	}
	var me map[string]any
	_ = json.Unmarshal(recPaula.Body.Bytes(), &me)
	if me["name"] != "Paula Rojas" || me["cuitMasked"] != "20-****5678-6" || me["category"] != "A" {
		t.Errorf("datos me incorrectos: %+v", me)
	}
}

func TestHTTP_ClosingsAndPaid_OwnershipAndValidation(t *testing.T) {
	h, _, closingRepo, _, inputRepo, _ := setupClientTestHandler(t)
	ctx := context.Background()
	pOct := period.MustParse("2026-10")
	pSep := period.MustParse("2026-09")
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, clock.ARLocation())

	// Cierre de Paula en Sep (READY)
	cSep := closing.NewClosing("clo_paula_sep", "cli_demo_15", pSep, true, "h_sep", signal.SeverityGreen, nil, "Sep", nil, nil, closing.Reconciliation{}, now)
	_ = cSep.Approve(closing.Approval{Version: 1})
	_ = cSep.Publish(closing.Published{
		Version: 1, Title: "Septiembre listo",
		Figures: closing.PublishedFigures{Total: money.New(2660000, money.ARS)},
	})
	_ = closingRepo.Save(ctx, cSep)
	_ = inputRepo.Save(ctx, clientinput.ClientPeriodInput{ClientID: "cli_demo_15", Period: pSep, DeclaredNoExternalIncome: true, ConfirmedAt: &now})

	// Cierre de Paula en Oct (ACTION_REQUIRED)
	cOct := closing.NewClosing("clo_paula_oct", "cli_demo_15", pOct, false, "h_oct", signal.SeverityYellow, nil, "Oct", nil, nil, closing.Reconciliation{}, now)
	_ = closingRepo.Save(ctx, cOct)

	// Cierre de Lucía en Sep (READY)
	cLucia := closing.NewClosing("clo_lucia_sep", "cli_demo_01", pSep, true, "h_luc", signal.SeverityGreen, nil, "Lucia Sep", nil, nil, closing.Reconciliation{}, now)
	_ = cLucia.Approve(closing.Approval{Version: 1})
	_ = cLucia.Publish(closing.Published{Version: 1, Title: "Lucia listo"})
	_ = closingRepo.Save(ctx, cLucia)

	// 1. GET /api/v1/me/closings para Paula: sólo devuelve los de Paula
	reqList := httptest.NewRequest("GET", "/api/v1/me/closings", nil)
	reqList.Header.Set("X-Demo-User", "client:cli_demo_15")
	recList := httptest.NewRecorder()
	h.ServeHTTP(recList, reqList)
	if recList.Code != 200 {
		t.Fatalf("status = %d, want 200", recList.Code)
	}
	var closings []map[string]any
	_ = json.Unmarshal(recList.Body.Bytes(), &closings)
	if len(closings) != 2 {
		t.Fatalf("len = %d, want 2 (sólo Paula)", len(closings))
	}
	if closings[0]["closingId"] != "clo_paula_oct" || closings[0]["state"] != "ACTION_REQUIRED" {
		t.Errorf("cierre oct: %+v", closings[0])
	}
	if closings[1]["closingId"] != "clo_paula_sep" || closings[1]["state"] != "READY" {
		t.Errorf("cierre sep: %+v", closings[1])
	}

	// 2. Paula intenta pagar el cierre de Lucía -> 403 Forbidden!
	reqCross := httptest.NewRequest("POST", "/api/v1/me/closings/clo_lucia_sep/paid", nil)
	reqCross.Header.Set("X-Demo-User", "client:cli_demo_15")
	recCross := httptest.NewRecorder()
	h.ServeHTTP(recCross, reqCross)
	if recCross.Code != 403 {
		t.Fatalf("status = %d, want 403 al pagar cierre ajeno", recCross.Code)
	}

	// 3. Paula intenta pagar su propio cierre de Octubre que está en ACTION_REQUIRED -> 409 Conflict!
	reqNotReady := httptest.NewRequest("POST", "/api/v1/me/closings/clo_paula_oct/paid", nil)
	reqNotReady.Header.Set("X-Demo-User", "client:cli_demo_15")
	recNotReady := httptest.NewRecorder()
	h.ServeHTTP(recNotReady, reqNotReady)
	if recNotReady.Code != 409 {
		t.Fatalf("status = %d, want 409 al pagar cierre no READY", recNotReady.Code)
	}

	// 4. Paula paga su cierre de Septiembre (READY) -> 200 OK
	reqPaid := httptest.NewRequest("POST", "/api/v1/me/closings/clo_paula_sep/paid", nil)
	reqPaid.Header.Set("X-Demo-User", "client:cli_demo_15")
	recPaid := httptest.NewRecorder()
	h.ServeHTTP(recPaid, reqPaid)
	if recPaid.Code != 200 {
		t.Fatalf("status = %d, want 200 al pagar", recPaid.Code)
	}
	var paidResp map[string]any
	_ = json.Unmarshal(recPaid.Body.Bytes(), &paidResp)
	if paidResp["closingId"] != "clo_paula_sep" || paidResp["paidAt"] == nil {
		t.Errorf("respuesta pago: %+v", paidResp)
	}

	// 5. Idempotencia: segundo llamado a pagar devuelve 200 OK
	recPaid2 := httptest.NewRecorder()
	h.ServeHTTP(recPaid2, reqPaid)
	if recPaid2.Code != 200 {
		t.Fatalf("status = %d, want 200 en reintento idempotente", recPaid2.Code)
	}
}

func TestHTTP_Notifications_Ownership(t *testing.T) {
	h, _, _, notifRepo, _, _ := setupClientTestHandler(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 2, 10, 0, 0, 0, clock.ARLocation())

	nPaula := notification.Notification{
		ID:        "notif_paula",
		ClientID:  "cli_demo_15",
		Kind:      notification.KindClosingReady,
		Title:     "Paula notif",
		CreatedAt: now,
	}
	nLucia := notification.Notification{
		ID:        "notif_lucia",
		ClientID:  "cli_demo_01",
		Kind:      notification.KindClosingReady,
		Title:     "Lucia notif",
		CreatedAt: now,
	}
	_ = notifRepo.Save(ctx, nPaula)
	_ = notifRepo.Save(ctx, nLucia)

	// 1. GET /api/v1/me/notifications como Paula: sólo ve su notificación
	reqList := httptest.NewRequest("GET", "/api/v1/me/notifications", nil)
	reqList.Header.Set("X-Demo-User", "client:cli_demo_15")
	recList := httptest.NewRecorder()
	h.ServeHTTP(recList, reqList)
	if recList.Code != 200 {
		t.Fatalf("status = %d, want 200", recList.Code)
	}
	var notifs []map[string]any
	_ = json.Unmarshal(recList.Body.Bytes(), &notifs)
	if len(notifs) != 1 || notifs[0]["id"] != "notif_paula" {
		t.Fatalf("esperaba sólo notif de Paula, got: %+v", notifs)
	}

	// 2. Paula intenta marcar como leída la notificación de Lucía -> 403 Forbidden!
	reqCross := httptest.NewRequest("POST", "/api/v1/me/notifications/notif_lucia/read", nil)
	reqCross.Header.Set("X-Demo-User", "client:cli_demo_15")
	recCross := httptest.NewRecorder()
	h.ServeHTTP(recCross, reqCross)
	if recCross.Code != 403 {
		t.Fatalf("status = %d, want 403 al marcar notif ajena", recCross.Code)
	}

	// 3. Notificación inexistente -> 404
	req404 := httptest.NewRequest("POST", "/api/v1/me/notifications/notif_inexistente/read", nil)
	req404.Header.Set("X-Demo-User", "client:cli_demo_15")
	rec404 := httptest.NewRecorder()
	h.ServeHTTP(rec404, req404)
	if rec404.Code != 404 {
		t.Fatalf("status = %d, want 404", rec404.Code)
	}

	// 4. Paula marca como leída su notificación -> 200 OK
	reqOwn := httptest.NewRequest("POST", "/api/v1/me/notifications/notif_paula/read", nil)
	reqOwn.Header.Set("X-Demo-User", "client:cli_demo_15")
	recOwn := httptest.NewRecorder()
	h.ServeHTTP(recOwn, reqOwn)
	if recOwn.Code != 200 {
		t.Fatalf("status = %d, want 200", recOwn.Code)
	}
	saved, _ := notifRepo.Get(ctx, "notif_paula")
	if saved.ReadAt == nil {
		t.Fatal("ReadAt debió ser seteado")
	}
}
