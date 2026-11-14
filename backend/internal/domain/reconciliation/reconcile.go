// Package reconciliation implementa la conciliación determinística cobros ↔ facturas
// según domain-context.md §6.
package reconciliation

import (
	"sort"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// Params agrupa los parámetros de conciliación configurables por el contador.
type Params struct {
	ToleranceBp      money.BasisPoints
	WindowBeforeDays int
	WindowAfterDays  int
}

// Pair representa una conciliación exitosa 1 a 1 entre un movimiento y una factura.
type Pair struct {
	Movement   movement.Movement `json:"movement"`
	Invoice    invoice.Invoice   `json:"invoice"`
	DiffAmount int64             `json:"diffAmount"` // m.amount - inv.amount en centavos
}

// Result es el resultado final de la conciliación.
type Result struct {
	Matched           []Pair              `json:"matched"`
	UnmatchedIncome   []movement.Movement `json:"unmatchedIncome"`
	UnmatchedInvoices []invoice.Invoice   `json:"unmatchedInvoices"`
}

// Reconcile ejecuta el algoritmo 1 a 1 greedy de conciliación (domain-context.md §6).
func Reconcile(incomeMovements []movement.Movement, invoices []invoice.Invoice, params Params) Result {
	// Copiar y ordenar movimientos por fecha ascendente (tie-breaker: ID)
	movs := make([]movement.Movement, len(incomeMovements))
	copy(movs, incomeMovements)
	sort.Slice(movs, func(i, j int) bool {
		if movs[i].OccurredAt.Equal(movs[j].OccurredAt) {
			return movs[i].ID < movs[j].ID
		}
		return movs[i].OccurredAt.Before(movs[j].OccurredAt)
	})

	usedInvoices := make(map[string]bool, len(invoices))
	var matched []Pair
	var unmatchedIncome []movement.Movement

	for _, m := range movs {
		type candidate struct {
			inv      invoice.Invoice
			diffTime time.Duration
			diffAmt  int64
		}

		var candidates []candidate

		mDate := m.OccurredAt.In(period.Location())
		normMov := movement.NormalizeName(m.Counterparty.Name)

		for _, inv := range invoices {
			if usedInvoices[inv.ID] {
				continue
			}

			// 1. Misma moneda
			if inv.Currency != m.Amount.Currency {
				continue
			}

			// 2. Coincidencia de nombres normalizados (igualdad o contención mutua)
			normInv := movement.NormalizeName(inv.ReceiverName)
			if normMov == "" || normInv == "" {
				continue
			}
			if normMov != normInv && !strings.Contains(normMov, normInv) && !strings.Contains(normInv, normMov) {
				continue
			}

			// 3. Ventana de fechas:
			// movement.date in [invoice.date - windowAfterDays, invoice.date + windowBeforeDays]
			invDate := inv.IssueDate.In(period.Location())
			minDate := invDate.AddDate(0, 0, -params.WindowAfterDays)
			maxDate := invDate.AddDate(0, 0, params.WindowBeforeDays)

			// Comparación normalizada a días calendario
			mCal := time.Date(mDate.Year(), mDate.Month(), mDate.Day(), 0, 0, 0, 0, period.Location())
			minCal := time.Date(minDate.Year(), minDate.Month(), minDate.Day(), 0, 0, 0, 0, period.Location())
			maxCal := time.Date(maxDate.Year(), maxDate.Month(), maxDate.Day(), 23, 59, 59, 999999999, period.Location())

			if mCal.Before(minCal) || mCal.After(maxCal) {
				continue
			}

			// 4. Tolerancia de monto (|inv.amount - mov.amount| <= inv.amount * toleranceBp / 10000)
			diffAmt := m.Amount.Amount - inv.TotalAmount.Amount
			absDiffAmt := diffAmt
			if absDiffAmt < 0 {
				absDiffAmt = -absDiffAmt
			}

			maxTol := money.ApplyBp(inv.TotalAmount, params.ToleranceBp).Amount
			if absDiffAmt > maxTol {
				continue
			}

			diffTime := mDate.Sub(invDate)
			if diffTime < 0 {
				diffTime = -diffTime
			}

			candidates = append(candidates, candidate{
				inv:      inv,
				diffTime: diffTime,
				diffAmt:  absDiffAmt,
			})
		}

		if len(candidates) == 0 {
			unmatchedIncome = append(unmatchedIncome, m)
			continue
		}

		// Ordenar candidatos: menor diferencia de fecha, empate -> menor diferencia de monto
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].diffTime != candidates[j].diffTime {
				return candidates[i].diffTime < candidates[j].diffTime
			}
			if candidates[i].diffAmt != candidates[j].diffAmt {
				return candidates[i].diffAmt < candidates[j].diffAmt
			}
			return candidates[i].inv.ID < candidates[j].inv.ID
		})

		best := candidates[0]
		usedInvoices[best.inv.ID] = true
		matched = append(matched, Pair{
			Movement:   m,
			Invoice:    best.inv,
			DiffAmount: m.Amount.Amount - best.inv.TotalAmount.Amount,
		})
	}

	var unmatchedInvoices []invoice.Invoice
	for _, inv := range invoices {
		if !usedInvoices[inv.ID] {
			unmatchedInvoices = append(unmatchedInvoices, inv)
		}
	}

	return Result{
		Matched:           matched,
		UnmatchedIncome:   unmatchedIncome,
		UnmatchedInvoices: unmatchedInvoices,
	}
}

// KnownPayers extrae el conjunto de nombres normalizados de receptores de facturas previas.
func KnownPayers(invoicesBefore []invoice.Invoice) map[string]bool {
	res := make(map[string]bool)
	for _, inv := range invoicesBefore {
		norm := movement.NormalizeName(inv.ReceiverName)
		if norm != "" {
			res[norm] = true
		}
	}
	return res
}
