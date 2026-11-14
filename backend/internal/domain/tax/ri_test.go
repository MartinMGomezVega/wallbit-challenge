package tax

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestAssessRI_Errors(t *testing.T) {
	p := period.Period{Year: 2026, Month: 10}
	dueDate := time.Date(2026, 11, 18, 0, 0, 0, 0, period.Location())

	// VatDueDate zero
	_, err := AssessRI(RIInput{
		Period:   p,
		DueDates: RIDueDates{SelfEmployedDueDate: dueDate},
	})
	if err == nil {
		t.Error("expected error when VatDueDate is zero")
	}

	// SelfEmployedDueDate zero
	_, err = AssessRI(RIInput{
		Period:   p,
		DueDates: RIDueDates{VatDueDate: dueDate},
	})
	if err == nil {
		t.Error("expected error when SelfEmployedDueDate is zero")
	}
}

func TestAssessRI_Criterio4(t *testing.T) {
	loc := period.Location()
	p := period.Period{Year: 2026, Month: 10}
	vatDue := time.Date(2026, 11, 18, 0, 0, 0, 0, loc)
	selfDue := time.Date(2026, 11, 5, 0, 0, 0, 0, loc)
	advDue := time.Date(2026, 11, 15, 0, 0, 0, 0, loc)
	dueDates := RIDueDates{VatDueDate: vatDue, SelfEmployedDueDate: selfDue}
	params := parameters.Defaults()

	// 1. Solo exportación: Factura E (IVA 0), saldo a favor crece con compras
	t.Run("Solo exportación con compras genera crédito fiscal", func(t *testing.T) {
		usdAmount := money.New(300000, money.USD)
		invE := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:          "inv_e",
				Type:        invoice.TypeE,
				IssueDate:   time.Date(2026, 10, 10, 0, 0, 0, 0, loc),
				Currency:    money.USD,
				TotalAmount: usdAmount,
			},
			AmountARS: money.New(450000000, money.ARS), // $4,500,000 ARS
			AmountUSD: &usdAmount,
		}

		// Factura de compra con IVA crédito fiscal de $210,000 ARS
		purch := invoice.PurchaseInvoice{
			ID:        "pi_1",
			IssueDate: time.Date(2026, 10, 15, 0, 0, 0, 0, loc),
			NetAmount: money.New(100000000, money.ARS),
			VatAmount: money.New(21000000, money.ARS),
		}

		res, err := AssessRI(RIInput{
			Period:           p,
			Invoices:         []PricedInvoice{invE},
			PurchaseInvoices: []invoice.PurchaseInvoice{purch},
			Profile: RIProfile{
				SelfEmployedMonthlyContribution: money.New(5000000, money.ARS), // $50,000 ARS
				PreviousVatCredit:               money.Zero(money.ARS),
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if res.ExportSalesARS.Amount != 450000000 {
			t.Errorf("ExportSalesARS = %v, want 450000000", res.ExportSalesARS)
		}
		if res.OutputVat.Amount != 0 {
			t.Errorf("OutputVat = %v, want 0", res.OutputVat)
		}
		if res.InputVat.Amount != 21000000 {
			t.Errorf("InputVat = %v, want 21000000", res.InputVat)
		}
		if res.VatToPay.Amount != 0 {
			t.Errorf("VatToPay = %v, want 0", res.VatToPay)
		}
		if res.NewVatCredit.Amount != 21000000 {
			t.Errorf("NewVatCredit = %v, want 21000000", res.NewVatCredit)
		}
		if res.TotalToPay.Amount != 5000000 {
			t.Errorf("TotalToPay = %v, want 5000000 (solo autónomos)", res.TotalToPay)
		}
	})

	// 2. Factura A local (débito 21% del neto) sin y con VatAmount informado
	t.Run("Factura A local con débito fiscal", func(t *testing.T) {
		invA := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_a",
				Type:      invoice.TypeA,
				IssueDate: time.Date(2026, 10, 5, 0, 0, 0, 0, loc),
				NetAmount: money.New(100000000, money.ARS), // $1,000,000 ARS neto
				VatAmount: money.New(21000000, money.ARS),  // $210,000 ARS IVA informado
			},
			AmountARS: money.New(121000000, money.ARS),
		}

		res, err := AssessRI(RIInput{
			Period:           p,
			Invoices:         []PricedInvoice{invA},
			PurchaseInvoices: nil,
			Profile: RIProfile{
				SelfEmployedMonthlyContribution: money.Zero(money.ARS),
				PreviousVatCredit:               money.Zero(money.ARS),
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if res.LocalSalesNetARS.Amount != 100000000 {
			t.Errorf("LocalSalesNetARS = %v, want 100000000", res.LocalSalesNetARS)
		}
		if res.OutputVat.Amount != 21000000 {
			t.Errorf("OutputVat = %v, want 21000000", res.OutputVat)
		}
		if res.VatToPay.Amount != 21000000 {
			t.Errorf("VatToPay = %v, want 21000000", res.VatToPay)
		}
		if res.NewVatCredit.Amount != 0 {
			t.Errorf("NewVatCredit = %v, want 0", res.NewVatCredit)
		}
	})

	// 2b. Factura A con VatAmount no informado (calcula 21%)
	t.Run("Factura B sin VatAmount calcula 21%", func(t *testing.T) {
		invB := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_b",
				Type:      invoice.TypeB,
				IssueDate: time.Date(2026, 10, 5, 0, 0, 0, 0, loc),
				NetAmount: money.New(100000000, money.ARS), // $1,000,000 ARS neto
				VatAmount: money.Zero(money.ARS),
			},
			AmountARS: money.New(121000000, money.ARS),
		}

		res, err := AssessRI(RIInput{
			Period:   p,
			Invoices: []PricedInvoice{invB},
			Profile: RIProfile{
				SelfEmployedMonthlyContribution: money.Zero(money.ARS),
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if res.OutputVat.Amount != 21000000 {
			t.Errorf("OutputVat = %v, want 21000000", res.OutputVat)
		}
	})

	// 3. Saldo a favor previo que absorbe débito parcial y total
	t.Run("Saldo a favor previo absorbe débito parcial", func(t *testing.T) {
		invA := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_a",
				Type:      invoice.TypeA,
				IssueDate: time.Date(2026, 10, 5, 0, 0, 0, 0, loc),
				NetAmount: money.New(100000000, money.ARS),
				VatAmount: money.New(21000000, money.ARS), // $210,000 débito
			},
			AmountARS: money.New(121000000, money.ARS),
		}

		// Saldo a favor previo de $100,000 ARS -> paga $110,000 ARS
		res, err := AssessRI(RIInput{
			Period:   p,
			Invoices: []PricedInvoice{invA},
			Profile: RIProfile{
				PreviousVatCredit: money.New(10000000, money.ARS),
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if res.VatToPay.Amount != 11000000 {
			t.Errorf("VatToPay = %v, want 11000000", res.VatToPay)
		}
		if res.NewVatCredit.Amount != 0 {
			t.Errorf("NewVatCredit = %v, want 0", res.NewVatCredit)
		}
	})

	t.Run("Saldo a favor previo absorbe débito total", func(t *testing.T) {
		invA := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_a",
				Type:      invoice.TypeA,
				IssueDate: time.Date(2026, 10, 5, 0, 0, 0, 0, loc),
				NetAmount: money.New(100000000, money.ARS),
				VatAmount: money.New(21000000, money.ARS), // $210,000 débito
			},
			AmountARS: money.New(121000000, money.ARS),
		}

		// Saldo a favor previo de $300,000 ARS -> paga $0, nuevo saldo $90,000 ARS
		res, err := AssessRI(RIInput{
			Period:   p,
			Invoices: []PricedInvoice{invA},
			Profile: RIProfile{
				PreviousVatCredit: money.New(30000000, money.ARS),
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if res.VatToPay.Amount != 0 {
			t.Errorf("VatToPay = %v, want 0", res.VatToPay)
		}
		if res.NewVatCredit.Amount != 9000000 {
			t.Errorf("NewVatCredit = %v, want 9000000", res.NewVatCredit)
		}
	})

	// 4. Anticipo del período presente y ausente + TotalToPay
	t.Run("Anticipo presente y ausente con TotalToPay", func(t *testing.T) {
		advAmount := money.New(15000000, money.ARS)   // $150,000
		selfEmployed := money.New(8000000, money.ARS) // $80,000

		// Con anticipo
		resWithAdv, err := AssessRI(RIInput{
			Period: p,
			Profile: RIProfile{
				SelfEmployedMonthlyContribution: selfEmployed,
				IncomeTaxAdvances: []IncomeTaxAdvance{
					{
						Period:  p,
						Amount:  advAmount,
						DueDate: advDue,
					},
					{
						Period:  period.Period{Year: 2026, Month: 11},
						Amount:  money.New(9999999, money.ARS),
						DueDate: advDue,
					},
				},
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if resWithAdv.IncomeTaxAdvance == nil || resWithAdv.IncomeTaxAdvance.Amount != 15000000 {
			t.Errorf("IncomeTaxAdvance = %v, want 15000000", resWithAdv.IncomeTaxAdvance)
		}
		if resWithAdv.IncomeTaxAdvanceDueDate == nil || !resWithAdv.IncomeTaxAdvanceDueDate.Equal(advDue) {
			t.Errorf("IncomeTaxAdvanceDueDate = %v, want %v", resWithAdv.IncomeTaxAdvanceDueDate, advDue)
		}
		// TotalToPay = 0 (IVA) + 80,000 (Autónomos) + 150,000 (Anticipo) = 230,000 ARS
		if resWithAdv.TotalToPay.Amount != 23000000 {
			t.Errorf("TotalToPay = %v, want 23000000", resWithAdv.TotalToPay)
		}

		// Sin anticipo para este período
		resWithoutAdv, err := AssessRI(RIInput{
			Period: p,
			Profile: RIProfile{
				SelfEmployedMonthlyContribution: selfEmployed,
			},
			DueDates: dueDates,
			Params:   params,
		})
		if err != nil {
			t.Fatalf("AssessRI failed: %v", err)
		}

		if resWithoutAdv.IncomeTaxAdvance != nil {
			t.Errorf("expected IncomeTaxAdvance to be nil, got %v", resWithoutAdv.IncomeTaxAdvance)
		}
		if resWithoutAdv.TotalToPay.Amount != 8000000 {
			t.Errorf("TotalToPay = %v, want 8000000", resWithoutAdv.TotalToPay)
		}
	})
}
