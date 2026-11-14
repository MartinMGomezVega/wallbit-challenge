// Package parameters define los parámetros globales de criterio impositivo
// configurables por el contador (domain-context.md §9, architecture.md §2.2).
package parameters

import (
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

// FXRule es el criterio de selección de cotización BNA para pesificar facturas.
type FXRule string

const (
	FXRuleBNABuyPrevBusinessDay  FXRule = "BNA_BUY_PREV_BUSINESS_DAY"
	FXRuleBNASellPrevBusinessDay FXRule = "BNA_SELL_PREV_BUSINESS_DAY"
	FXRuleBNABuySameDay          FXRule = "BNA_BUY_SAME_DAY"
)

// AccountantParameters son las variables de negocio editables por el contador matriculado.
type AccountantParameters struct {
	InvoiceECountsForCap           bool              `json:"invoiceECountsForCap"`
	FXRule                         FXRule            `json:"fxRule"`
	CategoryCapWarningBp           money.BasisPoints `json:"categoryCapWarningBp"`
	KCapWarningBp                  money.BasisPoints `json:"kCapWarningBp"`
	AtypicalMultiplierBp           money.BasisPoints `json:"atypicalMultiplierBp"`
	AtypicalWindowMonths           int               `json:"atypicalWindowMonths"`
	MissingInvoiceMinUSD           money.Money       `json:"missingInvoiceMinUsd"`
	ReconciliationToleranceBp      money.BasisPoints `json:"reconciliationToleranceBp"`
	ReconciliationWindowBeforeDays int               `json:"reconciliationWindowBeforeDays"`
	ReconciliationWindowAfterDays  int               `json:"reconciliationWindowAfterDays"`
	CardSpendKCapBp                money.BasisPoints `json:"cardSpendKCapBp"`
	ReminderDaysBeforeDue          int               `json:"reminderDaysBeforeDue"`
	UpdatedAt                      time.Time         `json:"updatedAt"`
	UpdatedBy                      string            `json:"updatedBy"`
}

// Defaults devuelve los parámetros por defecto según domain-context.md §9.
func Defaults() AccountantParameters {
	return AccountantParameters{
		InvoiceECountsForCap:           true,
		FXRule:                         FXRuleBNABuyPrevBusinessDay,
		CategoryCapWarningBp:           8000,  // 80%
		KCapWarningBp:                  8000,  // 80%
		AtypicalMultiplierBp:           25000, // 2.5x
		AtypicalWindowMonths:           6,
		MissingInvoiceMinUSD:           money.New(5000, money.USD), // US$ 50
		ReconciliationToleranceBp:      300,                        // 3%
		ReconciliationWindowBeforeDays: 45,
		ReconciliationWindowAfterDays:  10,
		CardSpendKCapBp:                4000, // 40%
		ReminderDaysBeforeDue:          3,
		UpdatedBy:                      "system",
	}
}

// Validate verifica que los parámetros sean consistentes (spec 001 criterio 11).
func (p AccountantParameters) Validate() error {
	var errs domain.ValidationErrors

	checkBp := func(field string, bp money.BasisPoints) {
		if bp < 0 || bp > 100000 {
			errs = append(errs, domain.ValidationError{
				Field: field,
				Msg:   "debe estar entre 0 y 100.000 (0% a 1.000%)",
			})
		}
	}

	checkBp("categoryCapWarningBp", p.CategoryCapWarningBp)
	checkBp("kCapWarningBp", p.KCapWarningBp)
	checkBp("atypicalMultiplierBp", p.AtypicalMultiplierBp)
	checkBp("reconciliationToleranceBp", p.ReconciliationToleranceBp)
	checkBp("cardSpendKCapBp", p.CardSpendKCapBp)

	if p.AtypicalWindowMonths <= 0 {
		errs = append(errs, domain.ValidationError{Field: "atypicalWindowMonths", Msg: "debe ser mayor a 0"})
	}
	if p.ReconciliationWindowBeforeDays <= 0 {
		errs = append(errs, domain.ValidationError{Field: "reconciliationWindowBeforeDays", Msg: "debe ser mayor a 0"})
	}
	if p.ReconciliationWindowAfterDays <= 0 {
		errs = append(errs, domain.ValidationError{Field: "reconciliationWindowAfterDays", Msg: "debe ser mayor a 0"})
	}
	if p.ReminderDaysBeforeDue <= 0 {
		errs = append(errs, domain.ValidationError{Field: "reminderDaysBeforeDue", Msg: "debe ser mayor a 0"})
	}
	if p.MissingInvoiceMinUSD.Amount < 0 || p.MissingInvoiceMinUSD.Currency != money.USD {
		errs = append(errs, domain.ValidationError{Field: "missingInvoiceMinUsd", Msg: "debe ser un monto positivo en USD"})
	}

	switch p.FXRule {
	case FXRuleBNABuyPrevBusinessDay, FXRuleBNASellPrevBusinessDay, FXRuleBNABuySameDay:
		// ok
	default:
		errs = append(errs, domain.ValidationError{Field: "fxRule", Msg: "regla de tipo de cambio desconocida"})
	}

	return errs.OrNil()
}
