package http

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
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
	computeUC := &usecase.ComputeClosing{
		Clients: clientRepo, Movements: movRepo, Classifications: classRepo,
		Invoices: invRepo, PurchaseInvoices: purchRepo, ClientInputs: inputRepo,
		Parameters: paramRepo, Closings: closingRepo,
		TaxScales: arcaClient, DueDates: arcaClient, ExchangeRates: fxClient,
		Clock: clk, IDGen: idGen, Tx: txMgr,
	}
	runUC := &usecase.RunMonthlyClosing{Clients: clientRepo, Compute: computeUC, Workers: 1}
	if _, err := runUC.Execute(ctx, p); err != nil {
		t.Fatalf("RunMonthlyClosing: %v", err)
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
		LatestExclusionDateReader: &fakes.LatestExclusionDateReader{},
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
		return rec.Body.Bytes()
	}

	closingsBytes := writeFixture("closings.json", "/api/v1/accountant/periods/2026-10/closings?page=1&pageSize=50")
	writeFixture("closings_summary.json", "/api/v1/accountant/periods/2026-10/closings/summary")
	writeFixture("closings_ids.json", "/api/v1/accountant/periods/2026-10/closings/ids")
	writeFixture("stats.json", "/api/v1/accountant/periods/2026-10/stats")
	writeFixture("signals_catalog.json", "/api/v1/accountant/signals/catalog")
	writeFixture("parameters.json", "/api/v1/accountant/parameters")

	var page struct {
		Items []struct {
			ClosingID string `json:"closingId"`
		} `json:"items"`
	}
	if err := json.Unmarshal(closingsBytes, &page); err != nil || len(page.Items) == 0 {
		t.Fatalf("no items in closings.json: %v", err)
	}
	writeFixture("closing_detail.json", "/api/v1/accountant/closings/"+page.Items[0].ClosingID)
}
