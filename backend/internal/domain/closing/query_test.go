package closing_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func sampleRows() []closing.Row {
	catD := tax.CatD
	catE := tax.CatE
	now := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	return []closing.Row{
		{
			ClosingID:   "clo_1",
			ClientID:    "cli_01",
			ClientName:  "Ana Pérez",
			CUIT:        "27-11111111-4",
			Regime:      client.RegimeMonotributo,
			Category:    &catD,
			Severity:    signal.SeverityGreen,
			Description: "Todo ok",
			BilledInPeriod: closing.BilledInPeriod{
				USD: money.New(200000, money.USD),
				ARS: money.New(30000000, money.ARS),
			},
			Accumulated12m: closing.Accumulated12m{
				ARS:         money.New(2500000000, money.ARS),
				PctOfKCapBp: 2000,
			},
			ToPay: closing.ToPay{
				Amount:  money.New(8461300, money.ARS),
				DueDate: now,
			},
			Status:  closing.StatusReadyForReview,
			Version: 1,
		},
		{
			ClosingID:  "clo_2",
			ClientID:   "cli_02",
			ClientName: "Carlos Gómez",
			CUIT:       "20-22222222-3",
			Regime:     client.RegimeMonotributo,
			Category:   &catE,
			Severity:   signal.SeverityYellow,
			PrimarySignal: &closing.SignalSummary{
				Code:  signal.CodeMonoNearCategoryCap,
				Label: "Cerca de tope categoría",
			},
			OtherSignals: []closing.SignalSummary{
				{
					Code:  signal.CodeAtypicalIncome,
					Label: "Cobro atípico",
				},
			},
			Description: "Cerca de tope",
			BilledInPeriod: closing.BilledInPeriod{
				USD: money.New(400000, money.USD),
				ARS: money.New(60000000, money.ARS),
			},
			Accumulated12m: closing.Accumulated12m{
				ARS:         money.New(3200000000, money.ARS),
				PctOfKCapBp: 8500,
			},
			ToPay: closing.ToPay{
				Amount:  money.New(11981100, money.ARS),
				DueDate: now,
			},
			Status:  closing.StatusPendingClientInput,
			Version: 1,
		},
		{
			ClosingID:  "clo_3",
			ClientID:   "cli_03",
			ClientName: "Beatriz Rossi",
			CUIT:       "27-33333333-8",
			Regime:     client.RegimeRI,
			Severity:   signal.SeverityRed,
			PrimarySignal: &closing.SignalSummary{
				Code:  signal.CodeMissingInvoices,
				Label: "Faltan facturas",
			},
			Description: "Faltan facturas",
			BilledInPeriod: closing.BilledInPeriod{
				USD: money.New(500000, money.USD),
				ARS: money.New(75000000, money.ARS),
			},
			Accumulated12m: closing.Accumulated12m{
				ARS:         money.New(0, money.ARS),
				PctOfKCapBp: 0,
			},
			ToPay: closing.ToPay{
				Amount:  money.New(20000000, money.ARS),
				DueDate: now,
			},
			Status:  closing.StatusApproved,
			Version: 2,
		},
	}
}

func TestFilterSeverities(t *testing.T) {
	rows := sampleRows()
	res := closing.Apply(rows, closing.Filter{
		Severities: []signal.Severity{signal.SeverityRed, signal.SeverityYellow},
	}, closing.Sort{})

	if len(res) != 2 {
		t.Fatalf("esperaba 2 resultados, obtuvo %d", len(res))
	}
	if res[0].Severity != signal.SeverityRed || res[1].Severity != signal.SeverityYellow {
		t.Fatalf("orden esperado RED luego YELLOW, obtuvo %v, %v", res[0].Severity, res[1].Severity)
	}
}

func TestFilterSignals(t *testing.T) {
	rows := sampleRows()

	// Coincide con PrimarySignal de cli_03
	res1 := closing.Apply(rows, closing.Filter{
		Signals: []signal.Code{signal.CodeMissingInvoices},
	}, closing.Sort{})
	if len(res1) != 1 || res1[0].ClientID != "cli_03" {
		t.Fatalf("esperaba match de cli_03 por MISSING_INVOICES, obtuvo %+v", res1)
	}

	// Coincide con OtherSignals de cli_02
	res2 := closing.Apply(rows, closing.Filter{
		Signals: []signal.Code{signal.CodeAtypicalIncome},
	}, closing.Sort{})
	if len(res2) != 1 || res2[0].ClientID != "cli_02" {
		t.Fatalf("esperaba match de cli_02 por ATYPICAL_INCOME en OtherSignals, obtuvo %+v", res2)
	}
}

func TestFilterRegime(t *testing.T) {
	rows := sampleRows()

	resRI := closing.Apply(rows, closing.Filter{
		Regime: client.RegimeRI,
	}, closing.Sort{})
	if len(resRI) != 1 || resRI[0].ClientID != "cli_03" {
		t.Fatalf("esperaba 1 cliente RI (cli_03), obtuvo %+v", resRI)
	}

	resMono := closing.Apply(rows, closing.Filter{
		Regime: client.RegimeMonotributo,
	}, closing.Sort{})
	if len(resMono) != 2 {
		t.Fatalf("esperaba 2 clientes Monotributo, obtuvo %d", len(resMono))
	}
}

func TestFilterStatus(t *testing.T) {
	rows := sampleRows()

	res := closing.Apply(rows, closing.Filter{
		Statuses: []closing.Status{closing.StatusApproved},
	}, closing.Sort{})
	if len(res) != 1 || res[0].ClientID != "cli_03" {
		t.Fatalf("esperaba cli_03 con APPROVED, obtuvo %+v", res)
	}
}

func TestFilterQuery(t *testing.T) {
	rows := sampleRows()

	// Por nombre parcial insensible a mayúsculas
	resName := closing.Apply(rows, closing.Filter{Query: "pÉrez"}, closing.Sort{})
	if len(resName) != 1 || resName[0].ClientID != "cli_01" {
		t.Fatalf("esperaba match de Ana Pérez, obtuvo %+v", resName)
	}

	// Por CUIT sin guiones
	resCUIT := closing.Apply(rows, closing.Filter{Query: "22222222"}, closing.Sort{})
	if len(resCUIT) != 1 || resCUIT[0].ClientID != "cli_02" {
		t.Fatalf("esperaba match de Carlos Gómez por CUIT, obtuvo %+v", resCUIT)
	}
}

func TestSorts(t *testing.T) {
	rows := sampleRows()

	// 1. Sort by name asc
	resNameAsc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "name", Dir: "asc"})
	if resNameAsc[0].ClientName != "Ana Pérez" || resNameAsc[1].ClientName != "Beatriz Rossi" || resNameAsc[2].ClientName != "Carlos Gómez" {
		t.Fatalf("orden por nombre asc incorrecto: %s, %s, %s", resNameAsc[0].ClientName, resNameAsc[1].ClientName, resNameAsc[2].ClientName)
	}

	// 2. Sort by name desc
	resNameDesc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "name", Dir: "desc"})
	if resNameDesc[0].ClientName != "Carlos Gómez" || resNameDesc[2].ClientName != "Ana Pérez" {
		t.Fatalf("orden por nombre desc incorrecto")
	}

	// 3. Sort by billed asc / desc
	resBilledDesc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "billed", Dir: "desc"})
	if resBilledDesc[0].ClientID != "cli_03" || resBilledDesc[2].ClientID != "cli_01" {
		t.Fatalf("orden por billed desc incorrecto")
	}

	// 4. Sort by toPay asc
	resToPayAsc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "toPay", Dir: "asc"})
	if resToPayAsc[0].ClientID != "cli_01" || resToPayAsc[2].ClientID != "cli_03" {
		t.Fatalf("orden por toPay asc incorrecto")
	}

	// 5. Sort by pctKCap desc
	resPctDesc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "pctKCap", Dir: "desc"})
	if resPctDesc[0].ClientID != "cli_02" || resPctDesc[2].ClientID != "cli_03" {
		t.Fatalf("orden por pctKCap desc incorrecto")
	}

	// 6. Default sort: severity desc (RED > YELLOW > GREEN)
	resSevDefault := closing.Apply(rows, closing.Filter{}, closing.Sort{})
	if resSevDefault[0].ClientID != "cli_03" || resSevDefault[1].ClientID != "cli_02" || resSevDefault[2].ClientID != "cli_01" {
		t.Fatalf("orden default de severidad incorrecto")
	}

	// 7. Severity asc (GREEN < YELLOW < RED)
	resSevAsc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "severity", Dir: "asc"})
	if resSevAsc[0].ClientID != "cli_01" || resSevAsc[1].ClientID != "cli_02" || resSevAsc[2].ClientID != "cli_03" {
		t.Fatalf("orden severity asc incorrecto")
	}

	// 8. Sort by regime asc / desc
	resRegimeAsc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "regime", Dir: "asc"})
	if resRegimeAsc[0].ClientID != "cli_01" || resRegimeAsc[1].ClientID != "cli_02" || resRegimeAsc[2].ClientID != "cli_03" {
		t.Fatalf("orden regime asc incorrecto: %+v", resRegimeAsc)
	}
	resRegimeDesc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "regime", Dir: "desc"})
	if resRegimeDesc[0].ClientID != "cli_03" || resRegimeDesc[1].ClientID != "cli_02" || resRegimeDesc[2].ClientID != "cli_01" {
		t.Fatalf("orden regime desc incorrecto: %+v", resRegimeDesc)
	}

	// 9. Sort by status asc / desc
	resStatusAsc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "status", Dir: "asc"})
	if resStatusAsc[0].ClientID != "cli_02" || resStatusAsc[1].ClientID != "cli_01" || resStatusAsc[2].ClientID != "cli_03" {
		t.Fatalf("orden status asc incorrecto: %+v", resStatusAsc)
	}
	resStatusDesc := closing.Apply(rows, closing.Filter{}, closing.Sort{By: "status", Dir: "desc"})
	if resStatusDesc[0].ClientID != "cli_03" || resStatusDesc[1].ClientID != "cli_01" || resStatusDesc[2].ClientID != "cli_02" {
		t.Fatalf("orden status desc incorrecto: %+v", resStatusDesc)
	}
}

func TestApplyPurity(t *testing.T) {
	rows := sampleRows()
	original := make([]closing.Row, len(rows))
	copy(original, rows)

	_ = closing.Apply(rows, closing.Filter{Query: "Carlos"}, closing.Sort{By: "name"})

	if !reflect.DeepEqual(rows, original) {
		t.Fatal("closing.Apply no debe mutar el slice de entrada")
	}
}

func TestComputeSummary(t *testing.T) {
	rows := sampleRows()
	summary := closing.ComputeSummary(rows)

	if summary.Total != 3 {
		t.Fatalf("esperaba total 3, obtuvo %d", summary.Total)
	}
	if summary.BySeverity["RED"] != 1 || summary.BySeverity["YELLOW"] != 1 || summary.BySeverity["GREEN"] != 1 {
		t.Fatalf("bySeverity incorrecto: %+v", summary.BySeverity)
	}
	if summary.ByStatus[closing.StatusReadyForReview] != 1 || summary.ByStatus[closing.StatusPendingClientInput] != 1 || summary.ByStatus[closing.StatusApproved] != 1 {
		t.Fatalf("byStatus incorrecto: %+v", summary.ByStatus)
	}
	if summary.BySignal[string(signal.CodeMissingInvoices)] != 1 {
		t.Fatalf("esperaba 1 MISSING_INVOICES, obtuvo %d", summary.BySignal[string(signal.CodeMissingInvoices)])
	}
	if summary.BySignal[string(signal.CodeAtypicalIncome)] != 1 {
		t.Fatalf("esperaba 1 ATYPICAL_INCOME, obtuvo %d", summary.BySignal[string(signal.CodeAtypicalIncome)])
	}
}
