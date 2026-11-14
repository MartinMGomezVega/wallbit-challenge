package signal

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/reconciliation"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestEvaluate_EveryCodeTriggerAndBoundary(t *testing.T) {
	loc := period.Location()
	params := parameters.Defaults()
	params.CategoryCapWarningBp = 8000                       // 80%
	params.KCapWarningBp = 8000                              // 80%
	params.CardSpendKCapBp = 8000                            // 80%
	params.MissingInvoiceMinUSD = money.New(5000, money.USD) // 50 USD

	kCap := money.New(10000000000, money.ARS) // 100M ARS
	catD := tax.CatD
	catE := tax.CatE
	catC := tax.CatC

	// 1. MONO_EXCLUSION_K_CAP_EXCEEDED
	t.Run("MONO_EXCLUSION_K_CAP_EXCEEDED", func(t *testing.T) {
		exclDate := time.Date(2026, 10, 14, 0, 0, 0, 0, loc)
		excess := money.New(4000, money.ARS)

		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				Excluded:      true,
				ExclusionDate: &exclDate,
				ExcessARS:     &excess,
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Severity != SeverityRed || eval.Primary == nil || eval.Primary.Code != CodeMonoExclusionKCapExceeded {
			t.Fatalf("expected MONO_EXCLUSION_K_CAP_EXCEEDED primary, got %+v", eval)
		}
		expectedMsg := "Superó el tope de la categoría K el 14/10/2026 por $ 40,00. Corresponde pase a Responsable Inscripto."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}

		// Non-trigger: Excluded = false
		inNonTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				Excluded: false,
			},
			Params: params,
		}
		evalNon := Evaluate(inNonTrigger)
		for _, s := range evalNon.Signals {
			if s.Code == CodeMonoExclusionKCapExceeded {
				t.Error("should not trigger when Excluded is false")
			}
		}
	})

	// 2. POST_EXCLUSION_C_INVOICES
	t.Run("POST_EXCLUSION_C_INVOICES", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeRI,
			CInvoicesAfterRI: []CInvoiceAfterRI{
				{
					Invoice:   invoice.Invoice{ID: "inv_c1"},
					AmountARS: money.New(5000000, money.ARS),
				},
				{
					Invoice:   invoice.Invoice{ID: "inv_c2"},
					AmountARS: money.New(7000000, money.ARS),
				},
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Severity != SeverityRed || eval.Primary == nil || eval.Primary.Code != CodePostExclusionCInvoices {
			t.Fatalf("expected POST_EXCLUSION_C_INVOICES, got %+v", eval)
		}
		expectedMsg := "2 facturas C emitidas después de la exclusión ($ 120.000,00)."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}

		// Non-trigger: len == 0
		inNonTrigger := Input{
			Regime: client.RegimeRI,
			Params: params,
		}
		evalNon := Evaluate(inNonTrigger)
		for _, s := range evalNon.Signals {
			if s.Code == CodePostExclusionCInvoices {
				t.Error("should not trigger when CInvoicesAfterRI is empty")
			}
		}
	})

	// 3. MISSING_INVOICES
	t.Run("MISSING_INVOICES", func(t *testing.T) {
		// Borde: menor que mínimo (49.99 USD) no dispara
		inUnder := Input{
			Regime: client.RegimeMonotributo,
			Reconciliation: reconciliation.Result{
				UnmatchedIncome: []movement.Movement{
					{ID: "m1", Amount: money.New(4999, money.USD)},
				},
			},
			Params: params,
		}
		evalUnder := Evaluate(inUnder)
		for _, s := range evalUnder.Signals {
			if s.Code == CodeMissingInvoices {
				t.Error("should not trigger below missing invoice min")
			}
		}

		// Disparo: 2 cobros por 4200 USD
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Reconciliation: reconciliation.Result{
				UnmatchedIncome: []movement.Movement{
					{ID: "m1", Amount: money.New(200000, money.USD)},
					{ID: "m2", Amount: money.New(220000, money.USD)},
				},
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Severity != SeverityRed || eval.Primary == nil || eval.Primary.Code != CodeMissingInvoices {
			t.Fatalf("expected MISSING_INVOICES, got %+v", eval)
		}
		expectedMsg := "2 cobros por US$ 4.200,00 sin factura asociada."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}
	})

	// 4. DEPOSITS_EXCEED_K_CAP
	t.Run("DEPOSITS_EXCEED_K_CAP", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				KCapARS:       kCap,
				Inbound12mARS: money.New(kCap.Amount+1, money.ARS),
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		found := false
		for _, s := range eval.Signals {
			if s.Code == CodeDepositsExceedKCap {
				found = true
				if s.Severity != SeverityRed || s.Priority != 4 {
					t.Errorf("unexpected meta: %+v", s)
				}
			}
		}
		if !found {
			t.Fatal("expected DEPOSITS_EXCEED_K_CAP to trigger")
		}

		// Borde: Inbound12mARS == kCap no dispara
		inEdge := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				KCapARS:       kCap,
				Inbound12mARS: kCap,
			},
			Params: params,
		}
		evalEdge := Evaluate(inEdge)
		for _, s := range evalEdge.Signals {
			if s.Code == CodeDepositsExceedKCap {
				t.Error("should not trigger when Inbound12mARS == kCap")
			}
		}
	})

	// 5. MONO_NEAR_K_CAP
	t.Run("MONO_NEAR_K_CAP", func(t *testing.T) {
		// Borde: 7999 bp (79.99%) no dispara con umbral 8000 bp
		inUnder := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				PctOfKCapBp: 7999,
			},
			Params: params,
		}
		evalUnder := Evaluate(inUnder)
		for _, s := range evalUnder.Signals {
			if s.Code == CodeMonoNearKCap {
				t.Error("should not trigger when PctOfKCapBp < warning bp")
			}
		}

		// Borde: 8000 bp (80%) dispara
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				PctOfKCapBp: 8000,
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeMonoNearKCap {
			t.Fatalf("expected MONO_NEAR_K_CAP, got %+v", eval)
		}
		expectedMsg := "Acumulado 12 meses al 80% del tope de la categoría K."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}
	})

	// 6. MONO_CATEGORY_CHANGE_REQUIRED
	t.Run("MONO_CATEGORY_CHANGE_REQUIRED", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				CurrentCategory:  catD,
				RequiredCategory: &catE,
				Billed12mARS:     money.New(3500000000, money.ARS), // 35M
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeMonoCategoryChangeRequired {
			t.Fatalf("expected MONO_CATEGORY_CHANGE_REQUIRED, got %+v", eval)
		}
		expectedMsg := "Facturó $ 35.000.000,00 en 12 meses: supera el tope de la D. Recategorizar a E."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}

		// Non-trigger: required == current
		inSame := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				CurrentCategory:  catD,
				RequiredCategory: &catD,
			},
			Params: params,
		}
		evalSame := Evaluate(inSame)
		for _, s := range evalSame.Signals {
			if s.Code == CodeMonoCategoryChangeRequired {
				t.Error("should not trigger when RequiredCategory == CurrentCategory")
			}
		}
	})

	// 7. MONO_NEAR_CATEGORY_CAP
	t.Run("MONO_NEAR_CATEGORY_CAP", func(t *testing.T) {
		// Borde: 7999 bp no dispara
		inUnder := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				CurrentCategory:   catD,
				RequiredCategory:  &catD,
				PctOfCurrentCapBp: 7999,
			},
			Params: params,
		}
		evalUnder := Evaluate(inUnder)
		for _, s := range evalUnder.Signals {
			if s.Code == CodeMonoNearCategoryCap {
				t.Error("should not trigger when PctOfCurrentCapBp < warning bp")
			}
		}

		// Borde: 8000 bp dispara
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				CurrentCategory:   catD,
				RequiredCategory:  &catD,
				PctOfCurrentCapBp: 8000,
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeMonoNearCategoryCap {
			t.Fatalf("expected MONO_NEAR_CATEGORY_CAP, got %+v", eval)
		}
		expectedMsg := "Acumulado al 80% del tope de la categoría D."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}
	})

	// 8. UNCLASSIFIED_MOVEMENTS
	t.Run("UNCLASSIFIED_MOVEMENTS", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			PeriodMovements: []ClassifiedMovement{
				{
					Movement:       movement.Movement{ID: "m1", Amount: money.New(50000, money.USD)},
					Classification: movement.Classification{Kind: movement.KindUnclassified},
				},
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeUnclassifiedMovements {
			t.Fatalf("expected UNCLASSIFIED_MOVEMENTS, got %+v", eval)
		}
		expectedMsg := "1 movimientos sin clasificar (US$ 500,00)."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}

		// Non-trigger: 0 unclassified
		inNone := Input{
			Regime: client.RegimeMonotributo,
			PeriodMovements: []ClassifiedMovement{
				{
					Movement:       movement.Movement{ID: "m1", Amount: money.New(50000, money.USD)},
					Classification: movement.Classification{Kind: movement.KindIncome},
				},
			},
			Params: params,
		}
		evalNone := Evaluate(inNone)
		for _, s := range evalNone.Signals {
			if s.Code == CodeUnclassifiedMovements {
				t.Error("should not trigger when no unclassified movements")
			}
		}
	})

	// 9. EXTERNAL_INCOME_DECLARED
	t.Run("EXTERNAL_INCOME_DECLARED", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			ClientInput: clientinput.ClientPeriodInput{
				ExternalIncome: []clientinput.ExternalIncomeItem{
					{Amount: money.New(150000, money.USD), Description: "Consulting direct"},
				},
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeExternalIncomeDeclared {
			t.Fatalf("expected EXTERNAL_INCOME_DECLARED, got %+v", eval)
		}
		expectedMsg := "Declaró cobros fuera de Wallbit por US$ 1.500,00."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}

		// Non-trigger
		inNone := Input{
			Regime: client.RegimeMonotributo,
			ClientInput: clientinput.ClientPeriodInput{
				DeclaredNoExternalIncome: true,
			},
			Params: params,
		}
		evalNone := Evaluate(inNone)
		for _, s := range evalNone.Signals {
			if s.Code == CodeExternalIncomeDeclared {
				t.Error("should not trigger without external income")
			}
		}
	})

	// 10. INVOICE_WITHOUT_PAYMENT
	t.Run("INVOICE_WITHOUT_PAYMENT", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Reconciliation: reconciliation.Result{
				UnmatchedInvoices: []invoice.Invoice{
					{ID: "inv_1", TotalAmount: money.New(100000, money.USD)},
				},
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeInvoiceWithoutPayment {
			t.Fatalf("expected INVOICE_WITHOUT_PAYMENT, got %+v", eval)
		}
		expectedMsg := "1 facturas por US$ 1.000,00 sin cobro en Wallbit."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}

		// Non-trigger
		inNone := Input{
			Regime: client.RegimeMonotributo,
			Params: params,
		}
		evalNone := Evaluate(inNone)
		for _, s := range evalNone.Signals {
			if s.Code == CodeInvoiceWithoutPayment {
				t.Error("should not trigger without unmatched invoices")
			}
		}
	})

	// 11. HIGH_CARD_SPENDING
	t.Run("HIGH_CARD_SPENDING", func(t *testing.T) {
		// K cap = 100M. Umbral 80% = 80M.
		// Borde: 80M exacto no supera
		inEdge := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				KCapARS:         kCap,
				CardSpend12mARS: money.New(8000000000, money.ARS),
			},
			Params: params,
		}
		evalEdge := Evaluate(inEdge)
		for _, s := range evalEdge.Signals {
			if s.Code == CodeHighCardSpending {
				t.Error("should not trigger when CardSpend12m == threshold")
			}
		}

		// Disparo: 85M (85%)
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				KCapARS:         kCap,
				CardSpend12mARS: money.New(8500000000, money.ARS),
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeHighCardSpending {
			t.Fatalf("expected HIGH_CARD_SPENDING, got %+v", eval)
		}
		expectedMsg := "Gastos con tarjeta en 12 meses al 85% del tope de la categoría K."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}
	})

	// 12. MONO_CATEGORY_DOWNGRADE_POSSIBLE
	t.Run("MONO_CATEGORY_DOWNGRADE_POSSIBLE", func(t *testing.T) {
		inTrigger := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				CurrentCategory:  catD,
				RequiredCategory: &catC, // C < D
			},
			Params: params,
		}
		eval := Evaluate(inTrigger)
		if eval.Primary == nil || eval.Primary.Code != CodeMonoCategoryDowngradePossible {
			t.Fatalf("expected MONO_CATEGORY_DOWNGRADE_POSSIBLE, got %+v", eval)
		}
		expectedMsg := "Podría bajar a la categoría C en la próxima recategorización."
		if eval.Primary.Message != expectedMsg {
			t.Errorf("Message = %q, want %q", eval.Primary.Message, expectedMsg)
		}
	})
}

func TestEvaluate_MutualExclusions(t *testing.T) {
	loc := period.Location()
	params := parameters.Defaults()
	exclDate := time.Date(2026, 10, 14, 0, 0, 0, 0, loc)
	excess := money.New(4000, money.ARS)
	catD := tax.CatD
	catE := tax.CatE

	// Exclusión K presente suprime near K, change required, near category, downgrade
	t.Run("Exclusion K suprime alertas de categoría", func(t *testing.T) {
		in := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				Excluded:          true,
				ExclusionDate:     &exclDate,
				ExcessARS:         &excess,
				CurrentCategory:   catD,
				RequiredCategory:  &catE,
				PctOfKCapBp:       9500,
				PctOfCurrentCapBp: 9500,
			},
			Params: params,
		}
		eval := Evaluate(in)
		for _, s := range eval.Signals {
			if s.Code == CodeMonoNearKCap ||
				s.Code == CodeMonoCategoryChangeRequired ||
				s.Code == CodeMonoNearCategoryCap ||
				s.Code == CodeMonoCategoryDowngradePossible {
				t.Errorf("signal %s should be suppressed by exclusion K", s.Code)
			}
		}
	})

	// Change required suprime near category
	t.Run("Change required suprime near category", func(t *testing.T) {
		in := Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				CurrentCategory:   catD,
				RequiredCategory:  &catE,
				PctOfCurrentCapBp: 9500,
			},
			Params: params,
		}
		eval := Evaluate(in)
		foundChange := false
		for _, s := range eval.Signals {
			if s.Code == CodeMonoCategoryChangeRequired {
				foundChange = true
			}
			if s.Code == CodeMonoNearCategoryCap {
				t.Error("MONO_NEAR_CATEGORY_CAP should be suppressed by change required")
			}
		}
		if !foundChange {
			t.Error("expected MONO_CATEGORY_CHANGE_REQUIRED")
		}
	})
}

func TestEvaluate_GreenEvaluations(t *testing.T) {
	params := parameters.Defaults()

	t.Run("Green Monotributo", func(t *testing.T) {
		eval := Evaluate(Input{
			Regime: client.RegimeMonotributo,
			Mono: &tax.MonotributoAssessment{
				BilledInPeriodUSD: money.New(350000, money.USD),
				Fee:               money.New(7241410, money.ARS),
			},
			Params: params,
		})
		if eval.Severity != SeverityGreen {
			t.Errorf("Severity = %v, want GREEN", eval.Severity)
		}
		expectedDesc := "Facturó US$ 3.500,00 · cuota $ 72.414,10 · todo conciliado."
		if eval.Description != expectedDesc {
			t.Errorf("Description = %q, want %q", eval.Description, expectedDesc)
		}
	})

	t.Run("Green RI", func(t *testing.T) {
		eval := Evaluate(Input{
			Regime: client.RegimeRI,
			RI: &tax.RIAssessment{
				BilledInPeriodUSD: money.New(500000, money.USD),
				VatToPay:          money.New(21000000, money.ARS),
			},
			Params: params,
		})
		if eval.Severity != SeverityGreen {
			t.Errorf("Severity = %v, want GREEN", eval.Severity)
		}
		expectedDesc := "Facturó US$ 5.000,00 · IVA $ 210.000,00 · todo conciliado."
		if eval.Description != expectedDesc {
			t.Errorf("Description = %q, want %q", eval.Description, expectedDesc)
		}
	})
}
