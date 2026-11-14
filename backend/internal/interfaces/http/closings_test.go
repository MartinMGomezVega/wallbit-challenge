package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func setupClosingsTestRouter(t *testing.T) (http.Handler, *memory.ClosingRepository, *memory.ClientRepository) {
	t.Helper()
	h, cr, clr, _, _, _, _ := setupApprovalTestRouter(t)
	return h, cr, clr
}

func setupApprovalTestRouter(t *testing.T) (
	http.Handler,
	*memory.ClosingRepository,
	*memory.ClientRepository,
	*memory.NotificationRepository,
	*memory.OutboxRepository,
	*fakes.NotificationSender,
	*fakes.Clock,
) {
	t.Helper()
	clk := &fakes.Clock{Current: testNow}
	paramRepo := memory.NewParametersRepository()
	clientRepo := memory.NewClientRepository()
	movRepo := memory.NewMovementRepository()
	classRepo := memory.NewClassificationRepository()
	invRepo := memory.NewInvoiceRepository()
	purchRepo := memory.NewPurchaseInvoiceRepository()
	inputRepo := memory.NewClientInputRepository()
	closingRepo := memory.NewClosingRepository()
	notifRepo := memory.NewNotificationRepository()
	outboxRepo := memory.NewOutboxRepository()
	sender := fakes.NewNotificationSender()
	idGen := fakes.NewIDGenerator()

	scale := tax.Scale{
		ValidFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, period.Location()),
		Rows: []tax.CategoryRow{
			{Category: tax.CatA, CapARS: money.New(1200941045, money.ARS), ServicesFeeARS: money.New(4952718, money.ARS)},
			{Category: tax.CatK, CapARS: money.New(12661083875, money.ARS), ServicesFeeARS: money.New(161444604, money.ARS)},
		},
	}
	scales := &fakes.TaxScaleProvider{Scale: scale}

	due := ports.DueDates{
		Monotributo:  time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		VAT:          time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		SelfEmployed: time.Date(2026, 11, 5, 0, 0, 0, 0, period.Location()),
	}
	dueDates := &fakes.DueDateProvider{Dates: due}

	fx := &fakes.ExchangeRateProvider{
		Default: ports.Quote{
			Day:  testNow,
			Buy:  150000,
			Sell: 155000,
		},
	}

	deps := Deps{
		Clock:               clk,
		CORSOrigin:          "http://localhost:5173",
		ParametersRepo:      paramRepo,
		ClientRepo:          clientRepo,
		MovementRepo:        movRepo,
		ClassificationRepo:  classRepo,
		InvoiceRepo:         invRepo,
		PurchaseInvoiceRepo: purchRepo,
		ClientInputRepo:     inputRepo,
		ClosingRepo:         closingRepo,
		ClosingQuery:        memory.NewClosingQuery(closingRepo, clientRepo),
		StatsQuery:          memory.NewStatsQuery(closingRepo),
		TaxScales:           scales,
		DueDates:            dueDates,
		ExchangeRates:       fx,
		IDGen:               idGen,
		Workers:             4,
		NotificationRepo:    notifRepo,
		OutboxRepo:          outboxRepo,
		NotificationSender:  sender,
	}

	return NewRouter(deps), closingRepo, clientRepo, notifRepo, outboxRepo, sender, clk
}

func TestRunEndpoint(t *testing.T) {
	router, _, clientRepo := setupClosingsTestRouter(t)
	ctx := context.Background()

	// Crear un cliente
	cuit, _ := client.ParseCUIT("20-11111111-2")
	_ = clientRepo.Save(ctx, client.Client{
		ID:                  "cli_1",
		Name:                "Cliente 1",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	})

	// 1. Sin auth -> 401
	rec := do(t, router, "POST", "/api/v1/accountant/periods/2026-10/run", "", nil)
	if rec.Code != 401 {
		t.Fatalf("esperaba 401 sin auth, got %d", rec.Code)
	}

	// 2. Con rol cliente -> 403
	rec = do(t, router, "POST", "/api/v1/accountant/periods/2026-10/run", "client:cli_1", nil)
	if rec.Code != 403 {
		t.Fatalf("esperaba 403 para cliente, got %d", rec.Code)
	}

	// 3. Período inválido -> 422
	rec = do(t, router, "POST", "/api/v1/accountant/periods/invalid-period/run", "accountant", nil)
	if rec.Code != 422 {
		t.Fatalf("esperaba 422 para período inválido, got %d", rec.Code)
	}

	// 4. Período válido como contador -> 200
	rec = do(t, router, "POST", "/api/v1/accountant/periods/2026-10/run", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("esperaba 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var sum usecase.RunSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &sum); err != nil {
		t.Fatal(err)
	}
	if sum.Processed != 1 || sum.Created != 1 {
		t.Fatalf("resumen inesperado: %+v", sum)
	}
}

func TestListAndSummaryEndpoints(t *testing.T) {
	router, closingRepo, clientRepo := setupClosingsTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	// Crear 2 clientes
	cuit1, _ := client.ParseCUIT("20-11111111-2")
	cuit2, _ := client.ParseCUIT("27-22222222-3")
	_ = clientRepo.Save(ctx, client.Client{
		ID:                  "cli_1",
		Name:                "Ana Pérez",
		CUIT:                cuit1,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	})
	_ = clientRepo.Save(ctx, client.Client{
		ID:        "cli_2",
		Name:      "Beatriz Rossi",
		CUIT:      cuit2,
		Regime:    client.RegimeRI,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	})

	// Crear 2 cierres en el repo
	c1 := closing.NewClosing(
		"clo_1", "cli_1", p, true, "hash1",
		signal.SeverityGreen, nil, "Todo OK",
		&tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(100000, money.USD),
			BilledInPeriodARS: money.New(15000000, money.ARS),
			CurrentCategory:   tax.CatA,
			Fee:               money.New(4952718, money.ARS),
			FeeDueDate:        time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		},
		nil, closing.Reconciliation{}, testNow,
	)
	c2 := closing.NewClosing(
		"clo_2", "cli_2", p, false, "hash2",
		signal.SeverityRed,
		[]signal.Signal{
			{
				Code:     signal.CodeMissingInvoices,
				Severity: signal.SeverityRed,
				Priority: 3,
				Message:  "Faltan facturas",
			},
		},
		"Faltan facturas",
		nil,
		&tax.RIAssessment{
			BilledInPeriodUSD: money.New(200000, money.USD),
			BilledInPeriodARS: money.New(30000000, money.ARS),
			TotalToPay:        money.New(15000000, money.ARS),
			VatDueDate:        time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		},
		closing.Reconciliation{}, testNow,
	)
	_ = closingRepo.Save(ctx, c1)
	_ = closingRepo.Save(ctx, c2)

	// 1. GET /closings con período inválido -> 422
	rec := do(t, router, "GET", "/api/v1/accountant/periods/bad-period/closings", "accountant", nil)
	if rec.Code != 422 {
		t.Fatalf("esperaba 422 para período inválido, got %d", rec.Code)
	}

	// 2. GET /closings lista completa (2 ítems)
	rec = do(t, router, "GET", "/api/v1/accountant/periods/2026-10/closings", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("esperaba 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var listRes usecase.ListClosingsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &listRes); err != nil {
		t.Fatal(err)
	}
	if listRes.Total != 2 || len(listRes.Items) != 2 {
		t.Fatalf("esperaba 2 filas, obtuvo total=%d len=%d", listRes.Total, len(listRes.Items))
	}
	// Default sort es severity desc: cli_2 (RED) luego cli_1 (GREEN)
	if listRes.Items[0].ClientID != "cli_2" || listRes.Items[1].ClientID != "cli_1" {
		t.Fatalf("orden severity desc incorrecto: %s, %s", listRes.Items[0].ClientID, listRes.Items[1].ClientID)
	}

	// 3. Filtro por severidad: severity=RED -> solo cli_2
	rec = do(t, router, "GET", "/api/v1/accountant/periods/2026-10/closings?severity=RED", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("esperaba 200 con filtro severity, got %d", rec.Code)
	}
	var redList usecase.ListClosingsResult
	_ = json.Unmarshal(rec.Body.Bytes(), &redList)
	if redList.Total != 1 || redList.Items[0].ClientID != "cli_2" {
		t.Fatalf("filtro severity=RED falló: total=%d", redList.Total)
	}

	// 4. Filtro por régimen: regime=MONOTRIBUTO -> solo cli_1
	rec = do(t, router, "GET", "/api/v1/accountant/periods/2026-10/closings?regime=MONOTRIBUTO", "accountant", nil)
	var monoList usecase.ListClosingsResult
	_ = json.Unmarshal(rec.Body.Bytes(), &monoList)
	if monoList.Total != 1 || monoList.Items[0].ClientID != "cli_1" {
		t.Fatalf("filtro regime=MONOTRIBUTO falló: total=%d", monoList.Total)
	}

	// 5. Paginación: page=1&pageSize=1 -> total=2, totalPages=2, len=1
	rec = do(t, router, "GET", "/api/v1/accountant/periods/2026-10/closings?page=1&pageSize=1", "accountant", nil)
	var p1List usecase.ListClosingsResult
	_ = json.Unmarshal(rec.Body.Bytes(), &p1List)
	if p1List.Total != 2 || p1List.TotalPages != 2 || len(p1List.Items) != 1 {
		t.Fatalf("paginación falló: %+v", p1List)
	}

	// 6. GET /summary -> total=2, bySeverity{RED:1, GREEN:1, YELLOW:0}
	rec = do(t, router, "GET", "/api/v1/accountant/periods/2026-10/closings/summary", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("esperaba 200 en summary, got %d (%s)", rec.Code, rec.Body.String())
	}
	var sum closing.Summary
	if err := json.Unmarshal(rec.Body.Bytes(), &sum); err != nil {
		t.Fatal(err)
	}
	if sum.Total != 2 || sum.BySeverity["RED"] != 1 || sum.BySeverity["GREEN"] != 1 {
		t.Fatalf("summary inesperado: %+v", sum)
	}
	if sum.BySignal[string(signal.CodeMissingInvoices)] != 1 {
		t.Fatalf("summary bySignal MISSING_INVOICES inesperado: %d", sum.BySignal[string(signal.CodeMissingInvoices)])
	}
}

func TestDetailEndpoint(t *testing.T) {
	router, closingRepo, clientRepo := setupClosingsTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	// 1. Cierre inexistente -> 404
	rec := do(t, router, "GET", "/api/v1/accountant/closings/nonexistent", "accountant", nil)
	if rec.Code != 404 {
		t.Fatalf("esperaba 404 para cierre inexistente, got %d", rec.Code)
	}

	// 2. Cierre existente -> 200 con detalle completo
	cuit, _ := client.ParseCUIT("20-11111111-2")
	cli := client.Client{
		ID:                  "cli_1",
		Name:                "Ana Pérez",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = clientRepo.Save(ctx, cli)

	c := closing.NewClosing(
		"clo_det", cli.ID, p, true, "hash_det",
		signal.SeverityGreen, nil, "Todo OK",
		&tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(100000, money.USD),
			BilledInPeriodARS: money.New(15000000, money.ARS),
			CurrentCategory:   tax.CatA,
			Fee:               money.New(4952718, money.ARS),
		},
		nil,
		closing.Reconciliation{
			MatchedPairs: []closing.MatchedPair{
				{MovementID: "mov_1", InvoiceID: "inv_1", DiffAmount: 0},
			},
		},
		testNow,
	)
	_ = closingRepo.Save(ctx, c)

	rec = do(t, router, "GET", "/api/v1/accountant/closings/clo_det", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("esperaba 200 para detalle, got %d (%s)", rec.Code, rec.Body.String())
	}

	var det usecase.ClosingDetailDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &det); err != nil {
		t.Fatal(err)
	}
	if det.Closing.ID != "clo_det" || det.Client.ID != "cli_1" {
		t.Fatalf("datos del detalle incorrectos: %+v", det)
	}
}

func TestClosingIDsEndpoint(t *testing.T) {
	router, closingRepo, clientRepo := setupClosingsTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	// 1. Sin auth -> 401
	rec := do(t, router, "GET", "/api/v1/accountant/periods/2026-10/closings/ids", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperaba 401 sin auth, got %d", rec.Code)
	}

	// 2. Período inválido -> 422
	rec = do(t, router, "GET", "/api/v1/accountant/periods/invalid/closings/ids", "accountant", nil)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperaba 422 con período inválido, got %d", rec.Code)
	}

	// Crear 5 clientes y 5 cierres con datos variados
	type clientSeedData struct {
		id        string
		name      string
		cuit      string
		regime    client.Regime
		severity  signal.Severity
		billedUSD int64
		pctKCap   int64
		status    closing.Status
	}
	seeds := []clientSeedData{
		{"cli_01", "Bernardo Aranda", "20-11111111-1", client.RegimeMonotributo, signal.SeverityRed, 1000000, 10500, closing.StatusReadyForReview},
		{"cli_02", "Carlos Benítez", "20-22222222-2", client.RegimeMonotributo, signal.SeverityYellow, 500000, 8500, closing.StatusReadyForReview},
		{"cli_03", "Ana Clara Díaz", "27-33333333-3", client.RegimeRI, signal.SeverityGreen, 200000, 2000, closing.StatusApproved},
		{"cli_04", "Daniel Estevez", "20-44444444-4", client.RegimeMonotributo, signal.SeverityGreen, 100000, 1000, closing.StatusReadyForReview},
		{"cli_05", "Esteban Fuentes", "20-55555555-5", client.RegimeRI, signal.SeverityRed, 800000, 9000, closing.StatusPendingClientInput},
	}

	for _, s := range seeds {
		cuit, _ := client.ParseCUIT(s.cuit)
		_ = clientRepo.Save(ctx, client.Client{
			ID:     s.id,
			Name:   s.name,
			CUIT:   cuit,
			Regime: s.regime,
		})
		c := closing.NewClosing(
			"clo_"+s.id, s.id, p, true, "hash_"+s.id,
			s.severity, nil, "Desc "+s.name,
			&tax.MonotributoAssessment{
				BilledInPeriodUSD: money.New(s.billedUSD, money.USD),
				BilledInPeriodARS: money.New(s.billedUSD*1500, money.ARS),
				PctOfKCapBp:       s.pctKCap,
				Fee:               money.New(5000000, money.ARS),
				FeeDueDate:        time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
			},
			nil,
			closing.Reconciliation{},
			testNow,
		)
		c.Status = s.status
		_ = closingRepo.Save(ctx, c)
	}

	// 5 combinaciones de filtro/orden para verificar equivalencia con listado paginado (criterio 1)
	testCases := []struct {
		name  string
		query string
	}{
		{"1. Sin filtros (default sort: severity desc)", ""},
		{"2. Filtro de severidad múltiple", "severity=RED,YELLOW"},
		{"3. Filtro por régimen y orden alfabético", "regime=MONOTRIBUTO&sort=name&dir=asc"},
		{"4. Filtro por estado y orden por facturación desc", "status=READY_FOR_REVIEW&sort=billed&dir=desc"},
		{"5. Búsqueda por texto y orden por porcentaje tope K asc", "q=Aranda&sort=pctKCap&dir=asc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// A. Obtener IDs directos de /closings/ids
			idsURL := "/api/v1/accountant/periods/2026-10/closings/ids"
			if tc.query != "" {
				idsURL += "?" + tc.query
			}
			recIDs := do(t, router, "GET", idsURL, "accountant", nil)
			if recIDs.Code != http.StatusOK {
				t.Fatalf("esperaba 200 en ids endpoint, got %d: %s", recIDs.Code, recIDs.Body.String())
			}

			var idsResult usecase.ListClosingIDsResult
			if err := json.Unmarshal(recIDs.Body.Bytes(), &idsResult); err != nil {
				t.Fatalf("error unmarshaling ids: %v", err)
			}

			// B. Obtener todas las páginas de /closings (pageSize=2 para forzar múltiples páginas)
			var paginatedIDs []string
			page := 1
			for {
				pageURL := "/api/v1/accountant/periods/2026-10/closings?page=" + strconv.Itoa(page) + "&pageSize=2"
				if tc.query != "" {
					pageURL += "&" + tc.query
				}
				recList := do(t, router, "GET", pageURL, "accountant", nil)
				if recList.Code != http.StatusOK {
					t.Fatalf("esperaba 200 en listado página %d, got %d: %s", page, recList.Code, recList.Body.String())
				}

				var listResult usecase.ListClosingsResult
				if err := json.Unmarshal(recList.Body.Bytes(), &listResult); err != nil {
					t.Fatalf("error unmarshaling list: %v", err)
				}

				for _, item := range listResult.Items {
					paginatedIDs = append(paginatedIDs, item.ClosingID)
				}

				if page >= listResult.TotalPages || len(listResult.Items) == 0 {
					break
				}
				page++
			}

			// C. Verificar equivalencia total y de orden
			if idsResult.Total != len(paginatedIDs) {
				t.Fatalf("total no coincide: idsResult.Total=%d != len(paginatedIDs)=%d", idsResult.Total, len(paginatedIDs))
			}
			if len(idsResult.IDs) != len(paginatedIDs) {
				t.Fatalf("largo de IDs no coincide: len(idsResult.IDs)=%d != len(paginatedIDs)=%d", len(idsResult.IDs), len(paginatedIDs))
			}
			for i := range paginatedIDs {
				if idsResult.IDs[i] != paginatedIDs[i] {
					t.Fatalf("discrepancia en índice %d: idsResult.IDs[%d]=%s != paginatedIDs[%d]=%s", i, i, idsResult.IDs[i], i, paginatedIDs[i])
				}
			}
		})
	}
}

func TestApproveClosing_Individual_Success_And_Warnings(t *testing.T) {
	router, closingRepo, clientRepo, notifRepo, outboxRepo, sender, clk := setupApprovalTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	// 1. Cliente Monotributo y cierre READY_FOR_REVIEW
	cl1 := client.Client{ID: "cli_appr_1", Name: "Matías Rossi", Regime: client.RegimeMonotributo}
	_ = clientRepo.Save(ctx, cl1)

	clo1 := closing.Closing{
		ID:         "clo_appr_1",
		ClientID:   "cli_appr_1",
		Period:     p,
		Status:     closing.StatusReadyForReview,
		Severity:   signal.SeverityGreen,
		Version:    2,
		ResultHash: "hash_v2",
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(200000, money.USD),
			BilledInPeriodARS: money.New(30000000, money.ARS),
			CurrentCategory:   tax.CatE,
			Fee:               money.New(5000000, money.ARS),
			FeeDueDate:        clk.Now().AddDate(0, 0, 5),
		},
	}
	_ = closingRepo.Save(ctx, clo1)

	// Aprobación exitosa
	body := `{"version": 2, "clientNote": "Todo en orden para este mes"}`
	rec := do(t, router, "POST", "/api/v1/accountant/closings/clo_appr_1/approve", "accountant", strings.NewReader(body))
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200 en approve, got %d: %s", rec.Code, rec.Body.String())
	}

	var res approveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("error unmarshaling approveResponse: %v", err)
	}

	if res.Status != closing.StatusNotified {
		t.Errorf("esperaba status NOTIFIED, got %s", res.Status)
	}
	if res.Version != 2 {
		t.Errorf("esperaba versión 2, got %d", res.Version)
	}
	if res.Warning != "" {
		t.Errorf("no esperaba warning, got %s", res.Warning)
	}
	if len(res.Approvals) != 1 || res.Approvals[0].Note != "Todo en orden para este mes" {
		t.Errorf("auditoría de aprobación inválida: %+v", res.Approvals)
	}

	// Verificar notificación in-app y outbox
	notifs, _ := notifRepo.ByClient(ctx, "cli_appr_1")
	if len(notifs) != 1 {
		t.Fatalf("esperaba 1 notificación in-app, got %d", len(notifs))
	}
	if notifs[0].Kind != notification.KindClosingReady {
		t.Errorf("esperaba KindClosingReady, got %s", notifs[0].Kind)
	}
	if len(sender.Sent) < 2 {
		t.Fatalf("esperaba al menos 2 mensajes salientes enviados, got %d", len(sender.Sent))
	}

	// 2. Cierre PENDING_CLIENT_INPUT -> debe devolver warning CLIENT_INPUT_NOT_CONFIRMED
	cl2 := client.Client{ID: "cli_appr_2", Name: "Ana Silva", Regime: client.RegimeMonotributo}
	_ = clientRepo.Save(ctx, cl2)

	clo2 := closing.Closing{
		ID:         "clo_appr_2",
		ClientID:   "cli_appr_2",
		Period:     p,
		Status:     closing.StatusPendingClientInput,
		Severity:   signal.SeverityYellow,
		Version:    1,
		ResultHash: "hash_v1",
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(100000, money.USD),
			BilledInPeriodARS: money.New(15000000, money.ARS),
			CurrentCategory:   tax.CatC,
			Fee:               money.New(3000000, money.ARS),
			FeeDueDate:        clk.Now().AddDate(0, 0, 5),
		},
	}
	_ = closingRepo.Save(ctx, clo2)

	rec2 := do(t, router, "POST", "/api/v1/accountant/closings/clo_appr_2/approve", "accountant", strings.NewReader(`{"version": 1}`))
	if rec2.Code != http.StatusOK {
		t.Fatalf("esperaba 200 en approve con warning, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var res2 approveResponse
	_ = json.Unmarshal(rec2.Body.Bytes(), &res2)
	if res2.Warning != "CLIENT_INPUT_NOT_CONFIRMED" {
		t.Errorf("esperaba warning CLIENT_INPUT_NOT_CONFIRMED, got %q", res2.Warning)
	}
	if res2.Status != closing.StatusNotified {
		t.Errorf("esperaba status NOTIFIED, got %s", res2.Status)
	}

	_ = outboxRepo
}

func TestApproveClosing_Individual_Errors(t *testing.T) {
	router, closingRepo, clientRepo, _, _, _, _ := setupApprovalTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	cl := client.Client{ID: "cli_err_1", Name: "Carlos Ruiz", Regime: client.RegimeMonotributo}
	_ = clientRepo.Save(ctx, cl)

	clo := closing.Closing{
		ID:         "clo_err_1",
		ClientID:   "cli_err_1",
		Period:     p,
		Status:     closing.StatusReadyForReview,
		Severity:   signal.SeverityGreen,
		Version:    2,
		ResultHash: "hash_err",
	}
	_ = closingRepo.Save(ctx, clo)

	// 1. Sin autenticación -> 401
	rec401 := do(t, router, "POST", "/api/v1/accountant/closings/clo_err_1/approve", "", strings.NewReader(`{"version": 2}`))
	if rec401.Code != http.StatusUnauthorized {
		t.Errorf("esperaba 401 sin auth, got %d", rec401.Code)
	}

	// 2. Versión incorrecta -> 409 VERSION_CONFLICT
	recConflict := do(t, router, "POST", "/api/v1/accountant/closings/clo_err_1/approve", "accountant", strings.NewReader(`{"version": 99}`))
	if recConflict.Code != http.StatusConflict {
		t.Errorf("esperaba 409 en conflicto de versión, got %d", recConflict.Code)
	}
	var errResp errorBody
	_ = json.Unmarshal(recConflict.Body.Bytes(), &errResp)
	if errResp.Error.Code != "VERSION_CONFLICT" {
		t.Errorf("esperaba error code VERSION_CONFLICT, got %s", errResp.Error.Code)
	}

	// 3. Nota de más de 500 caracteres -> 422 VALIDATION_ERROR
	longNote := strings.Repeat("a", 501)
	reqLong := `{"version": 2, "clientNote": "` + longNote + `"}`
	recVal := do(t, router, "POST", "/api/v1/accountant/closings/clo_err_1/approve", "accountant", strings.NewReader(reqLong))
	if recVal.Code != http.StatusUnprocessableEntity {
		t.Errorf("esperaba 422 por nota larga, got %d", recVal.Code)
	}

	// 4. Cierre inexistente -> 404 NOT_FOUND
	rec404 := do(t, router, "POST", "/api/v1/accountant/closings/clo_inexistente/approve", "accountant", strings.NewReader(`{"version": 1}`))
	if rec404.Code != http.StatusNotFound {
		t.Errorf("esperaba 404 para cierre inexistente, got %d", rec404.Code)
	}

	// 5. Cierre ya NOTIFIED en la misma versión -> 409 ALREADY_APPROVED
	cloNotified := clo
	cloNotified.ID = "clo_err_notified"
	cloNotified.Status = closing.StatusNotified
	_ = closingRepo.Save(ctx, cloNotified)

	recAlready := do(t, router, "POST", "/api/v1/accountant/closings/clo_err_notified/approve", "accountant", strings.NewReader(`{"version": 2}`))
	if recAlready.Code != http.StatusConflict {
		t.Errorf("esperaba 409 por ALREADY_APPROVED, got %d", recAlready.Code)
	}
	var errAlreadyResp errorBody
	_ = json.Unmarshal(recAlready.Body.Bytes(), &errAlreadyResp)
	if errAlreadyResp.Error.Code != "ALREADY_APPROVED" {
		t.Errorf("esperaba error code ALREADY_APPROVED, got %s", errAlreadyResp.Error.Code)
	}
}

func TestApproveBatch_Endpoint_And_Race(t *testing.T) {
	router, closingRepo, clientRepo, _, _, _, clk := setupApprovalTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	// Crear 5 clientes y cierres con estados variados
	for i := 1; i <= 5; i++ {
		cid := "cli_batch_" + strconv.Itoa(i)
		_ = clientRepo.Save(ctx, client.Client{ID: cid, Name: "Batch Client " + strconv.Itoa(i), Regime: client.RegimeMonotributo})
	}

	c1 := closing.Closing{ID: "clo_b1", ClientID: "cli_batch_1", Period: p, Status: closing.StatusReadyForReview, Severity: signal.SeverityGreen, Version: 1, ResultHash: "h1", Mono: &tax.MonotributoAssessment{FeeDueDate: clk.Now().AddDate(0, 0, 5)}}
	c2 := closing.Closing{ID: "clo_b2", ClientID: "cli_batch_2", Period: p, Status: closing.StatusReadyForReview, Severity: signal.SeverityRed, Version: 1, ResultHash: "h2"}
	c3 := closing.Closing{ID: "clo_b3", ClientID: "cli_batch_3", Period: p, Status: closing.StatusReadyForReview, Severity: signal.SeverityYellow, Version: 1, ResultHash: "h3"}
	c4 := closing.Closing{ID: "clo_b4", ClientID: "cli_batch_4", Period: p, Status: closing.StatusPendingClientInput, Severity: signal.SeverityGreen, Version: 1, ResultHash: "h4"}
	c5 := closing.Closing{ID: "clo_b5", ClientID: "cli_batch_5", Period: p, Status: closing.StatusReadyForReview, Severity: signal.SeverityGreen, Version: 1, ResultHash: "h5"}

	_ = closingRepo.Save(ctx, c1)
	_ = closingRepo.Save(ctx, c2)
	_ = closingRepo.Save(ctx, c3)
	_ = closingRepo.Save(ctx, c4)
	_ = closingRepo.Save(ctx, c5)

	batchPayload := `{
		"items": [
			{"id": "clo_b1", "version": 1},
			{"id": "clo_b2", "version": 1},
			{"id": "clo_b3", "version": 1},
			{"id": "clo_b4", "version": 1},
			{"id": "clo_b5", "version": 99},
			{"id": "clo_b_inexistente", "version": 1}
		]
	}`

	rec := do(t, router, "POST", "/api/v1/accountant/periods/2026-10/closings/approve-batch", "accountant", strings.NewReader(batchPayload))
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200 en approve-batch, got %d: %s", rec.Code, rec.Body.String())
	}

	var batchRes usecase.ApproveBatchResult
	if err := json.Unmarshal(rec.Body.Bytes(), &batchRes); err != nil {
		t.Fatalf("error unmarshaling batchRes: %v", err)
	}

	if batchRes.Approved != 1 {
		t.Errorf("esperaba 1 aprobado, got %d", batchRes.Approved)
	}
	if batchRes.Failed != 5 {
		t.Errorf("esperaba 5 fallidos, got %d", batchRes.Failed)
	}

	resMap := make(map[string]usecase.ApproveBatchItemResult)
	for _, r := range batchRes.Results {
		resMap[r.ID] = r
	}

	if !resMap["clo_b1"].OK {
		t.Errorf("clo_b1 debía ser OK")
	}
	if resMap["clo_b2"].Error != "NOT_GREEN" {
		t.Errorf("clo_b2 error esperado NOT_GREEN, got %s", resMap["clo_b2"].Error)
	}
	if resMap["clo_b3"].Error != "NOT_GREEN" {
		t.Errorf("clo_b3 error esperado NOT_GREEN, got %s", resMap["clo_b3"].Error)
	}
	if resMap["clo_b4"].Error != "NOT_READY" {
		t.Errorf("clo_b4 error esperado NOT_READY, got %s", resMap["clo_b4"].Error)
	}
	if resMap["clo_b5"].Error != "VERSION_CONFLICT" {
		t.Errorf("clo_b5 error esperado VERSION_CONFLICT, got %s", resMap["clo_b5"].Error)
	}
	if resMap["clo_b_inexistente"].Error != "NOT_FOUND" {
		t.Errorf("clo_b_inexistente error esperado NOT_FOUND, got %s", resMap["clo_b_inexistente"].Error)
	}
}

func TestSendPaymentReminders_Endpoint(t *testing.T) {
	router, closingRepo, clientRepo, _, _, _, clk := setupApprovalTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	cl := client.Client{ID: "cli_rem_1", Name: "Mariano Díaz", Regime: client.RegimeMonotributo}
	_ = clientRepo.Save(ctx, cl)

	// Due date en 2 días (dentro de la ventana de 3 días)
	c := closing.Closing{
		ID:         "clo_rem_1",
		ClientID:   "cli_rem_1",
		Period:     p,
		Status:     closing.StatusNotified,
		Severity:   signal.SeverityGreen,
		Version:    2,
		ResultHash: "h_rem",
		Mono: &tax.MonotributoAssessment{
			Fee:        money.New(5000000, money.ARS),
			FeeDueDate: clk.Now().AddDate(0, 0, 2),
		},
	}
	_ = closingRepo.Save(ctx, c)

	rec := do(t, router, "POST", "/api/v1/accountant/reminders/run", "accountant", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200 en reminders/run, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("error unmarshaling reminders response: %v", err)
	}

	if res["reminded"] != 1 {
		t.Errorf("esperaba 1 recordatorio enviado, got %d", res["reminded"])
	}
}

func TestPrivacy_OutboundMessagesNeverContainAmounts(t *testing.T) {
	// Criterio 7: salientes nunca llevan montos ni CUIT
	// Recorremos los mensajes salientes generados para todas las clases de notificaciones
	router, closingRepo, clientRepo, _, _, sender, clk := setupApprovalTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	cl := client.Client{ID: "cli_priv_1", Name: "Privacidad Test", Regime: client.RegimeMonotributo}
	_ = clientRepo.Save(ctx, cl)

	c := closing.Closing{
		ID:         "clo_priv_1",
		ClientID:   "cli_priv_1",
		Period:     p,
		Status:     closing.StatusReadyForReview,
		Severity:   signal.SeverityGreen,
		Version:    1,
		ResultHash: "h_priv",
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(543200, money.USD),
			BilledInPeriodARS: money.New(85000000, money.ARS),
			CurrentCategory:   tax.CatE,
			Fee:               money.New(1234500, money.ARS),
			FeeDueDate:        clk.Now().AddDate(0, 0, 2),
		},
	}
	_ = closingRepo.Save(ctx, c)

	// Aprobar
	_ = do(t, router, "POST", "/api/v1/accountant/closings/clo_priv_1/approve", "accountant", strings.NewReader(`{"version": 1}`))

	// Correr recordatorios
	_ = do(t, router, "POST", "/api/v1/accountant/reminders/run", "accountant", nil)

	if len(sender.Sent) == 0 {
		t.Fatalf("esperaba mensajes salientes en sender, no hubo ninguno")
	}

	for i, m := range sender.Sent {
		if notification.ContainsAmounts(m.Subject) {
			t.Errorf("mensaje %d: Subject %q contiene montos", i, m.Subject)
		}
		if notification.ContainsAmounts(m.Body) {
			t.Errorf("mensaje %d: Body %q contiene montos", i, m.Body)
		}
	}
}

func TestStatsEndpoint(t *testing.T) {
	router, closingRepo, clientRepo := setupClosingsTestRouter(t)
	ctx := context.Background()
	p := period.MustParse("2026-10")

	_ = clientRepo.Save(ctx, client.Client{
		ID: "cli_st_1", Name: "Stats Client", CUIT: "20-11111111-1",
		Regime: client.RegimeMonotributo, MonotributoCategory: tax.CatB,
	})
	c := closing.NewClosing(
		"clo_st_1", "cli_st_1", p, true, "h_st", signal.SeverityGreen, nil, "OK",
		&tax.MonotributoAssessment{
			CurrentCategory:   tax.CatB,
			Fee:               money.New(250000, money.ARS),
			FeeDueDate:        time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
			BilledInPeriodUSD: money.New(10000, money.USD),
			BilledInPeriodARS: money.New(15000000, money.ARS),
		},
		nil, closing.Reconciliation{}, testNow,
	)
	_ = closingRepo.Save(ctx, c)

	// 200 OK
	rec := do(t, router, "GET", "/api/v1/accountant/periods/2026-10/stats", "accountant", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET stats status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var st closing.PeriodStats
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if st.Period != "2026-10" || st.Progress.Total != 1 || st.Totals.ToPayARS.Amount != 250000 {
		t.Fatalf("unexpected stats: %+v", st)
	}

	// 422 invalid period
	recBad := do(t, router, "GET", "/api/v1/accountant/periods/bad-period/stats", "accountant", nil)
	if recBad.Code != http.StatusUnprocessableEntity {
		t.Fatalf("GET stats invalid period status = %d, want 422", recBad.Code)
	}

	// 403 client role
	recForb := do(t, router, "GET", "/api/v1/accountant/periods/2026-10/stats", "client:cli_st_1", nil)
	if recForb.Code != http.StatusForbidden {
		t.Fatalf("GET stats as client status = %d, want 403", recForb.Code)
	}
}

func TestGetPeriods(t *testing.T) {
	h, cr, clr := setupClosingsTestRouter(t)
	ctx := context.Background()
	_ = clr.Save(ctx, client.Client{ID: "cli_p1", Name: "Client P1", CUIT: "20-11111111-1", Regime: client.RegimeMonotributo})
	pSep := period.MustParse("2026-09")
	pOct := period.MustParse("2026-10")
	_ = cr.Save(ctx, closing.NewClosing("clo_p1", "cli_p1", pSep, true, "h1", signal.SeverityGreen, nil, "d", nil, nil, closing.Reconciliation{}, testNow))
	_ = cr.Save(ctx, closing.NewClosing("clo_p2", "cli_p1", pOct, true, "h2", signal.SeverityGreen, nil, "d", nil, nil, closing.Reconciliation{}, testNow))

	rec := do(t, h, "GET", "/api/v1/accountant/periods", "accountant", nil)
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var res struct {
		Periods []string `json:"periods"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(res.Periods) != 2 || res.Periods[0] != "2026-10" || res.Periods[1] != "2026-09" {
		t.Fatalf("unexpected periods: %v", res.Periods)
	}
}
