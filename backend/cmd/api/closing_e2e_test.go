package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/config"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	arcamock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/arca/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	fxmock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
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

	// Criterio 17: Para cada cli_demo_XX la señal principal y el estado coinciden con la tabla de 002/plan.md
	expectedDemo := map[string]struct {
		Severity      signal.Severity
		Status        closing.Status
		PrimarySignal *signal.Code
	}{
		"cli_demo_01": {Severity: signal.SeverityGreen, Status: closing.StatusReadyForReview, PrimarySignal: nil},
		"cli_demo_02": {Severity: signal.SeverityRed, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoExclusionKCapExceeded)},
		"cli_demo_03": {Severity: signal.SeverityRed, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMissingInvoices)},
		"cli_demo_04": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoCategoryChangeRequired)},
		"cli_demo_05": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoNearCategoryCap)},
		"cli_demo_06": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeAtypicalIncome)},
		"cli_demo_07": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeUnclassifiedMovements)},
		"cli_demo_08": {Severity: signal.SeverityYellow, Status: closing.StatusReadyForReview, PrimarySignal: codePtr(signal.CodeMonoNearKCap)},
		"cli_demo_09": {Severity: signal.SeverityGreen, Status: closing.StatusReadyForReview, PrimarySignal: nil},
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
