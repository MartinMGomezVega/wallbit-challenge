// Package tax implementa las liquidaciones de Monotributo y Responsable Inscripto.
package tax

import (
	"errors"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// PricedInvoice representa una factura con sus importes pesificados calculados.
type PricedInvoice struct {
	Invoice   invoice.Invoice
	AmountARS money.Money
	AmountUSD *money.Money
}

// DatedAmount representa un importe monetario asociado a una fecha.
type DatedAmount struct {
	Day    time.Time
	Amount money.Money
}

// CivilDay representa un día del calendario civil sin componente horario.
type CivilDay struct {
	Year  int
	Month time.Month
	Day   int
}

// civilDay es un alias de CivilDay para compatibilidad con las especificaciones.
type civilDay = CivilDay

// Time devuelve un time.Time a las 00:00:00 en la ubicación indicada.
func (c CivilDay) Time(loc *time.Location) time.Time {
	if loc == nil {
		loc = period.Location()
	}
	return time.Date(c.Year, c.Month, c.Day, 0, 0, 0, 0, loc)
}

// CivilDayFromTime extrae el día civil de un time.Time en la ubicación dada.
func CivilDayFromTime(t time.Time, loc *time.Location) CivilDay {
	if loc != nil {
		t = t.In(loc)
	}
	y, m, d := t.Date()
	return CivilDay{Year: y, Month: m, Day: d}
}

// Parameters es un alias a los parámetros del contador.
type Parameters = parameters.AccountantParameters

// MonotributoInput define los datos de entrada requeridos para liquidar Monotributo.
type MonotributoInput struct {
	Period              period.Period
	CurrentCategory     Category
	Invoices            []PricedInvoice
	ScaleAt             func(time.Time) Scale
	FeeDueDate          time.Time
	Params              parameters.AccountantParameters
	InboundMovementsARS []DatedAmount
	CardSpendARS        []DatedAmount
}

// MonotributoAssessment contiene el resultado de la liquidación de Monotributo para un período.
type MonotributoAssessment struct {
	BilledInPeriodUSD money.Money
	BilledInPeriodARS money.Money
	Billed12mARS      money.Money
	CurrentCategory   Category
	RequiredCategory  *Category
	PctOfCurrentCapBp int64
	PctOfKCapBp       int64
	Excluded          bool
	ExclusionDate     *time.Time
	ExcessARS         *money.Money
	CanReturnFrom     *time.Time
	Fee               money.Money
	FeeDueDate        time.Time
	Inbound12mARS     money.Money
	CardSpend12mARS   money.Money
	KCapARS           money.Money
	ScaleValidFrom    time.Time
}

// toEpochDay devuelve el número de días transcurridos desde el epoch UTC (1970-01-01).
func toEpochDay(t time.Time, loc *time.Location) int {
	if loc != nil {
		t = t.In(loc)
	}
	y, m, d := t.Date()
	return int(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400)
}

// dailySeries mantiene sumas acumuladas por día en O(1).
type dailySeries struct {
	minEpochDay int
	prefixSums  []int64
	loc         *time.Location
}

// newDailySeries construye una serie diaria con sumas de prefijos.
func newDailySeries(minDate, maxDate time.Time, items []DatedAmount, loc *time.Location) *dailySeries {
	if loc == nil {
		loc = period.Location()
	}
	minEpoch := toEpochDay(minDate, loc)
	maxEpoch := toEpochDay(maxDate, loc)
	if maxEpoch < minEpoch {
		maxEpoch = minEpoch
	}
	size := maxEpoch - minEpoch + 2
	sums := make([]int64, size)

	for _, it := range items {
		e := toEpochDay(it.Day, loc)
		if e >= minEpoch && e <= maxEpoch {
			idx := e - minEpoch + 1
			sums[idx] += it.Amount.Amount
		}
	}

	for i := 1; i < len(sums); i++ {
		sums[i] += sums[i-1]
	}

	return &dailySeries{
		minEpochDay: minEpoch,
		prefixSums:  sums,
		loc:         loc,
	}
}

// Sum devuelve la suma en el intervalo semiabierto (fromExclusive, toInclusive].
func (s *dailySeries) Sum(fromExclusive, toInclusive time.Time) int64 {
	eFrom := toEpochDay(fromExclusive, s.loc)
	eTo := toEpochDay(toInclusive, s.loc)
	if eTo <= eFrom {
		return 0
	}

	idxTo := eTo - s.minEpochDay + 1
	if idxTo < 0 {
		return 0
	}
	if idxTo >= len(s.prefixSums) {
		idxTo = len(s.prefixSums) - 1
	}

	idxFrom := eFrom - s.minEpochDay + 1
	if idxFrom < 0 {
		idxFrom = 0
	}
	if idxFrom >= len(s.prefixSums) {
		idxFrom = len(s.prefixSums) - 1
	}

	return s.prefixSums[idxTo] - s.prefixSums[idxFrom]
}

// AssessMonotributo ejecuta la liquidación pura de Monotributo para un cliente y período.
func AssessMonotributo(in MonotributoInput) (MonotributoAssessment, error) {
	if in.ScaleAt == nil {
		return MonotributoAssessment{}, errors.New("ScaleAt provider es requerido")
	}
	if in.CurrentCategory == "" {
		return MonotributoAssessment{}, errors.New("CurrentCategory es requerida")
	}
	if in.FeeDueDate.IsZero() {
		return MonotributoAssessment{}, errors.New("FeeDueDate es requerida")
	}

	loc := period.Location()
	periodEnd := in.Period.End()
	scaleAtPeriodEnd := in.ScaleAt(periodEnd)
	maxCap := scaleAtPeriodEnd.MaxCap()

	// 1. Facturado en el período (USD y ARS)
	billedInPeriodUSD := money.Zero(money.USD)
	billedInPeriodARS := money.Zero(money.ARS)

	// Filtrar facturas computables para el acumulado 12 meses
	computableInvoices := make([]DatedAmount, 0, len(in.Invoices))

	for _, pi := range in.Invoices {
		p := period.FromTime(pi.Invoice.IssueDate)
		if p == in.Period {
			billedInPeriodARS.Amount += pi.AmountARS.Amount
			if pi.AmountUSD != nil {
				billedInPeriodUSD.Amount += pi.AmountUSD.Amount
			} else if pi.Invoice.Currency == money.USD {
				billedInPeriodUSD.Amount += pi.Invoice.TotalAmount.Amount
			}
		}

		// Computabilidad para tope 12m:
		// Facturas C siempre computan; facturas E solo si InvoiceECountsForCap == true.
		computes := false
		if pi.Invoice.Type == invoice.TypeC {
			computes = true
		} else if pi.Invoice.Type == invoice.TypeE && in.Params.InvoiceECountsForCap {
			computes = true
		}

		if computes {
			computableInvoices = append(computableInvoices, DatedAmount{
				Day:    pi.Invoice.IssueDate,
				Amount: pi.AmountARS,
			})
		}
	}

	// 2. Construir serie de facturación para el período y ventana histórica
	// Necesitamos desde periodEnd - 2 años hasta periodEnd
	minDate := periodEnd.AddDate(-2, -1, 0)
	maxDate := periodEnd.AddDate(0, 1, 0)
	invSeries := newDailySeries(minDate, maxDate, computableInvoices, loc)

	// 3. Acumulado 12 meses a periodEnd
	// El último día del período
	endDayCivil := CivilDayFromTime(periodEnd, loc)
	endDay := endDayCivil.Time(loc)
	from12mDay := endDay.AddDate(-1, 0, 0)

	billed12mVal := invSeries.Sum(from12mDay, endDay)
	billed12mARS := money.New(billed12mVal, money.ARS)

	// 4. Verificación de exclusión:
	// Revisar los 365 días que terminan en endDay para encontrar el primer día de cruce.
	var firstExclusionDay *time.Time
	for offset := 364; offset >= 0; offset-- {
		d := endDay.AddDate(0, 0, -offset)
		dFrom := d.AddDate(-1, 0, 0)
		rollingSum := invSeries.Sum(dFrom, d)
		scaleD := in.ScaleAt(d)
		if rollingSum > scaleD.MaxCap().Amount {
			dayCopy := d
			firstExclusionDay = &dayCopy
			break
		}
	}

	excluded := billed12mARS.Amount > maxCap.Amount || firstExclusionDay != nil
	var exclusionDate *time.Time
	var excessARS *money.Money
	var canReturnFrom *time.Time
	var requiredCategory *Category

	if excluded {
		if firstExclusionDay != nil {
			exclusionDate = firstExclusionDay
		} else {
			dayCopy := endDay
			exclusionDate = &dayCopy
		}

		excessVal := billed12mARS.Amount - maxCap.Amount
		if excessVal < 0 {
			excessVal = 0
		}
		excess := money.New(excessVal, money.ARS)
		excessARS = &excess

		canReturn := time.Date(exclusionDate.Year()+4, 1, 1, 0, 0, 0, 0, loc)
		canReturnFrom = &canReturn
		requiredCategory = nil
	} else {
		reqCat, err := scaleAtPeriodEnd.CategoryFor(billed12mARS)
		if err == nil {
			requiredCategory = &reqCat
		}
	}

	// 5. Porcentajes de topes en puntos básicos (redondeo hacia abajo)
	pctOfCurrentCapBp := int64(0)
	if row, ok := scaleAtPeriodEnd.Row(in.CurrentCategory); ok && row.CapARS.Amount > 0 {
		pctOfCurrentCapBp = billed12mARS.Amount * 10000 / row.CapARS.Amount
	}

	pctOfKCapBp := int64(0)
	if maxCap.Amount > 0 {
		pctOfKCapBp = billed12mARS.Amount * 10000 / maxCap.Amount
	}

	// 6. Cuota a pagar: cuota de la categoría actual según la escala vigente a FeeDueDate
	feeScale := in.ScaleAt(in.FeeDueDate)
	fee := money.Zero(money.ARS)
	if row, ok := feeScale.Row(in.CurrentCategory); ok {
		fee = row.ServicesFeeARS
	}

	// 7. Inbound y CardSpend de 12 meses
	inboundSeries := newDailySeries(minDate, maxDate, in.InboundMovementsARS, loc)
	inbound12mVal := inboundSeries.Sum(from12mDay, endDay)
	inbound12mARS := money.New(inbound12mVal, money.ARS)

	cardSeries := newDailySeries(minDate, maxDate, in.CardSpendARS, loc)
	card12mVal := cardSeries.Sum(from12mDay, endDay)
	cardSpend12mARS := money.New(card12mVal, money.ARS)

	return MonotributoAssessment{
		BilledInPeriodUSD: billedInPeriodUSD,
		BilledInPeriodARS: billedInPeriodARS,
		Billed12mARS:      billed12mARS,
		CurrentCategory:   in.CurrentCategory,
		RequiredCategory:  requiredCategory,
		PctOfCurrentCapBp: pctOfCurrentCapBp,
		PctOfKCapBp:       pctOfKCapBp,
		Excluded:          excluded,
		ExclusionDate:     exclusionDate,
		ExcessARS:         excessARS,
		CanReturnFrom:     canReturnFrom,
		Fee:               fee,
		FeeDueDate:        in.FeeDueDate,
		Inbound12mARS:     inbound12mARS,
		CardSpend12mARS:   cardSpend12mARS,
		KCapARS:           maxCap,
		ScaleValidFrom:    scaleAtPeriodEnd.ValidFrom,
	}, nil
}
