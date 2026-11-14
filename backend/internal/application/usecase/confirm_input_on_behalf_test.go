package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

type confirmNotifier struct {
	events []ports.ClosingInputsChanged
}

func (r *confirmNotifier) NotifyClosingInputsChanged(_ context.Context, e ports.ClosingInputsChanged) error {
	r.events = append(r.events, e)
	return nil
}

func TestConfirmInputOnBehalf(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, period.Location())
	clk := fakes.Clock{Current: now}
	inputRepo := fakes.NewClientInputRepository()
	closingRepo := fakes.NewClosingRepository()
	notifier := &confirmNotifier{}

	uc := &usecase.ConfirmInputOnBehalf{
		Inputs:   inputRepo,
		Closings: closingRepo,
		Clock:    clk,
		Notifier: notifier,
	}

	p := period.MustParse("2026-10")

	// 1. Nota obligatoria (vacía o espacios) -> 422 NOTE_REQUIRED
	_, err := uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID: "cli_1",
		Period:   p,
		None:     true,
		Note:     "   ",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for empty note, got %v", err)
	}

	// 2. Nota superior a 300 caracteres -> 422 NOTE_REQUIRED
	_, err = uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID: "cli_1",
		Period:   p,
		None:     true,
		Note:     strings.Repeat("a", 301),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for note > 300 chars, got %v", err)
	}

	// 3. None = false sin ítems -> 422 EXTERNAL_INCOME_NOT_ANSWERED
	_, err = uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID:       "cli_1",
		Period:         p,
		None:           false,
		ExternalIncome: nil,
		Note:           "Cliente confirmó sin ítems",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation when None=false and items empty, got %v", err)
	}

	// 4. Confirmación exitosa con None = true
	res, err := uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID: "cli_1",
		Period:   p,
		None:     true,
		Note:     "Confirmado por WhatsApp el 05/11",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.DeclaredNoExternalIncome {
		t.Errorf("expected DeclaredNoExternalIncome = true")
	}
	if res.ConfirmedBy != clientinput.ConfirmedByAccountant {
		t.Errorf("ConfirmedBy = %v, want ACCOUNTANT", res.ConfirmedBy)
	}
	if res.ConfirmationNote != "Confirmado por WhatsApp el 05/11" {
		t.Errorf("ConfirmationNote = %q, want 'Confirmado por WhatsApp el 05/11'", res.ConfirmationNote)
	}
	if len(notifier.events) != 1 || notifier.events[0].ClientID != "cli_1" {
		t.Fatalf("expected 1 notification event for cli_1, got %+v", notifier.events)
	}

	// 5. Segundo intento sobre el mismo período -> 409 ALREADY_CONFIRMED
	_, err = uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID: "cli_1",
		Period:   p,
		None:     true,
		Note:     "Reintento de confirmación",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict on already confirmed input, got %v", err)
	}

	// 6. Confirmación exitosa con ingresos externos
	res2, err := uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID: "cli_2",
		Period:   p,
		None:     false,
		ExternalIncome: []clientinput.ExternalIncomeItem{
			{Amount: money.New(50000, money.USD), Description: "Cobro Upwork directo"},
		},
		Note: "Me envió resumen de cuenta por mail",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res2.DeclaredNoExternalIncome {
		t.Errorf("expected DeclaredNoExternalIncome = false")
	}
	if len(res2.ExternalIncome) != 1 {
		t.Errorf("expected 1 external income item, got %d", len(res2.ExternalIncome))
	}
}

func TestConfirmInputOnBehalf_IntegrationWithClosingRecalc(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, period.Location())
	h := newTestHarness(now)

	p := period.MustParse("2026-10")

	// Crear cliente Monotributo
	c := client.Client{
		ID:                  "cli_test",
		Name:                "Paula Rojas",
		Email:               "paula@test.com",
		CUIT:                "27-33444444-4",
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatB,
		CreatedAt:           now.AddDate(-1, 0, 0),
	}
	_ = h.clients.Save(ctx, c)

	recalc := &usecase.RecalcOnChange{
		Compute:  h.uc,
		Closings: h.closings,
	}

	uc := &usecase.ConfirmInputOnBehalf{
		Inputs:   h.clientInputs,
		Closings: h.closings,
		Clock:    h.clock,
		Notifier: recalc,
	}

	// 1. Ejecutar liquidación inicial -> sin confirmación queda en PENDING_CLIENT_INPUT
	initRes, err := h.uc.Execute(ctx, "cli_test", p)
	if err != nil {
		t.Fatalf("computeUC: %v", err)
	}
	if initRes.Closing.Status != closing.StatusPendingClientInput {
		t.Fatalf("expected initial status PENDING_CLIENT_INPUT, got %v", initRes.Closing.Status)
	}

	// 2. Confirmar en nombre del cliente -> avanza a READY_FOR_REVIEW
	_, err = uc.Execute(ctx, usecase.ConfirmInputOnBehalfInput{
		ClientID: "cli_test",
		Period:   p,
		None:     true,
		Note:     "Confirmado telefónicamente",
	})
	if err != nil {
		t.Fatalf("ConfirmInputOnBehalf: %v", err)
	}

	cloAfter, found, err := h.closings.ByClientPeriod(ctx, "cli_test", p)
	if err != nil || !found {
		t.Fatalf("closing not found: %v", err)
	}
	if cloAfter.Status != closing.StatusReadyForReview {
		t.Errorf("expected closing status READY_FOR_REVIEW after confirmation, got %v", cloAfter.Status)
	}
}
