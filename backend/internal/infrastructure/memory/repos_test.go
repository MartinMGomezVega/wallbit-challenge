package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestClientRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewClientRepository()

	c1 := client.Client{ID: "cli_1", Name: "Client One", Regime: client.RegimeMonotributo, MonotributoCategory: tax.CatA}
	c2 := client.Client{ID: "cli_2", Name: "Client Two", Regime: client.RegimeRI}

	if err := repo.Save(ctx, c1); err != nil {
		t.Fatalf("Save c1 failed: %v", err)
	}
	if err := repo.SaveAll(ctx, []client.Client{c2}); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	got, err := repo.Get(ctx, "cli_1")
	if err != nil || got.ID != "cli_1" {
		t.Fatalf("Get cli_1 failed: %v", err)
	}

	_, err = repo.Get(ctx, "non_existent")
	if err != domain.ErrNotFound {
		t.Fatalf("Get non_existent want ErrNotFound, got %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("List want 2 clients, got %d", len(list))
	}
	if list[0].ID != "cli_1" || list[1].ID != "cli_2" {
		t.Fatalf("List order invalid: %+v", list)
	}
}

func TestMovementRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewMovementRepository()

	t1 := time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)

	m1 := movement.Movement{
		ID:         "mov_1",
		ClientID:   "cli_1",
		Type:       movement.TypeInboundTransfer,
		Amount:     money.New(100000, money.USD),
		OccurredAt: t2,
	}
	m2 := movement.Movement{
		ID:         "mov_2",
		ClientID:   "cli_1",
		Type:       movement.TypeCardSpend,
		Amount:     money.New(5000, money.USD),
		OccurredAt: t1,
	}
	m3 := movement.Movement{
		ID:         "mov_3",
		ClientID:   "cli_2",
		Type:       movement.TypeInboundTransfer,
		Amount:     money.New(200000, money.USD),
		OccurredAt: t3,
	}

	if err := repo.Save(ctx, m1); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if err := repo.SaveAll(ctx, []movement.Movement{m2, m3}); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	got, err := repo.Get(ctx, "mov_1")
	if err != nil || got.ID != "mov_1" {
		t.Fatalf("Get mov_1 failed: %v", err)
	}

	_, err = repo.Get(ctx, "mov_999")
	if err != domain.ErrNotFound {
		t.Fatalf("Get mov_999 want ErrNotFound, got %v", err)
	}

	// ByClient sorts ascending by OccurredAt (m2 then m1)
	byCli1, err := repo.ByClient(ctx, "cli_1")
	if err != nil || len(byCli1) != 2 {
		t.Fatalf("ByClient cli_1 want 2, got %d", len(byCli1))
	}
	if byCli1[0].ID != "mov_2" || byCli1[1].ID != "mov_1" {
		t.Fatalf("ByClient order invalid: %+v", byCli1)
	}

	// ByClientBetween
	rangeMovements, err := repo.ByClientBetween(ctx, "cli_1", t1, t1.Add(time.Hour))
	if err != nil || len(rangeMovements) != 1 || rangeMovements[0].ID != "mov_2" {
		t.Fatalf("ByClientBetween want mov_2, got %+v", rangeMovements)
	}

	// ListAll
	all, err := repo.ListAll(ctx)
	if err != nil || len(all) != 3 {
		t.Fatalf("ListAll want 3 movements, got %d", len(all))
	}
}

func TestClassificationRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewClassificationRepository()

	c1 := movement.Classification{
		MovementID: "mov_1",
		Kind:       movement.KindIncome,
		Confidence: movement.ConfidenceHigh,
		Rule:       movement.RuleForeignCompany,
	}
	c2 := movement.Classification{
		MovementID: "mov_2",
		Kind:       movement.KindNotIncome,
		Confidence: movement.ConfidenceHigh,
		Rule:       movement.RuleCardExpense,
	}

	if err := repo.Save(ctx, c1); err != nil {
		t.Fatalf("Save c1 failed: %v", err)
	}
	if err := repo.SaveAll(ctx, []movement.Classification{c2}); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	got, found, err := repo.Get(ctx, "mov_1")
	if err != nil || !found || got.Rule != movement.RuleForeignCompany {
		t.Fatalf("Get mov_1 failed: %v, %v, %+v", err, found, got)
	}

	_, found, err = repo.Get(ctx, "mov_none")
	if err != nil || found {
		t.Fatalf("Get mov_none should not be found: %v, %v", err, found)
	}

	byIDs, err := repo.ByMovementIDs(ctx, []string{"mov_1", "mov_2", "mov_missing"})
	if err != nil || len(byIDs) != 2 {
		t.Fatalf("ByMovementIDs want 2 results, got %d", len(byIDs))
	}
}

func TestInvoiceRepositories(t *testing.T) {
	ctx := context.Background()
	invRepo := NewInvoiceRepository()
	piRepo := NewPurchaseInvoiceRepository()
	ciRepo := NewClientInputRepository()

	d1 := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	// InvoiceRepo
	inv1 := invoice.Invoice{
		ID:          "inv_1",
		ClientID:    "cli_1",
		Type:        invoice.TypeE,
		PointOfSale: 1,
		Number:      100,
		IssueDate:   d1,
		TotalAmount: money.New(100000, money.USD),
	}
	inv2 := invoice.Invoice{
		ID:          "inv_2",
		ClientID:    "cli_1",
		Type:        invoice.TypeE,
		PointOfSale: 1,
		Number:      101,
		IssueDate:   d2,
		TotalAmount: money.New(200000, money.USD),
	}

	_ = invRepo.Save(ctx, inv1)
	_ = invRepo.SaveAll(ctx, []invoice.Invoice{inv2})

	gotInv, err := invRepo.Get(ctx, "inv_1")
	if err != nil || gotInv.ID != "inv_1" {
		t.Fatalf("Get invoice failed: %v", err)
	}

	_, err = invRepo.Get(ctx, "inv_missing")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, err := invRepo.Exists(ctx, "cli_1", inv1.Key())
	if err != nil || !exists {
		t.Fatalf("Exists want true, got %v %v", exists, err)
	}
	existsMissing, _ := invRepo.Exists(ctx, "cli_1", invoice.Key{Type: invoice.TypeA, PointOfSale: 99, Number: 99})
	if existsMissing {
		t.Fatal("Exists want false")
	}

	byCli, err := invRepo.ByClient(ctx, "cli_1")
	if err != nil || len(byCli) != 2 {
		t.Fatalf("ByClient want 2, got %d", len(byCli))
	}
	between, err := invRepo.ByClientBetween(ctx, "cli_1", d1, d1.Add(time.Hour))
	if err != nil || len(between) != 1 || between[0].ID != "inv_1" {
		t.Fatalf("ByClientBetween want inv_1, got %+v", between)
	}

	// PurchaseInvoiceRepo
	pi := invoice.PurchaseInvoice{
		ID:           "pi_1",
		ClientID:     "cli_1",
		IssueDate:    d1,
		SupplierName: "Proveedor SRL",
		SupplierCUIT: "30-70000000-1",
		NetAmount:    money.New(5000000, money.ARS),
		VatAmount:    money.New(1050000, money.ARS),
	}
	_ = piRepo.Save(ctx, pi)
	_ = piRepo.SaveAll(ctx, []invoice.PurchaseInvoice{})
	piList, err := piRepo.ByClient(ctx, "cli_1")
	if err != nil || len(piList) != 1 {
		t.Fatalf("ByClient purchase invoices want 1, got %d", len(piList))
	}
	piBetween, err := piRepo.ByClientBetween(ctx, "cli_1", d1, d1.Add(time.Hour))
	if err != nil || len(piBetween) != 1 {
		t.Fatalf("ByClientBetween purchase invoices want 1, got %d", len(piBetween))
	}

	// ClientInputRepo
	p, _ := period.Parse("2026-10")
	input := clientinput.ClientPeriodInput{
		ClientID:                 "cli_1",
		Period:                   p,
		DeclaredNoExternalIncome: true,
	}
	_ = ciRepo.Save(ctx, input)
	_ = ciRepo.SaveAll(ctx, []clientinput.ClientPeriodInput{})
	gotInput, err := ciRepo.Get(ctx, "cli_1", p)
	if err != nil || !gotInput.DeclaredNoExternalIncome {
		t.Fatalf("Get client input failed: %v", err)
	}
	pMissing, _ := period.Parse("2026-09")
	_, err = ciRepo.Get(ctx, "cli_1", pMissing)
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound for missing client input, got %v", err)
	}
}

func TestRepositoriesConcurrencyWithRace(t *testing.T) {
	ctx := context.Background()
	clientRepo := NewClientRepository()
	movementRepo := NewMovementRepository()
	classRepo := NewClassificationRepository()
	invRepo := NewInvoiceRepository()
	piRepo := NewPurchaseInvoiceRepository()
	ciRepo := NewClientInputRepository()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	p, _ := period.Parse("2026-10")

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			cid := fmt.Sprintf("cli_%d", idx%10)
			mid := fmt.Sprintf("mov_%d", idx)
			invid := fmt.Sprintf("inv_%d", idx)
			piid := fmt.Sprintf("pi_%d", idx)

			// Clients
			_ = clientRepo.Save(ctx, client.Client{ID: cid, Name: "Name"})
			_, _ = clientRepo.Get(ctx, cid)
			_, _ = clientRepo.List(ctx)

			// Movements
			m := movement.Movement{
				ID:         mid,
				ClientID:   cid,
				Type:       movement.TypeInboundTransfer,
				Amount:     money.New(int64(100+idx), money.USD),
				OccurredAt: time.Date(2026, 10, 1+(idx%25), 10, 0, 0, 0, time.UTC),
			}
			_ = movementRepo.Save(ctx, m)
			_, _ = movementRepo.Get(ctx, mid)
			_, _ = movementRepo.ByClient(ctx, cid)
			_, _ = movementRepo.ListAll(ctx)

			// Classifications
			c := movement.Classification{
				MovementID: mid,
				Kind:       movement.KindIncome,
				Confidence: movement.ConfidenceHigh,
				Rule:       movement.RuleForeignCompany,
			}
			_ = classRepo.Save(ctx, c)
			_, _, _ = classRepo.Get(ctx, mid)
			_, _ = classRepo.ByMovementIDs(ctx, []string{mid})

			// Invoices
			inv := invoice.Invoice{
				ID:          invid,
				ClientID:    cid,
				Type:        invoice.TypeE,
				PointOfSale: 1,
				Number:      idx + 1,
				IssueDate:   time.Date(2026, 10, 1+(idx%25), 10, 0, 0, 0, time.UTC),
				TotalAmount: money.New(int64(100+idx)*100, money.USD),
			}
			_ = invRepo.Save(ctx, inv)
			_, _ = invRepo.Get(ctx, invid)
			_, _ = invRepo.Exists(ctx, cid, inv.Key())
			_, _ = invRepo.ByClient(ctx, cid)

			// Purchase Invoices
			pi := invoice.PurchaseInvoice{
				ID:        piid,
				ClientID:  cid,
				IssueDate: time.Date(2026, 10, 1+(idx%25), 10, 0, 0, 0, time.UTC),
				NetAmount: money.New(100000, money.ARS),
			}
			_ = piRepo.Save(ctx, pi)
			_, _ = piRepo.ByClient(ctx, cid)

			// Client Inputs
			ci := clientinput.ClientPeriodInput{
				ClientID:                 cid,
				Period:                   p,
				DeclaredNoExternalIncome: true,
			}
			_ = ciRepo.Save(ctx, ci)
			_, _ = ciRepo.Get(ctx, cid, p)
		}(i)
	}

	wg.Wait()
}
