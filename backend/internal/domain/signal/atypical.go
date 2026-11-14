package signal

import (
	"fmt"
	"sort"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
)

// Median calcula la mediana de una serie de montos monetarios en USD.
// Si la longitud es par, devuelve el promedio de los dos valores centrales.
func Median(values []money.Money) money.Money {
	n := len(values)
	if n == 0 {
		return money.Zero(money.USD)
	}

	copied := make([]int64, n)
	for i, v := range values {
		copied[i] = v.Amount
	}
	sort.Slice(copied, func(i, j int) bool {
		return copied[i] < copied[j]
	})

	if n%2 == 1 {
		return money.New(copied[n/2], money.USD)
	}

	mid1 := copied[n/2-1]
	mid2 := copied[n/2]
	// Promedio entero
	return money.New((mid1+mid2)/2, money.USD)
}

// checkAtypicalIncome evalúa si hubo algún cobro atípico en el período según domain-context.md §7.5.
func checkAtypicalIncome(in Input) (Signal, bool) {
	// Se requieren al menos 3 movimientos históricos para calcular la mediana
	if len(in.HistoricIncomeUSD) < 3 {
		return Signal{}, false
	}

	median := Median(in.HistoricIncomeUSD)
	if median.Amount <= 0 {
		return Signal{}, false
	}

	multiplier := int64(in.Params.AtypicalMultiplierBp)
	threshold := (median.Amount * multiplier) / 10000

	var maxAtypical money.Money
	found := false

	for _, pm := range in.PeriodMovements {
		if pm.Classification.Kind == movement.KindIncome {
			if pm.Movement.Amount.Amount > threshold {
				if !found || pm.Movement.Amount.Amount > maxAtypical.Amount {
					maxAtypical = pm.Movement.Amount
					found = true
				}
			}
		}
	}

	if !found {
		return Signal{}, false
	}

	ratioStr := formatRatio(maxAtypical.Amount, median.Amount)
	msg := fmt.Sprintf("Cobro de %s, %s su mediana de %d meses.",
		formatUSD(maxAtypical),
		ratioStr,
		in.Params.AtypicalWindowMonths,
	)

	meta, _ := MetaFor(CodeAtypicalIncome)
	return Signal{
		Code:     CodeAtypicalIncome,
		Severity: meta.Severity,
		Priority: meta.Priority,
		Message:  msg,
		Data: map[string]string{
			"amountUSD": formatCents(maxAtypical.Amount),
			"medianUSD": formatCents(median.Amount),
			"ratio":     ratioStr,
			"window":    fmt.Sprintf("%d", in.Params.AtypicalWindowMonths),
		},
	}, true
}
