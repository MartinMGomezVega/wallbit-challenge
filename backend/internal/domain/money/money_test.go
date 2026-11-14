package money

import (
	"errors"
	"math/big"
	"testing"
)

func TestMoneyBasic(t *testing.T) {
	m1 := New(10050, ARS) // $100.50
	m2 := New(20025, ARS) // $200.25

	if !m1.IsPositive() || m1.IsZero() {
		t.Fatal("esperaba positivo")
	}
	if !Zero(ARS).IsZero() {
		t.Fatal("esperaba zero")
	}

	sum, err := m1.Add(m2)
	if err != nil || sum.Amount != 30075 || sum.Currency != ARS {
		t.Fatalf("Add falló: %v, %v", sum, err)
	}

	diff, err := m2.Sub(m1)
	if err != nil || diff.Amount != 9975 || diff.Currency != ARS {
		t.Fatalf("Sub falló: %v, %v", diff, err)
	}

	mUSD := New(5000, USD)
	if _, err := m1.Add(mUSD); !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("esperaba ErrCurrencyMismatch, got %v", err)
	}
	if _, err := m1.Sub(mUSD); !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("esperaba ErrCurrencyMismatch, got %v", err)
	}
}

func TestRate(t *testing.T) {
	r, err := RateFromDecimal("1539.87")
	if err != nil || r != 15398700 {
		t.Fatalf("RateFromDecimal falló: %v %v", r, err)
	}
	if r.DecimalString() != "1539.8700" {
		t.Fatalf("DecimalString = %s", r.DecimalString())
	}

	rComma, err := RateFromDecimal("1539,87")
	if err != nil || rComma != 15398700 {
		t.Fatalf("RateFromDecimal con coma falló: %v", rComma)
	}

	rInt, err := RateFromDecimal("1500")
	if err != nil || rInt != 15000000 {
		t.Fatalf("RateFromDecimal entero falló: %v", rInt)
	}

	rNeg, err := RateFromDecimal("-15.5")
	if err != nil || rNeg != -155000 {
		t.Fatalf("RateFromDecimal negativo: %v %v", rNeg, err)
	}
	if rNeg.DecimalString() != "-15.5000" {
		t.Fatalf("DecimalString negativo = %s", rNeg.DecimalString())
	}

	if _, err := RateFromDecimal("abc"); err == nil {
		t.Fatal("esperaba error")
	}
	if _, err := RateFromDecimal("1.2.3"); err == nil {
		t.Fatal("esperaba error")
	}
}

func TestParseAndFormat(t *testing.T) {
	cases := []struct {
		dec   string
		cents int64
	}{
		{"1234.56", 123456},
		{"0.05", 5},
		{"100", 10000},
		{"-50.25", -5025},
		{"0.00", 0},
	}
	for _, c := range cases {
		m, err := ParseDecimal(c.dec, ARS)
		if err != nil {
			t.Fatalf("ParseDecimal(%s) falló: %v", c.dec, err)
		}
		if m.Amount != c.cents {
			t.Fatalf("ParseDecimal(%s) = %d, want %d", c.dec, m.Amount, c.cents)
		}
		if s := m.DecimalString(); s != c.dec && (c.dec != "100" || s != "100.00") {
			t.Fatalf("DecimalString() = %s, want %s", s, c.dec)
		}
	}

	// ParseLocalized
	mLoc, err := ParseLocalized("1.234.567,89", ARS)
	if err != nil || mLoc.Amount != 123456789 {
		t.Fatalf("ParseLocalized con puntos de miles falló: %v %v", mLoc, err)
	}

	mLocSimple, err := ParseLocalized("1234,56", USD)
	if err != nil || mLocSimple.Amount != 123456 || mLocSimple.Currency != USD {
		t.Fatalf("ParseLocalized simple falló: %v", mLocSimple)
	}

	if _, err := ParseDecimal("", ARS); err == nil {
		t.Fatal("esperaba error con string vacío")
	}
	if _, err := ParseDecimal("abc", ARS); err == nil {
		t.Fatal("esperaba error")
	}
	if _, err := ParseDecimal("1.2.3", ARS); err == nil {
		t.Fatal("esperaba error")
	}
}

func TestRoundHalfEven(t *testing.T) {
	// Casos clásicos half-even:
	// 1 / 2 = 0.5 -> 0 (par)
	// 3 / 2 = 1.5 -> 2 (par)
	// 5 / 2 = 2.5 -> 2 (par)
	// 7 / 2 = 3.5 -> 4 (par)
	// -1 / 2 = -0.5 -> 0
	// -3 / 2 = -1.5 -> -2
	check := func(num, den int64, want int64) {
		t.Helper()
		got := RoundHalfEven(big.NewInt(num), big.NewInt(den))
		if got != want {
			t.Fatalf("RoundHalfEven(%d, %d) = %d, want %d", num, den, got, want)
		}
	}

	check(1, 2, 0)
	check(3, 2, 2)
	check(5, 2, 2)
	check(7, 2, 4)
	check(-1, 2, 0)
	check(-3, 2, -2)
	check(-5, 2, -2)

	// No mitades exactas:
	check(14, 10, 1) // 1.4 -> 1
	check(16, 10, 2) // 1.6 -> 2
	check(-14, 10, -1)
	check(-16, 10, -2)
}

func TestConvertUSDToARS(t *testing.T) {
	// Rate = 1539.87 -> 15398700
	rate, err := RateFromDecimal("1539.87")
	if err != nil {
		t.Fatal(err)
	}

	// 100 USD (10000 cents): 10000 * 15398700 / 10000 = 15398700 cents ($ 153.987,00)
	usd := New(10000, USD)
	ars, err := ConvertUSDToARS(usd, rate)
	if err != nil {
		t.Fatalf("ConvertUSDToARS falló: %v", err)
	}
	if ars.Amount != 15398700 || ars.Currency != ARS {
		t.Fatalf("ars = %v", ars)
	}

	// Error si no es USD
	if _, err := ConvertUSDToARS(New(100, ARS), rate); err == nil {
		t.Fatal("esperaba error al convertir moneda no USD")
	}

	// Safe con montos grandes de 10^12 centavos sin overflow
	bigUSD := New(1_000_000_000_000, USD)
	bigARS, err := ConvertUSDToARS(bigUSD, rate)
	if err != nil {
		t.Fatalf("big conversion error: %v", err)
	}
	if bigARS.Amount != 1539870000000000 {
		t.Fatalf("bigARS = %d", bigARS.Amount)
	}
}

func TestBasisPoints(t *testing.T) {
	m := New(10000, ARS)    // $100.00
	bp := BasisPoints(8000) // 80%

	applied := ApplyBp(m, bp)
	if applied.Amount != 8000 {
		t.Fatalf("ApplyBp = %d, want 8000", applied.Amount)
	}

	// RatioBp
	r, err := RatioBp(applied, m)
	if err != nil || r != 8000 {
		t.Fatalf("RatioBp = %v, %v", r, err)
	}

	// RatioBp errors
	if _, err := RatioBp(New(10, ARS), New(10, USD)); !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("esperaba ErrCurrencyMismatch, got %v", err)
	}
	if _, err := RatioBp(New(10, ARS), Zero(ARS)); !errors.Is(err, ErrDivisionByZero) {
		t.Fatalf("esperaba ErrDivisionByZero, got %v", err)
	}
}
