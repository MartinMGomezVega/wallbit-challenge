package closing_test

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestComputeHashDeterminism(t *testing.T) {
	due := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())
	mono := &tax.MonotributoAssessment{
		BilledInPeriodUSD: money.New(100000, money.USD),
		BilledInPeriodARS: money.New(15000000, money.ARS),
		Billed12mARS:      money.New(50000000, money.ARS),
		CurrentCategory:   tax.CatC,
		Fee:               money.New(6602000, money.ARS),
		FeeDueDate:        due,
	}

	sig1 := []signal.Signal{
		{
			Code:     signal.CodeMonoNearCategoryCap,
			Severity: signal.SeverityYellow,
			Priority: 7,
			Message:  "Cerca del tope",
			Data: map[string]string{
				"alpha": "1",
				"beta":  "2",
				"gamma": "3",
			},
		},
	}

	sig2 := []signal.Signal{
		{
			Code:     signal.CodeMonoNearCategoryCap,
			Severity: signal.SeverityYellow,
			Priority: 7,
			Message:  "Cerca del tope",
			Data: map[string]string{
				"gamma": "3",
				"alpha": "1",
				"beta":  "2",
			},
		},
	}

	h1 := closing.ComputeHash(mono, nil, sig1)
	h2 := closing.ComputeHash(mono, nil, sig2)

	if h1 != h2 {
		t.Fatalf("hashes deben ser idénticos a pesar de distinto orden de inserción en map: h1=%s, h2=%s", h1, h2)
	}

	if len(h1) != 64 {
		t.Fatalf("longitud de hash SHA-256 en hex debe ser 64, obtuvo %d", len(h1))
	}
}

func TestComputeHashOneCentDifference(t *testing.T) {
	due := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())
	mono1 := &tax.MonotributoAssessment{
		BilledInPeriodUSD: money.New(100000, money.USD),
		BilledInPeriodARS: money.New(15000000, money.ARS),
		Billed12mARS:      money.New(50000000, money.ARS),
		CurrentCategory:   tax.CatC,
		Fee:               money.New(6602000, money.ARS),
		FeeDueDate:        due,
	}

	// 1 centavo más en BilledInPeriodARS
	mono2 := &tax.MonotributoAssessment{
		BilledInPeriodUSD: money.New(100000, money.USD),
		BilledInPeriodARS: money.New(15000001, money.ARS),
		Billed12mARS:      money.New(50000000, money.ARS),
		CurrentCategory:   tax.CatC,
		Fee:               money.New(6602000, money.ARS),
		FeeDueDate:        due,
	}

	h1 := closing.ComputeHash(mono1, nil, nil)
	h2 := closing.ComputeHash(mono2, nil, nil)

	if h1 == h2 {
		t.Fatalf("un centavo de diferencia debe producir hash distinto: %s", h1)
	}
}

func TestComputeHashRI(t *testing.T) {
	due := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())
	ri := &tax.RIAssessment{
		BilledInPeriodUSD:        money.New(200000, money.USD),
		BilledInPeriodARS:        money.New(30000000, money.ARS),
		OutputVat:                money.New(6300000, money.ARS),
		VatToPay:                 money.New(5000000, money.ARS),
		VatDueDate:               due,
		SelfEmployedContribution: money.New(10000000, money.ARS),
		SelfEmployedDueDate:      due,
		TotalToPay:               money.New(15000000, money.ARS),
	}

	h1 := closing.ComputeHash(nil, ri, nil)

	// Mismo RI -> mismo hash
	h2 := closing.ComputeHash(nil, ri, nil)
	if h1 != h2 {
		t.Fatalf("h1 y h2 deben coincidir: %s vs %s", h1, h2)
	}

	// 1 centavo más en TotalToPay
	ri2 := *ri
	ri2.TotalToPay = money.New(15000001, money.ARS)
	h3 := closing.ComputeHash(nil, &ri2, nil)
	if h1 == h3 {
		t.Fatalf("cambio de 1 centavo en RI debe cambiar hash: %s vs %s", h1, h3)
	}
}
