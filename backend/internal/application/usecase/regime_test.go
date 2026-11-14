package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestChangeRegimeUseCase(t *testing.T) {
	ctx := context.Background()
	loc := period.Location()
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, loc)
	clk := fakes.Clock{Current: now}
	effective := time.Date(2026, 10, 14, 0, 0, 0, 0, loc)

	t.Run("Exitoso con fecha coincidente", func(t *testing.T) {
		cliRepo := fakes.NewClientRepository()
		_ = cliRepo.Save(ctx, client.Client{
			ID:        "cli_1",
			Regime:    client.RegimeMonotributo,
			CreatedAt: time.Date(2025, 10, 1, 0, 0, 0, 0, loc),
		})

		exclReader := fakes.NewLatestExclusionDateReader()
		exclReader.Dates["cli_1"] = &effective
		notifier := &recordingNotifier{}

		uc := ChangeRegime{
			Clients:    cliRepo,
			Exclusions: exclReader,
			Clock:      clk,
			Notifier:   notifier,
		}

		res, err := uc.Execute(ctx, ChangeRegimeInput{
			ClientID:      "cli_1",
			To:            client.RegimeRI,
			EffectiveDate: effective,
			Note:          "Pase formal por superar tope",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Regime != client.RegimeRI {
			t.Errorf("Regime = %v, want RI", res.Regime)
		}
		if len(notifier.events) != 1 || notifier.events[0].ClientID != "cli_1" {
			t.Errorf("expected 1 ClosingInputsChanged event for cli_1, got %v", notifier.events)
		}
	})

	t.Run("Mismatch de fecha sin force falla con ErrValidation", func(t *testing.T) {
		cliRepo := fakes.NewClientRepository()
		_ = cliRepo.Save(ctx, client.Client{
			ID:        "cli_1",
			Regime:    client.RegimeMonotributo,
			CreatedAt: time.Date(2025, 10, 1, 0, 0, 0, 0, loc),
		})

		exclReader := fakes.NewLatestExclusionDateReader()
		otherDate := time.Date(2026, 10, 20, 0, 0, 0, 0, loc)
		exclReader.Dates["cli_1"] = &otherDate

		uc := ChangeRegime{
			Clients:    cliRepo,
			Exclusions: exclReader,
			Clock:      clk,
		}

		_, err := uc.Execute(ctx, ChangeRegimeInput{
			ClientID:      "cli_1",
			To:            client.RegimeRI,
			EffectiveDate: effective,
			Force:         false,
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("Mismatch de fecha con force = true procede y anota la nota", func(t *testing.T) {
		cliRepo := fakes.NewClientRepository()
		_ = cliRepo.Save(ctx, client.Client{
			ID:        "cli_1",
			Regime:    client.RegimeMonotributo,
			CreatedAt: time.Date(2025, 10, 1, 0, 0, 0, 0, loc),
		})

		exclReader := fakes.NewLatestExclusionDateReader()
		otherDate := time.Date(2026, 10, 20, 0, 0, 0, 0, loc)
		exclReader.Dates["cli_1"] = &otherDate

		uc := ChangeRegime{
			Clients:    cliRepo,
			Exclusions: exclReader,
			Clock:      clk,
		}

		res, err := uc.Execute(ctx, ChangeRegimeInput{
			ClientID:      "cli_1",
			To:            client.RegimeRI,
			EffectiveDate: effective,
			Note:          "Decisión del contador",
			Force:         true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Regime != client.RegimeRI {
			t.Errorf("Regime = %v, want RI", res.Regime)
		}
		lastH := res.RegimeHistory[len(res.RegimeHistory)-1]
		if lastH.Note == "" {
			t.Error("expected note to contain forced tag")
		}
	})

	t.Run("Cliente no encontrado falla con ErrNotFound", func(t *testing.T) {
		cliRepo := fakes.NewClientRepository()
		uc := ChangeRegime{Clients: cliRepo, Clock: clk}

		_, err := uc.Execute(ctx, ChangeRegimeInput{
			ClientID:      "cli_none",
			To:            client.RegimeRI,
			EffectiveDate: effective,
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestUpdateRIProfileUseCase(t *testing.T) {
	ctx := context.Background()
	cliRepo := fakes.NewClientRepository()
	_ = cliRepo.Save(ctx, client.Client{
		ID:     "cli_ri",
		Regime: client.RegimeRI,
	})
	notifier := &recordingNotifier{}

	uc := UpdateRIProfile{
		Clients:  cliRepo,
		Notifier: notifier,
	}

	t.Run("Actualización exitosa", func(t *testing.T) {
		advDueDate := time.Date(2026, 11, 15, 0, 0, 0, 0, period.Location())
		p := period.Period{Year: 2026, Month: 10}

		prof, err := uc.Execute(ctx, UpdateRIProfileInput{
			ClientID:                        "cli_ri",
			SelfEmployedMonthlyContribution: money.New(5000000, money.ARS),
			PreviousVatCredit:               money.New(1200000, money.ARS),
			IncomeTaxAdvances: []tax.IncomeTaxAdvance{
				{
					Period:  p,
					Amount:  money.New(10000000, money.ARS),
					DueDate: advDueDate,
				},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if prof.SelfEmployedMonthlyContribution.Amount != 5000000 {
			t.Errorf("SelfEmployed = %v, want 5000000", prof.SelfEmployedMonthlyContribution)
		}
		if prof.PreviousVatCredit.Amount != 1200000 {
			t.Errorf("PreviousVatCredit = %v, want 1200000", prof.PreviousVatCredit)
		}
		if len(prof.IncomeTaxAdvances) != 1 {
			t.Fatalf("expected 1 advance, got %d", len(prof.IncomeTaxAdvances))
		}
		if len(notifier.events) == 0 {
			t.Error("expected notifier event")
		}
	})

	t.Run("Validaciones de importes negativos y monedas inválidas", func(t *testing.T) {
		_, err := uc.Execute(ctx, UpdateRIProfileInput{
			ClientID:                        "cli_ri",
			SelfEmployedMonthlyContribution: money.New(-100, money.ARS),
			PreviousVatCredit:               money.New(-200, money.ARS),
			IncomeTaxAdvances: []tax.IncomeTaxAdvance{
				{
					Period: period.Period{},
					Amount: money.New(-50, money.ARS),
				},
			},
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("Cliente no encontrado falla", func(t *testing.T) {
		_, err := uc.Execute(ctx, UpdateRIProfileInput{
			ClientID: "cli_none",
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}
