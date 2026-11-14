package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestListClosingsAndSummaryAndDetail(t *testing.T) {
	ctx := context.Background()
	p, _ := period.Parse("2026-10")

	closingRepo := fakes.NewClosingRepository()
	clientRepo := fakes.NewClientRepository()
	movementRepo := fakes.NewMovementRepository()
	invoiceRepo := fakes.NewInvoiceRepository()
	clientInputRepo := fakes.NewClientInputRepository()
	paramRepo := &fakes.ParametersRepository{Params: parameters.Defaults()}
	classRepo := fakes.NewClassificationRepository()
	fx := &fakes.ExchangeRateProvider{
		Default: ports.Quote{
			Buy:  150000,
			Sell: 155000,
		},
	}

	c1 := client.Client{
		ID:                  "cli_01",
		Name:                "Client One",
		CUIT:                "20-12345678-9",
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatC,
	}
	_ = clientRepo.Save(ctx, c1)

	cl1 := closing.Closing{
		ID:          "cls_01",
		ClientID:    "cli_01",
		Period:      p,
		Status:      closing.StatusReadyForReview,
		Version:     1,
		ResultHash:  "hash1",
		Severity:    signal.SeverityGreen,
		Description: "Todo en orden",
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodARS: money.New(150000000, money.ARS),
			BilledInPeriodUSD: money.New(100000, money.USD),
			Billed12mARS:      money.New(1500000000, money.ARS),
			PctOfKCapBp:       3000,
			Fee:               money.New(5000000, money.ARS),
			FeeDueDate:        time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		},
		ComputedAt: time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = closingRepo.Save(ctx, cl1)

	// ListClosings
	closingQuery := fakes.NewClosingQuery(closingRepo, clientRepo)
	listUC := &usecase.ListClosings{
		Query: closingQuery,
	}
	res, err := listUC.Execute(ctx, usecase.ListClosingsQuery{
		Period:   p,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListClosings failed: %v", err)
	}
	if res.Total != 1 || len(res.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", res.Total)
	}
	if res.Items[0].ClientName != "Client One" {
		t.Errorf("expected Client One, got %s", res.Items[0].ClientName)
	}

	// ClosingsSummary
	summaryUC := &usecase.ClosingsSummary{
		Query: closingQuery,
	}
	summary, err := summaryUC.Execute(ctx, p, closing.Filter{})
	if err != nil {
		t.Fatalf("ClosingsSummary failed: %v", err)
	}
	if summary.Total != 1 || summary.BySeverity["GREEN"] != 1 {
		t.Errorf("unexpected summary: %+v", summary)
	}

	// GetClosingDetail
	detailUC := &usecase.GetClosingDetail{
		Closings:        closingRepo,
		Clients:         clientRepo,
		Movements:       movementRepo,
		Classifications: classRepo,
		Invoices:        invoiceRepo,
		ClientInputs:    clientInputRepo,
		Parameters:      paramRepo,
		ExchangeRates:   fx,
	}

	// Not found
	_, err = detailUC.Execute(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent closing")
	}

	// Found
	detail, err := detailUC.Execute(ctx, "cls_01")
	if err != nil {
		t.Fatalf("GetClosingDetail failed: %v", err)
	}
	if detail.Closing.ID != "cls_01" || detail.Client.ID != "cli_01" {
		t.Errorf("unexpected detail: %+v", detail)
	}

	// Also test RI row building
	c2 := client.Client{
		ID:     "cli_02",
		Name:   "Client Two RI",
		CUIT:   "20-87654321-0",
		Regime: client.RegimeRI,
	}
	_ = clientRepo.Save(ctx, c2)
	dueDate := time.Date(2026, 11, 18, 0, 0, 0, 0, period.Location())
	cl2 := closing.Closing{
		ID:          "cls_02",
		ClientID:    "cli_02",
		Period:      p,
		Status:      closing.StatusPendingClientInput,
		Version:     1,
		ResultHash:  "hash2",
		Severity:    signal.SeverityYellow,
		Description: "Atención requerida",
		Signals: []signal.Signal{
			{Code: signal.CodeInvoiceWithoutPayment, Severity: signal.SeverityYellow, Priority: 11},
			{Code: signal.CodeAtypicalIncome, Severity: signal.SeverityYellow, Priority: 8},
		},
		RI: &tax.RIAssessment{
			ExportSalesARS: money.New(750000000, money.ARS),
			TotalToPay:     money.New(10000000, money.ARS),
			VatDueDate:     dueDate,
		},
		ComputedAt: time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = closingRepo.Save(ctx, cl2)

	// List again with 2 items
	res2, err := listUC.Execute(ctx, usecase.ListClosingsQuery{
		Period:   p,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListClosings failed: %v", err)
	}
	if res2.Total != 2 {
		t.Fatalf("expected 2 items, got %d", res2.Total)
	}

	// Detail for RI with movements and invoices
	_ = movementRepo.Save(ctx, movement.Movement{
		ID:         "mov_01",
		ClientID:   "cli_02",
		Amount:     money.New(100000, money.USD),
		OccurredAt: time.Date(2026, 10, 10, 0, 0, 0, 0, period.Location()),
	})
	_ = invoiceRepo.Save(ctx, invoice.Invoice{
		ID:          "inv_01",
		ClientID:    "cli_02",
		IssueDate:   time.Date(2026, 10, 8, 0, 0, 0, 0, period.Location()),
		TotalAmount: money.New(100000, money.USD),
	})
	_ = clientInputRepo.Save(ctx, clientinput.ClientPeriodInput{
		ClientID: "cli_02",
		Period:   p,
	})
	detail2, err := detailUC.Execute(ctx, "cls_02")
	if err != nil {
		t.Fatalf("GetClosingDetail for RI failed: %v", err)
	}
	if len(detail2.PeriodMovements) != 1 || len(detail2.PeriodInvoices) != 1 {
		t.Errorf("unexpected counts in RI detail: %+v", detail2)
	}

	// ListClosingIDs
	idsUC := &usecase.ListClosingIDs{
		Query: closingQuery,
	}
	idsRes, err := idsUC.Execute(ctx, usecase.ListClosingIDsQuery{
		Period: p,
	})
	if err != nil {
		t.Fatalf("ListClosingIDs failed: %v", err)
	}
	if idsRes.Total != 2 || len(idsRes.IDs) != 2 {
		t.Fatalf("expected 2 IDs, got %d", idsRes.Total)
	}
	// Default sort is severity desc (RED cls_02 first, then GREEN cls_01)
	if idsRes.IDs[0] != "cls_02" || idsRes.IDs[1] != "cls_01" {
		t.Errorf("unexpected IDs order: %+v", idsRes.IDs)
	}
}
