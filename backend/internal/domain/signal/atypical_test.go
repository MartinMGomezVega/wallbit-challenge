package signal

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
)

func TestMedian(t *testing.T) {
	// Empty
	if m := Median(nil); m.Amount != 0 {
		t.Errorf("Median(nil) = %v, want 0", m)
	}

	// Odd
	odd := []money.Money{
		money.New(300, money.USD),
		money.New(100, money.USD),
		money.New(200, money.USD),
	}
	if m := Median(odd); m.Amount != 200 {
		t.Errorf("Median(odd) = %v, want 200", m)
	}

	// Even
	even := []money.Money{
		money.New(100, money.USD),
		money.New(400, money.USD),
		money.New(200, money.USD),
		money.New(300, money.USD),
	}
	if m := Median(even); m.Amount != 250 {
		t.Errorf("Median(even) = %v, want 250", m)
	}
}

func TestCheckAtypicalIncome(t *testing.T) {
	params := parameters.Defaults()
	params.AtypicalMultiplierBp = 25000 // 2.5x
	params.AtypicalWindowMonths = 6

	// Menos de 3 históricos -> no dispara
	res, ok := checkAtypicalIncome(Input{
		HistoricIncomeUSD: []money.Money{
			money.New(100000, money.USD),
			money.New(100000, money.USD),
		},
		PeriodMovements: []ClassifiedMovement{
			{
				Movement: movement.Movement{
					Amount: money.New(500000, money.USD),
				},
				Classification: movement.Classification{
					Kind: movement.KindIncome,
				},
			},
		},
		Params: params,
	})
	if ok {
		t.Errorf("expected no signal when < 3 historical items, got %+v", res)
	}

	// 3 históricos con mediana 1000 USD, cobro de 2500 USD (borde: = umbral 2.5x no dispara)
	resEdge, okEdge := checkAtypicalIncome(Input{
		HistoricIncomeUSD: []money.Money{
			money.New(100000, money.USD),
			money.New(100000, money.USD),
			money.New(100000, money.USD),
		},
		PeriodMovements: []ClassifiedMovement{
			{
				Movement: movement.Movement{
					Amount: money.New(250000, money.USD), // 2.5x exactamente -> no es >
				},
				Classification: movement.Classification{
					Kind: movement.KindIncome,
				},
			},
		},
		Params: params,
	})
	if okEdge {
		t.Errorf("expected no signal when amount == threshold, got %+v", resEdge)
	}

	// Cobro de 3100 USD (3.1x) -> dispara
	resAtypical, okAtypical := checkAtypicalIncome(Input{
		HistoricIncomeUSD: []money.Money{
			money.New(100000, money.USD),
			money.New(100000, money.USD),
			money.New(100000, money.USD),
		},
		PeriodMovements: []ClassifiedMovement{
			{
				Movement: movement.Movement{
					Amount:     money.New(310000, money.USD),
					OccurredAt: time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC),
				},
				Classification: movement.Classification{
					Kind: movement.KindIncome,
				},
			},
		},
		Params: params,
	})
	if !okAtypical {
		t.Fatal("expected atypical signal to trigger")
	}
	if resAtypical.Code != CodeAtypicalIncome {
		t.Errorf("code = %s, want ATYPICAL_INCOME", resAtypical.Code)
	}
	expectedMsg := "Cobro de US$ 3.100,00, 3,1× su mediana de 6 meses."
	if resAtypical.Message != expectedMsg {
		t.Errorf("Message = %s, want %s", resAtypical.Message, expectedMsg)
	}
}
