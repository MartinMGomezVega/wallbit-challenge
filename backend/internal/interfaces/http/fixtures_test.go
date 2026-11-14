package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	arcamock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/arca/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	fxmock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
)

func TestWriteFrontendFixtures(t *testing.T) {
	if os.Getenv("WRITE_FIXTURES") != "1" {
		t.Skip("set WRITE_FIXTURES=1 to generate frontend API fixtures")
	}

	ctx := context.Background()
	appNow := time.Date(2026, 11, 5, 9, 0, 0, 0, period.Location())
	clk := clock.NewFixed(appNow)

	fxClient, err := fxmock.New()
	if err != nil {
		t.Fatalf("fxmock.New: %v", err)
	}
	arcaClient, err := arcamock.New()
	if err != nil {
		t.Fatalf("arcamock.New: %v", err)
	}
	idGen := idgen.NewSequential()

	ds, err := seed.Generate(ctx, seed.SeedConfig{
		Seed:    42,
		Clients: 300,
		AppNow:  appNow,
	}, fxClient, idGen)
	if err != nil {
		t.Fatalf("seed.Generate: %v", err)
	}

	classifications := make([]movement.Classification, 0, len(ds.Movements))
	for _, m := range ds.Movements {
		classifications = append(classifications, movement.Classification{
			MovementID: m.ID,
			Kind:       movement.KindIncome,
			Confidence: movement.ConfidenceHigh,
			Rule:       movement.RuleForeignCompany,
		})
	}

	clientRepo := memory.NewClientRepository()
	movRepo := memory.NewMovementRepository()
	classRepo := memory.NewClassificationRepository()
	invRepo := memory.NewInvoiceRepository()
	purchRepo := memory.NewPurchaseInvoiceRepository()
	inputRepo := memory.NewClientInputRepository()
	paramRepo := memory.NewParametersRepository()
	closingRepo := memory.NewClosingRepository()
	notifRepo := memory.NewNotificationRepository()
	outboxRepo := memory.NewOutboxRepository()
	txMgr := memory.NewTxManager()

	_ = clientRepo.SaveAll(ctx, ds.Clients)
	_ = movRepo.SaveAll(ctx, ds.Movements)
	_ = invRepo.SaveAll(ctx, ds.Invoices)
	_ = purchRepo.SaveAll(ctx, ds.PurchaseInvoices)
	_ = inputRepo.SaveAll(ctx, ds.ClientInputs)
	_ = classRepo.SaveAll(ctx, classifications)

	p := period.MustParse("2026-10")
	pSep := period.MustParse("2026-09")

	computeUC := &usecase.ComputeClosing{
		Clients: clientRepo, Movements: movRepo, Classifications: classRepo,
		Invoices: invRepo, PurchaseInvoices: purchRepo, ClientInputs: inputRepo,
		Parameters: paramRepo, Closings: closingRepo,
		TaxScales: arcaClient, DueDates: arcaClient, ExchangeRates: fxClient,
		Clock: clk, IDGen: idGen, Tx: txMgr,
	}
	runUC := &usecase.RunMonthlyClosing{Clients: clientRepo, Compute: computeUC, Workers: 1}
	if _, err := runUC.Execute(ctx, p); err != nil {
		t.Fatalf("RunMonthlyClosing 2026-10: %v", err)
	}
	if _, err := runUC.Execute(ctx, pSep); err != nil {
		t.Fatalf("RunMonthlyClosing 2026-09: %v", err)
	}
	if err := usecase.SeedSeptemberClosings(ctx, closingRepo, clientRepo, idGen, 42); err != nil {
		t.Fatalf("SeedSeptemberClosings: %v", err)
	}

	h := NewRouter(Deps{
		Clock:                     clk,
		ParametersRepo:            paramRepo,
		ClientRepo:                clientRepo,
		MovementRepo:              movRepo,
		ClassificationRepo:        classRepo,
		InvoiceRepo:               invRepo,
		PurchaseInvoiceRepo:       purchRepo,
		ClientInputRepo:           inputRepo,
		LatestExclusionDateReader: closingRepo,
		ClosingRepo:               closingRepo,
		ClosingQuery:              memory.NewClosingQuery(closingRepo, clientRepo),
		StatsQuery:                memory.NewStatsQuery(closingRepo),
		TxManager:                 txMgr,
		TaxScales:                 arcaClient,
		DueDates:                  arcaClient,
		ExchangeRates:             fxClient,
		IDGen:                     idGen,
		ChangeNotifier:            &ports.NoOpChangeNotifier{},
		NotificationRepo:          notifRepo,
		OutboxRepo:                outboxRepo,
		NotificationSender:        &fakes.NotificationSender{},
	})

	outDir := filepath.Join("..", "..", "..", "..", "frontend", "src", "test", "fixtures", "api")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir fixtures: %v", err)
	}

	writeFixture := func(name, path string) []byte {
		rec := do(t, h, "GET", path, "accountant", nil)
		if rec.Code != 200 {
			t.Fatalf("GET %s status %d: %s", path, rec.Code, rec.Body.String())
		}
		var buf bytes.Buffer
		if err := json.Indent(&buf, rec.Body.Bytes(), "", "  "); err != nil {
			t.Fatalf("indent %s: %v", name, err)
		}
		buf.WriteByte('\n')
		if err := os.WriteFile(filepath.Join(outDir, name), buf.Bytes(), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return buf.Bytes()
	}

	writePostFixture := func(name, path, user string, body io.Reader, contentType string) []byte {
		req := httptest.NewRequest("POST", path, body)
		if user != "" {
			req.Header.Set("X-Demo-User", user)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("POST %s status %d: %s", path, rec.Code, rec.Body.String())
		}
		var buf bytes.Buffer
		if err := json.Indent(&buf, rec.Body.Bytes(), "", "  "); err != nil {
			t.Fatalf("indent %s: %v", name, err)
		}
		buf.WriteByte('\n')
		if err := os.WriteFile(filepath.Join(outDir, name), buf.Bytes(), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return buf.Bytes()
	}

	closingsBytes := writeFixture("closings.json", "/api/v1/accountant/periods/2026-10/closings?page=1&pageSize=50")
	writeFixture("closings_summary.json", "/api/v1/accountant/periods/2026-10/closings/summary")
	writeFixture("closings_ids.json", "/api/v1/accountant/periods/2026-10/closings/ids")
	writeFixture("stats.json", "/api/v1/accountant/periods/2026-10/stats")
	writeFixture("signals_catalog.json", "/api/v1/accountant/signals/catalog")
	writeFixture("parameters.json", "/api/v1/accountant/parameters")

	getClosing := func(clientID string, per period.Period) closing.Closing {
		c, found, err := closingRepo.ByClientPeriod(ctx, clientID, per)
		if err != nil || !found {
			t.Fatalf("closing not found for client %s, period %s: err=%v, found=%v", clientID, per, err, found)
		}
		return c
	}

	// 1. Monotributo estándar (cli_demo_01)
	cloMono := getClosing("cli_demo_01", p)
	monoBytes := writeFixture("closing_detail_mono.json", "/api/v1/accountant/closings/"+cloMono.ID)
	if err := os.WriteFile(filepath.Join(outDir, "closing_detail.json"), monoBytes, 0o644); err != nil {
		t.Fatalf("write closing_detail.json: %v", err)
	}

	// 2. Monotributo con exclusión (cli_demo_02 - Ignacio Sosa)
	cloExcl := getClosing("cli_demo_02", p)
	writeFixture("closing_detail_mono_excluded.json", "/api/v1/accountant/closings/"+cloExcl.ID)

	// 3. Responsable Inscripto (cli_demo_09 - Camila Torres)
	cloRI := getClosing("cli_demo_09", p)
	writeFixture("closing_detail_ri.json", "/api/v1/accountant/closings/"+cloRI.ID)

	// 4. Pendiente de confirmación del cliente (cli_demo_15 - Paula Rojas)
	cloPending := getClosing("cli_demo_15", p)
	writeFixture("closing_detail_pending_input.json", "/api/v1/accountant/closings/"+cloPending.ID)

	// 5. Cierre notificado de septiembre 2026 (cli_demo_01)
	cloNotified := getClosing("cli_demo_01", pSep)
	writeFixture("closing_detail_notified.json", "/api/v1/accountant/closings/"+cloNotified.ID)

	// 6. Cierre reabierto v2 con aprobaciones previas (cli_demo_05)
	cloReopen := getClosing("cli_demo_05", p)
	cloReopen.Status = closing.StatusApproved
	cloReopen.Approvals = []closing.Approval{{
		ID:         idGen.New("apr"),
		Version:    cloReopen.Version,
		ResultHash: cloReopen.ResultHash,
		ApprovedBy: "accountant",
		ApprovedAt: clk.Now().Add(-2 * time.Hour),
		Note:       "Primera aprobación antes de nuevo comprobante",
	}}
	cloReopen.ApplyResult("reopened_hash_val_v2", cloReopen.Severity, cloReopen.Signals, cloReopen.Description, cloReopen.Mono, cloReopen.RI, cloReopen.Reconciliation, clk.Now())
	if err := closingRepo.Save(ctx, cloReopen); err != nil {
		t.Fatalf("save reopened closing: %v", err)
	}
	writeFixture("closing_detail_reopened.json", "/api/v1/accountant/closings/"+cloReopen.ID)

	// 7. Endpoint de escritura: approve_response.json (cli_demo_04)
	cloApprove := getClosing("cli_demo_04", p)
	approveBody, _ := json.Marshal(map[string]any{
		"version": cloApprove.Version,
	})
	writePostFixture("approve_response.json", "/api/v1/accountant/closings/"+cloApprove.ID+"/approve", "accountant", bytes.NewReader(approveBody), "application/json")

	// 8. Endpoint de escritura: approve_batch_response.json (2 cierres verdes)
	var green1, green2 closing.Closing
	closingsList, err := closingRepo.ListByPeriod(ctx, p)
	if err != nil {
		t.Fatalf("ListByPeriod: %v", err)
	}
	for _, c := range closingsList {
		if c.Severity == signal.SeverityGreen && c.Status == closing.StatusReadyForReview {
			if green1.ID == "" {
				green1 = c
			} else if green2.ID == "" {
				green2 = c
				break
			}
		}
	}
	if green1.ID == "" || green2.ID == "" {
		t.Fatalf("no green closings found for batch approve fixture")
	}
	batchReqBody, _ := json.Marshal(map[string]any{
		"items": []map[string]any{
			{"id": green1.ID, "version": green1.Version},
			{"id": green2.ID, "version": green2.Version},
		},
	})
	writePostFixture("approve_batch_response.json", "/api/v1/accountant/periods/2026-10/closings/approve-batch", "accountant", bytes.NewReader(batchReqBody), "application/json")

	// 9. Endpoint de escritura: regime_change_response.json (cli_demo_02 - Ignacio Sosa)
	latestExcl, _ := closingRepo.LatestExclusionDate(ctx, "cli_demo_02")
	effDateStr := "2026-10-15"
	if latestExcl != nil {
		effDateStr = latestExcl.In(period.Location()).Format("2006-01-02")
	}
	regimeReqBody, _ := json.Marshal(map[string]any{
		"to":            "RI",
		"effectiveDate": effDateStr,
		"force":         true,
		"note":          "Pase a Responsable Inscripto por superar límite de facturación",
	})
	writePostFixture("regime_change_response.json", "/api/v1/accountant/clients/cli_demo_02/regime-change", "accountant", bytes.NewReader(regimeReqBody), "application/json")

	// 10. Endpoint de escritura: invoices_import_client_response.json (cli_demo_03)
	csvContent := "Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total\n12/10/2026,19 - Factura E,00001,00000099,00000099,74392819289901,80,DE329182391,Brightloop GmbH,1482.15,USD,850.00,0.00,0.00,0.00,850.00\n"
	bodyBuf := &bytes.Buffer{}
	mpWriter := multipart.NewWriter(bodyBuf)
	part, _ := mpWriter.CreateFormFile("file", "mis_comprobantes.csv")
	_, _ = part.Write([]byte(csvContent))
	_ = mpWriter.Close()
	writePostFixture("invoices_import_client_response.json", "/api/v1/me/invoices/import", "client:cli_demo_03", bodyBuf, mpWriter.FormDataContentType())

	// 11. Endpoint de escritura: invoices_import_accountant_response.json (cli_demo_03)
	csvAccContent := "Fecha,Tipo,Punto de Venta,Número Desde,Número Hasta,Cód. Autorización,Tipo Doc. Receptor,Nro. Doc. Receptor,Denominación Receptor,Tipo Cambio,Moneda,Imp. Neto Gravado,Imp. Neto No Gravado,Imp. Op. Exentas,IVA,Imp. Total\n14/10/2026,19 - Factura E,00001,00000098,00000098,74392819289902,80,DE329182392,NextGen AI Inc,1482.15,USD,1200.00,0.00,0.00,0.00,1200.00\n"
	bodyAccBuf := &bytes.Buffer{}
	mpAccWriter := multipart.NewWriter(bodyAccBuf)
	partAcc, _ := mpAccWriter.CreateFormFile("file", "mis_comprobantes_acc.csv")
	_, _ = partAcc.Write([]byte(csvAccContent))
	_ = mpAccWriter.Close()
	writePostFixture("invoices_import_accountant_response.json", "/api/v1/accountant/clients/cli_demo_03/invoices/import", "accountant", bodyAccBuf, mpAccWriter.FormDataContentType())

	// 12. Endpoint de escritura: confirm_input_response.json (cli_demo_15 - Paula Rojas)
	confirmBody, _ := json.Marshal(map[string]any{
		"externalIncome": map[string]any{
			"none": true,
		},
		"note": "Me lo confirmó por mail el 06/11",
	})
	writePostFixture("confirm_input_response.json", "/api/v1/accountant/clients/cli_demo_15/periods/2026-10/confirm-input", "accountant", bytes.NewReader(confirmBody), "application/json")

	_ = closingsBytes
}
