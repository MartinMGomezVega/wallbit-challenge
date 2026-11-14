package signal

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestFormatters(t *testing.T) {
	loc := period.Location()

	// formatDate
	t.Run("formatDate", func(t *testing.T) {
		d := time.Date(2026, 10, 14, 15, 30, 0, 0, loc)
		formatted := formatDate(d)
		if formatted != "14/10/2026" {
			t.Errorf("formatDate = %s, want 14/10/2026", formatted)
		}
	})

	// formatARS
	t.Run("formatARS", func(t *testing.T) {
		tests := []struct {
			amount int64
			want   string
		}{
			{0, "$ 0,00"},
			{50, "$ 0,50"},
			{100, "$ 1,00"},
			{123456, "$ 1.234,56"},
			{123456789, "$ 1.234.567,89"},
			{-100000, "-$ 1.000,00"},
		}
		for _, tt := range tests {
			got := formatARS(money.New(tt.amount, money.ARS))
			if got != tt.want {
				t.Errorf("formatARS(%d) = %s, want %s", tt.amount, got, tt.want)
			}
		}
	})

	// formatUSD
	t.Run("formatUSD", func(t *testing.T) {
		tests := []struct {
			amount int64
			want   string
		}{
			{0, "US$ 0,00"},
			{420000, "US$ 4.200,00"},
			{150, "US$ 1,50"},
			{-100000, "-US$ 1.000,00"},
		}
		for _, tt := range tests {
			got := formatUSD(money.New(tt.amount, money.USD))
			if got != tt.want {
				t.Errorf("formatUSD(%d) = %s, want %s", tt.amount, got, tt.want)
			}
		}
	})

	// formatCents
	t.Run("formatCents", func(t *testing.T) {
		if got := formatCents(12345); got != "123.45" {
			t.Errorf("formatCents(12345) = %s, want 123.45", got)
		}
		if got := formatCents(-12345); got != "-123.45" {
			t.Errorf("formatCents(-12345) = %s, want -123.45", got)
		}
	})

	// formatPct
	t.Run("formatPct", func(t *testing.T) {
		if got := formatPct(8700); got != "87%" {
			t.Errorf("formatPct(8700) = %s, want 87%%", got)
		}
		if got := formatPct(8765); got != "87%" {
			t.Errorf("formatPct(8765) = %s, want 87%%", got)
		}
	})

	// formatRatio
	t.Run("formatRatio", func(t *testing.T) {
		if got := formatRatio(310, 100); got != "3,1×" {
			t.Errorf("formatRatio(310, 100) = %s, want 3,1×", got)
		}
		if got := formatRatio(125, 100); got != "1,3×" {
			t.Errorf("formatRatio(125, 100) = %s, want 1,3×", got)
		}
		if got := formatRatio(100, 0); got != "0,0×" {
			t.Errorf("formatRatio(100, 0) = %s, want 0,0×", got)
		}
	})
}
