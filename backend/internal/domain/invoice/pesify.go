package invoice

import (
	"errors"
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

var (
	// ErrUnsupportedInvoiceCurrency indica moneda distinta de ARS o USD.
	ErrUnsupportedInvoiceCurrency = errors.New("moneda de factura no soportada")
	// ErrQuoteLookupRequired indica que no se proveyó la función de cotizaciones para USD sin TC propio.
	ErrQuoteLookupRequired = errors.New("se requiere quote lookup para pesificar USD sin tipo de cambio propio")
)

// Quote representa la cotización BNA para pesificación de comprobantes.
type Quote struct {
	Buy  money.Rate
	Sell money.Rate
}

// QuoteLookup es la función desacoplada para obtener cotizaciones en la capa de dominio.
type QuoteLookup func(day time.Time) (Quote, error)

// AmountARS calcula el importe pesificado de una factura según domain-context.md §5:
// 1. Si la factura trae fxRate propio, se usa ese.
// 2. Si no, se consulta el tipo de cambio oficial BNA según la fx_rule del parámetro:
//   - BNA_BUY_PREV_BUSINESS_DAY: compra del día hábil anterior a la emisión.
//   - BNA_SELL_PREV_BUSINESS_DAY: venta del día hábil anterior a la emisión.
//   - BNA_BUY_SAME_DAY: compra del mismo día (o día hábil anterior si no hay cotización).
//
// 3. Facturas en ARS: rate = 1 (mismo importe).
func AmountARS(inv Invoice, rule parameters.FXRule, lookup QuoteLookup) (money.Money, error) {
	if inv.Currency == money.ARS {
		return inv.TotalAmount, nil
	}
	if inv.Currency != money.USD {
		return money.Money{}, fmt.Errorf("%w: %s", ErrUnsupportedInvoiceCurrency, inv.Currency)
	}

	// 1. Si la factura trae fxRate propio
	if inv.FXRate != nil && *inv.FXRate > 0 {
		return money.ConvertUSDToARS(inv.TotalAmount, *inv.FXRate)
	}

	if lookup == nil {
		return money.Money{}, ErrQuoteLookupRequired
	}

	var targetDate time.Time
	switch rule {
	case parameters.FXRuleBNABuySameDay:
		targetDate = period.PrevBusinessDay(inv.IssueDate)
	case parameters.FXRuleBNASellPrevBusinessDay, parameters.FXRuleBNABuyPrevBusinessDay:
		targetDate = period.PrevBusinessDay(inv.IssueDate.AddDate(0, 0, -1))
	default:
		targetDate = period.PrevBusinessDay(inv.IssueDate.AddDate(0, 0, -1))
	}

	q, err := lookup(targetDate)
	if err != nil {
		return money.Money{}, fmt.Errorf("error obteniendo cotización para %s: %w", targetDate.Format("2006-01-02"), err)
	}

	rate := q.Buy
	if rule == parameters.FXRuleBNASellPrevBusinessDay {
		rate = q.Sell
	}

	return money.ConvertUSDToARS(inv.TotalAmount, rate)
}
