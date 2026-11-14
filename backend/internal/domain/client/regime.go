package client

import (
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ErrInvalidTransition indica una transición de régimen no permitida.
var ErrInvalidTransition = domain.ErrInvalidTransition

// ChangeRegime realiza la transición de régimen de un cliente (spec 005 criterio 5).
func (c *Client) ChangeRegime(to Regime, effective time.Time, reason, note string, now time.Time) error {
	if to != RegimeRI {
		return fmt.Errorf("%w: solo se soporta transición a RI", ErrInvalidTransition)
	}

	loc := period.Location()
	effAR := effective.In(loc)
	nowAR := now.In(loc)

	// Validar que la fecha efectiva no sea futura
	// Consideramos fecha calendario (día civil)
	effDay := time.Date(effAR.Year(), effAR.Month(), effAR.Day(), 0, 0, 0, 0, loc)
	nowDay := time.Date(nowAR.Year(), nowAR.Month(), nowAR.Day(), 0, 0, 0, 0, loc)
	if effDay.After(nowDay) {
		return fmt.Errorf("%w: la fecha efectiva no puede ser futura", ErrInvalidTransition)
	}

	// Solo si el cliente es MONO a esa fecha
	if c.RegimeAt(effective) != RegimeMonotributo {
		return fmt.Errorf("%w: el cliente no es Monotributo a la fecha efectiva", ErrInvalidTransition)
	}

	if reason == "" {
		reason = "EXCLUSION_K_CAP"
	}

	// Cerrar el tramo Monotributo en effectiveDate - 1 día
	closedTo := effDay.AddDate(0, 0, -1)
	endOfClosedDay := time.Date(closedTo.Year(), closedTo.Month(), closedTo.Day(), 23, 59, 59, 999999999, loc)

	if len(c.RegimeHistory) == 0 {
		c.RegimeHistory = append(c.RegimeHistory, RegimeHistoryItem{
			Regime: RegimeMonotributo,
			From:   c.CreatedAt,
			To:     &endOfClosedDay,
			Reason: "INICIAL",
		})
	} else {
		for i := range c.RegimeHistory {
			if c.RegimeHistory[i].To == nil || c.RegimeHistory[i].To.After(endOfClosedDay) {
				c.RegimeHistory[i].To = &endOfClosedDay
			}
		}
	}

	// Abrir nuevo tramo RI desde effective
	effStart := time.Date(effAR.Year(), effAR.Month(), effAR.Day(), 0, 0, 0, 0, loc)
	c.RegimeHistory = append(c.RegimeHistory, RegimeHistoryItem{
		Regime: RegimeRI,
		From:   effStart,
		To:     nil,
		Reason: reason,
		Note:   note,
	})

	c.Regime = RegimeRI
	if c.RIProfile == nil {
		c.RIProfile = &RIProfile{
			SelfEmployedMonthlyContribution: money.Zero(money.ARS),
			PreviousVatCredit:               money.Zero(money.ARS),
		}
	}

	return nil
}
