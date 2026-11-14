// Package mock implementa los clientes mockeados de ARCA para escalas impositivas
// y vencimientos (ports.TaxScaleProvider y ports.DueDateProvider) embebiendo datos reales
// (domain-context.md §2.1, §2.2, §2.3, architecture.md §2.2).
package mock

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

//go:embed data/monotributo_scales.json
var scalesJSON []byte

//go:embed data/due_dates.json
var dueDatesJSON []byte

var (
	_ ports.TaxScaleProvider = (*Client)(nil)
	_ ports.DueDateProvider  = (*Client)(nil)
)

// ErrSimulatedFailure se produce cuando failEvery se activa.
var ErrSimulatedFailure = errors.New("falla simulada de ARCA")

type scaleEntryJSON struct {
	ValidFrom string `json:"validFrom"`
	Rows      []struct {
		Category    string `json:"category"`
		Cap         string `json:"cap"`
		ServicesFee string `json:"servicesFee"`
	} `json:"rows"`
}

type dueDatesMonthJSON struct {
	Monotributo  string            `json:"monotributo"`
	VAT          map[string]string `json:"vat"`
	SelfEmployed map[string]string `json:"selfEmployed"`
}

// Client simula el cliente HTTP de ARCA. La implementación real haría GET a <servicio> y mapearía la respuesta a tax.Scale.
type Client struct {
	latency   time.Duration
	failEvery int
	calls     atomic.Int64
	scales    []tax.Scale
	dueDates  map[string]dueDatesMonthJSON
}

// Option configura el cliente mock.
type Option func(*Client)

// WithLatency añade una latencia simulada a cada llamada.
func WithLatency(d time.Duration) Option {
	return func(c *Client) { c.latency = d }
}

// WithFailEvery hace fallar cada n-ésima llamada.
func WithFailEvery(n int) Option {
	return func(c *Client) { c.failEvery = n }
}

// New construye un Client cargando los JSONs embebidos.
func New(opts ...Option) (*Client, error) {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}

	// Parsear escalas
	var rawScales []scaleEntryJSON
	if err := json.Unmarshal(scalesJSON, &rawScales); err != nil {
		return nil, fmt.Errorf("error parseando monotributo_scales.json: %w", err)
	}
	for _, raw := range rawScales {
		validFrom, err := time.Parse("2006-01-02", raw.ValidFrom)
		if err != nil {
			return nil, err
		}
		var rows []tax.CategoryRow
		for _, r := range raw.Rows {
			capMoney, err := money.ParseDecimal(r.Cap, money.ARS)
			if err != nil {
				return nil, err
			}
			feeMoney, err := money.ParseDecimal(r.ServicesFee, money.ARS)
			if err != nil {
				return nil, err
			}
			rows = append(rows, tax.CategoryRow{
				Category:       tax.Category(r.Category),
				CapARS:         capMoney,
				ServicesFeeARS: feeMoney,
			})
		}
		c.scales = append(c.scales, tax.Scale{
			ValidFrom: validFrom.In(period.Location()),
			Rows:      rows,
		})
	}
	// Ordenar escalas por ValidFrom ascendente
	sort.Slice(c.scales, func(i, j int) bool {
		return c.scales[i].ValidFrom.Before(c.scales[j].ValidFrom)
	})

	// Parsear due_dates
	if err := json.Unmarshal(dueDatesJSON, &c.dueDates); err != nil {
		return nil, fmt.Errorf("error parseando due_dates.json: %w", err)
	}

	return c, nil
}

func (c *Client) check(ctx context.Context) error {
	call := c.calls.Add(1)
	if c.failEvery > 0 && call%int64(c.failEvery) == 0 {
		return ErrSimulatedFailure
	}
	if c.latency > 0 {
		select {
		case <-time.After(c.latency):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return ctx.Err()
}

// MonotributoScale devuelve la escala vigente en at.
func (c *Client) MonotributoScale(ctx context.Context, at time.Time) (tax.Scale, error) {
	if err := c.check(ctx); err != nil {
		return tax.Scale{}, err
	}
	if len(c.scales) == 0 {
		return tax.Scale{}, errors.New("no hay escalas cargadas")
	}

	atAR := at.In(period.Location())
	// Si at es anterior a la primera escala, devuelve la primera (supuesto §2.2)
	selected := c.scales[0]
	for _, s := range c.scales {
		if !atAR.Before(s.ValidFrom) {
			selected = s
		}
	}
	return selected, nil
}

// DueDates devuelve los vencimientos de las obligaciones que vencen en duePeriod para el CUIT dado.
func (c *Client) DueDates(ctx context.Context, duePeriod period.Period, cuit client.CUIT) (ports.DueDates, error) {
	if err := c.check(ctx); err != nil {
		return ports.DueDates{}, err
	}

	term := fmt.Sprintf("%d", cuit.Terminacion())
	pKey := duePeriod.String()

	// Si está en due_dates.json (ej: 2026-11 real)
	if entry, ok := c.dueDates[pKey]; ok {
		monoT, err := time.ParseInLocation("2006-01-02", entry.Monotributo, period.Location())
		if err != nil {
			return ports.DueDates{}, err
		}
		vatStr, ok := entry.VAT[term]
		if !ok {
			return ports.DueDates{}, fmt.Errorf("terminación %s no encontrada en VAT", term)
		}
		vatT, err := time.ParseInLocation("2006-01-02", vatStr, period.Location())
		if err != nil {
			return ports.DueDates{}, err
		}
		seStr, ok := entry.SelfEmployed[term]
		if !ok {
			return ports.DueDates{}, fmt.Errorf("terminación %s no encontrada en autónomos", term)
		}
		seT, err := time.ParseInLocation("2006-01-02", seStr, period.Location())
		if err != nil {
			return ports.DueDates{}, err
		}
		return ports.DueDates{Monotributo: monoT, VAT: vatT, SelfEmployed: seT}, nil
	}

	// Patrón referencial (domain-context.md §2.3)
	// Monotributo: día 20 corrido
	monoDay := period.NextBusinessDay(time.Date(duePeriod.Year, duePeriod.Month, 20, 0, 0, 0, 0, period.Location()))

	// IVA: terminación 0-1 (día 18), 2-3 (19), 4-5 (20), 6-7 (21), 8-9 (22)
	termInt := cuit.Terminacion()
	vatBaseDay := 18 + (termInt / 2)
	vatDay := period.NextBusinessDay(time.Date(duePeriod.Year, duePeriod.Month, vatBaseDay, 0, 0, 0, 0, period.Location()))

	// Autónomos: terminación 0-3 (día 5), 4-9 (día 9)
	seBaseDay := 5
	if termInt >= 4 {
		seBaseDay = 9
	}
	seDay := period.NextBusinessDay(time.Date(duePeriod.Year, duePeriod.Month, seBaseDay, 0, 0, 0, 0, period.Location()))

	return ports.DueDates{
		Monotributo:  monoDay,
		VAT:          vatDay,
		SelfEmployed: seDay,
	}, nil
}
