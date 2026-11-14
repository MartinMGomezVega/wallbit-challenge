package seed

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
)

func hashDataset(t *testing.T, ds Dataset) string {
	t.Helper()
	b, err := json.Marshal(ds)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}

func TestSeedDeterminism(t *testing.T) {
	ctx := context.Background()
	fxClient, err := mock.New()
	if err != nil {
		t.Fatalf("mock.New failed: %v", err)
	}

	cfg := SeedConfig{
		Seed:    42,
		Clients: 50,
		AppNow:  time.Date(2026, 11, 13, 12, 0, 0, 0, period.Location()),
	}

	ds1, err := Generate(ctx, cfg, fxClient, idgen.NewSequential())
	if err != nil {
		t.Fatalf("Generate 1 failed: %v", err)
	}

	ds2, err := Generate(ctx, cfg, fxClient, idgen.NewSequential())
	if err != nil {
		t.Fatalf("Generate 2 failed: %v", err)
	}

	h1 := hashDataset(t, ds1)
	h2 := hashDataset(t, ds2)

	if h1 != h2 {
		t.Fatalf("Seed is not deterministic! Hash 1 = %s, Hash 2 = %s", h1, h2)
	}
	if len(ds1.Clients) != 50 {
		t.Fatalf("expected 50 clients, got %d", len(ds1.Clients))
	}
}

func TestDemoClients(t *testing.T) {
	ctx := context.Background()
	fxClient, err := mock.New()
	if err != nil {
		t.Fatalf("mock.New failed: %v", err)
	}

	demoData, err := BuildDemoClients(ctx, fxClient)
	if err != nil {
		t.Fatalf("BuildDemoClients failed: %v", err)
	}

	if len(demoData.Clients) != 15 {
		t.Fatalf("expected 15 demo clients, got %d", len(demoData.Clients))
	}

	expectedClients := map[string]client.Regime{
		"cli_demo_01": client.RegimeMonotributo,
		"cli_demo_02": client.RegimeMonotributo,
		"cli_demo_03": client.RegimeMonotributo,
		"cli_demo_04": client.RegimeMonotributo,
		"cli_demo_05": client.RegimeMonotributo,
		"cli_demo_06": client.RegimeMonotributo,
		"cli_demo_07": client.RegimeMonotributo,
		"cli_demo_08": client.RegimeMonotributo,
		"cli_demo_09": client.RegimeRI,
		"cli_demo_10": client.RegimeRI,
		"cli_demo_11": client.RegimeRI,
		"cli_demo_12": client.RegimeMonotributo,
		"cli_demo_13": client.RegimeMonotributo,
		"cli_demo_14": client.RegimeMonotributo,
		"cli_demo_15": client.RegimeMonotributo,
	}

	foundMap := make(map[string]bool)
	for _, c := range demoData.Clients {
		expectedRegime, ok := expectedClients[c.ID]
		if !ok {
			t.Fatalf("unexpected demo client ID: %s", c.ID)
		}
		if c.Regime != expectedRegime {
			t.Errorf("client %s has regime %s, want %s", c.ID, c.Regime, expectedRegime)
		}
		foundMap[c.ID] = true

		// Check CUIT valid
		if _, err := client.ParseCUIT(string(c.CUIT)); err != nil {
			t.Errorf("client %s has invalid CUIT %s: %v", c.ID, c.CUIT, err)
		}
	}

	if len(foundMap) != 15 {
		t.Fatalf("not all demo clients found: %d of 15", len(foundMap))
	}

	// Verify movements exist for each demo client
	movCountByClient := make(map[string]int)
	for _, m := range demoData.Movements {
		movCountByClient[m.ClientID]++
		if m.FXRate <= 0 {
			t.Errorf("movement %s has invalid FXRate %v", m.ID, m.FXRate)
		}
	}

	for id := range expectedClients {
		if movCountByClient[id] == 0 {
			t.Errorf("demo client %s has 0 movements", id)
		}
	}
}

func TestSeedDistributionAndCUITs(t *testing.T) {
	ctx := context.Background()
	fxClient, err := mock.New()
	if err != nil {
		t.Fatalf("mock.New failed: %v", err)
	}

	cfg := SeedConfig{
		Seed:    12345,
		Clients: 300,
		AppNow:  time.Date(2026, 11, 13, 12, 0, 0, 0, period.Location()),
	}

	ds, err := Generate(ctx, cfg, fxClient, idgen.NewSequential())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(ds.Clients) != 300 {
		t.Fatalf("expected 300 clients, got %d", len(ds.Clients))
	}

	monoCount := 0
	riCount := 0
	terminations := make(map[int]int)

	for _, c := range ds.Clients {
		if c.Regime == client.RegimeMonotributo {
			monoCount++
		} else if c.Regime == client.RegimeRI {
			riCount++
		}

		cuit, err := client.ParseCUIT(string(c.CUIT))
		if err != nil {
			t.Errorf("invalid CUIT for %s (%s): %v", c.ID, c.CUIT, err)
		}
		terminations[cuit.Terminacion()]++
	}

	monoPct := float64(monoCount) / float64(len(ds.Clients)) * 100
	riPct := float64(riCount) / float64(len(ds.Clients)) * 100

	// 85% Monotributo, 15% RI with ±5 pp tolerance (80-90% mono, 10-20% ri)
	if monoPct < 80 || monoPct > 90 {
		t.Errorf("Monotributo percentage %.2f%% outside [80%%, 90%%]", monoPct)
	}
	if riPct < 10 || riPct > 20 {
		t.Errorf("RI percentage %.2f%% outside [10%%, 20%%]", riPct)
	}

	// Terminations 0..9 all present
	for term := 0; term <= 9; term++ {
		if terminations[term] == 0 {
			t.Errorf("termination %d has 0 clients", term)
		}
	}

	// Movements checks
	if len(ds.Movements) < 3000 {
		t.Errorf("expected at least 3000 movements for 300 clients across 13 months, got %d", len(ds.Movements))
	}

	// Monthly inbound income ranges for clients
	inboundByClient := make(map[string][]int64)
	for _, m := range ds.Movements {
		if m.Type == movement.TypeInboundTransfer {
			inboundByClient[m.ClientID] = append(inboundByClient[m.ClientID], m.Amount.Amount/100)
		}
	}

	// Check that incomes are in 800 - 9000 range with majority between 1500 and 4500
	inMiddleRange := 0
	totalClientsChecked := 0
	for _, amounts := range inboundByClient {
		if len(amounts) == 0 {
			continue
		}
		totalClientsChecked++
		avg := int64(0)
		for _, a := range amounts {
			avg += a
		}
		avg /= int64(len(amounts))
		if avg >= 1500 && avg <= 4500 {
			inMiddleRange++
		}
	}

	midPct := float64(inMiddleRange) / float64(totalClientsChecked) * 100
	if midPct < 50 {
		t.Errorf("expected majority of clients in 1500-4500 range, got %.2f%%", midPct)
	}

	// Invoices checks (T003.9)
	if len(ds.Invoices) < 3000 {
		t.Errorf("expected at least 3000 invoices for 300 clients across 13 months, got %d", len(ds.Invoices))
	}
	for _, inv := range ds.Invoices {
		if inv.TotalAmount.Amount <= 0 {
			t.Errorf("invoice %s has non-positive amount %v", inv.ID, inv.TotalAmount)
		}
	}

	// Purchase invoices checks (T003.9)
	if len(ds.PurchaseInvoices) < 100 {
		t.Errorf("expected at least 100 purchase invoices for RI clients, got %d", len(ds.PurchaseInvoices))
	}

	// Client inputs checks (T003.9)
	if len(ds.ClientInputs) != 300 {
		t.Errorf("expected exactly 300 client inputs for October 2026, got %d", len(ds.ClientInputs))
	}
	confirmedCount := 0
	for _, ci := range ds.ClientInputs {
		if ci.ConfirmedAt != nil {
			confirmedCount++
		}
	}
	confPct := float64(confirmedCount) / float64(len(ds.ClientInputs)) * 100
	if confPct < 80 || confPct > 98 {
		t.Errorf("confirmed client input pct %.2f%% outside [80%%, 98%%]", confPct)
	}
}
