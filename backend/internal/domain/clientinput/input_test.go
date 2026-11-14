package clientinput

import (
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestClientPeriodInput(t *testing.T) {
	p, _ := period.Parse("2026-10")
	now := time.Date(2026, 11, 2, 14, 0, 0, 0, time.UTC)

	// 1. Inicial: Answered = false, Confirm falla
	input := ClientPeriodInput{
		ClientID: "cli_1",
		Period:   p,
	}
	if input.Answered() {
		t.Fatal("expected Answered = false initially")
	}
	err := input.Confirm(now)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation from Confirm when unanswered, got %v", err)
	}

	// 2. Declara que no tuvo ingresos externos
	input.DeclaredNoExternalIncome = true
	if !input.Answered() {
		t.Fatal("expected Answered = true after declaring none")
	}
	if err := input.Confirm(now); err != nil {
		t.Fatalf("Confirm unexpected error: %v", err)
	}
	if input.ConfirmedAt == nil || !input.ConfirmedAt.Equal(now) {
		t.Fatalf("ConfirmedAt mismatch: %v", input.ConfirmedAt)
	}

	// 3. Con ítems de ingresos externos
	inputWithItems := ClientPeriodInput{
		ClientID: "cli_2",
		Period:   p,
		ExternalIncome: []ExternalIncomeItem{
			{Amount: money.New(50000, money.USD), Description: "Consultoría banco local"},
			{Amount: money.New(30000, money.USD), Description: "Clases universitarias"},
		},
	}
	if !inputWithItems.Answered() {
		t.Fatal("expected Answered = true when items present")
	}
	totalUSD := inputWithItems.TotalExternalIncomeUSD()
	expectedTotal := money.New(80000, money.USD)
	if totalUSD != expectedTotal {
		t.Fatalf("TotalExternalIncomeUSD = %v, want %v", totalUSD, expectedTotal)
	}
	if err := inputWithItems.Confirm(now); err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}
}

