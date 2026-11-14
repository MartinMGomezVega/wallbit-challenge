package client

import (
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// Regímenes tributarios.
const (
	RegimeMonotributo Regime = "MONOTRIBUTO"
	RegimeRI          Regime = "RI"
)

// Regime representa el régimen impositivo del cliente.
type Regime string

// IncomeTaxAdvance representa un anticipo de impuesto a las ganancias (solo RI).
type IncomeTaxAdvance = tax.IncomeTaxAdvance

// RIProfile guarda la configuración impositiva específica de un Responsable Inscripto cargada por el contador.
type RIProfile = tax.RIProfile

// RegimeHistoryItem registra un tramo de pertenencia a un régimen.
type RegimeHistoryItem struct {
	Regime Regime     `json:"regime"`
	From   time.Time  `json:"from"`
	To     *time.Time `json:"to,omitempty"`
	Reason string     `json:"reason"`
	Note   string     `json:"note,omitempty"`
}

// Client representa un usuario de Wallbit que contrató el servicio contable.
type Client struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Email               string              `json:"email"`
	CUIT                CUIT                `json:"cuit"`
	Regime              Regime              `json:"regime"`
	MonotributoCategory tax.Category        `json:"monotributoCategory,omitempty"`
	RegimeHistory       []RegimeHistoryItem `json:"regimeHistory"`
	RIProfile           *RIProfile          `json:"riProfile,omitempty"`
	CreatedAt           time.Time           `json:"createdAt"`
}

// RegimeAt determina qué régimen regía para el cliente en el instante t (spec 002 criterio 1).
func (c Client) RegimeAt(t time.Time) Regime {
	if len(c.RegimeHistory) == 0 {
		return c.Regime
	}
	tAR := t.In(period.Location())
	for _, item := range c.RegimeHistory {
		fromAR := item.From.In(period.Location())
		if tAR.Before(fromAR) {
			continue
		}
		if item.To != nil {
			toAR := item.To.In(period.Location())
			if tAR.After(toAR) {
				continue
			}
		}
		return item.Regime
	}
	// Fallback si no matchea ningún tramo específico
	return c.Regime
}

// RIFrom devuelve la fecha desde la cual el cliente es Responsable Inscripto si aplica.
func (c Client) RIFrom() (time.Time, bool) {
	for _, item := range c.RegimeHistory {
		if item.Regime == RegimeRI {
			return item.From, true
		}
	}
	if c.Regime == RegimeRI {
		return c.CreatedAt, true
	}
	return time.Time{}, false
}
