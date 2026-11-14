package invoice

import (
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestAmountARS(t *testing.T) {
	// 1. Factura en ARS
	invARS := Invoice{
		Currency:    money.ARS,
		TotalAmount: money.New(150000000, money.ARS), // 1.5M ARS
	}
	ars, err := AmountARS(invARS, parameters.FXRuleBNABuyPrevBusinessDay, nil)
	if err != nil || ars != invARS.TotalAmount {
		t.Fatalf("Factura en ARS = %v, err %v", ars, err)
	}

	// 2. Factura con TC propio
	ownRate := money.Rate(15200000) // 1520.0000 ARS/USD
	invWithTC := Invoice{
		Currency:    money.USD,
		TotalAmount: money.New(100000, money.USD), // 1000.00 USD
		FXRate:      &ownRate,
	}
	arsOwn, err := AmountARS(invWithTC, parameters.FXRuleBNABuyPrevBusinessDay, nil)
	if err != nil {
		t.Fatalf("Factura con TC propio error: %v", err)
	}
	// 1000 * 1520 = 1,520,000 ARS -> 152000000 cents
	expectedOwn := money.New(152000000, money.ARS)
	if arsOwn != expectedOwn {
		t.Fatalf("Factura con TC propio = %v, want %v", arsOwn, expectedOwn)
	}

	// 3. Reglas BNA con QuoteLookup
	// Emisión: Lunes 2026-10-19
	// Día hábil anterior: Viernes 2026-10-16
	mon19 := time.Date(2026, 10, 19, 10, 0, 0, 0, period.Location())
	fri16 := time.Date(2026, 10, 16, 10, 0, 0, 0, period.Location())

	mockQuotes := map[string]Quote{
		mon19.Format("2006-01-02"): {
			Buy:  money.Rate(14876400), // 1487.64
			Sell: money.Rate(15410900), // 1541.09
		},
		fri16.Format("2006-01-02"): {
			Buy:  money.Rate(14814300), // 1481.43
			Sell: money.Rate(15348000), // 1534.80
		},
	}

	var requestedDay time.Time
	lookup := func(d time.Time) (Quote, error) {
		requestedDay = d
		q, ok := mockQuotes[d.Format("2006-01-02")]
		if !ok {
			return Quote{}, errors.New("quote not found")
		}
		return q, nil
	}

	invUSD := Invoice{
		Currency:    money.USD,
		TotalAmount: money.New(100000, money.USD), // 1000.00 USD
		IssueDate:   mon19,
	}

	// BNA_BUY_PREV_BUSINESS_DAY: viernes 16, compra = 1481.43
	arsBuyPrev, err := AmountARS(invUSD, parameters.FXRuleBNABuyPrevBusinessDay, lookup)
	if err != nil {
		t.Fatalf("BNA_BUY_PREV_BUSINESS_DAY error: %v", err)
	}
	if requestedDay.Format("2006-01-02") != "2026-10-16" {
		t.Errorf("requestedDay = %s, want 2026-10-16", requestedDay.Format("2006-01-02"))
	}
	expectedBuyPrev := money.New(148143000, money.ARS)
	if arsBuyPrev != expectedBuyPrev {
		t.Errorf("BNA_BUY_PREV_BUSINESS_DAY = %v, want %v", arsBuyPrev, expectedBuyPrev)
	}

	// BNA_SELL_PREV_BUSINESS_DAY: viernes 16, venta = 1534.80
	arsSellPrev, err := AmountARS(invUSD, parameters.FXRuleBNASellPrevBusinessDay, lookup)
	if err != nil {
		t.Fatalf("BNA_SELL_PREV_BUSINESS_DAY error: %v", err)
	}
	expectedSellPrev := money.New(153480000, money.ARS)
	if arsSellPrev != expectedSellPrev {
		t.Errorf("BNA_SELL_PREV_BUSINESS_DAY = %v, want %v", arsSellPrev, expectedSellPrev)
	}

	// BNA_BUY_SAME_DAY: lunes 19, compra = 1487.64
	arsSameDay, err := AmountARS(invUSD, parameters.FXRuleBNABuySameDay, lookup)
	if err != nil {
		t.Fatalf("BNA_BUY_SAME_DAY error: %v", err)
	}
	if requestedDay.Format("2006-01-02") != "2026-10-19" {
		t.Errorf("requestedDay = %s, want 2026-10-19", requestedDay.Format("2006-01-02"))
	}
	expectedSameDay := money.New(148764000, money.ARS)
	if arsSameDay != expectedSameDay {
		t.Errorf("BNA_BUY_SAME_DAY = %v, want %v", arsSameDay, expectedSameDay)
	}

	// 4. Errores
	// Moneda inválida
	invEUR := Invoice{Currency: "EUR"}
	if _, err := AmountARS(invEUR, parameters.FXRuleBNABuyPrevBusinessDay, lookup); err == nil {
		t.Fatal("expected error for EUR")
	}

	// Lookup nil para USD sin TC
	if _, err := AmountARS(invUSD, parameters.FXRuleBNABuyPrevBusinessDay, nil); !errors.Is(err, ErrQuoteLookupRequired) {
		t.Fatalf("expected ErrQuoteLookupRequired, got %v", err)
	}

	// Lookup con error
	failLookup := func(d time.Time) (Quote, error) {
		return Quote{}, errors.New("network error")
	}
	if _, err := AmountARS(invUSD, parameters.FXRuleBNABuyPrevBusinessDay, failLookup); err == nil {
		t.Fatal("expected error when lookup fails")
	}
}
