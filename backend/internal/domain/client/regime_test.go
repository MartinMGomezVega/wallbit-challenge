package client

import (
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestChangeRegime(t *testing.T) {
	loc := period.Location()
	createdAt := time.Date(2025, 10, 1, 0, 0, 0, 0, loc)
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, loc)
	effective := time.Date(2026, 10, 14, 0, 0, 0, 0, loc)

	t.Run("MONO a RI exitoso", func(t *testing.T) {
		c := Client{
			ID:        "cli_01",
			Regime:    RegimeMonotributo,
			CreatedAt: createdAt,
		}

		err := c.ChangeRegime(RegimeRI, effective, "EXCLUSION_K_CAP", "Exclusión por superar tope K", now)
		if err != nil {
			t.Fatalf("ChangeRegime failed: %v", err)
		}

		if c.Regime != RegimeRI {
			t.Errorf("Regime = %v, want RI", c.Regime)
		}
		if c.RIProfile == nil {
			t.Fatal("expected RIProfile to be initialized")
		}
		if len(c.RegimeHistory) != 2 {
			t.Fatalf("expected 2 history items, got %d", len(c.RegimeHistory))
		}

		// Tramo 1: Monotributo cerrado el 13/10/2026
		h1 := c.RegimeHistory[0]
		if h1.Regime != RegimeMonotributo {
			t.Errorf("h1.Regime = %v, want MONOTRIBUTO", h1.Regime)
		}
		if h1.To == nil || h1.To.Day() != 13 || h1.To.Month() != 10 || h1.To.Year() != 2026 {
			t.Errorf("h1.To = %v, want 2026-10-13", h1.To)
		}

		// Tramo 2: RI abierto desde 14/10/2026
		h2 := c.RegimeHistory[1]
		if h2.Regime != RegimeRI {
			t.Errorf("h2.Regime = %v, want RI", h2.Regime)
		}
		if !h2.From.Equal(effective) {
			t.Errorf("h2.From = %v, want %v", h2.From, effective)
		}
		if h2.To != nil {
			t.Errorf("h2.To = %v, want nil", h2.To)
		}
		if h2.Reason != "EXCLUSION_K_CAP" {
			t.Errorf("h2.Reason = %v, want EXCLUSION_K_CAP", h2.Reason)
		}

		// RIFrom()
		riFrom, ok := c.RIFrom()
		if !ok || !riFrom.Equal(effective) {
			t.Errorf("RIFrom() = (%v, %v), want (%v, true)", riFrom, ok, effective)
		}

		// RegimeAt en fechas antes y después
		before := time.Date(2026, 10, 10, 0, 0, 0, 0, loc)
		after := time.Date(2026, 10, 15, 0, 0, 0, 0, loc)
		if c.RegimeAt(before) != RegimeMonotributo {
			t.Errorf("RegimeAt(before) = %v, want MONOTRIBUTO", c.RegimeAt(before))
		}
		if c.RegimeAt(after) != RegimeRI {
			t.Errorf("RegimeAt(after) = %v, want RI", c.RegimeAt(after))
		}
	})

	t.Run("Ya es RI falla con ErrInvalidTransition", func(t *testing.T) {
		c := Client{
			ID:        "cli_02",
			Regime:    RegimeRI,
			CreatedAt: createdAt,
			RegimeHistory: []RegimeHistoryItem{
				{Regime: RegimeRI, From: createdAt},
			},
		}

		err := c.ChangeRegime(RegimeRI, effective, "", "", now)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})

	t.Run("Fecha efectiva futura falla", func(t *testing.T) {
		c := Client{
			ID:        "cli_03",
			Regime:    RegimeMonotributo,
			CreatedAt: createdAt,
		}

		futureEffective := now.AddDate(0, 0, 5)
		err := c.ChangeRegime(RegimeRI, futureEffective, "", "", now)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition for future effective date, got %v", err)
		}
	})

	t.Run("Transición a régimen distinto de RI falla", func(t *testing.T) {
		c := Client{
			ID:        "cli_04",
			Regime:    RegimeMonotributo,
			CreatedAt: createdAt,
		}

		err := c.ChangeRegime(RegimeMonotributo, effective, "", "", now)
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("expected ErrInvalidTransition, got %v", err)
		}
	})
}

