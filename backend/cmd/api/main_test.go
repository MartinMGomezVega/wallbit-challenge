package main

import (
	"fmt"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
)

func TestClassifyMovementsWorkerPool(t *testing.T) {
	movements := []movement.Movement{
		{
			ID:   "mov_1",
			Type: movement.TypeCardSpend,
		},
		{
			ID:   "mov_2",
			Type: movement.TypeInboundTransfer,
			Counterparty: movement.Counterparty{
				Name:    "Northwind Labs LLC",
				Country: "US",
				Kind:    movement.CounterpartyCompany,
			},
		},
		{
			ID:   "mov_3",
			Type: movement.TypeStablecoinBuy,
		},
	}

	// Test with workers <= 0 (fallback to default 4)
	classifications := classifyMovementsWorkerPool(movements, 0)
	if len(classifications) != 3 {
		t.Fatalf("expected 3 classifications, got %d", len(classifications))
	}
	if classifications[0].Rule != movement.RuleCardExpense {
		t.Errorf("class 0 rule = %s", classifications[0].Rule)
	}
	if classifications[1].Rule != movement.RuleForeignCompany {
		t.Errorf("class 1 rule = %s", classifications[1].Rule)
	}
	if classifications[2].Rule != movement.RuleStablecoinConversion {
		t.Errorf("class 2 rule = %s", classifications[2].Rule)
	}

	// Test with empty slice
	empty := classifyMovementsWorkerPool(nil, 2)
	if len(empty) != 0 {
		t.Fatalf("expected 0 classifications for empty input, got %d", len(empty))
	}

	// Test with larger slice
	many := make([]movement.Movement, 100)
	for i := 0; i < 100; i++ {
		many[i] = movement.Movement{
			ID:   fmt.Sprintf("mov_%d", i),
			Type: movement.TypeFee,
		}
	}
	resMany := classifyMovementsWorkerPool(many, 4)
	if len(resMany) != 100 {
		t.Fatalf("expected 100 classifications, got %d", len(resMany))
	}
	for _, c := range resMany {
		if c.Rule != movement.RuleOutbound {
			t.Errorf("expected OUTBOUND for fee, got %s", c.Rule)
		}
	}
}
