package mock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestFixtureIntegrity(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatalf("New() falló: %v", err)
	}

	if len(c.quotes) == 0 {
		t.Fatal("quotes vacío")
	}

	// 1. Ordenado y sin duplicados, buy < sell
	seen := make(map[string]bool)
	var prev time.Time
	for _, q := range c.quotes {
		dayStr := q.Day.Format("2006-01-02")
		if seen[dayStr] {
			t.Fatalf("duplicado en fixture: %s", dayStr)
		}
		seen[dayStr] = true

		if !prev.IsZero() && !q.Day.After(prev) {
			t.Fatalf("fixture no ordenado: %s después de %s", dayStr, prev.Format("2006-01-02"))
		}
		prev = q.Day

		if q.Buy >= q.Sell {
			t.Fatalf("en %s: buy (%d) >= sell (%d)", dayStr, q.Buy, q.Sell)
		}
	}

	// 2. 2026-10-01 = 1486.11 / 1539.87
	qOct1, ok := c.byDay["2026-10-01"]
	if !ok {
		t.Fatal("falta 2026-10-01")
	}
	if qOct1.Buy != 14861100 || qOct1.Sell != 15398700 {
		t.Fatalf("2026-10-01 = buy %d, sell %d", qOct1.Buy, qOct1.Sell)
	}
}

func TestUSDQuoteFallback(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// 1. Feriado 2026-10-12 (Día del Respeto a la Diversidad Cultural) -> cotización del 2026-10-09
	feriado := time.Date(2026, 10, 12, 14, 0, 0, 0, period.Location())
	qFeriado, err := c.USDQuote(ctx, feriado)
	if err != nil {
		t.Fatalf("USDQuote feriado: %v", err)
	}
	qViernes, err := c.USDQuote(ctx, time.Date(2026, 10, 9, 0, 0, 0, 0, period.Location()))
	if err != nil {
		t.Fatal(err)
	}
	if qFeriado.Day.Format("2006-01-02") != "2026-10-09" || qFeriado.Buy != qViernes.Buy {
		t.Fatalf("Feriado 12/10 debió devolver 09/10, got %s", qFeriado.Day.Format("2006-01-02"))
	}

	// 2. Domingo 2026-10-04 -> cotización del viernes 2026-10-02
	domingo := time.Date(2026, 10, 4, 18, 0, 0, 0, period.Location())
	qDom, err := c.USDQuote(ctx, domingo)
	if err != nil {
		t.Fatalf("USDQuote domingo: %v", err)
	}
	if qDom.Day.Format("2006-01-02") != "2026-10-02" {
		t.Fatalf("Domingo 04/10 debió devolver 02/10, got %s", qDom.Day.Format("2006-01-02"))
	}

	// 3. Fecha muy lejana (año 2020) -> ErrRateNotFound
	lejos := time.Date(2020, 1, 1, 0, 0, 0, 0, period.Location())
	_, err = c.USDQuote(ctx, lejos)
	if !errors.Is(err, ErrRateNotFound) {
		t.Fatalf("esperaba ErrRateNotFound para 2020, got %v", err)
	}
}

func TestClientOptions(t *testing.T) {
	c, err := New(WithFailEvery(2))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	day := time.Date(2026, 10, 1, 0, 0, 0, 0, period.Location())

	if _, err := c.USDQuote(ctx, day); err != nil {
		t.Fatal(err)
	}
	if _, err := c.USDQuote(ctx, day); !errors.Is(err, ErrSimulatedFailure) {
		t.Fatalf("esperaba ErrSimulatedFailure, got %v", err)
	}

	cLat, _ := New(WithLatency(50 * time.Millisecond))
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := cLat.USDQuote(cancelCtx, day); !errors.Is(err, context.Canceled) {
		t.Fatalf("esperaba context.Canceled, got %v", err)
	}
}
