package parameters

import (
	"errors"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func TestDefaultsAreValid(t *testing.T) {
	d := Defaults()
	if err := d.Validate(); err != nil {
		t.Fatalf("Defaults() no es válido: %v", err)
	}
	if !d.InvoiceECountsForCap {
		t.Fatal("InvoiceECountsForCap default debe ser true")
	}
	if d.FXRule != FXRuleBNABuyPrevBusinessDay {
		t.Fatalf("FXRule default = %s", d.FXRule)
	}
	if d.MissingInvoiceMinUSD.Amount != 5000 {
		t.Fatalf("MissingInvoiceMinUSD = %v", d.MissingInvoiceMinUSD)
	}
}

func TestValidationErrors(t *testing.T) {
	p := Defaults()
	p.CategoryCapWarningBp = -1
	p.KCapWarningBp = 100001
	p.AtypicalMultiplierBp = -10
	p.AtypicalWindowMonths = 0
	p.MissingInvoiceMinUSD = money.New(-10, money.USD)
	p.ReconciliationToleranceBp = -5
	p.ReconciliationWindowBeforeDays = -1
	p.ReconciliationWindowAfterDays = 0
	p.CardSpendKCapBp = 200000
	p.ReminderDaysBeforeDue = 0
	p.FXRule = "UNKNOWN"

	err := p.Validate()
	if err == nil {
		t.Fatal("esperaba errores de validación")
	}
	var ves domain.ValidationErrors
	if !errors.As(err, &ves) {
		t.Fatalf("esperaba ValidationErrors, got %v", err)
	}
	fields := ves.Fields()
	expectedFields := []string{
		"categoryCapWarningBp",
		"kCapWarningBp",
		"atypicalMultiplierBp",
		"atypicalWindowMonths",
		"missingInvoiceMinUsd",
		"reconciliationToleranceBp",
		"reconciliationWindowBeforeDays",
		"reconciliationWindowAfterDays",
		"cardSpendKCapBp",
		"reminderDaysBeforeDue",
		"fxRule",
	}
	for _, f := range expectedFields {
		if _, ok := fields[f]; !ok {
			t.Errorf("no validó campo %s", f)
		}
	}

	pUSD := Defaults()
	pUSD.MissingInvoiceMinUSD = money.New(100, money.ARS)
	if err := pUSD.Validate(); err == nil {
		t.Fatal("esperaba error por moneda no USD")
	}
}
