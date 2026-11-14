package fakes

import (
	"context"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestFakes(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	clk := Clock{Current: now}
	if !clk.Now().Equal(now) {
		t.Errorf("Clock.Now() = %v, want %v", clk.Now(), now)
	}
	if clk.Location() == nil {
		t.Error("Clock.Location() should not be nil")
	}

	clkZero := Clock{}
	if clkZero.Location() != time.UTC {
		t.Errorf("Clock{}.Location() = %v, want UTC", clkZero.Location())
	}

	// Movement repo
	mRepo := NewMovementRepository()
	m := movement.Movement{ID: "mov_1", ClientID: "cli_1"}
	_ = mRepo.Save(ctx, m)
	_ = mRepo.SaveAll(ctx, []movement.Movement{{ID: "mov_2", ClientID: "cli_1"}})
	_, _ = mRepo.Get(ctx, "mov_1")
	_, _ = mRepo.Get(ctx, "non_existent")
	_, _ = mRepo.ByClient(ctx, "cli_1")
	_, _ = mRepo.ByClientBetween(ctx, "cli_1", time.Time{}, time.Now())
	list, _ := mRepo.ListAll(ctx)
	if len(list) != 2 {
		t.Errorf("ListAll() = %d, want 2", len(list))
	}

	// Classification repo
	cRepo := NewClassificationRepository()
	c := movement.Classification{MovementID: "mov_1"}
	_ = cRepo.Save(ctx, c)
	_ = cRepo.SaveAll(ctx, []movement.Classification{{MovementID: "mov_2"}})
	_, _, _ = cRepo.Get(ctx, "mov_1")
	_, _, _ = cRepo.Get(ctx, "non_existent")
	byIDs, _ := cRepo.ByMovementIDs(ctx, []string{"mov_1", "mov_2"})
	if len(byIDs) != 2 {
		t.Errorf("ByMovementIDs() = %d, want 2", len(byIDs))
	}

	// IDGenerator
	gen := NewIDGenerator()
	id1 := gen.New("test")
	id2 := gen.New("test")
	if id1 == id2 {
		t.Errorf("expected distinct IDs, got %s and %s", id1, id2)
	}

	// Client repo
	cliRepo := NewClientRepository()
	_ = cliRepo.Save(ctx, client.Client{ID: "cli_1", Name: "Client 1"})
	_ = cliRepo.SaveAll(ctx, []client.Client{{ID: "cli_2", Name: "Client 2"}})
	if c, err := cliRepo.Get(ctx, "cli_1"); err != nil || c.ID != "cli_1" {
		t.Errorf("cliRepo.Get failed: %v", err)
	}
	if _, err := cliRepo.Get(ctx, "non_existent"); err == nil {
		t.Error("cliRepo.Get non_existent should fail")
	}
	allClients, _ := cliRepo.List(ctx)
	if len(allClients) != 2 {
		t.Errorf("cliRepo.List() = %d, want 2", len(allClients))
	}

	// Invoice repo
	invRepo := NewInvoiceRepository()
	inv1 := invoice.Invoice{
		ID:          "inv_1",
		ClientID:    "cli_1",
		Type:        invoice.TypeE,
		PointOfSale: 1,
		Number:      10,
		IssueDate:   now,
	}
	_ = invRepo.Save(ctx, inv1)
	_ = invRepo.SaveAll(ctx, []invoice.Invoice{{
		ID:          "inv_2",
		ClientID:    "cli_1",
		Type:        invoice.TypeE,
		PointOfSale: 1,
		Number:      11,
		IssueDate:   now.AddDate(0, 0, 1),
	}})
	if _, err := invRepo.Get(ctx, "inv_1"); err != nil {
		t.Errorf("invRepo.Get failed: %v", err)
	}
	if _, err := invRepo.Get(ctx, "non_existent"); err == nil {
		t.Error("invRepo.Get non_existent should fail")
	}
	exists, _ := invRepo.Exists(ctx, "cli_1", inv1.Key())
	if !exists {
		t.Error("invRepo.Exists should be true")
	}
	existsFalse, _ := invRepo.Exists(ctx, "cli_1", invoice.Key{Type: invoice.TypeE, PointOfSale: 1, Number: 99})
	if existsFalse {
		t.Error("invRepo.Exists should be false")
	}
	invs, _ := invRepo.ByClient(ctx, "cli_1")
	if len(invs) != 2 {
		t.Errorf("invRepo.ByClient = %d, want 2", len(invs))
	}
	invsBet, _ := invRepo.ByClientBetween(ctx, "cli_1", now.AddDate(0, 0, -1), now.AddDate(0, 0, 2))
	if len(invsBet) != 2 {
		t.Errorf("invRepo.ByClientBetween = %d, want 2", len(invsBet))
	}

	// Purchase invoice repo
	piRepo := NewPurchaseInvoiceRepository()
	pi1 := invoice.PurchaseInvoice{
		ID:        "pi_1",
		ClientID:  "cli_1",
		IssueDate: now,
	}
	_ = piRepo.Save(ctx, pi1)
	_ = piRepo.SaveAll(ctx, []invoice.PurchaseInvoice{{
		ID:        "pi_2",
		ClientID:  "cli_1",
		IssueDate: now.AddDate(0, 0, 1),
	}})
	pis, _ := piRepo.ByClient(ctx, "cli_1")
	if len(pis) != 2 {
		t.Errorf("piRepo.ByClient = %d, want 2", len(pis))
	}
	pisBet, _ := piRepo.ByClientBetween(ctx, "cli_1", now.AddDate(0, 0, -1), now.AddDate(0, 0, 2))
	if len(pisBet) != 2 {
		t.Errorf("piRepo.ByClientBetween = %d, want 2", len(pisBet))
	}

	// Client input repo
	ciRepo := NewClientInputRepository()
	p := period.Period{Year: 2026, Month: 10}
	ci1 := clientinput.ClientPeriodInput{
		ClientID: "cli_1",
		Period:   p,
	}
	_ = ciRepo.Save(ctx, ci1)
	_ = ciRepo.SaveAll(ctx, []clientinput.ClientPeriodInput{{
		ClientID: "cli_2",
		Period:   p,
	}})
	if ci, err := ciRepo.Get(ctx, "cli_1", p); err != nil || ci.ClientID != "cli_1" {
		t.Errorf("ciRepo.Get failed: %v", err)
	}
	if _, err := ciRepo.Get(ctx, "cli_none", p); err == nil {
		t.Errorf("ciRepo.Get non-existent should fail")
	}

	// LatestExclusionDateReader
	exclReader := NewLatestExclusionDateReader()
	exclDate := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	exclReader.Dates["cli_1"] = &exclDate
	d, err := exclReader.LatestExclusionDate(ctx, "cli_1")
	if err != nil || d == nil || !d.Equal(exclDate) {
		t.Errorf("LatestExclusionDate = (%v, %v), want %v", d, err, exclDate)
	}
	dNone, err := exclReader.LatestExclusionDate(ctx, "cli_none")
	if err != nil || dNone != nil {
		t.Errorf("LatestExclusionDate non-existent = (%v, %v), want nil", dNone, err)
	}
}
