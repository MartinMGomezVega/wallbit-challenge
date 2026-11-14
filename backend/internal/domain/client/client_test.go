package client

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestClientRegimeAt(t *testing.T) {
	fromMono := time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location())
	toMono := time.Date(2026, 8, 11, 23, 59, 59, 0, period.Location())
	fromRI := time.Date(2026, 8, 12, 0, 0, 0, 0, period.Location())

	c := Client{
		ID:                  "cli_test",
		Name:                "Agustina Castro",
		Email:               "agustina@test.com",
		CUIT:                "27-35000004-7",
		Regime:              RegimeRI,
		MonotributoCategory: tax.CatK,
		RegimeHistory: []RegimeHistoryItem{
			{Regime: RegimeMonotributo, From: fromMono, To: &toMono, Reason: "INICIO"},
			{Regime: RegimeRI, From: fromRI, To: nil, Reason: "EXCLUSION_K_CAP"},
		},
		RIProfile: &RIProfile{
			SelfEmployedMonthlyContribution: money.New(4500000, money.ARS),
			PreviousVatCredit:               money.Zero(money.ARS),
		},
		CreatedAt: fromMono,
	}

	// 1. En mayo 2026 era Monotributo
	inMay := time.Date(2026, 5, 10, 0, 0, 0, 0, period.Location())
	if got := c.RegimeAt(inMay); got != RegimeMonotributo {
		t.Fatalf("en mayo 2026 = %s, want MONOTRIBUTO", got)
	}

	// 2. En octubre 2026 es RI
	inOct := time.Date(2026, 10, 15, 0, 0, 0, 0, period.Location())
	if got := c.RegimeAt(inOct); got != RegimeRI {
		t.Fatalf("en octubre 2026 = %s, want RI", got)
	}

	// 3. Cliente sin historial usa c.Regime
	cSimple := Client{Regime: RegimeMonotributo}
	if got := cSimple.RegimeAt(inOct); got != RegimeMonotributo {
		t.Fatalf("cliente simple = %s", got)
	}

	// 4. RIFrom
	riDate, ok := c.RIFrom()
	if !ok || !riDate.Equal(fromRI) {
		t.Fatalf("RIFrom = %v %v", riDate, ok)
	}
	cMono := Client{Regime: RegimeMonotributo}
	if _, ok := cMono.RIFrom(); ok {
		t.Fatal("MONO no debe tener RIFrom")
	}

	cRIDirect := Client{Regime: RegimeRI, CreatedAt: fromRI}
	if riDirectDate, ok := cRIDirect.RIFrom(); !ok || !riDirectDate.Equal(fromRI) {
		t.Fatalf("RIFrom directo = %v %v", riDirectDate, ok)
	}
}
