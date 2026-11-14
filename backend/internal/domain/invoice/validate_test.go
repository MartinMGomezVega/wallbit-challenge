package invoice

import (
	"errors"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func TestValidateManual(t *testing.T) {
	now := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	validDate := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)

	baseInv := Invoice{
		ClientID:     "cli_01",
		PointOfSale:  1,
		Number:       10,
		IssueDate:    validDate,
		Currency:     money.USD,
		TotalAmount:  money.New(100000, money.USD),
		ReceiverName: "Acme Corp",
	}

	t.Run("Valid Monotributo Factura E", func(t *testing.T) {
		inv := baseInv
		inv.Type = TypeE
		if err := ValidateManual(inv, "MONOTRIBUTO", now); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("Valid Monotributo Factura C", func(t *testing.T) {
		inv := baseInv
		inv.Type = TypeC
		inv.Currency = money.ARS
		inv.TotalAmount = money.New(50000000, money.ARS)
		if err := ValidateManual(inv, "MONOTRIBUTO", now); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("Invalid Monotributo Factura A", func(t *testing.T) {
		inv := baseInv
		inv.Type = TypeA
		err := ValidateManual(inv, "MONOTRIBUTO", now)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for Monotributo Factura A, got %v", err)
		}
	})

	t.Run("Valid RI Factura A and B", func(t *testing.T) {
		invA := baseInv
		invA.Type = TypeA
		if err := ValidateManual(invA, "RI", now); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}

		invB := baseInv
		invB.Type = TypeB
		if err := ValidateManual(invB, "RI", now); err != nil {
			t.Fatalf("unexpected validation error: %v", err)
		}
	})

	t.Run("Invalid RI Factura C", func(t *testing.T) {
		inv := baseInv
		inv.Type = TypeC
		err := ValidateManual(inv, "RI", now)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for RI Factura C, got %v", err)
		}
	})

	t.Run("Future date", func(t *testing.T) {
		inv := baseInv
		inv.Type = TypeE
		inv.IssueDate = now.Add(24 * time.Hour)
		err := ValidateManual(inv, "MONOTRIBUTO", now)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for future date, got %v", err)
		}
	})

	t.Run("Zero or negative amounts and numbers", func(t *testing.T) {
		inv := Invoice{}
		err := ValidateManual(inv, "MONOTRIBUTO", now)
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("expected ErrValidation for empty invoice, got %v", err)
		}
		var ves domain.ValidationErrors
		if errors.As(err, &ves) {
			fields := ves.Fields()
			if _, ok := fields["clientId"]; !ok {
				t.Error("missing clientId error")
			}
			if _, ok := fields["pointOfSale"]; !ok {
				t.Error("missing pointOfSale error")
			}
			if _, ok := fields["number"]; !ok {
				t.Error("missing number error")
			}
			if _, ok := fields["totalAmount"]; !ok {
				t.Error("missing totalAmount error")
			}
			if _, ok := fields["receiverName"]; !ok {
				t.Error("missing receiverName error")
			}
		}
	})
}
