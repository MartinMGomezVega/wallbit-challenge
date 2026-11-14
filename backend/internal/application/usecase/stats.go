package usecase

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// GetPeriodStats devuelve las estadísticas agregadas del período para el tablero de resumen.
type GetPeriodStats struct {
	Query ports.StatsQuery
}

// Execute ejecuta la consulta de estadísticas del período.
func (uc *GetPeriodStats) Execute(ctx context.Context, p period.Period) (closing.PeriodStats, error) {
	return uc.Query.PeriodStats(ctx, p)
}
