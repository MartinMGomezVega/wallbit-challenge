// Package mock implementa ports.ExchangeRateProvider embebiendo las cotizaciones oficiales
// BNA reales (domain-context.md §2.4, architecture.md §2.2).
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
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

//go:embed data/usd_bna.json
var usdBNAJSON []byte

var _ ports.ExchangeRateProvider = (*Client)(nil)

// Errores del proveedor de tipo de cambio.
var (
	ErrRateNotFound     = errors.New("cotización no encontrada")
	ErrSimulatedFailure = errors.New("falla simulada de tipo de cambio")
)

type rawQuote struct {
	Day  string `json:"day"`
	Buy  string `json:"buy"`
	Sell string `json:"sell"`
}

// Client simula el cliente de tipo de cambio BNA.
type Client struct {
	latency   time.Duration
	failEvery int
	calls     atomic.Int64
	quotes    []ports.Quote
	byDay     map[string]ports.Quote
}

// Option configura el Client.
type Option func(*Client)

// WithLatency añade una latencia simulada a cada llamada.
func WithLatency(d time.Duration) Option {
	return func(c *Client) { c.latency = d }
}

// WithFailEvery hace fallar cada n-ésima llamada.
func WithFailEvery(n int) Option {
	return func(c *Client) { c.failEvery = n }
}

// New construye un Client cargando el fixture embebido.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		byDay: make(map[string]ports.Quote),
	}
	for _, opt := range opts {
		opt(c)
	}

	var raw []rawQuote
	if err := json.Unmarshal(usdBNAJSON, &raw); err != nil {
		return nil, fmt.Errorf("error parseando usd_bna.json: %w", err)
	}

	for _, r := range raw {
		day, err := time.ParseInLocation("2006-01-02", r.Day, period.Location())
		if err != nil {
			return nil, fmt.Errorf("fecha inválida %s: %w", r.Day, err)
		}
		buy, err := money.RateFromDecimal(r.Buy)
		if err != nil {
			return nil, err
		}
		sell, err := money.RateFromDecimal(r.Sell)
		if err != nil {
			return nil, err
		}
		q := ports.Quote{Day: day, Buy: buy, Sell: sell}
		c.quotes = append(c.quotes, q)
		c.byDay[r.Day] = q
	}

	sort.Slice(c.quotes, func(i, j int) bool {
		return c.quotes[i].Day.Before(c.quotes[j].Day)
	})

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

// USDQuote devuelve la cotización BNA para el día dado. Si el día no tiene cotización
// (fin de semana o feriado), busca hacia atrás hasta 10 días el día hábil anterior más cercano
// (spec 001 criterio 8).
func (c *Client) USDQuote(ctx context.Context, day time.Time) (ports.Quote, error) {
	if err := c.check(ctx); err != nil {
		return ports.Quote{}, err
	}

	cur := day.In(period.Location())
	for i := 0; i <= 10; i++ {
		key := cur.Format("2006-01-02")
		if q, ok := c.byDay[key]; ok {
			return q, nil
		}
		cur = cur.AddDate(0, 0, -1)
	}

	return ports.Quote{}, fmt.Errorf("%w para %s (y 10 días previos)", ErrRateNotFound, day.Format("2006-01-02"))
}
