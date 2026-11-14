package tax

import (
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// IncomeTaxAdvance representa un anticipo de impuesto a las ganancias (solo RI).
type IncomeTaxAdvance struct {
	Period  period.Period `json:"period"`
	Amount  money.Money   `json:"amount"`
	DueDate time.Time     `json:"dueDate"`
}

// RIProfile guarda la configuración impositiva específica de un Responsable Inscripto cargada por el contador.
type RIProfile struct {
	SelfEmployedMonthlyContribution money.Money        `json:"selfEmployedMonthlyContribution"`
	IncomeTaxAdvances               []IncomeTaxAdvance `json:"incomeTaxAdvances"`
	PreviousVatCredit               money.Money        `json:"previousVatCredit"`
}

// RIDueDates agrupa los vencimientos impositivos de un Responsable Inscripto.
type RIDueDates struct {
	VatDueDate              time.Time
	SelfEmployedDueDate     time.Time
	IncomeTaxAdvanceDueDate *time.Time
}

// RIInput define las entradas necesarias para la liquidación mensual de Responsable Inscripto.
type RIInput struct {
	Period           period.Period
	Invoices         []PricedInvoice
	PurchaseInvoices []invoice.PurchaseInvoice
	Profile          RIProfile
	DueDates         RIDueDates
	Params           parameters.AccountantParameters
}

// RIAssessment contiene el resultado de la liquidación mensual de un Responsable Inscripto.
type RIAssessment struct {
	BilledInPeriodUSD        money.Money  `json:"billedInPeriodUSD"`
	BilledInPeriodARS        money.Money  `json:"billedInPeriodARS"`
	ExportSalesARS           money.Money  `json:"exportSalesARS"`
	LocalSalesNetARS         money.Money  `json:"localSalesNetARS"`
	OutputVat                money.Money  `json:"outputVat"`
	InputVat                 money.Money  `json:"inputVat"`
	PreviousVatCredit        money.Money  `json:"previousVatCredit"`
	VatToPay                 money.Money  `json:"vatToPay"`
	NewVatCredit             money.Money  `json:"newVatCredit"`
	VatDueDate               time.Time    `json:"vatDueDate"`
	SelfEmployedContribution money.Money  `json:"selfEmployedContribution"`
	SelfEmployedDueDate      time.Time    `json:"selfEmployedDueDate"`
	IncomeTaxAdvance         *money.Money `json:"incomeTaxAdvance,omitempty"`
	IncomeTaxAdvanceDueDate  *time.Time   `json:"incomeTaxAdvanceDueDate,omitempty"`
	TotalToPay               money.Money  `json:"totalToPay"`
}

// AssessRI ejecuta la liquidación pura de Responsable Inscripto según domain-context.md §7.2.
func AssessRI(in RIInput) (RIAssessment, error) {
	if in.DueDates.VatDueDate.IsZero() {
		return RIAssessment{}, errors.New("VatDueDate es requerida")
	}
	if in.DueDates.SelfEmployedDueDate.IsZero() {
		return RIAssessment{}, errors.New("SelfEmployedDueDate es requerida")
	}

	billedInPeriodUSD := money.Zero(money.USD)
	billedInPeriodARS := money.Zero(money.ARS)
	exportSalesARS := money.Zero(money.ARS)
	localSalesNetARS := money.Zero(money.ARS)
	outputVatARS := money.Zero(money.ARS)

	// Procesar facturas de venta del período
	for _, pi := range in.Invoices {
		p := period.FromTime(pi.Invoice.IssueDate)
		if p != in.Period {
			continue
		}

		billedInPeriodARS.Amount += pi.AmountARS.Amount
		if pi.AmountUSD != nil {
			billedInPeriodUSD.Amount += pi.AmountUSD.Amount
		} else if pi.Invoice.Currency == money.USD {
			billedInPeriodUSD.Amount += pi.Invoice.TotalAmount.Amount
		}

		switch pi.Invoice.Type {
		case invoice.TypeE:
			// Exportaciones: IVA 0%
			exportSalesARS.Amount += pi.AmountARS.Amount
		case invoice.TypeA, invoice.TypeB:
			// Ventas locales: Neto e IVA débito fiscal
			localSalesNetARS.Amount += pi.Invoice.NetAmount.Amount
			if pi.Invoice.VatAmount.IsPositive() {
				outputVatARS.Amount += pi.Invoice.VatAmount.Amount
			} else {
				// 21% por defecto si no viene desglosado
				outputVatARS.Amount += (pi.Invoice.NetAmount.Amount * 21) / 100
			}
		}
	}

	// Procesar facturas de compra del período (crédito fiscal IVA)
	inputVatARS := money.Zero(money.ARS)
	for _, pi := range in.PurchaseInvoices {
		p := period.FromTime(pi.IssueDate)
		if p == in.Period {
			inputVatARS.Amount += pi.VatAmount.Amount
		}
	}

	// Cálculo de saldo de IVA
	prevCredit := in.Profile.PreviousVatCredit.Amount
	balance := outputVatARS.Amount - inputVatARS.Amount - prevCredit

	var vatToPay money.Money
	var newVatCredit money.Money

	if balance < 0 {
		vatToPay = money.Zero(money.ARS)
		newVatCredit = money.New(-balance, money.ARS)
	} else {
		vatToPay = money.New(balance, money.ARS)
		newVatCredit = money.Zero(money.ARS)
	}

	// Autónomos
	selfEmployed := in.Profile.SelfEmployedMonthlyContribution

	// Anticipo de impuesto a las ganancias del período (si fue cargado)
	var incomeTaxAdvance *money.Money
	var incomeTaxAdvanceDueDate *time.Time

	for _, adv := range in.Profile.IncomeTaxAdvances {
		if adv.Period == in.Period {
			advAmount := adv.Amount
			incomeTaxAdvance = &advAmount
			due := adv.DueDate
			incomeTaxAdvanceDueDate = &due
			break
		}
	}

	// Total a pagar = IVA a pagar + Autónomos + Anticipo
	totalToPayAmount := vatToPay.Amount + selfEmployed.Amount
	if incomeTaxAdvance != nil {
		totalToPayAmount += incomeTaxAdvance.Amount
	}

	return RIAssessment{
		BilledInPeriodUSD:        billedInPeriodUSD,
		BilledInPeriodARS:        billedInPeriodARS,
		ExportSalesARS:           exportSalesARS,
		LocalSalesNetARS:         localSalesNetARS,
		OutputVat:                outputVatARS,
		InputVat:                 inputVatARS,
		PreviousVatCredit:        money.New(prevCredit, money.ARS),
		VatToPay:                 vatToPay,
		NewVatCredit:             newVatCredit,
		VatDueDate:               in.DueDates.VatDueDate,
		SelfEmployedContribution: selfEmployed,
		SelfEmployedDueDate:      in.DueDates.SelfEmployedDueDate,
		IncomeTaxAdvance:         incomeTaxAdvance,
		IncomeTaxAdvanceDueDate:  incomeTaxAdvanceDueDate,
		TotalToPay:               money.New(totalToPayAmount, money.ARS),
	}, nil
}
