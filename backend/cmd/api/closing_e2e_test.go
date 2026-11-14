package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/config"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	arcamock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/arca/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	fxmock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
	notifylog "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/notify/log"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
	apihttp "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/interfaces/http"
)

func TestClosingE2E_SeedAndPerformance(t *testing.T) {
	ctx := context.Background()
	appNow, _ := time.Parse(time.RFC3339, config.DefaultAppNow)
	clk := clock.NewFixed(appNow)

	arcaClient, err := arcamock.New()
	if err != nil {
		t.Fatal(err)
	}
	fxClient, err := fxmock.New()
	if err != nil {
		t.Fatal(err)
	}

	paramsRepo := memory.NewParametersRepository()
	clientRepo := memory.NewClientRepository()
	movementRepo := memory.NewMovementRepository()
	classificationRepo := memory.NewClassificationRepository()
	invoiceRepo := memory.NewInvoiceRepository()
	purchaseInvoiceRepo := memory.NewPurchaseInvoiceRepository()
	clientInputRepo := memory.NewClientInputRepository()
	closingRepo := memory.NewClosingRepository()

	idGen := idgen.NewSequential()
	ds, err := seed.Generate(ctx, seed.SeedConfig{
		Seed:    42,
		Clients: 300,
		AppNow:  clk.Now(),
	}, fxClient, idGen)
	if err != nil {
		t.Fatalf("error generando seed: %v", err)
	}

	_ = clientRepo.SaveAll(ctx, ds.Clients)
	_ = movementRepo.SaveAll(ctx, ds.Movements)
	_ = invoiceRepo.SaveAll(ctx, ds.Invoices)
	_ = purchaseInvoiceRepo.SaveAll(ctx, ds.PurchaseInvoices)
	_ = clientInputRepo.SaveAll(ctx, ds.ClientInputs)

	classifications := classifyMovementsWorkerPool(ds.Movements, 8)
	_ = classificationRepo.SaveAll(ctx, classifications)

	computeClosing := &usecase.ComputeClosing{
		Clients:          clientRepo,
		Movements:        movementRepo,
		Classifications:  classificationRepo,
		Invoices:         invoiceRepo,
		PurchaseInvoices: purchaseInvoiceRepo,
		ClientInputs:     clientInputRepo,
		Parameters:       paramsRepo,
		Closings:         closingRepo,
		TaxScales:        arcaClient,
		DueDates:         arcaClient,
		ExchangeRates:    fxClient,
		Clock:            clk,
		IDGen:            idGen,
	}

	targetPeriod := period.MustParse("2026-10")
	runner := &usecase.RunMonthlyClosing{
		Clients: clientRepo,
		Compute: computeClosing,
		Workers: 8,
	}

	// Performance: 300 clientes en < 2 s (criterio 10)
	start := time.Now()
	summary, err := runner.Execute(ctx, targetPeriod)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("error en RunMonthlyClosing: %v", err)
	}
	if summary.Processed != 300 {
		t.Fatalf("esperaba 300 procesados, obtuvo %d", summary.Processed)
	}
	if len(summary.Failed) > 0 {
		t.Fatalf("hubo %d fallos inesperados: %+v", len(summary.Failed), summary.Failed)
	}

	t.Logf("300 clientes procesados en %v (duración < 2s requerida)", duration)
	if !testing.Short() && duration > 2*time.Second {
		t.Errorf("performance: tardó %v, esperado < 2s", duration)
	}

	// Sembrar aprobaciones de octubre para clientes demo cli_demo_01 y cli_demo_09 (spec 017 criterio 3)
	octSeedClk := clock.NewFixed(time.Date(2026, 11, 6, 10, 0, 0, 0, period.Location()))
	octSeedApproveUC := &usecase.ApproveClosing{
		Closings:          closingRepo,
		Clients:           clientRepo,
		Notifications:     memory.NewNotificationRepository(),
		Outbox:            memory.NewOutboxRepository(),
		Sender:            notifylog.NewSender(slog.Default()),
		Clock:             octSeedClk,
		IDGen:             idGen,
		AccountantName:    "C.P. Mariana Quiroga",
		AccountantLicense: "Matrícula de ejemplo",
	}
	if err := usecase.SeedOctoberApprovedClosings(ctx, octSeedApproveUC, closingRepo); err != nil {
		t.Fatalf("error sembrando aprobaciones de octubre: %v", err)
	}

	// Criterio 17: Para cada cli_demo_XX la señal principal y el estado coinciden con la tabla de 002/plan.md y spec 017
	expectedDemo := map[string]struct {
		Severity      signal.Severity
		Status        closing.Status
		PrimarySignal *signal.Code
	}{
		"cli_demo_01": {Severity: signal.SeverityGreen, Status: closing.StatusNotified, PrimarySignal: nil},
		"cli_demo_02": {Severity: signal.SeverityRed, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoExclusionKCapExceeded)},
		"cli_demo_03": {Severity: signal.SeverityRed, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMissingInvoices)},
		"cli_demo_04": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoCategoryChangeRequired)},
		"cli_demo_05": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoNearCategoryCap)},
		"cli_demo_06": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeAtypicalIncome)},
		"cli_demo_07": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeUnclassifiedMovements)},
		"cli_demo_08": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoNearKCap)},
		"cli_demo_09": {Severity: signal.SeverityGreen, Status: closing.StatusNotified, PrimarySignal: nil},
		"cli_demo_10": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeInvoiceWithoutPayment)},
		"cli_demo_11": {Severity: signal.SeverityRed, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodePostExclusionCInvoices)},
		"cli_demo_12": {Severity: signal.SeverityRed, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeDepositsExceedKCap)},
		"cli_demo_13": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeExternalIncomeDeclared)},
		"cli_demo_14": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoCategoryDowngradePossible)},
		"cli_demo_15": {Severity: signal.SeverityRed, Status: closing.StatusPendingClientInput, PrimarySignal: codePtr(signal.CodeMissingInvoices)},
	}

	for id, exp := range expectedDemo {
		c, found, err := closingRepo.ByClientPeriod(ctx, id, targetPeriod)
		if err != nil || !found {
			t.Fatalf("cierre para %s no encontrado: %v", id, err)
		}
		if c.Severity != exp.Severity {
			t.Errorf("%s severidad = %v, esperada %v", id, c.Severity, exp.Severity)
		}
		if c.Status != exp.Status {
			t.Errorf("%s status = %v, esperado %v", id, c.Status, exp.Status)
		}
		if exp.PrimarySignal == nil {
			if len(c.Signals) > 0 {
				t.Errorf("%s esperaba sin señales, obtuvo %+v", id, c.Signals)
			}
		} else {
			if len(c.Signals) == 0 {
				t.Errorf("%s esperaba señal principal %s, pero no tuvo señales", id, *exp.PrimarySignal)
			} else if c.Signals[0].Code != *exp.PrimarySignal {
				t.Errorf("%s señal principal = %s, esperada %s", id, c.Signals[0].Code, *exp.PrimarySignal)
			}
		}
	}

	// Distribución global de semáforos: aprox 70–80% verde, 15–22% amarillo, 5–10% rojo
	allClosings, _ := closingRepo.ListByPeriod(ctx, targetPeriod)
	greenCount := 0
	yellowCount := 0
	redCount := 0

	for _, c := range allClosings {
		switch c.Severity {
		case signal.SeverityGreen:
			greenCount++
		case signal.SeverityYellow:
			yellowCount++
		case signal.SeverityRed:
			redCount++
		}
	}

	total := len(allClosings)
	greenPct := (greenCount * 100) / total
	yellowPct := (yellowCount * 100) / total
	redPct := (redCount * 100) / total

	sigCounts := make(map[signal.Code]int)
	for _, c := range allClosings {
		for _, s := range c.Signals {
			sigCounts[s.Code]++
		}
	}
	t.Logf("Señales encontradas: %+v", sigCounts)

	t.Logf("Distribución global semáforos (total %d): Verdes: %d (%d%%), Amarillos: %d (%d%%), Rojos: %d (%d%%)",
		total, greenCount, greenPct, yellowCount, yellowPct, redCount, redPct)

	if greenPct < 65 || greenPct > 85 {
		t.Errorf("distribución verde fuera de rango: %d%%", greenPct)
	}
	if yellowPct < 10 || yellowPct > 28 {
		t.Errorf("distribución amarillo fuera de rango: %d%%", yellowPct)
	}
	if redPct < 3 || redPct > 15 {
		t.Errorf("distribución rojo fuera de rango: %d%%", redPct)
	}

	stats := closing.ComputeStats(allClosings, period.MustParse("2026-10"))
	if len(stats.Totals.ToPayByDueDate) == 0 {
		t.Fatal("esperaba vencimientos en el resumen de octubre")
	}
	hasNov20 := false
	for _, dd := range stats.Totals.ToPayByDueDate {
		if strings.HasPrefix(dd.Date, "2026-10-") {
			t.Errorf("vencimiento de octubre encontrado en el cierre de octubre: %s", dd.Date)
		}
		if dd.Date == "2026-11-20" {
			hasNov20 = true
		}
	}
	if !hasNov20 {
		t.Errorf("esperaba vencimiento 2026-11-20 en ToPayByDueDate, obtuvo %+v", stats.Totals.ToPayByDueDate)
	}
}

func codePtr(c signal.Code) *signal.Code {
	return &c
}

func setupTestApp(t *testing.T) (http.Handler, *memory.ClosingRepository, *memory.ClientRepository) {
	t.Helper()
	ctx := context.Background()
	appNow, _ := time.Parse(time.RFC3339, config.DefaultAppNow)
	clk := clock.NewFixed(appNow)

	arcaClient, err := arcamock.New()
	if err != nil {
		t.Fatal(err)
	}
	fxClient, err := fxmock.New()
	if err != nil {
		t.Fatal(err)
	}

	paramsRepo := memory.NewParametersRepository()
	clientRepo := memory.NewClientRepository()
	movementRepo := memory.NewMovementRepository()
	classificationRepo := memory.NewClassificationRepository()
	invoiceRepo := memory.NewInvoiceRepository()
	purchaseInvoiceRepo := memory.NewPurchaseInvoiceRepository()
	clientInputRepo := memory.NewClientInputRepository()
	closingRepo := memory.NewClosingRepository()
	closingQuery := memory.NewClosingQuery(closingRepo, clientRepo)
	statsQuery := memory.NewStatsQuery(closingRepo)
	notifRepo := memory.NewNotificationRepository()
	outboxRepo := memory.NewOutboxRepository()
	notifSender := notifylog.NewSender(slog.Default())
	txManager := memory.NewTxManager()

	idGen := idgen.NewSequential()
	ds, err := seed.Generate(ctx, seed.SeedConfig{
		Seed:    42,
		Clients: 300,
		AppNow:  clk.Now(),
	}, fxClient, idGen)
	if err != nil {
		t.Fatalf("error generando seed: %v", err)
	}

	_ = clientRepo.SaveAll(ctx, ds.Clients)
	_ = movementRepo.SaveAll(ctx, ds.Movements)
	_ = invoiceRepo.SaveAll(ctx, ds.Invoices)
	_ = purchaseInvoiceRepo.SaveAll(ctx, ds.PurchaseInvoices)
	_ = clientInputRepo.SaveAll(ctx, ds.ClientInputs)

	classifications := classifyMovementsWorkerPool(ds.Movements, 8)
	_ = classificationRepo.SaveAll(ctx, classifications)

	computeClosing := &usecase.ComputeClosing{
		Clients:          clientRepo,
		Movements:        movementRepo,
		Classifications:  classificationRepo,
		Invoices:         invoiceRepo,
		PurchaseInvoices: purchaseInvoiceRepo,
		ClientInputs:     clientInputRepo,
		Parameters:       paramsRepo,
		Closings:         closingRepo,
		TaxScales:        arcaClient,
		DueDates:         arcaClient,
		ExchangeRates:    fxClient,
		Clock:            clk,
		IDGen:            idGen,
	}

	runClosingUC := &usecase.RunMonthlyClosing{
		Clients: clientRepo,
		Compute: computeClosing,
		Workers: 8,
	}

	pSep := period.MustParse("2026-09")
	pOct := period.MustParse("2026-10")

	if _, err := runClosingUC.Execute(ctx, pSep); err != nil {
		t.Fatalf("error ejecutando cierre de septiembre: %v", err)
	}
	if err := usecase.SeedSeptemberClosings(ctx, closingRepo, clientRepo, idGen, 42); err != nil {
		t.Fatalf("error sembrando septiembre: %v", err)
	}
	if _, err := runClosingUC.Execute(ctx, pOct); err != nil {
		t.Fatalf("error ejecutando cierre de octubre: %v", err)
	}

	octSeedClk := clock.NewFixed(time.Date(2026, 11, 6, 10, 0, 0, 0, period.Location()))
	octSeedApproveUC := &usecase.ApproveClosing{
		Closings:          closingRepo,
		Clients:           clientRepo,
		Notifications:     notifRepo,
		Outbox:            outboxRepo,
		Sender:            notifSender,
		Clock:             octSeedClk,
		IDGen:             idGen,
		AccountantName:    "C.P. Mariana Quiroga",
		AccountantLicense: "Matrícula de ejemplo",
	}
	if err := usecase.SeedOctoberApprovedClosings(ctx, octSeedApproveUC, closingRepo); err != nil {
		t.Fatalf("error sembrando aprobaciones de octubre: %v", err)
	}

	notifier := &usecase.RecalcOnChange{
		Compute:  computeClosing,
		Closings: closingRepo,
	}

	handler := apihttp.NewRouter(apihttp.Deps{
		Clock:                     clk,
		Logger:                    slog.Default(),
		CORSOrigin:                "*",
		ParametersRepo:            paramsRepo,
		ClientRepo:                clientRepo,
		MovementRepo:              movementRepo,
		ClassificationRepo:        classificationRepo,
		InvoiceRepo:               invoiceRepo,
		PurchaseInvoiceRepo:       purchaseInvoiceRepo,
		ClientInputRepo:           clientInputRepo,
		LatestExclusionDateReader: closingRepo,
		ClosingRepo:               closingRepo,
		ClosingQuery:              closingQuery,
		StatsQuery:                statsQuery,
		TxManager:                 txManager,
		TaxScales:                 arcaClient,
		DueDates:                  arcaClient,
		ExchangeRates:             fxClient,
		Workers:                   8,
		IDGen:                     idGen,
		ChangeNotifier:            notifier,
		NotificationRepo:          notifRepo,
		OutboxRepo:                outboxRepo,
		NotificationSender:        notifSender,
		RealNow:                   clk.Now,
	})

	return handler, closingRepo, clientRepo
}

func TestSeptemberClosedAndOctoberCurrent(t *testing.T) {
	ctx := context.Background()
	_, closingRepo, _ := setupTestApp(t)

	periods, err := closingRepo.Periods(ctx)
	if err != nil {
		t.Fatalf("Periods error: %v", err)
	}
	if len(periods) < 2 {
		t.Fatalf("esperaba al menos 2 periodos, obtuvo %v", periods)
	}
	if periods[0] != "2026-10" || periods[1] != "2026-09" {
		t.Errorf("periodos = %v, esperado [2026-10, 2026-09]", periods)
	}

	// 2026-09 tiene 300 cierres en NOTIFIED
	sepClosings, err := closingRepo.ListByPeriod(ctx, period.MustParse("2026-09"))
	if err != nil {
		t.Fatalf("ListByPeriod 2026-09 error: %v", err)
	}
	if len(sepClosings) != 300 {
		t.Fatalf("2026-09 cierres = %d, esperado 300", len(sepClosings))
	}
	for _, c := range sepClosings {
		if c.Status != closing.StatusNotified {
			t.Errorf("cierre %s de septiembre tiene estado %s, esperado NOTIFIED", c.ID, c.Status)
			break
		}
	}

	// Métricas de éxito de septiembre > 0
	stats := closing.ComputeStats(sepClosings, period.MustParse("2026-09"))
	if stats.Success.NotifiedBeforeDay10Pct <= 0 {
		t.Errorf("NotifiedBeforeDay10Pct = %v, esperado > 0", stats.Success.NotifiedBeforeDay10Pct)
	}
	if stats.Success.PaidOnTimePct <= 0 {
		t.Errorf("PaidOnTimePct = %v, esperado > 0", stats.Success.PaidOnTimePct)
	}

	// Octubre demo clients siguen en estado esperado
	c2, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_02", period.MustParse("2026-10"))
	if err != nil || !found {
		t.Fatalf("cli_demo_02 no encontrado en octubre: %v", err)
	}
	if c2.Severity != signal.SeverityRed {
		t.Errorf("cli_demo_02 severidad = %v, esperada RED", c2.Severity)
	}
}

func TestAllClosingsDetail200(t *testing.T) {
	ctx := context.Background()
	handler, closingRepo, _ := setupTestApp(t)

	for _, pStr := range []string{"2026-09", "2026-10"} {
		p := period.MustParse(pStr)
		closings, err := closingRepo.ListByPeriod(ctx, p)
		if err != nil {
			t.Fatalf("ListByPeriod %s error: %v", pStr, err)
		}
		if len(closings) != 300 {
			t.Fatalf("%s cierres = %d, esperado 300", pStr, len(closings))
		}

		for _, c := range closings {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/accountant/closings/%s", c.ID), nil)
			req.Header.Set("X-Demo-User", "accountant")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET /closings/%s en periodo %s devolvió %d: %s", c.ID, pStr, rec.Code, rec.Body.String())
			}

			var resp struct {
				Closing struct {
					ID string `json:"id"`
				} `json:"closing"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("cuerpo inválido para %s: %v", c.ID, err)
			}
			if resp.Closing.ID != c.ID {
				t.Fatalf("closing.id = %s, esperado %s", resp.Closing.ID, c.ID)
			}
		}
	}
}

func TestRegimeChangeExcludedDemoClient(t *testing.T) {
	ctx := context.Background()
	handler, closingRepo, clientRepo := setupTestApp(t)

	c2, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_02", period.MustParse("2026-10"))
	if err != nil || !found {
		t.Fatalf("cli_demo_02 no encontrado en octubre: %v", err)
	}
	if c2.Mono == nil || c2.Mono.ExclusionDate == nil {
		t.Fatalf("cli_demo_02 no tiene fecha de exclusión en octubre")
	}
	effDateStr := c2.Mono.ExclusionDate.Format("2006-01-02")

	body := []byte(fmt.Sprintf(`{"to": "RI", "effectiveDate": "%s"}`, effDateStr))
	req := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_demo_02/regime-change", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Demo-User", "accountant")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST regime-change devolvió %d: %s", rec.Code, rec.Body.String())
	}

	// Cliente actualizado a RI
	cl, err := clientRepo.Get(ctx, "cli_demo_02")
	if err != nil {
		t.Fatalf("cliente cli_demo_02 no encontrado: %v", err)
	}
	if cl.Regime != client.RegimeRI {
		t.Errorf("cliente regime = %v, esperado RI", cl.Regime)
	}

	// Cierre de octubre recalculado como RI
	c, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_02", period.MustParse("2026-10"))
	if err != nil || !found {
		t.Fatalf("cierre de octubre para cli_demo_02 no encontrado: %v", err)
	}
	if c.RI == nil {
		t.Errorf("cierre de octubre para cli_demo_02 debería tener RI, pero es nil")
	}
}

func TestPaulaRojasCompleteFlowE2E(t *testing.T) {
	ctx := context.Background()
	handler, closingRepo, _ := setupTestApp(t)

	pOct := period.MustParse("2026-10")

	// 1. Estado inicial de Paula Rojas (cli_demo_15)
	cInit, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_15", pOct)
	if err != nil || !found {
		t.Fatalf("cierre de Paula Rojas no encontrado: %v", err)
	}
	if cInit.Status != closing.StatusPendingClientInput {
		t.Fatalf("estado inicial = %v, esperado PENDING_CLIENT_INPUT", cInit.Status)
	}

	// 2. Importar facturas CSV en su nombre como contador (usando el CSV de ejemplo mis_comprobantes_ok.csv)
	csvContent := "Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total\n12/10/2026,19 - Factura E,00001,00000001,00000001,74392819289901,80,DE329182391,Brightloop GmbH,1482.15,USD,850.00,0.00,0.00,0.00,850.00\n"
	bodyBuf := &bytes.Buffer{}
	mpWriter := multipart.NewWriter(bodyBuf)
	part, _ := mpWriter.CreateFormFile("file", "mis_comprobantes_paula.csv")
	_, _ = part.Write([]byte(csvContent))
	_ = mpWriter.Close()

	reqImport := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_demo_15/invoices/import", bodyBuf)
	reqImport.Header.Set("Content-Type", mpWriter.FormDataContentType())
	reqImport.Header.Set("X-Demo-User", "accountant")
	recImport := httptest.NewRecorder()
	handler.ServeHTTP(recImport, reqImport)

	if recImport.Code != http.StatusOK {
		t.Fatalf("POST import invoices devolvió %d: %s", recImport.Code, recImport.Body.String())
	}

	// Cierre recalculado tras la importación: pasa a verde
	cAfterImport, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_15", pOct)
	if err != nil || !found {
		t.Fatalf("cierre tras import no encontrado: %v", err)
	}
	if cAfterImport.Severity != signal.SeverityGreen {
		t.Errorf("severidad tras importar = %v, esperado GREEN", cAfterImport.Severity)
	}
	if cAfterImport.Status != closing.StatusPendingClientInput {
		t.Errorf("estado tras importar = %v, esperado PENDING_CLIENT_INPUT (aún falta confirmar)", cAfterImport.Status)
	}

	// 3. Confirmar en nombre del cliente ("No tuvo cobros fuera de Wallbit", nota obligatoria)
	confirmBody := []byte(`{"externalIncome": {"none": true}, "note": "Confirmó por WhatsApp el 05/11 que no tuvo otros ingresos"}`)
	reqConfirm := httptest.NewRequest("POST", "/api/v1/accountant/clients/cli_demo_15/periods/2026-10/confirm-input", bytes.NewReader(confirmBody))
	reqConfirm.Header.Set("Content-Type", "application/json")
	reqConfirm.Header.Set("X-Demo-User", "accountant")
	recConfirm := httptest.NewRecorder()
	handler.ServeHTTP(recConfirm, reqConfirm)

	if recConfirm.Code != http.StatusOK {
		t.Fatalf("POST confirm-input devolvió %d: %s", recConfirm.Code, recConfirm.Body.String())
	}

	// Cierre pasa a READY_FOR_REVIEW
	cAfterConfirm, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_15", pOct)
	if err != nil || !found {
		t.Fatalf("cierre tras confirmación no encontrado: %v", err)
	}
	if cAfterConfirm.Status != closing.StatusReadyForReview {
		t.Errorf("estado tras confirmar = %v, esperado READY_FOR_REVIEW", cAfterConfirm.Status)
	}

	// 4. Detalle del cierre expone datos auditados
	reqDet := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/accountant/closings/%s", cAfterConfirm.ID), nil)
	reqDet.Header.Set("X-Demo-User", "accountant")
	recDet := httptest.NewRecorder()
	handler.ServeHTTP(recDet, reqDet)

	if recDet.Code != http.StatusOK {
		t.Fatalf("GET closing detail devolvió %d: %s", recDet.Code, recDet.Body.String())
	}

	var detResp struct {
		PeriodInvoices []struct {
			Invoice struct {
				UploadedBy string `json:"uploadedBy"`
			} `json:"invoice"`
		} `json:"periodInvoices"`
		ClientInput struct {
			ConfirmedBy      string `json:"confirmedBy"`
			ConfirmationNote string `json:"confirmationNote"`
		} `json:"clientInput"`
	}
	if err := json.Unmarshal(recDet.Body.Bytes(), &detResp); err != nil {
		t.Fatalf("unmarshal detail response: %v", err)
	}
	if detResp.ClientInput.ConfirmedBy != "ACCOUNTANT" || detResp.ClientInput.ConfirmationNote != "Confirmó por WhatsApp el 05/11 que no tuvo otros ingresos" {
		t.Errorf("clientInput audit mismatch: %+v", detResp.ClientInput)
	}
	foundAccountantInv := false
	for _, pi := range detResp.PeriodInvoices {
		if pi.Invoice.UploadedBy == "ACCOUNTANT" {
			foundAccountantInv = true
			break
		}
	}
	if !foundAccountantInv {
		t.Errorf("no se encontró factura con uploadedBy = ACCOUNTANT en el detalle")
	}

	// 5. Aprobar cierre
	approveBody := []byte(fmt.Sprintf(`{"version": %d, "clientNote": "Aprobado por el contador"}`, cAfterConfirm.Version))
	reqApprove := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/accountant/closings/%s/approve", cAfterConfirm.ID), bytes.NewReader(approveBody))
	reqApprove.Header.Set("Content-Type", "application/json")
	reqApprove.Header.Set("X-Demo-User", "accountant")
	recApprove := httptest.NewRecorder()
	handler.ServeHTTP(recApprove, reqApprove)

	if recApprove.Code != http.StatusOK {
		t.Fatalf("POST approve devolvió %d: %s", recApprove.Code, recApprove.Body.String())
	}

	// Cierre aprobado y notificado
	cApproved, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_15", pOct)
	if err != nil || !found {
		t.Fatalf("cierre aprobado no encontrado: %v", err)
	}
	if cApproved.Status != closing.StatusNotified {
		t.Errorf("estado final = %v, esperado NOTIFIED", cApproved.Status)
	}
	if cApproved.Published == nil {
		t.Errorf("cierre notificado debe contener snapshot publicado (published != nil)")
	}
}

// TestPaulaCircuitE2E_ClientDemo cubre el criterio 13 de la spec 017:
// 1) Cliente (Paula): carga comprobante, declara no ingresos externos, confirma -> "En revisión"
// 2) Contador: busca a Paula (verde, Para revisar) -> Aprueba
// 3) Cliente (Paula): ve tarjeta validada con el badge del contador y registra el pago
func TestPaulaCircuitE2E_ClientDemo(t *testing.T) {
	ctx := context.Background()
	handler, closingRepo, _ := setupTestApp(t)

	pOct := period.MustParse("2026-10")

	// Paso 1: Cliente (Paula) verifica que su cierre inicial está en ACTION_REQUIRED
	reqClosingsInit := httptest.NewRequest("GET", "/api/v1/me/closings", nil)
	reqClosingsInit.Header.Set("X-Demo-User", "client:cli_demo_15")
	recClosingsInit := httptest.NewRecorder()
	handler.ServeHTTP(recClosingsInit, reqClosingsInit)

	if recClosingsInit.Code != http.StatusOK {
		t.Fatalf("GET /me/closings inicial devolvió %d: %s", recClosingsInit.Code, recClosingsInit.Body.String())
	}
	var closingsInit []struct {
		ClosingID string `json:"closingId"`
		Period    string `json:"period"`
		State     string `json:"state"`
	}
	_ = json.Unmarshal(recClosingsInit.Body.Bytes(), &closingsInit)
	var paulaOctInitState string
	var paulaClosingID string
	for _, c := range closingsInit {
		if c.Period == "2026-10" {
			paulaOctInitState = c.State
			paulaClosingID = c.ClosingID
			break
		}
	}
	if paulaOctInitState != "ACTION_REQUIRED" {
		t.Fatalf("estado inicial de Paula en /me/closings = %s, esperado ACTION_REQUIRED", paulaOctInitState)
	}

	// 1a. Paula carga un comprobante con "Alta de comprobante" (POST /me/invoices)
	invBody, _ := json.Marshal(map[string]any{
		"type":         "E",
		"pointOfSale":  1,
		"number":       1,
		"issueDate":    "2026-10-12",
		"currency":     "USD",
		"netAmount":    map[string]any{"amount": "850.00", "currency": "USD"},
		"vatAmount":    map[string]any{"amount": "0.00", "currency": "USD"},
		"totalAmount":  map[string]any{"amount": "850.00", "currency": "USD"},
		"fxRate":       "1482.15",
		"receiverName": "Brightloop GmbH",
		"receiverDoc":  "DE329182391",
	})
	reqInv := httptest.NewRequest("POST", "/api/v1/me/invoices", bytes.NewReader(invBody))
	reqInv.Header.Set("Content-Type", "application/json")
	reqInv.Header.Set("X-Demo-User", "client:cli_demo_15")
	recInv := httptest.NewRecorder()
	handler.ServeHTTP(recInv, reqInv)

	if recInv.Code != http.StatusCreated {
		t.Fatalf("POST /me/invoices devolvió %d: %s", recInv.Code, recInv.Body.String())
	}

	// 1b. Paula responde "No" a cobros fuera de Wallbit (PUT /me/periods/2026-10/external-income)
	extBody, _ := json.Marshal(map[string]any{
		"items": []any{},
		"none":  true,
	})
	reqExt := httptest.NewRequest("PUT", "/api/v1/me/periods/2026-10/external-income", bytes.NewReader(extBody))
	reqExt.Header.Set("Content-Type", "application/json")
	reqExt.Header.Set("X-Demo-User", "client:cli_demo_15")
	recExt := httptest.NewRecorder()
	handler.ServeHTTP(recExt, reqExt)

	if recExt.Code != http.StatusOK {
		t.Fatalf("PUT /me/periods/2026-10/external-income devolvió %d: %s", recExt.Code, recExt.Body.String())
	}

	// 1c. Paula confirma la carga con "Listo, ya cargué todo" (POST /me/periods/2026-10/confirm)
	reqConf := httptest.NewRequest("POST", "/api/v1/me/periods/2026-10/confirm", nil)
	reqConf.Header.Set("X-Demo-User", "client:cli_demo_15")
	recConf := httptest.NewRecorder()
	handler.ServeHTTP(recConf, reqConf)

	if recConf.Code != http.StatusOK {
		t.Fatalf("POST /me/periods/2026-10/confirm devolvió %d: %s", recConf.Code, recConf.Body.String())
	}

	// Paula verifica que ahora su estado es "IN_REVIEW"
	reqClosingsReview := httptest.NewRequest("GET", "/api/v1/me/closings", nil)
	reqClosingsReview.Header.Set("X-Demo-User", "client:cli_demo_15")
	recClosingsReview := httptest.NewRecorder()
	handler.ServeHTTP(recClosingsReview, reqClosingsReview)

	var closingsReview []struct {
		Period string `json:"period"`
		State  string `json:"state"`
	}
	_ = json.Unmarshal(recClosingsReview.Body.Bytes(), &closingsReview)
	for _, c := range closingsReview {
		if c.Period == "2026-10" && c.State != "IN_REVIEW" {
			t.Fatalf("estado tras confirmar = %s, esperado IN_REVIEW", c.State)
		}
	}

	// Paso 2: Contador busca a Paula en el tablero: verde y Para revisar
	cAfterClientInput, found, err := closingRepo.ByClientPeriod(ctx, "cli_demo_15", pOct)
	if err != nil || !found {
		t.Fatalf("cierre no encontrado en repo: %v", err)
	}
	if cAfterClientInput.Severity != signal.SeverityGreen {
		t.Errorf("severidad esperada GREEN, obtenida %v", cAfterClientInput.Severity)
	}
	if cAfterClientInput.Status != closing.StatusReadyForReview {
		t.Errorf("estado esperado READY_FOR_REVIEW, obtenido %v", cAfterClientInput.Status)
	}

	// Contador aprueba el cierre de Paula
	approveBody := []byte(fmt.Sprintf(`{"version": %d, "clientNote": "Revisado y aprobado ok"}`, cAfterClientInput.Version))
	reqApprove := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/accountant/closings/%s/approve", paulaClosingID), bytes.NewReader(approveBody))
	reqApprove.Header.Set("Content-Type", "application/json")
	reqApprove.Header.Set("X-Demo-User", "accountant")
	recApprove := httptest.NewRecorder()
	handler.ServeHTTP(recApprove, reqApprove)

	if recApprove.Code != http.StatusOK {
		t.Fatalf("POST /accountant/closings/{id}/approve devolvió %d: %s", recApprove.Code, recApprove.Body.String())
	}

	// Paso 3: Cliente (Paula) consulta su cierre: estado READY con tarjeta validada
	reqClosingsReady := httptest.NewRequest("GET", "/api/v1/me/closings", nil)
	reqClosingsReady.Header.Set("X-Demo-User", "client:cli_demo_15")
	recClosingsReady := httptest.NewRecorder()
	handler.ServeHTTP(recClosingsReady, reqClosingsReady)

	if recClosingsReady.Code != http.StatusOK {
		t.Fatalf("GET /me/closings tras aprobación devolvió %d: %s", recClosingsReady.Code, recClosingsReady.Body.String())
	}

	var closingsReady []struct {
		ClosingID string     `json:"closingId"`
		Period    string     `json:"period"`
		State     string     `json:"state"`
		PaidAt    *time.Time `json:"paidAt"`
		Published *struct {
			ValidatedBy struct {
				Name    string `json:"name"`
				License string `json:"license"`
			} `json:"validatedBy"`
			Regime             string `json:"regime"`
			Category           string `json:"category"`
			PctOfCategoryCapBp *int64 `json:"pctOfCategoryCapBp"`
		} `json:"published"`
	}
	if err := json.Unmarshal(recClosingsReady.Body.Bytes(), &closingsReady); err != nil {
		t.Fatalf("unmarshal closingsReady: %v", err)
	}

	var paulaReadyClosing struct {
		ClosingID string     `json:"closingId"`
		Period    string     `json:"period"`
		State     string     `json:"state"`
		PaidAt    *time.Time `json:"paidAt"`
		Published *struct {
			ValidatedBy struct {
				Name    string `json:"name"`
				License string `json:"license"`
			} `json:"validatedBy"`
			Regime             string `json:"regime"`
			Category           string `json:"category"`
			PctOfCategoryCapBp *int64 `json:"pctOfCategoryCapBp"`
		} `json:"published"`
	}
	for _, c := range closingsReady {
		if c.Period == "2026-10" {
			paulaReadyClosing = c
			break
		}
	}

	if paulaReadyClosing.State != "READY" {
		t.Fatalf("estado esperado READY, obtenido %s", paulaReadyClosing.State)
	}
	if paulaReadyClosing.Published == nil {
		t.Fatal("published es nil, esperado snapshot publicado")
	}
	if paulaReadyClosing.Published.ValidatedBy.Name != "C.P. Mariana Quiroga" || paulaReadyClosing.Published.ValidatedBy.License != "Matrícula de ejemplo" {
		t.Errorf("validatedBy mismatch: %+v", paulaReadyClosing.Published.ValidatedBy)
	}
	if paulaReadyClosing.Published.Regime != "MONOTRIBUTO" {
		t.Errorf("regime mismatch: %s", paulaReadyClosing.Published.Regime)
	}
	if paulaReadyClosing.Published.PctOfCategoryCapBp == nil {
		t.Errorf("pctOfCategoryCapBp debería estar poblado")
	}

	// 3b. Paula marca su cierre como pagado (POST /me/closings/{id}/paid)
	reqPaid := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/me/closings/%s/paid", paulaClosingID), nil)
	reqPaid.Header.Set("X-Demo-User", "client:cli_demo_15")
	recPaid := httptest.NewRecorder()
	handler.ServeHTTP(recPaid, reqPaid)

	if recPaid.Code != http.StatusOK {
		t.Fatalf("POST /me/closings/{id}/paid devolvió %d: %s", recPaid.Code, recPaid.Body.String())
	}

	var paidResp struct {
		ClosingID string     `json:"closingId"`
		PaidAt    *time.Time `json:"paidAt"`
	}
	if err := json.Unmarshal(recPaid.Body.Bytes(), &paidResp); err != nil {
		t.Fatalf("unmarshal paid response: %v", err)
	}
	if paidResp.PaidAt == nil {
		t.Errorf("paidAt es nil tras marcar como pagado")
	}
}
