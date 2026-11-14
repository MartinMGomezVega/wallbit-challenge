package ports

import (
	"context"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// TaxScaleProvider provee las escalas de Monotributo vigentes.
type TaxScaleProvider interface {
	MonotributoScale(ctx context.Context, at time.Time) (tax.Scale, error)
}

// DueDates agrupa los vencimientos de un período para un contribuyente.
type DueDates struct {
	Monotributo  time.Time `json:"monotributo"`
	VAT          time.Time `json:"vat"`
	SelfEmployed time.Time `json:"selfEmployed"`
}

// DueDateProvider provee los vencimientos oficiales según terminación de CUIT.
type DueDateProvider interface {
	DueDates(ctx context.Context, p period.Period, cuit client.CUIT) (DueDates, error)
}

// Quote es la cotización oficial BNA (billete) de un día hábil.
type Quote struct {
	Day  time.Time  `json:"day"`
	Buy  money.Rate `json:"buy"`
	Sell money.Rate `json:"sell"`
}

// ExchangeRateProvider provee cotizaciones históricas de tipo de cambio.
type ExchangeRateProvider interface {
	USDQuote(ctx context.Context, day time.Time) (Quote, error)
}

// ParametersRepository persiste los parámetros globales del contador.
type ParametersRepository interface {
	Get(ctx context.Context) (parameters.AccountantParameters, error)
	Save(ctx context.Context, p parameters.AccountantParameters) error
}
