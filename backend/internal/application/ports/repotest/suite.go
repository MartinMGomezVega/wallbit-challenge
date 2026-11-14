// Package repotest define suites de contrato reutilizables para memoria y Postgres.
package repotest

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// Bundle agrupa las implementaciones concretas bajo test de contrato.
type Bundle struct {
	TxManager        ports.TxManager
	Clients          ports.ClientRepository
	Movements        ports.MovementRepository
	Classifications  ports.ClassificationRepository
	Invoices         ports.InvoiceRepository
	PurchaseInvoices ports.PurchaseInvoiceRepository
	ClientInputs     ports.ClientInputRepository
	Closings         ports.ClosingRepository
	ClosingQuery     ports.ClosingQuery
	StatsQuery       ports.StatsQuery
	Notifications    ports.NotificationRepository
	Outbox           ports.OutboxRepository
	Parameters       ports.ParametersRepository
}

// RunAll ejecuta todas las suites de contrato de repositorios, query y transacciones.
func RunAll(t *testing.T, newBundle func(t *testing.T) Bundle, supportsRollback bool) {
	t.Helper()
	t.Run("Clients", func(t *testing.T) { RunClientRepo(t, newBundle(t).Clients) })
	t.Run("Movements", func(t *testing.T) { RunMovementRepo(t, newBundle(t)) })
	t.Run("Classifications", func(t *testing.T) { RunClassificationRepo(t, newBundle(t)) })
	t.Run("Invoices", func(t *testing.T) { RunInvoiceRepo(t, newBundle(t)) })
	t.Run("PurchaseInvoices", func(t *testing.T) { RunPurchaseInvoiceRepo(t, newBundle(t)) })
	t.Run("ClientInputs", func(t *testing.T) { RunClientInputRepo(t, newBundle(t)) })
	t.Run("Closings", func(t *testing.T) { RunClosingRepo(t, newBundle(t)) })
	t.Run("ClosingQuery", func(t *testing.T) { RunClosingQuery(t, newBundle(t)) })
	t.Run("StatsQuery", func(t *testing.T) { RunStatsQuery(t, newBundle(t)) })
	t.Run("Notifications", func(t *testing.T) { RunNotificationRepo(t, newBundle(t)) })
	t.Run("Outbox", func(t *testing.T) { RunOutboxRepo(t, newBundle(t).Outbox) })
	t.Run("Parameters", func(t *testing.T) { RunParametersRepo(t, newBundle(t).Parameters) })
	t.Run("TxManager", func(t *testing.T) { RunTxManager(t, newBundle(t), supportsRollback) })
}

func sampleClient(id, name, cuit string, regime client.Regime, cat tax.Category) client.Client {
	loc := period.Location()
	created := time.Date(2025, 1, 1, 0, 0, 0, 0, loc)
	return client.Client{
		ID:                  id,
		Name:                name,
		CUIT:                client.CUIT(cuit),
		Email:               id + "@example.com",
		Regime:              regime,
		MonotributoCategory: cat,
		CreatedAt:           created,
		RegimeHistory: []client.RegimeHistoryItem{{
			Regime: regime,
			From:   created,
			Reason: "INIT",
		}},
	}
}

func RunClientRepo(t *testing.T, repo ports.ClientRepository) {
	t.Helper()
	ctx := context.Background()

	c1 := sampleClient("cli_1", "Ana López", "20-11111111-1", client.RegimeMonotributo, tax.CatB)
	c2 := sampleClient("cli_2", "Beto Gómez", "20-22222222-2", client.RegimeRI, "")
	if err := repo.SaveAll(ctx, []client.Client{c1, c2}); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	got, err := repo.Get(ctx, "cli_1")
	if err != nil || got.Name != "Ana López" {
		t.Fatalf("Get cli_1 = %+v, err=%v", got, err)
	}
	if _, err := repo.Get(ctx, "cli_missing"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("List len = %d, err=%v", len(list), err)
	}

	c1.Name = "Ana María López"
	if err := repo.Save(ctx, c1); err != nil {
		t.Fatalf("Save update failed: %v", err)
	}
	got2, err := repo.Get(ctx, "cli_1")
	if err != nil || got2.Name != "Ana María López" {
		t.Fatalf("updated Get = %+v, err=%v", got2, err)
	}
}

func RunMovementRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_m", "Mov Client", "20-33333333-3", client.RegimeMonotributo, tax.CatA))

	t1 := time.Date(2026, 10, 5, 10, 0, 0, 0, loc)
	t2 := time.Date(2026, 10, 20, 10, 0, 0, 0, loc)
	t3 := time.Date(2026, 11, 2, 10, 0, 0, 0, loc)
	ms := []movement.Movement{
		{ID: "mov_1", ClientID: "cli_m", Type: movement.TypeInboundTransfer, Amount: money.New(10000, money.USD), FXRate: 150000, OccurredAt: t1, Counterparty: movement.Counterparty{Name: "ACME", Country: "US", Kind: movement.CounterpartyCompany}, Description: "REF1"},
		{ID: "mov_2", ClientID: "cli_m", Type: movement.TypeCardSpend, Amount: money.New(5000, money.USD), FXRate: 150000, OccurredAt: t2, Counterparty: movement.Counterparty{Name: "SHOP", Country: "AR", Kind: movement.CounterpartyCompany}, Description: "REF2"},
		{ID: "mov_3", ClientID: "cli_m", Type: movement.TypeInboundTransfer, Amount: money.New(2000, money.USD), FXRate: 150000, OccurredAt: t3, Counterparty: movement.Counterparty{Name: "GLOBEX", Country: "US", Kind: movement.CounterpartyCompany}, Description: "REF3"},
	}
	if err := b.Movements.SaveAll(ctx, ms); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}
	got, err := b.Movements.Get(ctx, "mov_2")
	if err != nil || got.Description != "REF2" {
		t.Fatalf("Get mov_2 = %+v, err=%v", got, err)
	}
	byCli, err := b.Movements.ByClient(ctx, "cli_m")
	if err != nil || len(byCli) != 3 {
		t.Fatalf("ByClient len = %d, err=%v", len(byCli), err)
	}
	between, err := b.Movements.ByClientBetween(ctx, "cli_m", time.Date(2026, 10, 1, 0, 0, 0, 0, loc), time.Date(2026, 10, 31, 23, 59, 59, 0, loc))
	if err != nil || len(between) != 2 {
		t.Fatalf("ByClientBetween len = %d, err=%v", len(between), err)
	}
	all, err := b.Movements.ListAll(ctx)
	if err != nil || len(all) != 3 {
		t.Fatalf("ListAll len = %d, err=%v", len(all), err)
	}
}

func RunClassificationRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_c", "Cls Client", "20-44444444-4", client.RegimeMonotributo, tax.CatA))
	_ = b.Movements.Save(ctx, movement.Movement{ID: "mov_c1", ClientID: "cli_c", Type: movement.TypeInboundTransfer, Amount: money.New(1000, money.USD), FXRate: 150000, OccurredAt: time.Date(2026, 10, 1, 0, 0, 0, 0, loc)})

	cls := movement.Classification{MovementID: "mov_c1", Kind: movement.KindIncome, Confidence: movement.ConfidenceHigh, Rule: movement.RuleForeignCompany, Note: "Rule 1"}
	if err := b.Classifications.Save(ctx, cls); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	got, ok, err := b.Classifications.Get(ctx, "mov_c1")
	if err != nil || !ok || got.Kind != movement.KindIncome {
		t.Fatalf("Get = %+v, ok=%v, err=%v", got, ok, err)
	}
	byIDs, err := b.Classifications.ByMovementIDs(ctx, []string{"mov_c1", "mov_missing"})
	if err != nil || len(byIDs) != 1 {
		t.Fatalf("ByMovementIDs = %+v, err=%v", byIDs, err)
	}
}

func RunInvoiceRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_i", "Inv Client", "20-55555555-5", client.RegimeMonotributo, tax.CatA))

	fxRate := money.Rate(100)
	inv := invoice.Invoice{
		ID: "inv_1", ClientID: "cli_i", Type: invoice.TypeC, PointOfSale: 1, Number: 10,
		IssueDate: time.Date(2026, 10, 10, 0, 0, 0, 0, loc), Currency: money.ARS,
		NetAmount: money.New(10000, money.ARS), VatAmount: money.New(0, money.ARS), TotalAmount: money.New(10000, money.ARS),
		FXRate: &fxRate, ReceiverName: "Client X", ReceiverDoc: "20123456789", Source: invoice.SourceManual, CreatedAt: time.Date(2026, 10, 10, 1, 0, 0, 0, loc),
	}
	if err := b.Invoices.Save(ctx, inv); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	exists, err := b.Invoices.Exists(ctx, "cli_i", invoice.Key{ClientID: "cli_i", Type: invoice.TypeC, PointOfSale: 1, Number: 10})
	if err != nil || !exists {
		t.Fatalf("Exists = %v, err=%v", exists, err)
	}
	got, err := b.Invoices.Get(ctx, "inv_1")
	if err != nil || got.Number != 10 {
		t.Fatalf("Get = %+v, err=%v", got, err)
	}
	byCli, err := b.Invoices.ByClient(ctx, "cli_i")
	if err != nil || len(byCli) != 1 {
		t.Fatalf("ByClient len = %d, err=%v", len(byCli), err)
	}
	between, err := b.Invoices.ByClientBetween(ctx, "cli_i", time.Date(2026, 10, 1, 0, 0, 0, 0, loc), time.Date(2026, 10, 31, 23, 59, 59, 0, loc))
	if err != nil || len(between) != 1 {
		t.Fatalf("ByClientBetween len = %d, err=%v", len(between), err)
	}
}

func RunPurchaseInvoiceRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_pi", "PI Client", "20-66666666-6", client.RegimeRI, ""))

	pi := invoice.PurchaseInvoice{
		ID: "pinv_1", ClientID: "cli_pi",
		IssueDate: time.Date(2026, 10, 15, 0, 0, 0, 0, loc), SupplierName: "Supplier SA", SupplierCUIT: "30-77777777-7",
		NetAmount: money.New(10000, money.ARS), VatAmount: money.New(2100, money.ARS),
		CreatedAt: time.Date(2026, 10, 15, 1, 0, 0, 0, loc),
	}
	if err := b.PurchaseInvoices.Save(ctx, pi); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	byCli, err := b.PurchaseInvoices.ByClient(ctx, "cli_pi")
	if err != nil || len(byCli) != 1 {
		t.Fatalf("ByClient len = %d, err=%v", len(byCli), err)
	}
	between, err := b.PurchaseInvoices.ByClientBetween(ctx, "cli_pi", time.Date(2026, 10, 1, 0, 0, 0, 0, loc), time.Date(2026, 10, 31, 23, 59, 59, 0, loc))
	if err != nil || len(between) != 1 {
		t.Fatalf("ByClientBetween len = %d, err=%v", len(between), err)
	}
}

func RunClientInputRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_ci", "CI Client", "20-77777777-7", client.RegimeMonotributo, tax.CatA))
	p := period.MustParse("2026-10")
	now := time.Date(2026, 10, 28, 10, 0, 0, 0, loc)

	in := clientinput.ClientPeriodInput{
		ClientID: "cli_ci",
		Period:   p,
		ExternalIncome: []clientinput.ExternalIncomeItem{
			{Amount: money.New(50000, money.ARS), Description: "Local work"},
		},
		ConfirmedAt: &now,
	}
	if err := b.ClientInputs.Save(ctx, in); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	got, err := b.ClientInputs.Get(ctx, "cli_ci", p)
	if err != nil || len(got.ExternalIncome) != 1 || got.ConfirmedAt == nil {
		t.Fatalf("Get = %+v, err=%v", got, err)
	}
}

func RunClosingRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_cl", "Closing Client", "20-88888888-8", client.RegimeMonotributo, tax.CatK))
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, loc)
	excl := time.Date(2026, 10, 25, 0, 0, 0, 0, loc)

	c := closing.NewClosing(
		"clo_1", "cli_cl", p, true, "hash1", signal.SeverityRed,
		[]signal.Signal{{Code: signal.CodeMonoExclusionKCapExceeded, Severity: signal.SeverityRed, Priority: 1, Message: "Superó tope K"}},
		"Superó tope K",
		&tax.MonotributoAssessment{
			CurrentCategory: tax.CatK, Fee: money.New(500000, money.ARS), FeeDueDate: now.AddDate(0, 0, 19),
			BilledInPeriodUSD: money.New(10000, money.USD), BilledInPeriodARS: money.New(15000000, money.ARS),
			Billed12mARS: money.New(150000000, money.ARS), PctOfKCapBp: 10500, Excluded: true, ExclusionDate: &excl,
		},
		nil, closing.Reconciliation{}, now,
	)
	if err := b.Closings.Save(ctx, c); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := b.Closings.Get(ctx, "clo_1")
	if err != nil || got.Version != 1 || len(got.Signals) != 1 {
		t.Fatalf("Get = %+v, err=%v", got, err)
	}
	byCP, found, err := b.Closings.ByClientPeriod(ctx, "cli_cl", p)
	if err != nil || !found || byCP.ID != "clo_1" {
		t.Fatalf("ByClientPeriod = %+v, found=%v, err=%v", byCP, found, err)
	}
	latestExcl, err := b.Closings.LatestExclusionDate(ctx, "cli_cl")
	if err != nil || latestExcl == nil {
		t.Fatalf("LatestExclusionDate = %v, err=%v", latestExcl, err)
	}

	// SaveIfVersion: mismatch version -> ErrVersionConflict
	if err := b.Closings.SaveIfVersion(ctx, got, 99); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("expected ErrVersionConflict, got %v", err)
	}

	// SaveIfVersion: concurrent approval on same version -> 1 ok, 1 ErrVersionConflict
	cA := got
	cB := got
	_ = cA.Approve(closing.Approval{ID: "ap_a", Version: 1, ResultHash: cA.ResultHash, ApprovedBy: "acc", ApprovedAt: now})
	_ = cA.Publish(closing.Published{Version: 1, Title: "t", Body: "b", PublishedAt: now})
	_ = cB.Approve(closing.Approval{ID: "ap_b", Version: 1, ResultHash: cB.ResultHash, ApprovedBy: "acc", ApprovedAt: now})
	_ = cB.Publish(closing.Published{Version: 1, Title: "t", Body: "b", PublishedAt: now})

	errs := make([]error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); errs[0] = b.Closings.SaveIfVersion(ctx, cA, 1) }()
	go func() { defer wg.Done(); errs[1] = b.Closings.SaveIfVersion(ctx, cB, 1) }()
	wg.Wait()

	okCnt, confCnt := 0, 0
	for _, e := range errs {
		if e == nil {
			okCnt++
		} else if errors.Is(e, domain.ErrVersionConflict) {
			confCnt++
		} else {
			t.Fatalf("unexpected concurrent error: %v", e)
		}
	}
	if okCnt != 1 || confCnt != 1 {
		t.Fatalf("expected 1 ok & 1 ErrVersionConflict, got ok=%d conf=%d errs=%v", okCnt, confCnt, errs)
	}
}

func RunClosingQuery(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, loc)

	clients := []client.Client{
		sampleClient("cq_1", "Alvarez, Juan", "20-10000001-1", client.RegimeMonotributo, tax.CatK),
		sampleClient("cq_2", "Bianchi, Sofía", "20-20000002-2", client.RegimeMonotributo, tax.CatD),
		sampleClient("cq_3", "Castro, Diego", "20-30000003-3", client.RegimeRI, ""),
		sampleClient("cq_4", "Domínguez, Laura", "20-40000004-4", client.RegimeMonotributo, tax.CatA),
	}
	if err := b.Clients.SaveAll(ctx, clients); err != nil {
		t.Fatalf("SaveAll clients failed: %v", err)
	}

	closings := []closing.Closing{
		closing.NewClosing("cqc_1", "cq_1", p, true, "h1", signal.SeverityRed,
			[]signal.Signal{{Code: signal.CodeMonoExclusionKCapExceeded, Severity: signal.SeverityRed, Priority: 1}}, "Excedió K",
			&tax.MonotributoAssessment{CurrentCategory: tax.CatK, Fee: money.New(900000, money.ARS), FeeDueDate: now, BilledInPeriodARS: money.New(3000000, money.ARS), BilledInPeriodUSD: money.New(2000, money.USD), Billed12mARS: money.New(120000000, money.ARS), PctOfKCapBp: 10500},
			nil, closing.Reconciliation{}, now),
		closing.NewClosing("cqc_2", "cq_2", p, false, "h2", signal.SeverityYellow,
			[]signal.Signal{{Code: signal.CodeMissingInvoices, Severity: signal.SeverityYellow, Priority: 3}}, "Falta factura",
			&tax.MonotributoAssessment{CurrentCategory: tax.CatD, Fee: money.New(400000, money.ARS), FeeDueDate: now, BilledInPeriodARS: money.New(1000000, money.ARS), BilledInPeriodUSD: money.New(700, money.USD), Billed12mARS: money.New(40000000, money.ARS), PctOfKCapBp: 4500},
			nil, closing.Reconciliation{}, now),
		closing.NewClosing("cqc_3", "cq_3", p, true, "h3", signal.SeverityGreen,
			nil, "Sin novedades",
			nil, &tax.RIAssessment{VatToPay: money.New(1200000, money.ARS), VatDueDate: now, TotalToPay: money.New(1200000, money.ARS), BilledInPeriodARS: money.New(5000000, money.ARS), BilledInPeriodUSD: money.New(3500, money.USD)},
			closing.Reconciliation{}, now),
		closing.NewClosing("cqc_4", "cq_4", p, true, "h4", signal.SeverityGreen,
			nil, "Todo OK",
			&tax.MonotributoAssessment{CurrentCategory: tax.CatA, Fee: money.New(100000, money.ARS), FeeDueDate: now, BilledInPeriodARS: money.New(500000, money.ARS), BilledInPeriodUSD: money.New(300, money.USD), Billed12mARS: money.New(10000000, money.ARS), PctOfKCapBp: 1200},
			nil, closing.Reconciliation{}, now),
	}
	for _, c := range closings {
		if err := b.Closings.Save(ctx, c); err != nil {
			t.Fatalf("Save closing %s failed: %v", c.ID, err)
		}
	}

	// Combinaciones de filtro y orden (incluye verificación de defaults cuando Dir es "")
	cases := []struct {
		name    string
		filter  closing.Filter
		sort    closing.Sort
		wantIDs []string
	}{
		{"01_default_severity_desc", closing.Filter{}, closing.Sort{By: "severity", Dir: "desc"}, []string{"cqc_1", "cqc_2", "cqc_3", "cqc_4"}},
		{"01b_default_sort_empty_dir", closing.Filter{}, closing.Sort{By: "severity"}, []string{"cqc_1", "cqc_2", "cqc_3", "cqc_4"}},
		{"01c_default_empty_sort", closing.Filter{}, closing.Sort{}, []string{"cqc_1", "cqc_2", "cqc_3", "cqc_4"}},
		{"02_severity_asc", closing.Filter{}, closing.Sort{By: "severity", Dir: "asc"}, []string{"cqc_4", "cqc_3", "cqc_2", "cqc_1"}},
		{"03_name_asc", closing.Filter{}, closing.Sort{By: "name", Dir: "asc"}, []string{"cqc_1", "cqc_2", "cqc_3", "cqc_4"}},
		{"03b_name_empty_dir", closing.Filter{}, closing.Sort{By: "name"}, []string{"cqc_1", "cqc_2", "cqc_3", "cqc_4"}},
		{"04_name_desc", closing.Filter{}, closing.Sort{By: "name", Dir: "desc"}, []string{"cqc_4", "cqc_3", "cqc_2", "cqc_1"}},
		{"05_toPay_desc", closing.Filter{}, closing.Sort{By: "toPay", Dir: "desc"}, []string{"cqc_3", "cqc_1", "cqc_2", "cqc_4"}},
		{"05b_toPay_empty_dir", closing.Filter{}, closing.Sort{By: "toPay"}, []string{"cqc_3", "cqc_1", "cqc_2", "cqc_4"}},
		{"06_pctKCap_desc", closing.Filter{}, closing.Sort{By: "pctKCap", Dir: "desc"}, []string{"cqc_1", "cqc_2", "cqc_4", "cqc_3"}},
		{"06b_pctKCap_empty_dir", closing.Filter{}, closing.Sort{By: "pctKCap"}, []string{"cqc_1", "cqc_2", "cqc_4", "cqc_3"}},
		{"07_filter_severity_red_yellow", closing.Filter{Severities: []signal.Severity{signal.SeverityRed, signal.SeverityYellow}}, closing.Sort{By: "severity", Dir: "desc"}, []string{"cqc_1", "cqc_2"}},
		{"08_filter_regime_RI", closing.Filter{Regime: client.RegimeRI}, closing.Sort{}, []string{"cqc_3"}},
		{"09_filter_status_pending_input", closing.Filter{Statuses: []closing.Status{closing.StatusPendingClientInput}}, closing.Sort{}, []string{"cqc_2"}},
		{"10_filter_signal_and_query", closing.Filter{Signals: []signal.Code{signal.CodeMonoExclusionKCapExceeded}, Query: "alvarez"}, closing.Sort{}, []string{"cqc_1"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ids, err := b.ClosingQuery.IDs(ctx, p, tc.filter, tc.sort)
			if err != nil {
				t.Fatalf("IDs failed: %v", err)
			}
			if len(ids) != len(tc.wantIDs) {
				t.Fatalf("IDs = %v, want %v", ids, tc.wantIDs)
			}
			for i := range ids {
				if ids[i] != tc.wantIDs[i] {
					t.Fatalf("IDs[%d] = %s, want %s (all=%v)", i, ids[i], tc.wantIDs[i], ids)
				}
			}

			pageRes, err := b.ClosingQuery.Page(ctx, p, tc.filter, tc.sort, ports.PageRequest{Page: 1, PageSize: 2})
			if err != nil {
				t.Fatalf("Page failed: %v", err)
			}
			if pageRes.Total != len(tc.wantIDs) {
				t.Fatalf("Page.Total = %d, want %d", pageRes.Total, len(tc.wantIDs))
			}
			if pageRes.Page != 1 || pageRes.PageSize != 2 || pageRes.TotalPages != (len(tc.wantIDs)+1)/2 {
				t.Fatalf("Page meta = (%d, %d, %d), want (1, 2, %d)", pageRes.Page, pageRes.PageSize, pageRes.TotalPages, (len(tc.wantIDs)+1)/2)
			}
			if len(pageRes.Items) > 0 {
				if pageRes.Items[0].ClosingID != tc.wantIDs[0] {
					t.Fatalf("Page.Items[0] = %s, want %s", pageRes.Items[0].ClosingID, tc.wantIDs[0])
				}
				if pageRes.Items[0].Regime == "" {
					t.Fatalf("Page.Items[0].Regime is empty")
				}
				if pageRes.Items[0].Regime == client.RegimeMonotributo && pageRes.Items[0].Category == nil {
					t.Fatalf("Page.Items[0].Category is nil for Monotributo")
				}
			}
		})
	}

	sum, err := b.ClosingQuery.Summary(ctx, p, closing.Filter{})
	if err != nil {
		t.Fatalf("Summary failed: %v", err)
	}
	if sum.Total != 4 || sum.BySeverity["RED"] != 1 || sum.BySeverity["YELLOW"] != 1 || sum.BySeverity["GREEN"] != 2 {
		t.Fatalf("unexpected Summary: %+v", sum)
	}
	if sum.ByStatus == nil || sum.BySignal == nil {
		t.Fatalf("Summary maps must not be nil: ByStatus=%v BySignal=%v", sum.ByStatus, sum.BySignal)
	}
	if sum.ByStatus[closing.StatusReadyForReview] != 3 || sum.ByStatus[closing.StatusPendingClientInput] != 1 {
		t.Fatalf("unexpected Summary.ByStatus: %+v", sum.ByStatus)
	}

	periods, err := b.ClosingQuery.Periods(ctx)
	if err != nil {
		t.Fatalf("Periods failed: %v", err)
	}
	if len(periods) == 0 || periods[0] != p.String() {
		t.Fatalf("unexpected Periods: got %v, want at least %s", periods, p.String())
	}
}

// CompareClosingQueries verifica que dos implementaciones de ClosingQuery devuelvan
// exactamente los mismos resultados (reflect.DeepEqual) en Page y Summary.
func CompareClosingQueries(t *testing.T, memQ, pgQ ports.ClosingQuery, p period.Period) {
	t.Helper()
	ctx := context.Background()

	combos := []struct {
		name   string
		filter closing.Filter
		sort   closing.Sort
		page   ports.PageRequest
	}{
		{"01_default_p1", closing.Filter{}, closing.Sort{By: "severity", Dir: "desc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"02_default_p2", closing.Filter{}, closing.Sort{By: "severity", Dir: "desc"}, ports.PageRequest{Page: 2, PageSize: 50}},
		{"03_severity_asc", closing.Filter{}, closing.Sort{By: "severity", Dir: "asc"}, ports.PageRequest{Page: 1, PageSize: 25}},
		{"04_name_asc", closing.Filter{}, closing.Sort{By: "name", Dir: "asc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"05_name_desc", closing.Filter{}, closing.Sort{By: "name", Dir: "desc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"06_toPay_desc", closing.Filter{}, closing.Sort{By: "toPay", Dir: "desc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"07_pctKCap_desc", closing.Filter{}, closing.Sort{By: "pctKCap", Dir: "desc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"08_filter_red_yellow", closing.Filter{Severities: []signal.Severity{signal.SeverityRed, signal.SeverityYellow}}, closing.Sort{By: "severity", Dir: "desc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"09_filter_regime_RI", closing.Filter{Regime: client.RegimeRI}, closing.Sort{By: "name", Dir: "asc"}, ports.PageRequest{Page: 1, PageSize: 50}},
		{"10_filter_signal_and_query", closing.Filter{Signals: []signal.Code{signal.CodeMonoExclusionKCapExceeded}, Query: "sosa"}, closing.Sort{}, ports.PageRequest{Page: 1, PageSize: 50}},
	}

	for _, tc := range combos {
		t.Run(tc.name, func(t *testing.T) {
			memPage, err := memQ.Page(ctx, p, tc.filter, tc.sort, tc.page)
			if err != nil {
				t.Fatalf("mem Page: %v", err)
			}
			pgPage, err := pgQ.Page(ctx, p, tc.filter, tc.sort, tc.page)
			if err != nil {
				t.Fatalf("pg Page: %v", err)
			}
			if !reflect.DeepEqual(memPage, pgPage) {
				t.Fatalf("Page mismatch\nmem: %+v\npg:  %+v", memPage, pgPage)
			}

			memSum, err := memQ.Summary(ctx, p, tc.filter)
			if err != nil {
				t.Fatalf("mem Summary: %v", err)
			}
			pgSum, err := pgQ.Summary(ctx, p, tc.filter)
			if err != nil {
				t.Fatalf("pg Summary: %v", err)
			}
			if !reflect.DeepEqual(memSum, pgSum) {
				t.Fatalf("Summary mismatch\nmem: %+v\npg:  %+v", memSum, pgSum)
			}
		})
	}

	memPeriods, err := memQ.Periods(ctx)
	if err != nil {
		t.Fatalf("mem Periods: %v", err)
	}
	pgPeriods, err := pgQ.Periods(ctx)
	if err != nil {
		t.Fatalf("pg Periods: %v", err)
	}
	if !reflect.DeepEqual(memPeriods, pgPeriods) {
		t.Fatalf("Periods mismatch\nmem: %v\npg:  %v", memPeriods, pgPeriods)
	}
}

func RunNotificationRepo(t *testing.T, b Bundle) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	_ = b.Clients.Save(ctx, sampleClient("cli_n", "Notif Client", "20-99999999-9", client.RegimeMonotributo, tax.CatA))
	now := time.Date(2026, 11, 1, 10, 0, 0, 0, loc)

	n := notification.Notification{
		ID: "not_1", ClientID: "cli_n", ClosingID: "clo_n1", Kind: notification.KindClosingReady,
		Title: "Cierre listo", Body: "Detalle", Key: "k1", CreatedAt: now,
	}
	if err := b.Notifications.Save(ctx, n); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	exists, err := b.Notifications.ExistsFor(ctx, "cli_n", "clo_n1", notification.KindClosingReady, "k1")
	if err != nil || !exists {
		t.Fatalf("ExistsFor = %v, err=%v", exists, err)
	}
	readAt := now.Add(time.Minute)
	if err := b.Notifications.MarkRead(ctx, "not_1", readAt); err != nil {
		t.Fatalf("MarkRead failed: %v", err)
	}
	got, err := b.Notifications.Get(ctx, "not_1")
	if err != nil || got.ReadAt == nil {
		t.Fatalf("Get = %+v, err=%v", got, err)
	}
}

func RunOutboxRepo(t *testing.T, repo ports.OutboxRepository) {
	t.Helper()
	ctx := context.Background()
	loc := period.Location()
	now := time.Date(2026, 11, 1, 10, 0, 0, 0, loc)

	msg := notification.OutboundMessage{
		ID: "out_1", Channel: notification.ChannelEmail, ClientID: "cli_o",
		Subject: "Subj", Body: "Body", CreatedAt: now,
	}
	if err := repo.Enqueue(ctx, msg); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	pending, err := repo.Pending(ctx, 10)
	if err != nil || len(pending) != 1 {
		t.Fatalf("Pending len = %d, err=%v", len(pending), err)
	}
	sentAt := now.Add(time.Second)
	msg.Attempts = 1
	msg.SentAt = &sentAt
	if err := repo.Update(ctx, msg); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	pending2, err := repo.Pending(ctx, 10)
	if err != nil || len(pending2) != 0 {
		t.Fatalf("Pending after sent = %d, err=%v", len(pending2), err)
	}
}

func RunParametersRepo(t *testing.T, repo ports.ParametersRepository) {
	t.Helper()
	ctx := context.Background()
	p, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	p.KCapWarningBp = 8750
	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	got, err := repo.Get(ctx)
	if err != nil || got.KCapWarningBp != 8750 {
		t.Fatalf("Get updated = %+v, err=%v", got, err)
	}
}

func RunTxManager(t *testing.T, b Bundle, supportsRollback bool) {
	t.Helper()
	ctx := context.Background()

	// Commit happy path
	err := b.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		return b.Clients.Save(ctx, sampleClient("cli_tx_ok", "TX OK", "20-12121212-1", client.RegimeMonotributo, tax.CatA))
	})
	if err != nil {
		t.Fatalf("WithinTx commit failed: %v", err)
	}
	if _, err := b.Clients.Get(ctx, "cli_tx_ok"); err != nil {
		t.Fatalf("expected committed client, got %v", err)
	}

	if !supportsRollback {
		return
	}

	// Rollback on error
	sentinel := errors.New("abort tx")
	err = b.TxManager.WithinTx(ctx, func(ctx context.Context) error {
		if err := b.Clients.Save(ctx, sampleClient("cli_tx_rb", "TX RB", "20-13131313-1", client.RegimeMonotributo, tax.CatA)); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if _, err := b.Clients.Get(ctx, "cli_tx_rb"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after rollback, got %v", err)
	}
}

func RunStatsQuery(t *testing.T, b Bundle) {
	t.Helper()
	if b.StatsQuery == nil {
		return
	}
	ctx := context.Background()
	loc := period.Location()
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 5, 9, 0, 0, 0, loc)
	due20 := time.Date(2026, 11, 20, 0, 0, 0, 0, loc)
	due10 := time.Date(2026, 11, 10, 0, 0, 0, 0, loc)

	// Empty period
	emptyStats, err := b.StatsQuery.PeriodStats(ctx, p)
	if err != nil || emptyStats.Progress.Total != 0 {
		t.Fatalf("empty PeriodStats = %+v, err=%v", emptyStats, err)
	}

	_ = b.Clients.SaveAll(ctx, []client.Client{
		sampleClient("sq_1", "Client A", "20-10000011-1", client.RegimeMonotributo, tax.CatK),
		sampleClient("sq_2", "Client B", "20-20000022-2", client.RegimeMonotributo, tax.CatB),
		sampleClient("sq_3", "Client C", "20-30000033-3", client.RegimeRI, ""),
	})

	c1 := closing.NewClosing("sqc_1", "sq_1", p, true, "h1", signal.SeverityRed,
		[]signal.Signal{{Code: signal.CodeMonoExclusionKCapExceeded, Severity: signal.SeverityRed, Priority: 1}}, "Excluido",
		&tax.MonotributoAssessment{CurrentCategory: tax.CatK, Fee: money.New(500000, money.ARS), FeeDueDate: due20, BilledInPeriodARS: money.New(10000000, money.ARS), BilledInPeriodUSD: money.New(10000, money.USD)},
		nil, closing.Reconciliation{}, now)
	c2 := closing.NewClosing("sqc_2", "sq_2", p, false, "h2", signal.SeverityYellow,
		[]signal.Signal{{Code: signal.CodeMissingInvoices, Severity: signal.SeverityYellow, Priority: 3}}, "Falta factura",
		&tax.MonotributoAssessment{CurrentCategory: tax.CatB, Fee: money.New(200000, money.ARS), FeeDueDate: due20, BilledInPeriodARS: money.New(5000000, money.ARS), BilledInPeriodUSD: money.New(5000, money.USD)},
		nil, closing.Reconciliation{}, now)
	c3 := closing.NewClosing("sqc_3", "sq_3", p, true, "h3", signal.SeverityGreen, nil, "OK",
		nil, &tax.RIAssessment{TotalToPay: money.New(800000, money.ARS), VatDueDate: due20, SelfEmployedDueDate: due10, BilledInPeriodARS: money.New(20000000, money.ARS), BilledInPeriodUSD: money.New(20000, money.USD)},
		closing.Reconciliation{}, now)
	_ = c3.Approve(closing.Approval{ID: "sq_a3", Version: 1, ResultHash: "h3", ApprovedBy: "acc", ApprovedAt: time.Date(2026, 11, 8, 12, 0, 0, 0, loc)})
	_ = c3.Publish(closing.Published{Version: 1, Title: "t", Body: "b", PublishedAt: time.Date(2026, 11, 8, 12, 0, 0, 0, loc)})
	_ = c3.MarkPaid(time.Date(2026, 11, 10, 18, 0, 0, 0, loc))

	for _, c := range []closing.Closing{c1, c2, c3} {
		if err := b.Closings.Save(ctx, c); err != nil {
			t.Fatalf("Save %s: %v", c.ID, err)
		}
	}

	got, err := b.StatsQuery.PeriodStats(ctx, p)
	if err != nil {
		t.Fatalf("PeriodStats failed: %v", err)
	}
	want := closing.ComputeStats([]closing.Closing{c1, c2, c3}, p)

	if got.Progress.Total != want.Progress.Total ||
		got.Progress.ApprovedPct != want.Progress.ApprovedPct ||
		got.Progress.EstimatedReviewMinutes != want.Progress.EstimatedReviewMinutes ||
		got.Totals.BilledUSD.Amount != want.Totals.BilledUSD.Amount ||
		got.Totals.ToPayARS.Amount != want.Totals.ToPayARS.Amount ||
		len(got.Totals.ToPayByDueDate) != len(want.Totals.ToPayByDueDate) ||
		got.Risk.NearKCap != want.Risk.NearKCap ||
		len(got.Risk.BySignal) != len(want.Risk.BySignal) ||
		got.Success != want.Success {
		t.Fatalf("PeriodStats mismatch:\ngot:  %+v\nwant: %+v", got, want)
	}
}
