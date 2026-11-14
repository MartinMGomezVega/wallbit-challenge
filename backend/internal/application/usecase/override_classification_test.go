package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

type recordingNotifier struct {
	events []ports.ClosingInputsChanged
}

func (r *recordingNotifier) NotifyClosingInputsChanged(_ context.Context, e ports.ClosingInputsChanged) error {
	r.events = append(r.events, e)
	return nil
}

func TestOverrideClassification(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 10, 0, 0, 0, period.Location())
	clk := fakes.Clock{Current: now}

	movRepo := fakes.NewMovementRepository()
	classRepo := fakes.NewClassificationRepository()
	notifier := &recordingNotifier{}

	uc := &OverrideClassification{
		Movements:       movRepo,
		Classifications: classRepo,
		Clock:           clk,
		Notifier:        notifier,
	}

	// 1. Validation error if kind is UNCLASSIFIED or invalid
	_, err := uc.Execute(ctx, OverrideClassificationInput{
		MovementID: "mov_1",
		Kind:       movement.KindUnclassified,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for UNCLASSIFIED, got %v", err)
	}

	// 2. Not found error if movement does not exist
	_, err = uc.Execute(ctx, OverrideClassificationInput{
		MovementID: "mov_missing",
		Kind:       movement.KindIncome,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing movement, got %v", err)
	}

	// 3. Success
	m := movement.Movement{
		ID:         "mov_1",
		ClientID:   "cli_1",
		Type:       movement.TypeInboundTransfer,
		Amount:     money.New(50000, money.USD),
		OccurredAt: time.Date(2026, 10, 15, 14, 0, 0, 0, period.Location()),
	}
	if err := movRepo.Save(ctx, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	res, err := uc.Execute(ctx, OverrideClassificationInput{
		MovementID: "mov_1",
		Kind:       movement.KindIncome,
		Note:       "Pago de honorarios del exterior confirmado con el cliente",
		Operator:   "accountant_maria",
	})
	if err != nil {
		t.Fatalf("Execute unexpected error: %v", err)
	}

	if res.MovementID != "mov_1" {
		t.Errorf("MovementID = %s, want mov_1", res.MovementID)
	}
	if res.Kind != movement.KindIncome {
		t.Errorf("Kind = %s, want INCOME", res.Kind)
	}
	if res.Confidence != movement.ConfidenceHigh {
		t.Errorf("Confidence = %s, want HIGH", res.Confidence)
	}
	if res.Rule != movement.RuleManualOverride {
		t.Errorf("Rule = %s, want MANUAL_OVERRIDE", res.Rule)
	}
	if res.OverriddenBy != "accountant_maria" {
		t.Errorf("OverriddenBy = %s, want accountant_maria", res.OverriddenBy)
	}
	if res.OverriddenAt == nil || !res.OverriddenAt.Equal(now) {
		t.Errorf("OverriddenAt = %v, want %v", res.OverriddenAt, now)
	}
	if res.Note != "Pago de honorarios del exterior confirmado con el cliente" {
		t.Errorf("Note = %s", res.Note)
	}

	// Verify persistence in repository
	saved, found, err := classRepo.Get(ctx, "mov_1")
	if err != nil || !found || saved.Rule != movement.RuleManualOverride {
		t.Fatalf("saved classification mismatch: %v, %v, %+v", err, found, saved)
	}

	// Verify notifier event
	if len(notifier.events) != 1 {
		t.Fatalf("expected 1 notifier event, got %d", len(notifier.events))
	}
	expectedPeriod, _ := period.Parse("2026-10")
	if notifier.events[0].ClientID != "cli_1" || notifier.events[0].Period != expectedPeriod {
		t.Errorf("event mismatch: %+v", notifier.events[0])
	}
}
