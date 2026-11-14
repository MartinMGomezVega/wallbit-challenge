package movement

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func TestMovementAmountARS(t *testing.T) {
	m := Movement{
		ID:         "mov_01",
		ClientID:   "cli_01",
		Type:       TypeInboundTransfer,
		Amount:     money.New(250000, money.USD), // 2500.00 USD
		FXRate:     money.Rate(15000000),         // 1500.0000 ARS/USD
		OccurredAt: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC),
		Counterparty: Counterparty{
			Name:    "Northwind Labs LLC",
			Country: "US",
			Kind:    CounterpartyCompany,
		},
		Description: "Payment for consulting services",
	}

	ars, err := m.AmountARS()
	if err != nil {
		t.Fatalf("AmountARS unexpected error: %v", err)
	}

	// 2500 * 1500 = 3,750,000 ARS -> 375000000 cents
	expected := money.New(375000000, money.ARS)
	if ars != expected {
		t.Fatalf("AmountARS = %v, want %v", ars, expected)
	}

	// Error when movement currency is not USD
	mBad := m
	mBad.Amount = money.New(10000, money.ARS)
	if _, err := mBad.AmountARS(); err == nil {
		t.Fatal("AmountARS should fail if currency is not USD")
	}
}
