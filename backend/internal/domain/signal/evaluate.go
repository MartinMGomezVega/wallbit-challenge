package signal

import (
	"fmt"
	"sort"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/reconciliation"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// ClassifiedMovement asocia un movimiento con su clasificación aplicada.
type ClassifiedMovement struct {
	Movement       movement.Movement
	Classification movement.Classification
}

// CInvoiceAfterRI representa una factura C emitida por un contribuyente tras haber pasado a RI.
type CInvoiceAfterRI struct {
	Invoice   invoice.Invoice
	AmountARS money.Money
}

// Input agrupa todas las fuentes de datos necesarias para evaluar señales de riesgo.
type Input struct {
	Regime            client.Regime
	Mono              *tax.MonotributoAssessment
	RI                *tax.RIAssessment
	Reconciliation    reconciliation.Result
	PeriodMovements   []ClassifiedMovement
	HistoricIncomeUSD []money.Money
	ClientInput       clientinput.ClientPeriodInput
	CInvoicesAfterRI  []CInvoiceAfterRI
	Params            parameters.AccountantParameters
}

// Signal representa una alerta o evento de riesgo detectado en el cierre.
type Signal struct {
	Code     Code              `json:"code"`
	Severity Severity          `json:"severity"`
	Priority int               `json:"priority"`
	Message  string            `json:"message"`
	Data     map[string]string `json:"data,omitempty"`
}

// Evaluation es el resultado integral de la evaluación de señales y semáforo.
type Evaluation struct {
	Severity    Severity `json:"severity"`
	Signals     []Signal `json:"signals"`
	Primary     *Signal  `json:"primary,omitempty"`
	Description string   `json:"description"`
}

// Evaluate evalúa de forma pura todas las reglas de riesgo del catálogo.
func Evaluate(in Input) Evaluation {
	var candidates []Signal

	// 1. MONO_EXCLUSION_K_CAP_EXCEEDED (Prio 1, RED, MONO)
	if s, ok := checkMonoExclusionKCapExceeded(in); ok {
		candidates = append(candidates, s)
	}

	// 2. POST_EXCLUSION_C_INVOICES (Prio 2, RED, RI)
	if s, ok := checkPostExclusionCInvoices(in); ok {
		candidates = append(candidates, s)
	}

	// 3. MISSING_INVOICES (Prio 3, RED, ambos)
	if s, ok := checkMissingInvoices(in); ok {
		candidates = append(candidates, s)
	}

	// 4. DEPOSITS_EXCEED_K_CAP (Prio 4, RED, MONO)
	if s, ok := checkDepositsExceedKCap(in); ok {
		candidates = append(candidates, s)
	}

	// 5. MONO_NEAR_K_CAP (Prio 5, YELLOW, MONO)
	if s, ok := checkMonoNearKCap(in); ok {
		candidates = append(candidates, s)
	}

	// 6. MONO_CATEGORY_CHANGE_REQUIRED (Prio 6, YELLOW, MONO)
	if s, ok := checkMonoCategoryChangeRequired(in); ok {
		candidates = append(candidates, s)
	}

	// 7. MONO_NEAR_CATEGORY_CAP (Prio 7, YELLOW, MONO)
	if s, ok := checkMonoNearCategoryCap(in); ok {
		candidates = append(candidates, s)
	}

	// 8. ATYPICAL_INCOME (Prio 8, YELLOW, ambos)
	if s, ok := checkAtypicalIncome(in); ok {
		candidates = append(candidates, s)
	}

	// 9. UNCLASSIFIED_MOVEMENTS (Prio 9, YELLOW, ambos)
	if s, ok := checkUnclassifiedMovements(in); ok {
		candidates = append(candidates, s)
	}

	// 10. EXTERNAL_INCOME_DECLARED (Prio 10, YELLOW, ambos)
	if s, ok := checkExternalIncomeDeclared(in); ok {
		candidates = append(candidates, s)
	}

	// 11. INVOICE_WITHOUT_PAYMENT (Prio 11, YELLOW, ambos)
	if s, ok := checkInvoiceWithoutPayment(in); ok {
		candidates = append(candidates, s)
	}

	// 12. HIGH_CARD_SPENDING (Prio 12, YELLOW, MONO)
	if s, ok := checkHighCardSpending(in); ok {
		candidates = append(candidates, s)
	}

	// 13. MONO_CATEGORY_DOWNGRADE_POSSIBLE (Prio 13, YELLOW, MONO)
	if s, ok := checkMonoCategoryDowngradePossible(in); ok {
		candidates = append(candidates, s)
	}

	// Exclusiones mutuas (criterio 4)
	hasExclusionK := false
	hasChangeRequired := false
	for _, c := range candidates {
		if c.Code == CodeMonoExclusionKCapExceeded {
			hasExclusionK = true
		}
		if c.Code == CodeMonoCategoryChangeRequired {
			hasChangeRequired = true
		}
	}

	filtered := make([]Signal, 0, len(candidates))
	for _, c := range candidates {
		if hasExclusionK {
			if c.Code == CodeMonoNearKCap ||
				c.Code == CodeMonoCategoryChangeRequired ||
				c.Code == CodeMonoNearCategoryCap ||
				c.Code == CodeMonoCategoryDowngradePossible {
				continue
			}
		}
		if hasChangeRequired {
			if c.Code == CodeMonoNearCategoryCap {
				continue
			}
		}
		filtered = append(filtered, c)
	}

	// Ordenar por severidad descendente (RED=2 > YELLOW=1 > GREEN=0) y prioridad ascendente (1 < 2 < 3...)
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Severity != filtered[j].Severity {
			return filtered[i].Severity > filtered[j].Severity
		}
		return filtered[i].Priority < filtered[j].Priority
	})

	if len(filtered) == 0 {
		greenDesc := greenDescription(in)
		return Evaluation{
			Severity:    SeverityGreen,
			Signals:     nil,
			Primary:     nil,
			Description: greenDesc,
		}
	}

	primary := filtered[0]
	// La severidad del cierre es la máxima severidad (es decir, la de la señal primaria)
	overallSeverity := primary.Severity

	return Evaluation{
		Severity:    overallSeverity,
		Signals:     filtered,
		Primary:     &primary,
		Description: primary.Message,
	}
}

func greenDescription(in Input) string {
	if in.Regime == client.RegimeRI && in.RI != nil {
		return fmt.Sprintf("Facturó %s · IVA %s · todo conciliado.",
			formatUSD(in.RI.BilledInPeriodUSD),
			formatARS(in.RI.VatToPay),
		)
	}
	if in.Mono != nil {
		return fmt.Sprintf("Facturó %s · cuota %s · todo conciliado.",
			formatUSD(in.Mono.BilledInPeriodUSD),
			formatARS(in.Mono.Fee),
		)
	}
	return "Todo conciliado."
}

func checkMonoExclusionKCapExceeded(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil || !in.Mono.Excluded {
		return Signal{}, false
	}
	if in.Mono.ExclusionDate == nil || in.Mono.ExcessARS == nil {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeMonoExclusionKCapExceeded)
	dateStr := formatDate(*in.Mono.ExclusionDate)
	excessStr := formatARS(*in.Mono.ExcessARS)

	msg := fmt.Sprintf("Superó el tope de la categoría K el %s por %s. Corresponde pase a Responsable Inscripto.",
		dateStr, excessStr)

	return Signal{
		Code:     CodeMonoExclusionKCapExceeded,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"exclusionDate": dateStr,
			"excessARS":     formatCents(in.Mono.ExcessARS.Amount),
		},
	}, true
}

func checkPostExclusionCInvoices(in Input) (Signal, bool) {
	if in.Regime != client.RegimeRI || len(in.CInvoicesAfterRI) == 0 {
		return Signal{}, false
	}

	totalARS := money.Zero(money.ARS)
	for _, ci := range in.CInvoicesAfterRI {
		totalARS.Amount += ci.AmountARS.Amount
	}

	meta, _ := MetaFor(CodePostExclusionCInvoices)
	msg := fmt.Sprintf("%d facturas C emitidas después de la exclusión (%s).",
		len(in.CInvoicesAfterRI),
		formatARS(totalARS),
	)

	return Signal{
		Code:     CodePostExclusionCInvoices,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"count":    fmt.Sprintf("%d", len(in.CInvoicesAfterRI)),
			"totalARS": formatCents(totalARS.Amount),
		},
	}, true
}

func checkMissingInvoices(in Input) (Signal, bool) {
	count := len(in.Reconciliation.UnmatchedIncome)
	if count == 0 {
		return Signal{}, false
	}

	totalUSD := money.Zero(money.USD)
	for _, m := range in.Reconciliation.UnmatchedIncome {
		totalUSD.Amount += m.Amount.Amount
	}

	if totalUSD.Amount < in.Params.MissingInvoiceMinUSD.Amount {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeMissingInvoices)
	msg := fmt.Sprintf("%d cobros por %s sin factura asociada.",
		count,
		formatUSD(totalUSD),
	)

	return Signal{
		Code:     CodeMissingInvoices,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"count":    fmt.Sprintf("%d", count),
			"totalUSD": formatCents(totalUSD.Amount),
		},
	}, true
}

func checkDepositsExceedKCap(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil {
		return Signal{}, false
	}

	if in.Mono.KCapARS.Amount <= 0 || in.Mono.Inbound12mARS.Amount <= in.Mono.KCapARS.Amount {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeDepositsExceedKCap)
	msg := fmt.Sprintf("Depósitos de los últimos 12 meses (%s) superan el tope de la categoría K.",
		formatARS(in.Mono.Inbound12mARS),
	)

	return Signal{
		Code:     CodeDepositsExceedKCap,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"inboundARS": formatCents(in.Mono.Inbound12mARS.Amount),
			"kCapARS":    formatCents(in.Mono.KCapARS.Amount),
		},
	}, true
}

func checkMonoNearKCap(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil || in.Mono.Excluded {
		return Signal{}, false
	}

	if in.Mono.PctOfKCapBp < int64(in.Params.KCapWarningBp) {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeMonoNearKCap)
	pctStr := formatPct(in.Mono.PctOfKCapBp)
	msg := fmt.Sprintf("Acumulado 12 meses al %s del tope de la categoría K.", pctStr)

	return Signal{
		Code:     CodeMonoNearKCap,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"pct": pctStr,
		},
	}, true
}

func checkMonoCategoryChangeRequired(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil || in.Mono.RequiredCategory == nil {
		return Signal{}, false
	}

	// requiredCategory > currentCategory
	if !in.Mono.CurrentCategory.Less(*in.Mono.RequiredCategory) {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeMonoCategoryChangeRequired)
	msg := fmt.Sprintf("Facturó %s en 12 meses: supera el tope de la %s. Recategorizar a %s.",
		formatARS(in.Mono.Billed12mARS),
		string(in.Mono.CurrentCategory),
		string(*in.Mono.RequiredCategory),
	)

	return Signal{
		Code:     CodeMonoCategoryChangeRequired,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"billed12m": formatARS(in.Mono.Billed12mARS),
			"current":   string(in.Mono.CurrentCategory),
			"required":  string(*in.Mono.RequiredCategory),
		},
	}, true
}

func checkMonoNearCategoryCap(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil || in.Mono.Excluded {
		return Signal{}, false
	}

	if in.Mono.PctOfCurrentCapBp < int64(in.Params.CategoryCapWarningBp) {
		return Signal{}, false
	}

	// Sin cambio requerido
	if in.Mono.RequiredCategory != nil && in.Mono.CurrentCategory.Less(*in.Mono.RequiredCategory) {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeMonoNearCategoryCap)
	pctStr := formatPct(in.Mono.PctOfCurrentCapBp)
	msg := fmt.Sprintf("Acumulado al %s del tope de la categoría %s.",
		pctStr,
		string(in.Mono.CurrentCategory),
	)

	return Signal{
		Code:     CodeMonoNearCategoryCap,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"pct":     pctStr,
			"current": string(in.Mono.CurrentCategory),
		},
	}, true
}

func checkUnclassifiedMovements(in Input) (Signal, bool) {
	count := 0
	totalUSD := money.Zero(money.USD)

	for _, pm := range in.PeriodMovements {
		if pm.Classification.Kind == movement.KindUnclassified {
			count++
			totalUSD.Amount += pm.Movement.Amount.Amount
		}
	}

	if count == 0 {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeUnclassifiedMovements)
	msg := fmt.Sprintf("%d movimientos sin clasificar (%s).",
		count,
		formatUSD(totalUSD),
	)

	return Signal{
		Code:     CodeUnclassifiedMovements,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"count":    fmt.Sprintf("%d", count),
			"totalUSD": formatCents(totalUSD.Amount),
		},
	}, true
}

func checkExternalIncomeDeclared(in Input) (Signal, bool) {
	if len(in.ClientInput.ExternalIncome) == 0 {
		return Signal{}, false
	}

	totalUSD := in.ClientInput.TotalExternalIncomeUSD()
	if totalUSD.Amount <= 0 {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeExternalIncomeDeclared)
	msg := fmt.Sprintf("Declaró cobros fuera de Wallbit por %s.", formatUSD(totalUSD))

	return Signal{
		Code:     CodeExternalIncomeDeclared,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"totalUSD": formatCents(totalUSD.Amount),
		},
	}, true
}

func checkInvoiceWithoutPayment(in Input) (Signal, bool) {
	count := len(in.Reconciliation.UnmatchedInvoices)
	if count == 0 {
		return Signal{}, false
	}

	totalUSD := money.Zero(money.USD)
	for _, inv := range in.Reconciliation.UnmatchedInvoices {
		totalUSD.Amount += inv.TotalAmount.Amount
	}

	meta, _ := MetaFor(CodeInvoiceWithoutPayment)
	msg := fmt.Sprintf("%d facturas por %s sin cobro en Wallbit.",
		count,
		formatUSD(totalUSD),
	)

	return Signal{
		Code:     CodeInvoiceWithoutPayment,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"count":    fmt.Sprintf("%d", count),
			"totalUSD": formatCents(totalUSD.Amount),
		},
	}, true
}

func checkHighCardSpending(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil {
		return Signal{}, false
	}

	if in.Mono.KCapARS.Amount <= 0 {
		return Signal{}, false
	}

	threshold := (in.Mono.KCapARS.Amount * int64(in.Params.CardSpendKCapBp)) / 10000
	if in.Mono.CardSpend12mARS.Amount <= threshold {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeHighCardSpending)
	pctBp := (in.Mono.CardSpend12mARS.Amount * 10000) / in.Mono.KCapARS.Amount
	pctStr := formatPct(pctBp)
	msg := fmt.Sprintf("Gastos con tarjeta en 12 meses al %s del tope de la categoría K.", pctStr)

	return Signal{
		Code:     CodeHighCardSpending,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"pct": pctStr,
		},
	}, true
}

func checkMonoCategoryDowngradePossible(in Input) (Signal, bool) {
	if in.Regime != client.RegimeMonotributo || in.Mono == nil || in.Mono.RequiredCategory == nil {
		return Signal{}, false
	}

	// requiredCategory < currentCategory
	if !in.Mono.RequiredCategory.Less(in.Mono.CurrentCategory) {
		return Signal{}, false
	}

	meta, _ := MetaFor(CodeMonoCategoryDowngradePossible)
	msg := fmt.Sprintf("Podría bajar a la categoría %s en la próxima recategorización.",
		string(*in.Mono.RequiredCategory),
	)

	return Signal{
		Code:     CodeMonoCategoryDowngradePossible,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"required": string(*in.Mono.RequiredCategory),
		},
	}, true
}
