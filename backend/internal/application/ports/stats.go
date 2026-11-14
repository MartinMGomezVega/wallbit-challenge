package ports

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// StatsQuery expone estadísticas agregadas del período para el tablero de resumen (spec 014).
type StatsQuery interface {
	PeriodStats(ctx context.Context, p period.Period) (closing.PeriodStats, error)
}
